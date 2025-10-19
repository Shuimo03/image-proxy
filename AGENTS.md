# Repository Guidelines

## Project Structure & Module Organization
The proxy is implemented in Go. Place runnable entry points under `cmd/image-proxy` and feature packages in `internal/` (for example `internal/http`, `internal/cache`). Shared utilities that may be reused externally belong in `pkg/`. Integration smoke tests can stage assets in `testdata/`, while developer scripts live in `scripts/`. Commit example configuration templates to `configs/` and keep secrets out of version control.

## Build, Test, and Development Commands
`go build ./cmd/image-proxy` compiles the binary and surfaces compile-time issues. `go run ./cmd/image-proxy --config configs/dev.yaml` is the fastest way to exercise the proxy locally; point it at a small sample origin host. `go test ./...` runs the unit and integration suites. Run `golangci-lint run` before opening a pull request; lint configuration will live in `.golangci.yml`.

## Coding Style & Naming Conventions
Format Go code with `go fmt` (or `gofumpt`) and organize imports using `goimports` before committing. Favour clear package names such as `cache`, `transform`, or `proxy`. Exported identifiers use CamelCase, while internals prefer lowerCamelCase. Keep files focused: handlers in `handler.go`, transport concerns in `transport.go`, etc. Document public functions with succinct GoDoc comments when behaviour is non-obvious.

## Testing Guidelines
Write table-driven tests named `<feature>_test.go` beside the code under test. Use sub-tests (`t.Run`) to highlight variants such as cache hits/misses or invalid URLs. Place golden images or sample responses under `testdata/` and reference them with relative paths. Maintain healthy coverage of edge cases such as timeouts, content-type mismatches, and redirect loops.

## Commit & Pull Request Guidelines
Use short, imperative commit messages following the pattern `<area>: <action>` (e.g., `proxy: add cache key hashing`). Squash fix-ups locally. Pull requests should describe the change, list manual or automated verification, and call out risk areas. Link issues when available and include before/after evidence (logs, curl samples, or screenshots) for behaviour changes.

## Security & Configuration Tips
Never proxy arbitrary hosts by default; update allowlists in `configs/` and keep credentials in environment variables. Sanitize user-supplied URLs and strip unsupported headers upstream. When adding new dependencies, prefer well-maintained libraries and document any licensing implications alongside GPLv2.
