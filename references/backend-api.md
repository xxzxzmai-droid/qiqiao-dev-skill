# Backend API and Self-Hosted API Reference

Use this for 七巧服务端代码, custom functions, `REST.API`, `applyApi`, `executeServiceAPI`, or calls to the user's own deployed API server. For OpenAPI token acquisition, 数智彩虹 intranet credentials, and UOS test tools, read `openapi-integration.md`.

Official docs:
- 七巧开发说明: https://qiqiao.do1.com.cn/help/develop_manual/%E4%B8%83%E5%B7%A7%E5%BC%80%E5%8F%91%E8%AF%B4%E6%98%8E.html
- 自定义页面使用教程: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2/%E4%BD%BF%E7%94%A8%E6%95%99%E7%A8%8B.html
- 页面 JS API: https://qiqiao.do1.com.cn/help/develop_manual/%E9%A1%B5%E9%9D%A2JS%E4%BA%8B%E4%BB%B6%E6%89%A9%E5%B1%95/JS-API.html

## Server code shape

Write Qiqiao server code in ES5-compatible JavaScript unless the target runtime has been verified.

```js
var FUNC = {
  log: function (message, data) {
    try {
      if (typeof $ !== "undefined" && $.log && $.log.info) {
        $.log.info(message + (data === undefined ? "" : " " + JSON.stringify(data)));
      }
    } catch (e) {}
  }
};

var API = {
  health: function () {
    FUNC.log("health called");
    return { ok: true, time: new Date().toISOString() };
  }
};
```

Rules:
- Expose only methods on `API`.
- Put helper methods in `FUNC` or another helper object.
- Return plain JSON-safe data. Avoid returning platform objects directly.
- Use `$.context` and Qiqiao function libraries only after confirming they exist in the current server-code surface.
- Do not hardcode user secrets, OpenAPI secrets, gateway tokens, or internal passwords in frontend code.

## Custom page frontend bridge

In runtime, Qiqiao's custom page tutorial calls:

```js
POST /dev-runtime/api/v1/runtime/business/{applicationId}/{businessId}/custompage/code/execute
Content-Type: application/json
X-Auth0-Token: <token when present>

{
  "code": "REST.API.methodName(arg1,arg2)",
  "methodName": "methodName",
  "applicationId": "...",
  "businessId": "..."
}
```

Use a wrapper that:
- Parses `applicationId` and `businessId` from URL path first, then query params.
- Treats `preview.html` as frontend-only and returns a mock/diagnostic result.
- Uses `parent.triggerSocket("REST.API.method(...)", "")` in Qiqiao debug iframe only when present.
- Sends the runtime POST only when IDs are available.
- Serializes args with `JSON.stringify`, never manual string concatenation.
- Reports HTTP status, content type, response body snippet, and tried execute URLs in the UI for debugging.

Do not hardcode one execute prefix. Deployments may mount runtime APIs under different path prefixes. Build candidates from the current page path and try the known variants:

```text
/dev-runtime/api/v1/runtime/business/{applicationId}/{businessId}/custompage/code/execute
/qiqiao/dev-runtime/api/v1/runtime/business/{applicationId}/{businessId}/custompage/code/execute
/runtime/api/v1/runtime/business/{applicationId}/{businessId}/custompage/code/execute
/qiqiao/runtime/api/v1/runtime/business/{applicationId}/{businessId}/custompage/code/execute
/runtime/api/v1/bpms-runtime/business/{applicationId}/{businessId}/custompage/code/execute
/qiqiao/runtime/api/v1/bpms-runtime/business/{applicationId}/{businessId}/custompage/code/execute
```

If the current page URL contains `/custompage/code/index.html`, put the same-path replacement first:

```text
location.pathname.replace("/custompage/code/index.html", "/custompage/code/execute")
```

This matters on deployments where the page is served from `/qiqiao/runtime/api/v1/bpms-runtime/business/.../custompage/code/index.html`; using only `/api/v1/runtime/business/...` will miss the real execute endpoint.

If `location.pathname` starts with another product prefix, also derive a prefixed candidate from that first segment.

Guard response parsing. Qiqiao gateways may return an HTML login, 404, reverse-proxy, or error page. This is the common cause of `Unexpected token '<', "<html> <h"... is not valid JSON`.

```js
function parseServerResponse(text, response) {
  var contentType = response && response.headers ? response.headers.get("content-type") || "" : "";
  var trimmed = (text || "").trim();
  if (!trimmed) return { ok: response.ok, empty: true };
  if (trimmed.charAt(0) === "<" || contentType.indexOf("text/html") >= 0) {
    return {
      ok: false,
      nonJson: true,
      message: "Qiqiao execute endpoint returned HTML/non-JSON. Check execute path, publish status, login state, and server-code deployment.",
      status: response.status,
      contentType: contentType,
      snippet: trimmed.slice(0, 500)
    };
  }
  try {
    return JSON.parse(trimmed);
  } catch (e) {
    return {
      ok: false,
      nonJson: true,
      message: e.message,
      status: response.status,
      contentType: contentType,
      snippet: trimmed.slice(0, 500)
    };
  }
}
```

For production CRUD custom pages, include a visible diagnostics action that runs `API.health()` and `API.diagnose()` and copies a JSON report containing environment, current page IDs, selected business inputs, execute URL candidates, recent frontend errors, recent backend errors, and backend probe results. Redact tokens, secrets, access keys, CorpID/CropID, and accounts.

## Fullstack form CRUD server methods

For protected form writes, prefer this method surface in `server-code.js`:

```js
var API = {
  health: function () { return { ok: true, time: new Date().toISOString() }; },
  diagnose: function () { /* token/schema/query/user probes only; no create */ },
  schema: function () { /* form model/component summary */ },
  queryReservations: function (params) { /* read form records */ },
  resolveUser: function () { /* runtime current user or configured-account fallback */ },
  createReservation: function (payload) { /* validate conflicts, then create */ },
  cancelReservation: function (payload) { /* status update, not destructive delete */ }
};
```

When `$.context` cannot resolve the current runtime user, a server-side fallback to a configured account is acceptable for smoke testing, but return a `source` such as `context` or `configured-account-fallback` so the page and diagnostic report make the identity path explicit.

## Self-hosted API strategy

When the user says future calls should hit their own deployed server API:

1. Prefer Qiqiao backend/server code or an approved gateway as the proxy when credentials, intranet access, or cross-origin restrictions are involved.
2. Use frontend direct calls only when the endpoint is public, CORS allows the Qiqiao runtime origin, and no secret is needed.
3. Keep API configuration in one place:

```js
var API_CONFIG = {
  baseUrl: "https://your-approved-api.example.com",
  timeoutMs: 15000
};
```

4. Never embed durable secrets in `index.js` or `index.html`.
5. Normalize errors from external APIs into `{ ok:false, message, detail }` so business users see a clear failure reason.

If the needed Qiqiao server-side HTTP client or Java bridge is unknown, inspect the current Qiqiao function library docs or ask the user for the deployed API access pattern before writing backend proxy code.
