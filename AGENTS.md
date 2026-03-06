# Repository Guidelines

## Project Structure & Module Organization

`papercli` is a Go CLI organized by responsibility:

- `cmd/papercli/main.go`: binary entrypoint.
- `internal/cmd/`: command parsing, dispatch, and help output.
- `internal/provider/`: provider adapters (`arxiv`, `semantic`, `serpapi`) and provider manager.
- `internal/network/`: shared HTTP client, retry/backoff logic.
- `internal/models/`: shared data models (`Paper`, query params).
- `internal/export/`, `internal/ui/`, `internal/seen/`, `internal/config/`: output, terminal rendering, seen-state, and config management.
- `docs/spec.md`: product/technical spec.

Tests live next to code as `*_test.go` (for example `internal/provider/arxiv_test.go`).

## Available Skills

| Skill         | Purpose                                                                                | Best use case                                                                       |
| ------------- | -------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `fast-search` | Rapid, evidence-grounded paper scouting with a referenced `findings.md`.               | Quick orientation on a topic with 3-6 papers and key equations.                     |
| `pro-search`  | Professional medium-depth literature synthesis with cross-paper comparisons.           | Serious research questions needing 8-12 papers and explicit tradeoff analysis.      |
| `deep-search` | Institutional-grade deep investigation with iterative waves and exhaustive references. | State-of-the-art reviews, strategy decisions, and high-confidence evidence mapping. |

## Build, Test, and Development Commands

Use the `Makefile` targets:

- `make build`: builds `bin/papercli`.
- `make run`: builds and prints version.
- `make test`: runs `go test ./...`.
- `make fmt`: runs `goimports` + `gofumpt` (after `make tools`).
- `make lint`: runs `golangci-lint`.
- `make clean`: removes `bin/` and local `.cache/`.

Direct Go command example: `go test ./...`.

## Coding Style & Naming Conventions

- Follow standard Go formatting (`gofmt`-compatible); run `make fmt`.
- Keep package names lowercase and focused (`provider`, `network`, `cmd`).
- Use clear exported names (`NewApp`, `SearchParams`) and concise unexported helpers.
- Wrap errors with context (`fmt.Errorf("...: %w", err)`).
- Keep command-specific logic in `internal/cmd` and provider-specific API logic in `internal/provider`.

## Testing Guidelines

- Framework: Go standard library `testing`.
- Naming: files `*_test.go`, tests `TestXxx`.
- Prefer fast, deterministic unit tests (parsers, exporters, seen-state logic).
- Avoid live API dependency in default tests; if adding integration tests, guard them with tags.
- Ensure `make test` passes before opening a PR.

## Commit & Pull Request Guidelines

- Use Conventional Commits format:
  - `<type>(<optional-scope>): <short imperative summary>`
  - Optional body for rationale and impact.
  - Optional footer for breaking changes and issue references.
- Preferred commit types:
  - `feat`: new user-facing behavior
  - `fix`: bug fix
  - `refactor`: internal change without behavior change
  - `docs`: documentation-only change
  - `test`: tests added or updated
  - `chore`: maintenance/tooling changes
- Example commit messages:
  - `feat(search): add provider fallback for empty responses`
  - `fix(arxiv): handle missing DOI without crashing`
  - `docs(readme): clarify semantic provider API key setup`
  - `test(seen): cover duplicate id merge behavior`
- Keep commits focused to one logical change.
- PRs should include:
  - concise summary of behavior changes,
  - key files touched,
  - validation output (for example `make build`, `make test`),
  - linked issue/task when applicable.

## Security & Configuration Tips

- Never commit API keys.
- Use environment variables for credentials:
  - `PAPERCLI_SEMANTIC_API_KEY`
  - `PAPERCLI_SERPAPI_KEY`
- Initialize user config with `papercli config init`.
