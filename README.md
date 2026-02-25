# 📄 papercli

<p align="center">
  <img src="docs/assets/papercli.png" alt="papercli banner" width="640"/>
</p>

CLI for searching academic papers across arXiv, Semantic Scholar, and Google Scholar (via SerpApi).

## Get the repository

### Option 1: Clone with git

```bash
git clone https://github.com/<your-user>/papercli.git
cd papercli
make build
cp ./bin/papercli ./papercli
./papercli --version
```

## Example usage

```bash
# Initialize config
./papercli config init

# Search papers
./papercli search "graph neural networks" --provider arxiv --limit 5

# Save results to Markdown
./papercli search "retrieval augmented generation" \
  --provider semantic \
  --year-from 2023 \
  --sort date \
  --limit 10 \
  --format md \
  --out rag-papers.md
```
