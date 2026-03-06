# 📚 papercli — Scientific papers from your terminal

<p align="center">
  <img src="docs/assets/papercli.png" alt="papercli banner" width="640"/>
</p>

<p align="center">
  <a href="https://github.com/jimezsa/papercli/actions/workflows/ci.yml?branch=main"><img src="https://img.shields.io/github/actions/workflow/status/jimezsa/papercli/ci.yml?branch=main&style=for-the-badge" alt="CI status"></a>
  <a href="https://github.com/jimezsa/papercli/releases"><img src="https://img.shields.io/github/v/release/jimezsa/papercli?include_prereleases&style=for-the-badge" alt="GitHub release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg?style=for-the-badge" alt="MIT License"></a>
</p>
CLI for searching academic papers across arXiv, Semantic Scholar, and Google Scholar (via SerpApi).

### Skills (perform deepresearch only on papers)

| Skill         | Purpose                                                                                | Best use case                                                                       |
| ------------- | -------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `fast-search` | Rapid, evidence-grounded paper scouting with a referenced `findings.md`.               | Quick orientation on a topic with 3-6 papers and key equations.                     |
| `pro-search`  | Professional medium-depth literature synthesis with cross-paper comparisons.           | Serious research questions needing 8-12 papers and explicit tradeoff analysis.      |
| `deep-search` | Institutional-grade deep investigation with iterative waves and exhaustive references. | State-of-the-art reviews, strategy decisions, and high-confidence evidence mapping. |

## Installation

### Option 1: Homebrew (macOS/Linux)

```bash
brew install jimezsa/tap/papercli
```

### Option 2: Clone with git

```bash
git clone https://github.com/jimezsa/papercli.git
cd papercli
make build
cp ./bin/papercli ./papercli
./papercli --version
```

## Example usage

```bash
# Initialize config
./papercli config init

# Overwrite existing config intentionally
./papercli config init --force

# Search papers
./papercli search "graph neural networks" --provider arxiv --limit 5

# Global flags can appear before or after the command
./papercli --json search "graph neural networks" --provider arxiv --limit 3
./papercli search "graph neural networks" --provider arxiv --limit 3 --json

# Save results to Markdown
./papercli search "retrieval augmented generation" \
  --provider semantic \
  --year-from 2023 \
  --sort date \
  --limit 10 \
  --format md \
  --out rag-papers.md
```

## More examples

```bash
# Find papers by author
./papercli author "Yoshua Bengio" \
  --provider semantic \
  --sort relevance \
  --limit 10

# Download a paper PDF by ID
./papercli download "1706.03762" \
  --provider arxiv \
  --out attention-is-all-you-need.pdf
```

## Commands and flags

### Commands

| Command                                                                    | Description                                                           |
| -------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `papercli version`                                                         | Print CLI version.                                                    |
| `papercli config init [--force]`                                           | Initialize default config file and print its path.                    |
| `papercli config path`                                                     | Print config file path.                                               |
| `papercli search <query>`                                                  | Search papers by query across one or more providers.                  |
| `papercli author <name>`                                                   | Search papers by author name.                                         |
| `papercli info <id>`                                                       | Fetch paper metadata by provider identifier.                          |
| `papercli download <id>`                                                   | Download paper PDF using provider metadata.                           |
| `papercli seen diff --new A.json --seen B.json --out C.json [--stats]`     | Write papers in `A.json` that are not present in seen store `B.json`. |
| `papercli seen update --seen B.json --input C.json --out B.json [--stats]` | Update seen store with papers from input JSON.                        |

### Global flags

| Flag           | Description                              | Values / Default                            |
| -------------- | ---------------------------------------- | ------------------------------------------- |
| `--color`      | Color output mode.                       | `auto`, `always`, `never` (default: `auto`) |
| `--json`       | Output JSON to stdout (disables colors). | boolean                                     |
| `--plain`      | Output TSV to stdout (disables colors).  | boolean                                     |
| `--verbose`    | Enable debug logging to stderr.          | boolean                                     |
| `--version`    | Print version and exit.                  | boolean                                     |
| `--help`, `-h` | Show help.                               | boolean                                     |

### Shared flags for `search` and `author`

| Flag                | Description                                                           | Values / Default                                        |
| ------------------- | --------------------------------------------------------------------- | ------------------------------------------------------- |
| `--provider`        | Provider to query.                                                    | `arxiv`, `semantic`, `scholar`, `all` (default: `all`)  |
| `--sort`            | Sort mode (provider dependent; unsupported sorts warn and fall back). | `relevance`, `date`, `citations` (default: `relevance`) |
| `--year-from`       | Lower publication year bound.                                         | integer                                                 |
| `--year-to`         | Upper publication year bound.                                         | integer                                                 |
| `--limit`           | Maximum number of results.                                            | integer (default from config or `20`)                   |
| `--offset`          | Result offset.                                                        | integer (default: `0`)                                  |
| `--format`          | Output format.                                                        | `csv`, `json`, `md`                                     |
| `--links`           | Link rendering mode for table output.                                 | `short`, `full` (default: `full`)                       |
| `--seen`            | Seen-history JSON file path.                                          | path                                                    |
| `--new-only`        | Output only unseen papers (requires `--seen`).                        | boolean                                                 |
| `--new-out`         | Always write unseen papers JSON (requires `--seen`).                  | path                                                    |
| `--out`, `--output` | Output file path.                                                     | path                                                    |

### Flags for `info`

| Flag                | Description          | Values / Default                                       |
| ------------------- | -------------------- | ------------------------------------------------------ |
| `--provider`        | Provider to query.   | `arxiv`, `semantic`, `scholar`, `all` (default: `all`) |
| `--format`          | Output format.       | `csv`, `json`, `md` (default: `json`)                  |
| `--links`           | Link rendering mode. | `short`, `full` (default: `full`)                      |
| `--out`, `--output` | Output file path.    | path                                                   |

### Flags for `download`

| Flag                          | Description        | Values / Default                                       |
| ----------------------------- | ------------------ | ------------------------------------------------------ |
| `--provider`                  | Provider to query. | `arxiv`, `semantic`, `scholar`, `all` (default: `all`) |
| `--out`, `--output`, `--file` | Output PDF path.   | path (default: derived from paper ID)                  |

### Flags for `seen diff`

| Flag                          | Description                 | Values / Default |
| ----------------------------- | --------------------------- | ---------------- |
| `--new`                       | Input papers JSON path.     | path (required)  |
| `--seen`                      | Seen JSON path.             | path (required)  |
| `--out`, `--output`, `--file` | Output JSON path.           | path (required)  |
| `--stats`                     | Print diff stats to stderr. | boolean          |

### Flags for `seen update`

| Flag                          | Description                   | Values / Default |
| ----------------------------- | ----------------------------- | ---------------- |
| `--seen`                      | Current seen JSON path.       | path (required)  |
| `--input`                     | Input papers JSON path.       | path (required)  |
| `--out`, `--output`, `--file` | Updated seen JSON path.       | path (required)  |
| `--stats`                     | Print update stats to stderr. | boolean          |

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
