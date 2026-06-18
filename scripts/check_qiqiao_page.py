#!/usr/bin/env python3
import argparse
import json
import re
from pathlib import Path


def read_text(path):
    return path.read_text(encoding="utf-8", errors="replace") if path.exists() else ""


def add(items, status, code, detail):
    items.append({"status": status, "code": code, "detail": detail})


SECRET_PATTERN = re.compile(
    r"("
    r"(?:corp(?:id)?|cropid|secret|password|passwd|accesskey|authorization|token)\s*[:=]\s*['\"][^'\"]{8,}['\"]"
    r"|bearer\s+[a-z0-9._-]{16,}"
    r"|x-auth0-token\s*[:=]\s*['\"][^'\"]{16,}['\"]"
    r")",
    re.I,
)


def has_secret_like(text):
    return bool(SECRET_PATTERN.search(text or ""))


def main():
    parser = argparse.ArgumentParser(description="Validate a Qiqiao injected custom page folder.")
    parser.add_argument("folder", help="Folder containing index.html, index.css, index.js, and optional server-code.js")
    args = parser.parse_args()

    root = Path(args.folder).expanduser().resolve()
    html_path = root / "index.html"
    css_path = root / "index.css"
    js_path = root / "index.js"
    server_path = root / "server-code.js"

    results = []
    for path in (html_path, css_path, js_path):
        add(results, "pass" if path.exists() else "fail", "file.exists", str(path.name))

    html = read_text(html_path)
    css = read_text(css_path)
    js = read_text(js_path)
    server = read_text(server_path)
    frontend = "\n".join([html, css, js])

    if html:
        add(results, "pass" if re.search(r"<!doctype html>|<html[\s>]", html, re.I) else "warn", "html.shell", "HTML shell detected")
        linked_css = re.search(r"<link[^>]+href=['\"][^'\"]*index\.css", html, re.I)
        linked_js = re.search(r"<script[^>]+src=['\"][^'\"]*index\.js", html, re.I)
        add(results, "fail" if linked_css else "pass", "html.no-index-css-link", "Do not link index.css manually in Qiqiao injection mode")
        add(results, "fail" if linked_js else "pass", "html.no-index-js-script", "Do not script-load index.js manually in Qiqiao injection mode")
        vite_assets = re.search(r"(?:src|href)=['\"]/?assets/[^'\"]+", html, re.I)
        add(results, "fail" if vite_assets else "pass", "html.no-vite-assets", "No Vite hashed asset references in HTML")

    if js:
        dom_ready = "DOMContentLoaded" in js or "document.readyState" in js
        add(results, "pass" if dom_ready else "warn", "js.dom-ready", "Initialize after DOM is ready")
        module_syntax = re.search(r"^\s*import\s+|^\s*export\s+", js, re.M) or "import.meta" in js
        add(results, "warn" if module_syntax else "pass", "js.no-module-assumption", "Avoid ESM-only code unless Qiqiao target is verified")
        dynamic_import = "import(" in js
        add(results, "warn" if dynamic_import else "pass", "js.no-dynamic-import", "Avoid dynamic chunks in injected index.js")
        fetch_own_files = re.search(r"fetch\s*\(\s*['\"][^'\"]*(index\.js|index\.css)", js)
        add(results, "fail" if fetch_own_files else "pass", "js.no-self-fetch", "Do not fetch injected frontend files by URL")
        bridge_markers = ("execute" in js) or ("REST.API" in js) or ("X-Auth0-Token" in js) or ("/open/" in js)
        preview_guard = "preview" in js.lower() and bridge_markers
        add(results, "pass" if preview_guard else "warn", "js.preview-guard", "Backend bridge should treat preview as frontend-only")

    if css:
        add(results, "pass", "css.present", "index.css is present")

    if server_path.exists():
        add(results, "pass" if re.search(r"\bvar\s+API\s*=", server) else "fail", "server.var-api", "Server code should expose var API = {...}")
        modern = re.search(r"^\s*(import|export)\s+", server, re.M) or re.search(r"\b(const|let)\s+", server)
        add(results, "warn" if modern else "pass", "server.es5", "Prefer ES5-compatible server code unless runtime is verified")
    else:
        add(results, "warn", "server.optional", "server-code.js not found; OK for frontend-only pages")

    frontend_secret = has_secret_like(frontend)
    add(
        results,
        "fail" if frontend_secret else "pass",
        "frontend.secrets.scan",
        "No obvious hardcoded durable secret patterns in index.html/index.css/index.js",
    )
    if server_path.exists():
        server_secret = has_secret_like(server)
        add(
            results,
            "warn" if server_secret else "pass",
            "server.secrets.scan",
            "Server code may contain private config; keep it out of public templates, logs, screenshots, and GitHub examples",
        )

    fail_count = sum(1 for item in results if item["status"] == "fail")
    warn_count = sum(1 for item in results if item["status"] == "warn")
    report = {
        "folder": str(root),
        "ok": fail_count == 0,
        "failures": fail_count,
        "warnings": warn_count,
        "checks": results,
    }
    print(json.dumps(report, ensure_ascii=False, indent=2))
    return 1 if fail_count else 0


if __name__ == "__main__":
    raise SystemExit(main())
