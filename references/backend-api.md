# Backend API and Self-Hosted API Reference

Use this for 七巧服务端代码, custom functions, `REST.API`, `applyApi`, `executeServiceAPI`, or calls to the user's own deployed API server. For OpenAPI token acquisition, 数智彩虹 intranet credentials, and UOS test tools, read `openapi-integration.md`.

Official docs:
- 七巧开发说明: https://qiqiao.do1.com.cn/help/develop_manual/%E4%B8%83%E5%B7%A7%E5%BC%80%E5%8F%91%E8%AF%B4%E6%98%8E.html
- 自定义页面使用教程: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2/%E4%BD%BF%E7%94%A8%E6%95%99%E7%A8%8B.html
- 页面 JS API: https://qiqiao.do1.com.cn/help/develop_manual/%E9%A1%B5%E9%9D%A2JS%E4%BA%8B%E4%BB%B6%E6%89%A9%E5%B1%95/JS-API.html
- 低代码函数库: https://qiqiao.do1.com.cn/help/develop_manual/%E4%BD%8E%E4%BB%A3%E7%A0%81%E5%87%BD%E6%95%B0%E5%BA%93.html

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

## Local runtime simulator

For fullstack pages, static local preview is not enough. A useful local simulator should:

- Serve the page under a Qiqiao-like path such as `/qiqiao/runtime/api/v1/bpms-runtime/business/local-app/local-business/custompage/code/index.html`.
- Inject `index.css` and `index.js` for local browser testing, because Qiqiao injects those files but a plain browser does not.
- Expose `POST .../custompage/code/execute` and dispatch `REST.API.methodName(...)` to the same `server-code.js`.
- Mock the Qiqiao `$` object minimally: `$.log`, `$.context`, and `$.httpclient`.
- Let `$.httpclient` call the real OpenAPI when the task is to verify form CRUD, while keeping secrets in local/server-side code only.
- Produce a smoke report for `health`, `diagnose`, `schema`, query, user resolution, create, and cancel/update paths.

Do not force every `localhost` page into preview mode. If the path matches `/bpms-runtime/business/.../custompage/code/`, treat it as a simulated runtime path so the frontend calls the local execute endpoint.

Limits of local simulation:
- It can prove frontend bridge code, server-code method dispatch, OpenAPI token/schema/query/create/update behavior, and conflict checks.
- It cannot prove the real Qiqiao gateway prefix, published-version routing, browser login state, `X-Auth0-Token`, or real `$.context` current-user behavior. Those still need runtime deployment diagnostics.

## Server `$.httpclient` adapter

Do not assume Qiqiao server-side HTTP client methods are lowercase `get`, `post`, and `put`. Some deployments expose uppercase names, Java-backed methods, or a generic `request`/`execute`/`send` method. Backend proxy code should:

- Prefer the documented Java-backed signatures when `send*` methods exist:
  - `$.httpclient.sendGet(uri, params, headers)`
  - `$.httpclient.sendPost(uri, params, headers, body)`
  - `$.httpclient.sendPut(uri, params, headers, body)`
  - `$.httpclient.sendDel(uri, params, headers)`
- For those documented `send*` methods, pass `params` and `headers` as `java.util.HashMap` in Qiqiao/Rhino, not ordinary JS objects and not `undefined`. A Rhino error like `Can't find method ... sendGet(org.mozilla.javascript.ConsString,object,org.mozilla.javascript.Undefined)` means the method exists but the argument types are wrong.
- For GET token/API calls, pass a plain `uri` plus a params map. Do not rely only on a URL with query string when the active method is `sendGet(uri, params, headers)`.
- Try lowercase, uppercase, and title-case method names for `GET`, `POST`, and `PUT`.
- Try generic calls such as `request({ url, method, headers, body })`, `execute(...)`, `send(...)`, or `fetch(...)` when method-specific functions are absent.
- Include `httpClientMethods` in `API.health()` / `API.diagnose()` using `Object.keys`, `for...in`, known method probes, and Java `getClass().getMethods()` when available.
- Return a clear diagnostic such as `$.httpclient does not support GET; available methods: ...` instead of only saying the backend is unavailable.

