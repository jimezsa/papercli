---
name: fast-search
description: Fast scientific paper scouting with papercli. Search, download, read, and produce a referenced markdown findings file with core ideas, concepts, and key math.
homepage: https://github.com/jimezsa/papercli
metadata:
  {
    "openclaw":
      {
        "emoji": "📄",
        "os": ["linux", "darwin"],
        "requires": { "bins": ["papercli"] },
        "install":
          [
            {
              "id": "homebrew",
              "kind": "shell",
              "script": "brew install jimezsa/tap/papercli",
              "bins": ["papercli"],
              "label": "Install PaperCLI with Homebrew",
            },
            {
              "id": "source",
              "kind": "shell",
              "script": "git clone https://github.com/jimezsa/papercli.git && cd papercli && make build && sudo install -m 0755 ./bin/papercli /usr/local/bin/papercli",
              "bins": ["papercli"],
              "label": "Build PaperCLI from source",
            },
          ],
      },
  }
---

# Fast Search Skill

Use this skill for a rapid, evidence-grounded literature brief when the user needs quick scientific orientation without sacrificing traceability.

## Update This Skill

Only do this if the user explicitly asks to update this skill from the GitHub repo.

To refresh this skill directly from the GitHub repo:

```bash
curl -fsSL https://raw.githubusercontent.com/jimezsa/papercli/main/SKILLS/fast-search/SKILL.md \
  -o SKILLS/fast-search/SKILL.md
```

## Mission

Given a research question, use `papercli` to:
1. Search relevant papers.
2. Download a focused core set of PDFs.
3. Read enough content to extract core ideas, concepts, and key equations.
4. Produce a detailed `findings.md` report with inline references tied to exact papers.

## Prerequisites

- `papercli` is installed and available in `PATH`.
- Optional provider keys:
  - `PAPERCLI_SEMANTIC_API_KEY`
  - `PAPERCLI_SERPAPI_KEY`
- If keys are missing, continue with `arxiv` coverage and document the reduced coverage.

## Required Inputs

- Research question or hypothesis.
- Optional scope constraints: years, domain, must-include authors, method family.

If inputs are missing, infer a minimal scope and proceed.

## Hard Requirements

- Always use `papercli` for retrieval (`search`, `info`, `download`).
- Download and read papers, not just metadata.
- Every factual claim must be grounded by references.
- Include key math when present in papers.
- Final output must be a markdown file named `findings.md`.

## Workflow

### 1. Setup workspace

```bash
mkdir -p research/{search,meta,pdf,text,notes}
```

Initialize config when needed:

```bash
papercli config init
```

### 2. Run fast retrieval pass

Use one tight query and one alternate phrasing:

```bash
papercli search "<query>" \
  --provider all \
  --sort relevance \
  --limit 15 \
  --format json \
  --out research/search/seed.json

papercli search "<alternate query>" \
  --provider all \
  --sort date \
  --year-from <optional_year> \
  --limit 10 \
  --format json \
  --out research/search/recency.json
```

### 3. Select and enrich 3-6 papers

Prioritize relevance, recency, and diversity of approach.

```bash
jq -r '.[].id' research/search/seed.json research/search/recency.json | \
  awk 'NF && !seen[$0]++' | head -n 6 > research/meta/selected_ids.txt
```

For each selected paper, fetch metadata and PDF:

```bash
while read -r id; do
  safe_id="$(echo "$id" | tr '/:' '__')"
  papercli info "$id" --provider all --format json --out "research/meta/${safe_id}.json" || true
  papercli download "$id" --provider all --out "research/pdf/${safe_id}.pdf" || true
done < research/meta/selected_ids.txt
```

### 4. Create agent-ready paper summaries

Read each PDF directly first so figures, tables, and layout cues remain available during analysis.

Direct-ingestion priorities:

- Inspect the abstract, introduction, method section, main architecture figure, result tables, ablations, and conclusion.
- Treat figures, captions, and tables as first-class evidence.
- When the method is clearer in a diagram than in prose, translate the diagram into equations and data flow in the summary.
- If the PDF cannot be inspected locally, keep the paper in references, use metadata where possible, and mark the summary as metadata-only evidence.

Then create one deterministic summary per paper at `research/notes/<safe_id>.md` using this exact schema:

