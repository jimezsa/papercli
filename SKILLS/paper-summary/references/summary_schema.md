# Paper Extraction Schema: <Paper Title>

Rules:
- Keep the section order unchanged for deterministic parsing.
- Ingest the PDF directly first so figures, tables, captions, and layout remain available.
- Anchor each section to observable PDF evidence such as figures, captions, equations, tables, and appendix material.
- Record exact figure, table, equation, algorithm, and page anchors whenever available. If you cannot locate one, write `Anchor not located in available evidence.`.
- Use metadata only as fallback and label it clearly.
- Do not invent equations, datasets, metrics, links, or foundation papers.
- If evidence is missing, write `Not clearly stated in available evidence.`.
- If a statement is an inference rather than an explicit claim, label it `Inference from available evidence: ...`.

## 1. The Why (Motivation & Core Problem)
- The Problem: What specific limitation in existing research or technology is this paper trying to solve? Keep this to 1-2 sentences.
- The Core Idea: What is the authors' main hypothesis or novel approach to solving this problem?
- Evidence Anchors: Exact page, figure, table, or equation anchors supporting the problem framing.

## 2. Main Architecture (Mathematical Formalization)
Agent instruction: Extract the core methodology and represent it strictly as a sequence of mathematical operations, data flows, and loss functions. Write equations in Markdown using fenced `math` blocks so they remain readable in raw `.md` and machine-extractable.
The main architecture can often be extracted directly from the architecture image or pipeline diagram in the paper.
If the method is explained primarily through a figure or diagram, use the PDF figure as evidence and translate it into equations and ordered data flow.
If the paper has no learnable architecture (for example a survey, benchmark, theorem, dataset, or systems paper), replace this section with `Algorithm / theorem / protocol flow` and formalize the central steps instead.
Use `Loss / Optimization: Not applicable.` when no training objective exists.
Formatting rules for this section:
- Use a short label such as `Input`, `Forward Pass`, `Update Rule`, `Objective`, or `Algorithm Step k` before each math block.
- Put each standalone equation or tightly related equation group in its own fenced `math` block.
- Use inline math such as `$X$` only for short variable mentions inside sentences or bullets.
- Prefer aligned multi-line math when the paper presents a sequence, derivation, or factorized objective.
- Keep notation faithful to the paper; if you rename a symbol for clarity, say `Inference from available evidence: renamed ... for consistency.`.
- Do not paste raw LaTeX delimiters like `\[` or `\]` directly into prose outside a fenced `math` block.

Input:

```math
X = \text{...}
```

Forward Pass:

```math
\begin{aligned}
H_1 &= f_{\text{module\_1}}(X) \\
H_2 &= f_{\text{module\_2}}(H_1) \\
\hat{Y} &= f_{\text{head}}(H_2)
\end{aligned}
```

Loss / Optimization:

```math
\mathcal{L}_{\text{total}} = \lambda_1 \mathcal{L}_{\text{task}} + \lambda_2 \mathcal{L}_{\text{reg}}
```

Evidence Anchors: Exact figure, equation, algorithm, and page anchors used for this formalization.

## 3. The Why of the Architecture (Component Rationale)
Agent instruction: For every variable and function defined in Section 2, explain exactly why it was chosen or designed that way.
If Section 2 is non-architectural, explain why each algorithmic step, theorem component, protocol stage, or evaluation stage exists instead of model modules.

- $X$: Why is the input represented this way?
- $f_{\text{module\_1}}$: Why use this specific module?
- $f_{\text{module\_2}}$: Why is this step necessary?
- $f_{\text{head}}$: Why this prediction head?
- $\mathcal{L}_{\text{task}}$: Why this task objective?
- $\mathcal{L}_{\text{reg}}$: Why use this specific regularizer?
- Evidence Anchors: Exact page, figure, or appendix anchors supporting the rationale.

## 4. Metrics & Evaluation
- Datasets Used: List the primary benchmarks.
- Key Metrics: How is success quantified?
- The Result: One sentence summarizing the paper's main performance claim.
- Visual Evidence: Note the key figure, table, or ablation that best supports the reported result when one is clearly present.
- Evidence Anchors: Exact table, figure, ablation, and page anchors supporting the reported results.

## 5. Relevant Links & Knowledge Anchors
- Project Page / GitHub: Link if available in the paper or metadata.
- Core Foundation Paper: The 1 or 2 most relied-upon prior papers, if the dependency is clear from the text.
- Evidence Anchors: Exact reference numbers, appendix pages, or metadata fields used to identify these links and foundation papers.

Summary requirements:
- Keep the section order unchanged.
- Express the main method as Markdown-native math: use fenced `math` blocks for standalone equations and inline math for short symbol references.
- Structure Section 2 as labeled math steps so both humans and parsers can recover the input, transformations, and objective.
- If no learnable architecture exists, switch Section 2 to `Algorithm / theorem / protocol flow` and write `Loss / Optimization: Not applicable.` instead of inventing modules.
- Explain why each Section 2 variable, module, and loss term exists.
- Preserve figure/table evidence when it carries the method, mechanism, or strongest empirical support.
- Record exact evidence anchors in each section.
- Label missing evidence explicitly instead of guessing.
- Mark inferred statements as `Inference from available evidence: ...`.
