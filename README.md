# papercli

CLI for searching academic papers across arXiv, Semantic Scholar, and Google Scholar (via SerpApi).

## Quick start

```bash
go run ./cmd/papercli --version
go run ./cmd/papercli search "graph neural networks" --provider arxiv --limit 5
go run ./cmd/papercli config init
```

## Example usage

```bash
# 1) Initialize config (creates default config file)
go run ./cmd/papercli config init

# 2) Search recent papers from Semantic Scholar and save as Markdown
go run ./cmd/papercli search "retrieval augmented generation" \
  --provider semantic \
  --year-from 2023 \
  --sort date \
  --limit 10 \
  --format md \
  --out rag-papers.md

# 3) Show only new papers compared with your seen database
go run ./cmd/papercli search "retrieval augmented generation" \
  --provider semantic \
  --seen .papercli-seen.json \
  --new-only
```

## Build binary with make

```bash
make build
./bin/papercli --version
```
