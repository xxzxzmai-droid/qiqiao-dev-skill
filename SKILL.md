---
name: qiqiao-dev
description: Build, debug, test, and package Qiqiao/七巧/数智彩虹/道一云低代码 custom pages, injected index.html/index.css/index.js frontends, server-side custom function code, REST.API/applyApi bridges, page JS event extensions, custom form/page components, form/table/OpenAPI integrations, UOS/Linux OpenAPI test tools, and self-hosted API calls. Use when working on 七巧 IDE, 数智彩虹, 自定义页面, 服务端代码, 页面JS事件扩展, 表单表格高级 API, custompage code, OpenAPI token/form CRUD, preview/debug/runtime issues, or Qiqiao delivery packages.
---

# Qiqiao Dev

## Core Rules

- Treat Qiqiao custom pages as injection-mode pages, not ordinary static sites.
- Keep `index.html` as the DOM shell. Put styles in `index.css` and behavior in `index.js`; do not add `<link href="index.css">` or `<script src="index.js">` to the HTML.
- Avoid hashed Vite assets, dynamic chunks, `import.meta.url`, relative asset fetches, and local `fetch("index.js")` / `fetch("index.css")`; Qiqiao injects the three frontend files but may not expose them as fetchable URLs.
- Initialize JavaScript idempotently on `DOMContentLoaded` or immediately when the DOM is ready.
- Write server code as `var API = { method: function (...) { ... } }`; keep helpers outside API, return JSON-safe plain data, and log through Qiqiao logging APIs when available.
- Keep secrets out of frontend files. Prefer a server-side proxy or approved gateway for self-hosted server APIs unless the endpoint is public, CORS-ready, and needs no secret.
- For production custom pages that write protected form data, default to a fullstack pattern: frontend `index.js` calls Qiqiao server `API` methods, and `server-code.js` handles OpenAPI token, form CRUD, conflict checks, and diagnostics.
- Distinguish modes: preview renders frontend only, debug uses Qiqiao debugger/socket behavior, runtime calls the execute endpoint with application/business IDs and token when available.
- Treat 数智彩虹 as the user's intranet deployment of Qiqiao. Use the same Qiqiao rules, but expect intranet base URLs, private credentials, and proxy/NO_PROXY issues.
- Never paste `CorpID`/`Secret`/admin account/token values into final answers, frontend code, public skill files, screenshots, or logs. Use private config files, environment variables, or local-only `qqkf.txt` parsing.
- Do not claim OpenAPI can create or design form models unless the target deployment exposes and verifies a form-design management endpoint. The documented OpenAPI surface covers form data CRUD, form model/component lookup, files, workflow, users/departments, and workflow design definitions.

## Workflow

1. Classify the task:
   - Custom page / 三文件 IDE delivery: read `references/custom-page.md`.
   - Server API, `applyApi`, REST bridge, or self-hosted API: read `references/backend-api.md`.
   - 页面JS事件扩展, 表单, 表格, custom components, or OpenAPI: read `references/forms-tables.md`.
   - 数智彩虹/OpenAPI token, intranet credentials, form CRUD, UOS Go test executable, or CRUD test page: read `references/openapi-integration.md`.
   - Loading failures, self-check reports, preview/debug/runtime issues, or final handoff: read `references/debugging-delivery.md`.
2. Pick the nearest starting asset:
   - Use `assets/custom-page-injection/` for a known-good injected frontend plus `server-code.js` test surface.
   - Use `assets/openapi-crud-custom-page/` for a Qiqiao-deployed form CRUD test page that uses the runtime token and stores no durable secret.
   - Use `assets/openapi-go-tool/` for a local/UOS Linux executable that reads private config or `qqkf.txt`, probes OpenAPI, serves a local CRUD test web UI, and never embeds credentials.
3. Run `scripts/check_qiqiao_page.py <folder>` before delivery.
4. When local visual testing is needed, run `scripts/make_injection_harness.py <folder> <out.html>` and open the generated harness. Do not paste the harness into Qiqiao; paste the original three files.
5. For UOS/OpenAPI tools, run `go test ./...`, cross-compile with `GOOS=linux GOARCH=amd64`, and provide the executable path plus the exact command using a private config path.
6. Deliver `index.html`, `index.css`, `index.js`, optional `server-code.js`, generated executable paths, and concise paste/run instructions. Include a single-file fallback only when Qiqiao injection is unavailable or the user asks for it.

