#!/usr/bin/env python3
"""Generate schema-conformant paper summaries from PDFs with Gemini."""

from __future__ import annotations

import argparse
import base64
import concurrent.futures
import json
import os
import sys
import time
import urllib.error
import urllib.parse
import urllib.request
from dataclasses import dataclass
from pathlib import Path
from typing import Any


DEFAULT_MODEL = os.getenv("PAPER_SUMMARY_GEMINI_MODEL", "gemini-2.5-pro")
DEFAULT_TIMEOUT_SECONDS = 300
DEFAULT_MAX_OUTPUT_TOKENS = 8192
RETRYABLE_STATUS_CODES = {429, 500, 502, 503, 504}
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
        "--timeout-seconds",
        type=int,
        default=DEFAULT_TIMEOUT_SECONDS,
        help="HTTP timeout for each Gemini request.",
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
            pdf_bytes=job.pdf_path.read_bytes(),
            timeout_seconds=args.timeout_seconds,
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
    pdf_bytes: bytes,
    timeout_seconds: int,
    max_output_tokens: int,
    retries: int,
) -> str:
    payload = {
        "contents": [
            {
                "role": "user",
                "parts": [
                    {"text": prompt},
                    {
                        "inline_data": {
                            "mime_type": "application/pdf",
                            "data": base64.b64encode(pdf_bytes).decode("ascii"),
                        }
                    },
                ],
            }
        ],
        "generationConfig": {
            "temperature": 0,
            "candidateCount": 1,
            "maxOutputTokens": max_output_tokens,
            "responseMimeType": "text/plain",
        },
    }

    url = (
        "https://generativelanguage.googleapis.com/v1beta/models/"
        f"{urllib.parse.quote(model, safe='')}:generateContent?key="
        f"{urllib.parse.quote(api_key, safe='')}"
    )
    request_body = json.dumps(payload).encode("utf-8")

    for attempt in range(1, max(retries, 1) + 1):
        try:
            request = urllib.request.Request(
                url,
                data=request_body,
                headers={"Content-Type": "application/json"},
                method="POST",
            )
            with urllib.request.urlopen(request, timeout=timeout_seconds) as response:
                response_body = response.read().decode("utf-8")
            return extract_text_from_response(response_body)
        except urllib.error.HTTPError as exc:
            detail = exc.read().decode("utf-8", errors="replace")
            if exc.code in RETRYABLE_STATUS_CODES and attempt < retries:
                time.sleep(2 ** (attempt - 1))
                continue
            raise RuntimeError(f"Gemini API error {exc.code}: {detail}") from exc
        except urllib.error.URLError as exc:
            if attempt < retries:
                time.sleep(2 ** (attempt - 1))
                continue
            raise RuntimeError(f"Gemini network error: {exc}") from exc

    raise RuntimeError("Gemini request failed after retries")


def extract_text_from_response(response_body: str) -> str:
    try:
        payload = json.loads(response_body)
    except json.JSONDecodeError as exc:
        raise RuntimeError(f"invalid Gemini response: {exc}") from exc

    if "error" in payload:
        raise RuntimeError(json.dumps(payload["error"], ensure_ascii=True))

    candidates = payload.get("candidates")
    if not isinstance(candidates, list) or not candidates:
        raise RuntimeError(f"Gemini returned no candidates: {response_body}")

    parts = candidates[0].get("content", {}).get("parts", [])
    text_parts = [part.get("text", "") for part in parts if isinstance(part, dict)]
    summary = "".join(text_parts).strip()
    if not summary:
        finish_reason = candidates[0].get("finishReason", "unknown")
        raise RuntimeError(f"Gemini returned empty content, finishReason={finish_reason}")
    return summary


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
