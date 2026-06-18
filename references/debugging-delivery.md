# Debugging and Delivery Reference

Use this for Qiqiao loading failures, self-check reports, packaging, and handoff.

## Known Qiqiao loading behavior

Observed and expected from the docs:
- `index.css` and `index.js` can be injected and run even when `fetch("index.css")` or `fetch("index.js")` fails.
- A self-check report with `css: pass`, `js: pass`, `auto-inject: pass`, and `css-fetch/js-fetch: fail` means injection works but static file URLs are not directly fetchable. Do not make app logic fetch those files.
- Preview mode may show backend as mocked or unavailable. That is expected because preview is frontend-only.
- Debug mode may call `parent.triggerSocket(...)`; the result is usually inspected in Qiqiao's debugger/log panel.
- Runtime mode should call the execute endpoint only when IDs and the runtime path are available.

## Failure triage

1. Plain HTML only, no styles or buttons:
   - Check whether `index.css` and `index.js` were pasted into the correct Qiqiao file panes.
   - Confirm the HTML has no malformed text before `<!doctype html>`.
   - Run the self-check template from `assets/custom-page-injection/`.

2. Large minified JS appears as visible text:
   - A script was pasted outside a `<script>` tag in a single-file HTML, or injected into the wrong file pane.
   - For Qiqiao three-file mode, put minified JS only in `index.js`; do not paste it after the HTML body.
   - For single-file mode, inline scripts must be inside `<script>...</script>` and any literal `</script>` inside strings must be escaped.

3. `xlsx.full.min.js 未加载` or similar:
   - Do not rely on external relative scripts.
   - Inline the library, use an approved material/CDN URL, or bundle it into the no-chunk `index.js`.

4. Backend call fails in preview:
   - Treat as expected unless testing debug/runtime.
   - The UI should display "preview frontend-only" rather than a raw network error.

5. Backend call fails in runtime:
   - Check `applicationId`, `businessId`, token, execute URL, published version, and server-code syntax.
   - Add a small `API.health()` first, then test business methods.

6. Browser shows `Unexpected token '<', "<html> <h"... is not valid JSON`:
   - The frontend parsed an HTML response as JSON. This usually means the execute endpoint path is wrong, the request hit a login/gateway/404 page, the server code is not published, or a reverse proxy changed the response.
   - Do not debug this from the visual page alone. Add guarded parsing and show/copy a diagnostic report with `status`, `content-type`, a short response snippet, `applicationId`, `businessId`, token presence, and every execute URL candidate tried.
   - Try the known execute prefixes: `/dev-runtime`, `/qiqiao/dev-runtime`, `/runtime`, `/qiqiao/runtime`, plus a candidate derived from the first segment of `location.pathname`.
   - If the page URL is under `/bpms-runtime/business/.../custompage/code/index.html`, try the same path with `index.html` replaced by `execute` before the generic `/runtime/business/...` candidates.
   - Confirm the page parses `applicationId` and `businessId` from the runtime URL before falling back to configured constants. If constants are read first, path extraction may never be used.

7. Page cannot read the current user:
   - First verify which runtime user APIs exist in `$.context` or the page context for the current deployment.
   - In Qiqiao server code, test `$.context.getCurrentUser()`, `$.context.getCurrentUserId()`, and `$.contact.getUserById(userId)` before falling back to a configured account.
   - If the user API is unavailable, let backend diagnostics report the failure and optionally resolve the configured backend account for testing.
   - Surface a `userProbe` in the diagnostic report with booleans for context user ID, context user object, contact-by-ID, configured-account fallback, and the final source.
   - Surface the user source in the UI/diagnostic report, for example `context-user`, `context-contact-by-id`, `url`, `configured-account-fallback`, or `unresolved`.

