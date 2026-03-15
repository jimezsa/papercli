# Changelog

## [Unreleased]

### Added

- Added concrete terminal usage examples to `papercli` global help output and `papercli --help`
- Added the `paper-summary` skill workflow with a default parallel summary count of 20
- Added rigorous mathematical formalization guidance to the `paper-summary` schema, including Markdown math formatting and explicit use of PDF figures and architecture diagrams as evidence
- Added local `.env.local` loading support to the `paper-summary` Gemini script
- Added support for metadata inputs that are either a single object or a list of objects

### Fixed

- Ignored the local `/research/` workspace path in Git tracking

## [0.1.1] - 2026-03-01

### Fixed

- Fixed `make fmt-check` to ignore Go files in local `.cache/` so CI format checks run only on repository source files

## [0.1.0] - 2026-03-01

Initial public release of `papercli` CLI for searching academic papers across arXiv, Semantic Scholar, and Google Scholar (via SerpApi).

### Added

- Initial release of PaperCLI
- Provider support for arXiv, Semantic Scholar, and SerpApi-backed Google Scholar
- Search and author workflows from the CLI
- Export options including Markdown output
- GoReleaser configuration for automated cross-platform builds
- GitHub Actions release workflow with Homebrew tap publishing support
