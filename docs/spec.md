# papercli spec

## Goal

Build a single, clean, high-performance Go CLI that aggregates academic papers and metadata from:

- arXiv
- Semantic Scholar
- Google Scholar (via SerpApi)

This follows the same engineering approach as `jobcli`, but tuned for research workflows:

- **Performance**: Uses Go concurrency (goroutines + context cancellation) to query providers in parallel.
- **Portability**: Single static binary distribution (no Python environment required).
- **Architecture**: Clean, interface-driven design with provider adapters and shared models.

## Non-goals

- A GUI or web app (CLI only).
- Full-text paper hosting or publisher-specific paywall bypassing.
- Citation graph analytics platform (focus is search + retrieval).
- Running a persistent server/API (automation is done via cron + CLI).

## Language/runtime

- Go `1.25` (target, matching `jobcli` conventions)

## CLI framework

- `github.com/alecthomas/kong`
- Root command: `papercli`
- Global flags:
  - `--color=auto|always|never` (default `auto`)
  - `--json` (JSON output to stdout; disables colors)
  - `--plain` (TSV output to stdout; stable/parseable; disables colors)
  - `--verbose` (enable debug logging to stderr)
  - `--version` (print version)

Notes:

- `SilenceUsage: true` behavior via custom error handling.
- `NO_COLOR` is respected.

Environment:

- `PAPERCLI_COLOR=auto|always|never` (default `auto`, overridden by `--color`)
- `PAPERCLI_JSON=1` (default JSON output; overridden by flags)
- `PAPERCLI_VERBOSE=1` (enable debug logs)

## Output (TTY-aware colors)

- `github.com/muesli/termenv` is used for TTY capability detection and color rendering.
- Colors are enabled when:
  - output is a rich terminal and `--color=auto`, and `NO_COLOR` is not set; or
  - `--color=always`
- Colors are disabled when:
  - `--color=never`; or
  - `--json` or `--plain` is set; or
  - `NO_COLOR` is set

Implementation target: `internal/ui/ui.go`.

## Network & Provider Strategy

### HTTP client and retries

- Shared HTTP layer with context deadlines, per-provider timeouts, and retry/backoff for transient failures.
- Respect provider rate limits and `429` handling with jittered backoff.
- User-Agent header set explicitly for each provider client.

Implementation target: `internal/network/client.go`.

### Provider adapters

- `arXiv`:
  - Endpoint: `https://export.arxiv.org/api/query`
  - Parsing: Atom XML feed
  - Author query mode uses `au:<name>` syntax
- `Semantic Scholar`:
  - Endpoint base: `https://api.semanticscholar.org/graph/v1`
  - Author search via `/author/search`
  - Optional API key for higher quota
- `Google Scholar` via SerpApi:
  - Endpoint: `https://serpapi.com/search`
  - `engine=google_scholar` for paper search
  - `engine=google_scholar_profiles` for author lookup
  - API key required

Implementation target: `internal/provider/*`.

## Config layout

- Base config dir: `$(os.UserConfigDir())/papercli/`
- Files:
  - `config.json` (JSON defaults + provider credentials)
  - `cache.db` (optional; local response metadata cache)

Environment:

- `PAPERCLI_SEMANTIC_API_KEY=...`
- `PAPERCLI_SERPAPI_KEY=...`
- `PAPERCLI_DEFAULT_PROVIDER=all`
- `PAPERCLI_DEFAULT_LIMIT=20`

Flag aliases:

- `--out` also accepts `--output`.
- `--file` also accepts `--output`.

## Commands (current + planned)

### Planned

- `papercli version`
- `papercli config init` (writes default `config.json`)
- `papercli config path`
- `papercli search <query> [--provider P] [--limit N] [--offset N]`
- `papercli author <name> [--provider P] [--limit N]`
- `papercli info <id> [--provider P]`
- `papercli download <id> [--provider P] [--out PATH]`
- `papercli seen diff --new A.json --seen B.json --out C.json [--stats]`
- `papercli seen update --seen B.json --input C.json --out B.json [--stats]`