## Implementation Checklist

- No frontend file depends on relative auto-loaded files being fetchable.
- No generated bundle uses code splitting unless every chunk is inlined or hosted from an approved stable URL.
- All buttons and uploads are tested after Qiqiao-style injection, not only by opening `index.html` directly.
- For fullstack custom pages, run a local runtime simulator when feasible: serve an injected page under a Qiqiao-like `/bpms-runtime/business/.../custompage/code/index.html` path and expose a local `/custompage/code/execute` endpoint that invokes the same `server-code.js`.
- For production fullstack pages, prefer a comprehensive diagnostic gate over single-endpoint smoke tests: static checks, local execute bridge, backend `health/diagnose`, OpenAPI schema/query/write/cancel, conflict and invalid-input rejection, redaction, and explicit manual gates for real Qiqiao publish/login/`$.context`/intranet-only checks.
- Backend calls degrade clearly in preview mode and report a useful diagnostic instead of throwing raw platform errors.
- Frontend runtime bridge code must handle non-JSON/HTML responses from Qiqiao gateways. Do not blindly `JSON.parse(text)`; surface status, content-type, response snippet, and tried execute URL candidates in a copyable diagnostic report.
- Custom pages with backend methods should expose at least `health` and `diagnose`; CRUD tools should also expose schema/query/create/update or cancel methods with no token/secret returned to the frontend.
- If using page JS event extensions, exported function names are unique and the code uses `this.utils`, `this.business`, and `this.context` patterns instead of custom-page globals.
- If using OpenAPI or external APIs, confirm the token, app IDs, model IDs, CORS, and secret-handling path before implementation.
- Avoid OpenAPI token bursts in Qiqiao server code: combine startup calls into `API.bootstrap`, reuse token within one backend invocation, use `$.contact` for user/account lookup when available, and avoid extra frontend pre-submit queries when the backend create method already performs conflict validation.
- For strict-rate OpenAPI deployments, do not automatically re-query after every create/cancel/date switch. Use the backend response for optimistic UI, add refresh/diagnose cooldowns, and let users manually refresh when they need read-back proof.
- For formal user-facing releases, separate diagnostics from production UX: use diagnostic builds while debugging, then remove visible diagnostic buttons/panels and avoid exposing `API.diagnose` to ordinary users.
- Queue write operations in strict-rate deployments, but do not let read/bootstrap calls delay the first user write. Submit/cancel should run through one frontend queue, try the first write immediately, add only a short spacing between consecutive queued writes, and back off only after a real rate-limit response.
- For cancellation flows, keep ledger rows and mark status canceled. Show a local "cancel syncing" state while retrying so users do not see raw rate-limit errors and other users do not assume the slot is fully released before backend confirmation.
- For time-slot or resource-occupancy UIs, the first free slot should select the minimal valid range immediately; a later compatible slot can extend the range. Do not force two clicks for the smallest valid booking duration.
- For mobile business flows, avoid crowding selection, detail, input, confirmation, and history into one screen. Use explicit staged screens and keep the PC page on the same state machine.
- Do not let large helper, step, or diagnostic panels stay sticky over the primary mobile content. If an idle hint competes with the main view, hide it until the user selects something; after selection, use a compact action bar that preserves the main viewing and tapping area.
- For record edits, expose backend update methods only for fields whose changes are safe without changing availability or identity constraints. Changes to constraint-defining fields need explicit server-side revalidation or a cancel-and-create flow.
- Validate required business fields both before confirmation in the frontend and again in `server-code.js`; frontend-only validation is not enough for Qiqiao custom pages.
- For current-user detection in fullstack custom pages, do not rely only on `$.context.getCurrentUserId()` or placeholder names. Try `$.context.getCurrentUser()`, then `$.contact.getUserById(userId)`, and include a `userProbe` in diagnostics.
- For intranet OpenAPI calls to `10.*`, check proxy environment first. A 502 from `127.0.0.1` usually means the local proxy intercepted an intranet URL; retry with direct/no-proxy before changing API code.
- Respect Qiqiao OpenAPI rate limits; cache token results, avoid repeated destructive smoke tests, and prefer schema/probe reads before create/update/delete.
