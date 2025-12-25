package httpapi

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"strconv"
	"strings"
	"time"

	"example.com/binance-pivot-monitor/internal/pivot"
	signalpkg "example.com/binance-pivot-monitor/internal/signal"
	"example.com/binance-pivot-monitor/internal/sse"
	"example.com/binance-pivot-monitor/internal/ticker"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	SignalBroker   *sse.Broker[signalpkg.Signal]
	History        *signalpkg.History
	AllowedOrigins []string
	PivotStatus    PivotStatusProvider
	TickerStore    *ticker.Store
	TickerMonitor  *ticker.Monitor
}

func New(signalBroker *sse.Broker[signalpkg.Signal], history *signalpkg.History, allowedOrigins []string) *Server {
	return &Server{SignalBroker: signalBroker, History: history, AllowedOrigins: allowedOrigins}
}

type PivotStatusProvider interface {
	PivotStatus() pivot.PivotStatusResponse
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDashboard)
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/api/sse", s.handleSSE)
	mux.HandleFunc("/api/history", s.handleHistory)
	mux.HandleFunc("/api/pivot-status", s.handlePivotStatus)
	mux.HandleFunc("/api/tickers", s.handleTickers)

	// 嵌入的静态文件
	staticContent, _ := fs.Sub(staticFS, "static")
	mux.Handle("/static/app.js", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))

	// 外部静态文件（图标等）
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	return s.cors(mux)
}

func (s *Server) handleTickers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.TickerStore == nil {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("{}"))
		return
	}

	// 可选：按 symbols 过滤
	q := r.URL.Query()
	symbolsParam := q.Get("symbols")

	var data map[string]*ticker.Ticker
	if symbolsParam != "" {
		symbols := strings.Split(symbolsParam, ",")
		data = s.TickerStore.GetBySymbols(symbols)
	} else {
		data = s.TickerStore.GetAll()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}

func (s *Server) handlePivotStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.PivotStatus == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	resp := s.PivotStatus.PivotStatus()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func (s *Server) handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.History == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	q := r.URL.Query()
	getFirstCI := func(key string) string {
		if v := q.Get(key); v != "" {
			return v
		}
		for k, vs := range q {
			if strings.EqualFold(k, key) && len(vs) > 0 {
				return vs[0]
			}
		}
		return ""
	}
	getAllCI := func(key string) string {
		var all []string
		for k, vs := range q {
			if strings.EqualFold(k, key) {
				all = append(all, vs...)
			}
		}
		return strings.Join(all, ",")
	}

	symbol := getFirstCI("symbol")
	period := getFirstCI("period")
	level := getAllCI("level")
	if level == "" {
		level = getAllCI("levels")
	}
	direction := getFirstCI("direction")
	source := getFirstCI("source")
	limitStr := getFirstCI("limit")
	limit := 200
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil {
			limit = v
		}
	}

	res := s.History.Query(symbol, period, level, direction, source, limit)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// 从嵌入的文件系统读取 index.html
	data, err := staticFS.ReadFile("static/index.html")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(data)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if s.SignalBroker == nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	// 订阅信号
	signalCh := s.SignalBroker.Subscribe(256)
	defer s.SignalBroker.Unsubscribe(signalCh)

	// 订阅 ticker（如果可用）
	var tickerCh chan ticker.TickerBatch
	if s.TickerMonitor != nil {
		tickerCh = s.TickerMonitor.Subscribe(64)
		defer s.TickerMonitor.Unsubscribe(tickerCh)
	}

	_, _ = fmt.Fprintf(w, ": connected %s\n\n", time.Now().UTC().Format(time.RFC3339))
	flusher.Flush()

	keepAlive := time.NewTicker(15 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-keepAlive.C:
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()

		case sig, ok := <-signalCh:
			if !ok {
				return
			}
			b, err := json.Marshal(sig)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: signal\n")
			_, _ = fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(string(b), "\n", ""))
			flusher.Flush()

		case batch, ok := <-tickerCh:
			if !ok {
				tickerCh = nil
				continue
			}
			b, err := json.Marshal(batch)
			if err != nil {
				continue
			}
			_, _ = fmt.Fprintf(w, "event: ticker\n")
			_, _ = fmt.Fprintf(w, "data: %s\n\n", strings.ReplaceAll(string(b), "\n", ""))
			flusher.Flush()
		}
	}
}

func ParseAllowedOrigins(v string) []string {
	v = strings.TrimSpace(v)
	if v == "" {
		return []string{"*"}
	}
	parts := strings.Split(v, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		out = append(out, p)
	}
	if len(out) == 0 {
		return []string{"*"}
	}
	return out
}

func (s *Server) cors(next http.Handler) http.Handler {
	allowed := s.AllowedOrigins
	if len(allowed) == 0 {
		allowed = []string{"*"}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(w, r)
			return
		}

		allowOrigin := ""
		for _, o := range allowed {
			if o == "*" {
				allowOrigin = "*"
				break
			}
			if o == origin {
				allowOrigin = origin
				break
			}
		}

		if allowOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			w.Header().Add("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
