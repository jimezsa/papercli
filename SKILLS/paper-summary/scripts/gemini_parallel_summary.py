#!/usr/bin/env python3
"""Generate schema-conformant paper summaries from PDFs with Gemini."""

from __future__ import annotations

import argparse
import concurrent.futures
import json
import os
import sys
import time
from dataclasses import dataclass
from pathlib import Path
from typing import Any

try:
    from google import genai
    from google.genai import types
except ImportError:
    genai = None
    types = None


DEFAULT_MODEL = os.getenv("PAPER_SUMMARY_GEMINI_MODEL", "gemini-2.5-pro")
DEFAULT_MAX_OUTPUT_TOKENS = 8192
REQUIRED_SECTION_PREFIXES = [
    "# Paper Extraction Schema:",
    "## 1.",
    "## 2.",
    "## 3.",
    "## 4.",
    "## 5.",
]


@dataclass(frozen=True)
class Job:
    pdf_path: Path
    output_path: Path
    metadata_path: Path | None

    @property
    def safe_id(self) -> str:
        return self.pdf_path.stem


@dataclass(frozen=True)
class Result:
    job: Job
    status: str
    original_id: str | None = None
    message: str | None = None


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate markdown paper summaries from one PDF or a directory of PDFs."
    )
    source_group = parser.add_mutually_exclusive_group(required=True)
    source_group.add_argument("--pdf", type=Path, help="Path to one PDF to summarize.")
    source_group.add_argument(
        "--pdf-dir",
        type=Path,
        help="Directory containing PDF files to summarize in parallel.",
    )
    parser.add_argument(
        "--output-dir",
        type=Path,
        help="Directory for generated markdown files. Defaults to the PDF directory.",
    )
    parser.add_argument(
        "--metadata-dir",
        type=Path,
        help="Directory containing papercli metadata JSON files named <safe_id>.json.",
    )
    parser.add_argument(
        "--summarized-ids",
        type=Path,
        help="Append successful original paper IDs to this file when metadata is available.",
    )
    parser.add_argument(
        "--failures-tsv",
        type=Path,
        help="Append summary failures to this TSV file as stage, id, and reason.",
    )
    parser.add_argument(
        "--model",
        default=DEFAULT_MODEL,
        help=f"Gemini model name to call. Default: {DEFAULT_MODEL}.",
    )
    parser.add_argument(
        "--concurrency",
        type=int,
        default=4,
        help="Maximum number of concurrent Gemini requests in batch mode.",
    )
    parser.add_argument(
        "--max-output-tokens",
        type=int,
        default=DEFAULT_MAX_OUTPUT_TOKENS,
        help="Maximum output tokens requested from Gemini.",
    )
    parser.add_argument(
        "--overwrite",
        action="store_true",
        help="Regenerate summaries even if the markdown file already exists.",
    )
    parser.add_argument(
        "--retries",
        type=int,
        default=3,
        help="Retry count for retryable Gemini API failures.",
    )
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    api_key = os.getenv("GEMINI_API_KEY")
    if not api_key:
        print("GEMINI_API_KEY is not set.", file=sys.stderr)
        return 2
    if genai is None or types is None:
        print(
            "google-genai is not installed. Install it with: python3 -m pip install google-genai",
            file=sys.stderr,
        )
        return 2

    schema_path = Path(__file__).resolve().parent.parent / "references" / "summary_schema.md"
    schema_text = schema_path.read_text(encoding="utf-8").strip()

    jobs = build_jobs(args)
    if not jobs:
        print("No PDF files found to summarize.", file=sys.stderr)
        return 1

    for job in jobs:
        job.output_path.parent.mkdir(parents=True, exist_ok=True)

    workers = min(max(args.concurrency, 1), len(jobs))
    results: list[Result] = []
    with concurrent.futures.ThreadPoolExecutor(max_workers=workers) as executor:
        future_map = {
            executor.submit(run_job, job, args, api_key, schema_text): job for job in jobs
        }
        for future in concurrent.futures.as_completed(future_map):
            results.append(future.result())

    ok_count = 0
    skipped_count = 0
    failed_count = 0
    for result in sorted(results, key=lambda item: item.job.safe_id):
        identifier = result.original_id or result.job.safe_id
        if result.status == "ok":
            ok_count += 1
            print(f"OK   {result.job.pdf_path} -> {result.job.output_path}", file=sys.stderr)
            append_line(args.summarized_ids, identifier)
        elif result.status == "skipped":
            skipped_count += 1
            print(f"SKIP {result.job.output_path} already exists", file=sys.stderr)
        else:
            failed_count += 1
            message = result.message or "unknown summary failure"
            print(f"FAIL {result.job.pdf_path}: {message}", file=sys.stderr)
            append_failure(args.failures_tsv, identifier, message)

    print(
        json.dumps(
            {
                "ok": ok_count,
                "skipped": skipped_count,
                "failed": failed_count,
                "total": len(results),
            },
            indent=2,
        )
    )
    return 0 if failed_count == 0 else 1


def build_jobs(args: argparse.Namespace) -> list[Job]:
    pdf_paths: list[Path]
    if args.pdf is not None:
        pdf_paths = [args.pdf]
    else:
        pdf_paths = sorted(path for path in args.pdf_dir.glob("*.pdf") if path.is_file())

    jobs: list[Job] = []
    for pdf_path in pdf_paths:
        output_dir = args.output_dir or pdf_path.parent
        output_path = output_dir / f"{pdf_path.stem}.md"
        metadata_path = None
        if args.metadata_dir is not None:
            candidate = args.metadata_dir / f"{pdf_path.stem}.json"
            if candidate.exists():
                metadata_path = candidate
        jobs.append(Job(pdf_path=pdf_path, output_path=output_path, metadata_path=metadata_path))
    return jobs