Search/author flags include:

- `--provider`: `arxiv|semantic|scholar|all` (default `all`)
- `--sort`: `relevance|date|citations` (provider-dependent)
- `--year-from`: lower publication year bound
- `--year-to`: upper publication year bound
- `--limit`: result cap
- `--format`: `csv|json|md`
- `--links`: `short|full` (table URL display)
- `--seen`: seen-history JSON path
- `--new-only`: output only unseen papers (requires `--seen`)
- `--new-out`: always write unseen JSON to file (requires `--seen`)

## Output formats

Default: Human-friendly tables (`text/tabwriter`) to `stdout` (columns: provider/title/authors/year/url).

- **JSON**: `--json` dumps unified `Paper` structs for piping to `jq`.
- **TSV**: `--plain` outputs stable tab-separated values.
- **CSV**: `--format=csv` writes standard CSV (default for file output).
- **Markdown**: `--format=md` writes a shareable literature-summary list.

## Code layout

- `cmd/papercli/main.go` - binary entrypoint
- `internal/cmd/*` - `kong` command structs
- `internal/ui/*` - color + rendering
- `internal/config/*` - config paths + file parsing
- `internal/network/*` - HTTP client, retry, timeout policy
- `internal/provider/*` - provider interfaces + implementations
  - `interface.go` (`Provider` interface)
  - `arxiv.go`
  - `semanticscholar.go`
  - `serpapi.go`
- `internal/models/*` - shared structs (`Paper`, `SearchParams`)
- `internal/seen/*` - seen-ID normalization, diff/merge, JSON I/O
- `internal/export/*` - writers for CSV/JSON/Markdown

## Dependencies (planned)

- `github.com/alecthomas/kong` (CLI framework)
- `github.com/muesli/termenv` (rich terminal output)
- `github.com/rs/zerolog` (high-performance logging)
- `github.com/mmcdole/gofeed` or stdlib XML parsing (Atom parsing for arXiv)

## Formatting, linting, tests

### Formatting

Pinned tools in local `.tools/` via `make tools`:

- `mvdan.cc/gofumpt@v0.7.0`
- `golang.org/x/tools/cmd/goimports@v0.38.0`
- `github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2`

Commands:

- `make fmt` - applies `goimports` + `gofumpt`
- `make fmt-check` - formats and fails if Go files or `go.mod/go.sum` change

### Lint

- `golangci-lint` with config in `.golangci.yml`
- `make lint`

### Tests

- stdlib `testing`
- `make test` - runs unit tests (parsers, normalization, exporters)

### Integration tests (local only)

Opt-in integration tests guarded by build tags (not run in CI). These hit live APIs.

- Requires:
  - working internet connection
  - API keys for optional providers
- Run:
  - `go test -tags=integration ./internal/provider/...`
  - note: can be flaky due to upstream quotas/latency

## CI (GitHub Actions)

Workflow target: `.github/workflows/ci.yml`

- runs on push + PR
- uses `actions/setup-go` with `go-version-file: go.mod`
- runs:
  - `make tools`
  - `make fmt-check`
  - `go test ./...` (unit tests only)
  - `golangci-lint`

## Next implementation steps

1.  **Skeleton**: Initialize `kong` command tree, logging, and configuration loading.
2.  **Models**: Define unified `Paper` and `SearchParams` structs.
3.  **Network Layer**: Implement shared HTTP client with timeout/retry/rate-limit handling.
4.  **First Provider (arXiv)**: Implement XML parsing, pagination, and author query mode.
5.  **Concurrency**: Add multi-provider fan-out/fan-in execution with context cancellation.
6.  **Export + Seen**: Implement JSON/CSV/Markdown writers and seen diff/update commands.