```markdown
# Paper Extraction Schema: <Paper Title>

Rules:
- Keep the section order unchanged for deterministic parsing.
- Ingest the PDF directly first so figures, tables, captions, and layout remain available.
- Anchor each section to observable PDF evidence such as figures, captions, equations, tables, and appendix material.
- Use metadata only as fallback and label it clearly.
- Do not invent equations, datasets, metrics, links, or foundation papers.
- If evidence is missing, write `Not clearly stated in available evidence.`.
- If a statement is an inference rather than an explicit claim, label it `Inference from available evidence: ...`.

## 1. The Why (Motivation & Core Problem)
- The Problem: What specific limitation in existing research or technology is this paper trying to solve? Keep this to 1-2 sentences.
- The Core Idea: What is the authors' main hypothesis or novel approach to solving this problem?

## 2. Main Architecture (Mathematical Formalization)
Agent instruction: Extract the core methodology and represent it strictly as a sequence of mathematical operations, data flows, and loss functions. Use standard LaTeX notation.
If the method is explained primarily through a figure or diagram, use the PDF figure as evidence and translate it into equations and ordered data flow.

Input:
\[
X = \text{...}
\]

Forward Pass:
\[
H_1 = f_{\text{module\_1}}(X)
\]
\[
H_2 = f_{\text{module\_2}}(H_1)
\]
\[
\hat{Y} = f_{\text{head}}(H_2)
\]

Loss / Optimization:
\[
\mathcal{L}_{\text{total}} = \lambda_1 \mathcal{L}_{\text{task}} + \lambda_2 \mathcal{L}_{\text{reg}}
\]

## 3. The Why of the Architecture (Component Rationale)
Agent instruction: For every variable and function defined in Section 2, explain exactly why it was chosen or designed that way.

- $X$: Why is the input represented this way?
- $f_{\text{module\_1}}$: Why use this specific module?
- $f_{\text{module\_2}}$: Why is this step necessary?
- $f_{\text{head}}$: Why this prediction head?
- $\mathcal{L}_{\text{task}}$: Why this task objective?
- $\mathcal{L}_{\text{reg}}$: Why use this specific regularizer?

## 4. Metrics & Evaluation
- Datasets Used: List the primary benchmarks.
- Key Metrics: How is success quantified?
- The Result: One sentence summarizing the paper's main performance claim.
- Visual Evidence: Note the key figure, table, or ablation that best supports the reported result when one is clearly present.

## 5. Relevant Links & Knowledge Anchors
- Project Page / GitHub: Link if available in the paper or metadata.
- Core Foundation Paper: The 1 or 2 most relied-upon prior papers, if the dependency is clear from the text.
```

Summary requirements:
- Keep the section order unchanged.
- Express the main method as LaTeX equations plus data flow and loss terms.
- Explain why each Section 2 variable, module, and loss term exists.
- Preserve figure/table evidence when it carries the method or key result.
- Label missing evidence explicitly instead of guessing.

### 5. Produce `findings.md`

Target quality: fast but technically useful.

- Include 3-6 referenced papers.
- Provide a compact synthesis of core ideas.
- Include at least 2 key equations from the corpus when available.
- Use the per-paper schemas in `research/notes/` as the primary synthesis substrate.

## Output Contract (`findings.md`)

Use this structure:

```markdown
# Findings: <topic>

## Scope
- Question: ...
- Coverage window: ...
- Selection criteria: ...

## Core Ideas
Claim with inline refs [R1][R3].
Claim with inline refs [R2].

## Key Concepts
- Concept A: definition and role [R1].
- Concept B: definition and trade-off [R2][R4].

## Key Math
\[
<equation>
\]
Meaning and why it matters [R3].

\[
<equation>
\]
Meaning and assumptions [R2].

## Paper Notes
### [R1] <title>
- Problem:
- Method:
- Main result:
- Limits:

### [R2] <title>
- Problem:
- Method:
- Main result:
- Limits:

## References
| Ref | Paper | Provider ID | Year | Evidence |
|---|---|---|---|---|
| R1 | Title... | arxiv:... | 2024 | `notes/...md`, `pdf/...pdf` |
| R2 | Title... | semantic:... | 2023 | `notes/...md`, `pdf/...pdf` |
```

## Referencing Rules

- Use `[R#]` inline citations in all analytical sections.
- Do not cite claims without evidence.
- For equation-based claims, cite the source paper on the same line.
- Keep quotes short; prefer paraphrase plus citation.

## Done Criteria

- `findings.md` exists and is detailed.
- Claims are referenced.
- Papers were downloaded and read.
- Each selected paper has an agent-ready summary in `research/notes/` unless extraction failed.
- Core ideas, concepts, and key math are covered.