ES5 helper pattern:

```js
function toJavaMap(value) {
  var map;
  var key;
  value = value || {};
  try {
    if (typeof Packages !== "undefined" && Packages.java && Packages.java.util && Packages.java.util.HashMap) {
      map = new Packages.java.util.HashMap();
      for (key in value) {
        if (value.hasOwnProperty(key) && value[key] !== undefined && value[key] !== null) {
          map.put(String(key), String(value[key]));
        }
      }
      return map;
    }
  } catch (e) {}
  return value;
}

// Qiqiao documented shape:
var params = toJavaMap({ pageSize: 10 });
var headers = toJavaMap({ Accept: "application/json", "X-Auth0-Token": token });
var raw = $.httpclient.sendGet(baseUrl + "/open/example", params, headers);
```

## Fullstack form CRUD server methods

For protected form writes, prefer this method surface in `server-code.js`:

```js
var API = {
  health: function () { return { ok: true, time: new Date().toISOString() }; },
  bootstrap: function () { /* current user + schema fields + first query in one call */ },
  diagnose: function () { /* token/schema/query/user probes only; no create */ },
  schema: function () { /* form model/component summary */ },
  queryReservations: function (params) { /* read form records */ },
  resolveUser: function () { /* runtime current user or configured-account fallback */ },
  createReservation: function (payload) { /* validate conflicts, then create */ },
  cancelReservation: function (payload) { /* status update, not destructive delete */ }
};
```

For the current runtime user, follow Qiqiao's documented custom-page pattern:

```js
var user = $.context.getCurrentUser();
var userId = $.context.getCurrentUserId();
var userById = $.contact.getUserById(userId);
```

Prefer `$.context.getCurrentUser()` when it returns a usable `UserDTO`. If it only returns an ID or no name, use `$.contact.getUserById(userId)` to resolve the display name/account. Do not stop at a placeholder such as `"当前用户"` when a real user ID is present.

When `$.context` cannot resolve the current runtime user, a server-side fallback to a configured account is acceptable for smoke testing, but return a `source` such as `context-user`, `context-contact-by-id`, or `configured-account-fallback` so the page and diagnostic report make the identity path explicit.

For user lookup, prefer Qiqiao's contact library before OpenAPI when available:

```js
var user = $.contact.getUserByUserAccount(account);
```

This avoids spending OpenAPI token requests on reserver/attendee parsing. If contact lookup is unavailable, fall back to `/open/users/account` and make the diagnostic source explicit.

For pages that load data immediately, expose a `bootstrap` method and have the frontend call it once on init. Returning user, field mapping, and first-page records together prevents a startup burst of `currentUser` + `schema` + `queryReservations` execute calls, which can trigger `access_key` frequency limits on deployments with strict token throttling.

When `queryReservations`, `createReservation`, or `cancelReservation` are called after bootstrap, let the frontend pass the effective field map back to the backend so those methods do not have to fetch the form model/schema every time. On create/cancel success, return enough record data for the frontend to update the local schedule immediately; avoid an automatic post-submit refresh on deployments that rate-limit `access_key`.

Do not split one business submit into multiple frontend server calls for each person field. For reservation-style pages, let `createReservation(payload)` resolve the reserver and write in one backend invocation. Free-text fields such as attendee names, departments, or phone numbers should be written as text and must not be resolved through `$.contact` or `/open/users/account`.

For strict-rate deployments, put create/cancel behind a frontend queue rather than letting repeated clicks call `REST.API.*` concurrently. Keep the queue fast-first: the first user write should run immediately even if bootstrap/read calls just happened; only consecutive queued writes need short spacing, and only real rate-limit failures should trigger delayed retries. Cancellation should show a pending-sync state until the backend status update succeeds.

For reservation grids with half-hour slots, selecting one free slot should produce a valid half-hour booking immediately. Let users extend the booking by clicking a later free slot, and block later candidates that would cross an occupied slot.

For user-facing diagnostics, format timestamps in the business timezone, usually `Asia/Shanghai +08:00`, instead of raw UTC `toISOString()` values. Keep epoch millisecond timestamps for token signing and IDs unchanged.

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