8. Report times look eight hours behind Beijing time:
   - A timestamp ending in `Z` is UTC. Do not treat `2026-06-18T05:37:00Z` as proof that the backend clock is wrong; it is `2026-06-18 13:37:00` in Beijing.
   - Format user-facing diagnostics and server health times as `Asia/Shanghai +08:00` for Chinese business users.
   - Keep epoch millisecond timestamps unchanged for OpenAPI token signing, IDs, and date field values.

9. Long backend errors stretch or break the mobile layout:
   - Do not render raw gateway URLs, HTML snippets, stack traces, or JSON blobs directly inside narrow cards.
   - Show a short business-readable message in the page, cap the message box height, and put full details in the copyable diagnostic report.
   - Add CSS wrapping constraints such as `min-width: 0`, `max-width: 100%`, `overflow-wrap: anywhere`, and a scrollable max-height for diagnostic/error panels.

10. Debug build behavior leaks into the formal release:
   - Keep a diagnostic build while the page is unstable, but produce a separate formal build for ordinary users.
   - Formal builds should remove visible diagnostic buttons/panels and should not expose destructive or high-frequency diagnostic actions to end users.
   - Keep guarded non-JSON parsing and friendly user messages in the formal build; remove only the visible diagnostic UI.

## Validation

Run:

```bash
python3 scripts/check_qiqiao_page.py /path/to/qiqiao-folder
```

For local simulation:

```bash
python3 scripts/make_injection_harness.py /path/to/qiqiao-folder /tmp/qiqiao-harness.html
```

Open the harness to verify basic rendering and event binding. Do not upload the harness as the Qiqiao three-file version.

For fullstack pages with `server-code.js`, prefer a local runtime simulator over a static harness. The simulator should serve a Qiqiao-like `/bpms-runtime/business/.../custompage/code/index.html` URL and a local `/custompage/code/execute` endpoint. Passing this local runtime smoke test means the app code, execute bridge, server API methods, and OpenAPI CRUD can work together; it does not prove Qiqiao's real gateway path, login token, published-version routing, or `$.context` user data.

For production-grade fullstack pages, add a comprehensive diagnostic gate instead of only testing the latest failing method. The gate should produce a JSON report and a short Markdown summary, and should include:

- Static checks: three-file injection contract, no manual `index.css` / `index.js` loading, no Vite chunk assumptions, frontend secret scan, server syntax, and core frontend logic tests.
- Execute checks: local Qiqiao-like runtime page URL, local `/custompage/code/execute`, `API.health()`, and guarded handling for non-JSON/HTML responses.
- Backend checks: `API.diagnose()`, `$.httpclient` method inventory, active OpenAPI base URL, candidate base URLs, and per-base failures with secrets redacted.
- OpenAPI checks: token, schema/component lookup, required field mapping, query, current-user or configured-account fallback, and exact value-shape validation for dates/times/options/users.
- Write-path checks when the user allows form pollution: create a test record, query it back, assert duplicate/conflict rejection, assert invalid inputs are rejected, then cancel/update the record and query the canceled ledger row.
- Manual-runtime gates: explicitly mark real Qiqiao published gateway routing, browser login state, true `$.context` current user, and intranet-only base URLs as requiring a deployed runtime diagnostic report.

Do not describe a local comprehensive diagnostic as proof that Qiqiao production is fully correct. State exactly which layers passed locally and which layers still require the page's deployed `诊断` report.

## Handoff checklist

- Provide exact files or zip path.
- State whether it is three-file injection mode or single-file fallback.
- Include whether backend code is required.
- Include whether local runtime simulation passed, and separately whether Qiqiao runtime diagnostics passed.
- Include whether the frontend has a copyable diagnostics panel and whether `API.health()` / `API.diagnose()` passed.
- For formal releases, state explicitly when diagnostic UI has been removed and whether a separate diagnostic build remains available for troubleshooting.
- Include expected Qiqiao test path: preview for frontend, debug for server breakpoints/logs, publish for runtime user testing.
- Ask the user to paste the self-check JSON report if Qiqiao behavior differs from local simulation.
