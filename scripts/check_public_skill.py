#!/usr/bin/env python3
import json
import re
import sys
from pathlib import Path

import yaml


ROOT = Path(__file__).resolve().parents[1]


def read(path):
    return (ROOT / path).read_text(encoding="utf-8")


def add(results, status, code, detail):
    results.append({"status": status, "code": code, "detail": detail})


def frontmatter():
    text = read("SKILL.md")
    parts = text.split("---", 2)
    if len(parts) < 3:
        return {}
    return yaml.safe_load(parts[1]) or {}


def main():
    results = []
    meta = frontmatter()
    desc = meta.get("description", "")

    add(results, "pass" if desc.startswith("Use when ") else "fail", "description.use-when", "SKILL.md description starts with 'Use when '")
    add(results, "pass" if len(desc) <= 500 else "fail", "description.length", f"SKILL.md description is concise ({len(desc)} chars <= 500)")

    readme = read("README.md")
    add(
        results,
        "fail" if re.search(r"Default OpenAPI base URL is `https?://", readme) else "pass",
        "docs.no-fixed-default-base",
        "README does not present a deployment URL as the universal default",
    )

    main_go = read("assets/openapi-go-tool/main.go")
    add(
        results,
        "fail" if re.search(r'const\s+DefaultBaseURL\s*=\s*"https?://', main_go) else "pass",
        "tool.no-fixed-default-base",
        "Go tool does not silently default to the observed deployment URL",
    )

    openapi = read("references/openapi-integration.md")
    deployment_phrases = [
        "The current tested OpenAPI base is",
        "On the current 数智彩虹 test app",
        "Against the user's current test form",
        "The observed test form field types",
    ]
    present = [phrase for phrase in deployment_phrases if phrase in openapi]
    add(
        results,
        "fail" if present else "pass",
        "docs.observations-isolated",
        "Deployment-specific observations are isolated from the general OpenAPI reference",
    )

    for path in sorted((ROOT / "references").glob("*.md")):
        rel = path.relative_to(ROOT)
        lines = path.read_text(encoding="utf-8").splitlines()
        if len(lines) > 100:
            head = "\n".join(lines[:25])
            add(
                results,
                "pass" if "## Contents" in head else "fail",
                "docs.long-reference-contents",
                f"{rel} has a short contents section",
            )

    failures = [item for item in results if item["status"] == "fail"]
    report = {"ok": not failures, "failures": len(failures), "checks": results}
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 1 if failures else 0


if __name__ == "__main__":
    sys.exit(main())
