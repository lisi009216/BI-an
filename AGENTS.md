# Repository Guidelines

## Project Structure & Module Organization
- `cmd/server/` is the main Go entry point for the backend service.
- `internal/` holds core backend packages (Binance clients, pivots, signals, SSE, ticker store).
- `extension/` contains the Chrome extension assets (background, popup, side panel, options).
- `static/` hosts web assets served by the dashboard.
- `data/` stores runtime pivots and signal history; use a separate path for local runs if you want to keep the repo clean.
- `packaging/` and `dist/` are for release packaging and build outputs.

## Build, Test, and Development Commands
- `go run ./cmd/server` starts the backend locally on the default address (`:8080`).
- `go build -o binance-pivot-monitor ./cmd/server` builds a local binary.
- `./build.sh` produces cross-platform binaries in `dist/` (uses `VERSION=...` if set).
- `go test ./...` runs the full Go test suite.

## Coding Style & Naming Conventions
- Go code follows `gofmt` and standard Go conventions (tabs for indentation).
- Package names are short, lowercase, and scoped to their domain (e.g., `internal/pivot`).
- Extension JavaScript uses 2-space indentation and camelCase identifiers; keep file names descriptive (`background.js`, `sidepanel.js`).

## Testing Guidelines
- Unit tests live alongside packages as `*_test.go` (e.g., `internal/monitor/monitor_test.go`).
- Add tests for new signal logic or data processing paths where feasible.
- Prefer `go test ./...` before submitting changes that touch backend logic.

## Commit & Pull Request Guidelines
- Commit messages in history are short, plain-language summaries (often in Chinese); keep them concise and focused.
- PRs should include a brief description, testing notes, and linked issues if relevant.
- Include screenshots or short clips for UI changes in `extension/` or dashboard assets.

## Configuration & Runtime Notes
- The server is configured via flags (e.g., `-addr`, `-data-dir`, `-history-max`).
- Use a custom `-data-dir` for local experiments to avoid committing runtime artifacts.
