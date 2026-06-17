---
name: qiqiao-dev
description: Build, debug, and package Qiqiao/七巧/道一云低代码 custom pages, injected index.html/index.css/index.js frontends, server-side custom function code, REST.API/applyApi bridges, page JS event extensions, custom form/page components, form/table/OpenAPI integrations, and self-hosted API calls. Use when working on 七巧 IDE, 自定义页面, 服务端代码, 页面JS事件扩展, 表单表格高级 API, custompage code, preview/debug/runtime issues, or Qiqiao delivery packages.
---

# Qiqiao Dev

## Core Rules

- Treat Qiqiao custom pages as injection-mode pages, not ordinary static sites.
- Keep `index.html` as the DOM shell. Put styles in `index.css` and behavior in `index.js`; do not add `<link href="index.css">` or `<script src="index.js">` to the HTML.
- Avoid hashed Vite assets, dynamic chunks, `import.meta.url`, relative asset fetches, and local `fetch("index.js")` / `fetch("index.css")`; Qiqiao injects the three frontend files but may not expose them as fetchable URLs.
- Initialize JavaScript idempotently on `DOMContentLoaded` or immediately when the DOM is ready.
- Write server code as `var API = { method: function (...) { ... } }`; keep helpers outside API, return JSON-safe plain data, and log through Qiqiao logging APIs when available.
- Keep secrets out of frontend files. Prefer a server-side proxy or approved gateway for self-hosted server APIs unless the endpoint is public, CORS-ready, and needs no secret.
- Distinguish modes: preview renders frontend only, debug uses Qiqiao debugger/socket behavior, runtime calls the execute endpoint with application/business IDs and token when available.

## Workflow

1. Classify the task:
   - Custom page / 三文件 IDE delivery: read `references/custom-page.md`.
   - Server API, `applyApi`, REST bridge, or self-hosted API: read `references/backend-api.md`.
   - 页面JS事件扩展, 表单, 表格, custom components, or OpenAPI: read `references/forms-tables.md`.
   - Loading failures, self-check reports, preview/debug/runtime issues, or final handoff: read `references/debugging-delivery.md`.
2. Start from `assets/custom-page-injection/` for a known-good injected frontend plus `server-code.js` test surface.
3. Run `scripts/check_qiqiao_page.py <folder>` before delivery.
4. When local visual testing is needed, run `scripts/make_injection_harness.py <folder> <out.html>` and open the generated harness. Do not paste the harness into Qiqiao; paste the original three files.
5. Deliver `index.html`, `index.css`, `index.js`, optional `server-code.js`, and concise paste/upload instructions. Include a single-file fallback only when Qiqiao injection is unavailable or the user asks for it.

## Implementation Checklist

- No frontend file depends on relative auto-loaded files being fetchable.
- No generated bundle uses code splitting unless every chunk is inlined or hosted from an approved stable URL.
- All buttons and uploads are tested after Qiqiao-style injection, not only by opening `index.html` directly.
- Backend calls degrade clearly in preview mode and report a useful diagnostic instead of throwing raw platform errors.
- If using page JS event extensions, exported function names are unique and the code uses `this.utils`, `this.business`, and `this.context` patterns instead of custom-page globals.
- If using OpenAPI or external APIs, confirm the token, app IDs, model IDs, CORS, and secret-handling path before implementation.