def run_job(
    job: Job,
    args: argparse.Namespace,
    api_key: str,
    schema_text: str,
) -> Result:
    if not job.pdf_path.exists():
        return Result(job=job, status="failed", message="pdf file does not exist")

    if job.output_path.exists() and not args.overwrite:
        return Result(job=job, status="skipped")

    original_id = None
    try:
        metadata = load_metadata(job.metadata_path)
        original_id = extract_original_id(metadata)
        prompt = build_prompt(job, metadata, schema_text)
        summary = generate_summary(
            api_key=api_key,
            model=args.model,
            prompt=prompt,
            pdf_path=job.pdf_path,
            max_output_tokens=args.max_output_tokens,
            retries=args.retries,
        )
        summary = strip_code_fences(summary).strip()
        validate_summary(summary)
        job.output_path.write_text(summary + "\n", encoding="utf-8")
    except Exception as exc:
        return Result(job=job, status="failed", original_id=original_id, message=str(exc))

    return Result(job=job, status="ok", original_id=original_id)


def load_metadata(metadata_path: Path | None) -> dict[str, Any] | None:
    if metadata_path is None:
        return None
    try:
        return json.loads(metadata_path.read_text(encoding="utf-8"))
    except json.JSONDecodeError as exc:
        raise ValueError(f"invalid metadata JSON at {metadata_path}: {exc}") from exc


def extract_original_id(metadata: dict[str, Any] | None) -> str | None:
    if not metadata:
        return None
    candidate = metadata.get("id")
    if isinstance(candidate, str) and candidate.strip():
        return candidate.strip()
    return None


def build_prompt(job: Job, metadata: dict[str, Any] | None, schema_text: str) -> str:
    metadata_block = "None available."
    if metadata is not None:
        metadata_block = json.dumps(metadata, indent=2, ensure_ascii=True)

    return f"""You are generating a deterministic markdown paper summary from a scientific PDF.

Return markdown only. Do not wrap the answer in code fences.
Follow the canonical schema exactly and preserve the section order.
Use the PDF as the primary source of evidence. Use metadata only as fallback and label it clearly.
Keep figures, captions, tables, equations, algorithms, and page anchors as first-class evidence.
If evidence is missing, write the required missing-evidence labels instead of guessing.

Local context:
- PDF path: {job.pdf_path}
- Safe ID: {job.safe_id}
- Metadata path: {job.metadata_path or "None"}

Metadata JSON:
{metadata_block}

Canonical schema:
{schema_text}
"""


def generate_summary(
    *,
    api_key: str,
    model: str,
    prompt: str,
    pdf_path: Path,
    max_output_tokens: int,
    retries: int,
) -> str:
    last_error: Exception | None = None
    for attempt in range(1, max(retries, 1) + 1):
        try:
            client = genai.Client(api_key=api_key)
            document = types.Part.from_bytes(
                data=pdf_path.read_bytes(),
                mime_type="application/pdf",
            )
            response = client.models.generate_content(
                model=model,
                contents=[prompt, document],
                config=types.GenerateContentConfig(
                    temperature=0,
                    max_output_tokens=max_output_tokens,
                    response_mime_type="text/plain",
                ),
            )
            summary = extract_text_from_response(response)
            if summary:
                return summary
            raise RuntimeError("Gemini returned empty content")
        except Exception as exc:
            last_error = exc
            if attempt < retries:
                time.sleep(2 ** (attempt - 1))
                continue
            break

    raise RuntimeError(f"Gemini request failed after retries: {last_error}")


def extract_text_from_response(response: Any) -> str:
    try:
        text = response.text
    except Exception:
        text = None

    if isinstance(text, str) and text.strip():
        return text.strip()

    candidates = getattr(response, "candidates", None)
    if not candidates:
        return ""

    text_parts: list[str] = []
    for candidate in candidates:
        content = getattr(candidate, "content", None)
        parts = getattr(content, "parts", None) or []
        for part in parts:
            part_text = getattr(part, "text", None)
            if isinstance(part_text, str) and part_text.strip():
                text_parts.append(part_text)

    return "".join(text_parts).strip()


def strip_code_fences(text: str) -> str:
    stripped = text.strip()
    if stripped.startswith("```"):
        lines = stripped.splitlines()
        if len(lines) >= 3 and lines[-1].strip() == "```":
            return "\n".join(lines[1:-1]).strip()
    return stripped


def validate_summary(summary: str) -> None:
    positions: list[int] = []
    for prefix in REQUIRED_SECTION_PREFIXES:
        position = summary.find(prefix)
        if position == -1:
            raise ValueError(f"generated summary is missing required section prefix: {prefix}")
        positions.append(position)

    if positions != sorted(positions):
        raise ValueError("generated summary does not preserve the required section order")


def append_line(path: Path | None, line: str | None) -> None:
    if path is None or line is None or not line.strip():
        return
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("a", encoding="utf-8") as handle:
        handle.write(line.strip() + "\n")


def append_failure(path: Path | None, identifier: str, reason: str) -> None:
    if path is None:
        return
    path.parent.mkdir(parents=True, exist_ok=True)
    safe_reason = reason.replace("\t", " ").replace("\n", " ").strip()
    with path.open("a", encoding="utf-8") as handle:
        handle.write(f"summary\t{identifier}\t{safe_reason}\n")


if __name__ == "__main__":
    raise SystemExit(main())
