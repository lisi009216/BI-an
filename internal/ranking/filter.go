package ranking

import "strings"

// IsUSDTPair checks if a symbol is a USDT trading pair.
// A USDT pair is identified by the symbol ending with "USDT" and having
// at least one character before "USDT" (the base currency).
func IsUSDTPair(symbol string) bool {
	return len(symbol) > 4 && strings.HasSuffix(symbol, "USDT")
}
