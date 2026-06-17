# Backend API and Self-Hosted API Reference

Use this for 七巧服务端代码, custom functions, `REST.API`, `applyApi`, `executeServiceAPI`, or calls to the user's own deployed API server.

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
- Reports HTTP status and response body in the UI for debugging.

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
