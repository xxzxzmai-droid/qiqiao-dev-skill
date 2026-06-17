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

## Handoff checklist

- Provide exact files or zip path.
- State whether it is three-file injection mode or single-file fallback.
- Include whether backend code is required.
- Include expected Qiqiao test path: preview for frontend, debug for server breakpoints/logs, publish for runtime user testing.
- Ask the user to paste the self-check JSON report if Qiqiao behavior differs from local simulation.
