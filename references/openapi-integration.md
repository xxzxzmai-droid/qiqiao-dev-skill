# OpenAPI, 数智彩虹, and UOS Tooling Reference

Use this for 数智彩虹 intranet OpenAPI work, token acquisition, form data CRUD, private `qqkf.txt` credentials, Linux/UOS test executables, and Qiqiao-deployed CRUD test pages.

Official docs:
- OpenAPI: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0/OpenAPI.html
- 开放平台: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0.html
- 自定义页面使用教程: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2/%E4%BD%BF%E7%94%A8%E6%95%99%E7%A8%8B.html

## Deployment vocabulary

- 数智彩虹 is the user's internal deployment of Qiqiao/七巧. Treat both names as the same platform unless the user says otherwise.
- The current tested OpenAPI base is:

```text
https://e.csg.cn/qiqiao/runtime/api/v1/bpms-integration
```

- Internal deployments may also expose `http://<intranet-host>/qiqiao/runtime/api/v1/bpms-integration`, but do not use `10.*` as the default unless the user asks for that path.
- On local macOS test environments, the public HTTPS path may require `--insecure-skip-verify` because the local CA chain can fail verification.
- Keep private `CorpID`/`CropID`, `Secret`, admin account, access key, and token in local files, environment variables, or server-side config only.

When the user supplies both internal and external OpenAPI bases, keep both in server-side configuration and auto-select:

- Local/offsite testing should prefer the public HTTPS base first.
- Qiqiao intranet deployment can prefer the internal HTTP base first.
- If token acquisition fails on one base, try the next base automatically before surfacing an error.
- Diagnostics should include the active base name, active base URL, candidate base URLs, and per-base failure messages, but never tokens or secrets.

## Token flow

Use the same `timestamp` and `random` for both calls:

```text
GET /securities/access_key?timestamp=...&random=...&corpId=...&secret=...&account=...
GET /securities/qiqiao_token?timestamp=...&random=...&corpId=...&secret=...&account=...&accessKey=...
```

Then call OpenAPI with:

```text
Content-Type: application/json
X-Auth0-Token: <qiqiaoToken>
```

When this flow runs inside Qiqiao server-side custom code through `$.httpclient.sendGet`, call it as `sendGet(uri, paramsMap, headersMap)` with Java `HashMap` params/headers. Do not pass ordinary JS objects or `undefined`; that fails in Rhino with a Java overload error even though `sendGet` is listed as available.

Rules:
- Cache the token during a test run; do not request it for every button click.
- Do not print token values. Redact as first 6 + `***` + last 4 if a diagnostic needs proof.
- Respect the documented low-frequency limit. Prefer one `probe` run, then targeted calls.
- Do not intentionally probe the exact production rate-limit threshold by "calling until it fails". That burns the same limit window real users need. Use conservative queue spacing and retry behavior instead.
- When calling the public `https://e.csg.cn/...` base from local scripts, use the reference script's browser-like `User-Agent` and `Accept` headers. A bare HTTP client may receive a `403` HTML gateway response even when credentials are correct.

If a deployed report first shows token/schema/query success and later fails with `access_key 失败：超出请求频率限制`, treat the OpenAPI path as connected but over-called. Reduce calls before changing credentials or base URLs:

- Combine first-page initialization into one backend method such as `API.bootstrap()` that returns current user, schema fields, and the first reservation query in one execute call.
- Avoid separate `schema` then `queryReservations` calls on every refresh; have `queryReservations` return the effective field map when needed.
- After bootstrap, pass the effective field map into later CRUD backend calls so they do not re-fetch `/form_models/{id}` on every query/create/cancel.
- Do not call `/open/users/account` for every reserver or attendee if Qiqiao server-side contact functions are available. Prefer `$.contact.getUserByUserAccount(account)` and fall back to OpenAPI only when contact lookup is unavailable.
- On submit, let backend `createReservation` perform the final conflict query and create in one method. Frontend should use its loaded records for immediate UX feedback, not as an extra pre-submit OpenAPI call.
- After create/cancel succeeds, update the visible schedule from the returned backend record first. Do not immediately call `queryReservations` again unless the user manually asks for read-back verification.
- Add cooldowns to manual refresh and diagnostic actions. A useful starting point is 15 seconds for refresh and 30 seconds for diagnostics, with a clear UI message instead of repeated backend calls.
- In formal pages, route create/cancel through a single frontend write queue. Persist the last backend-call timestamp in local storage so reloads do not immediately burst token calls again. For strict deployments, start with at least 60 seconds between write calls and retry rate-limit failures with a longer delay.
- If cancellation hits a rate limit, keep the record visually marked as "cancel syncing" and retry in the background. Do not claim the slot is fully released until the backend status update succeeds.
- After a rate-limit report, wait for the platform window to cool down before rerunning destructive or repeated diagnostics.

## Form data OpenAPI

Confirmed documented form endpoints include:

```text
GET    /open/applications/{applicationId}/form_models
GET    /open/applications/{applicationId}/form_models/{formModelId}
GET    /open/applications/{applicationId}/forms/{formModelId}/{id}
GET    /open/applications/{applicationId}/forms/{formModelId}
POST   /open/applications/{applicationId}/forms/{formModelId}
PUT    /open/applications/{applicationId}/forms/{formModelId}
DELETE /open/applications/{applicationId}/forms/{formModelId}/{id}
POST   /open/applications/{applicationId}/forms/{formModelId}/query
POST   /open/applications/{applicationId}/forms/{formModelId}/batch_save
POST   /open/applications/{applicationId}/forms/{formModelId}/batch_update
POST   /open/applications/{applicationId}/forms/{formModelId}/batch_delete
```

The user's captured API directory (`qqkf.txt`) also lists:

```text
PUT    /open/applications/{applicationId}/forms/batch_update_by_condition
```

Treat this as a conditional batch-update probe, not a confirmed stable endpoint, until the target deployment accepts the payload shape. On the current 数智彩虹 test app, the documented path returned `5500031 表单不存在` when passed `formModelId` in the body; common path/query variants returned `系统繁忙，请稍后再试`.

Single-record create payload shape:

```json
{
  "variables": {
    "字段名": "字段值"
  },
  "id": "32-char-record-id",
  "loginUserId": "user-id"
}
```

Update payload shape:

```json
{
  "variables": {
    "字段名": "新值"
  },
  "id": "record-id",
  "version": 1,
  "loginUserId": "user-id"
}
```

For the single-record `PUT /forms/{formModelId}` path, pass the current version returned by the create/query/get response. Do not increment it locally. For the tested `batch_update` endpoint, pass `version + 1`.

Use exact form component names and value shapes from `GET /form_models/{formModelId}` before writing filters or payloads. Dates commonly need timestamps; option fields may need stored option values or indexes, not display text.

If a field changes type, update the write payload shape and tests immediately. For example, when `参会人` changes from `multiUserSelect` to `textarea`, write a string such as `"张三\n李四"` rather than a user ID array, and stop resolving each attendee through contact/OpenAPI. Keep only the reserver/approver fields as user IDs when the form field type requires it.

For reservation-style tools, prefer a status field over destructive cancellation when business audit history matters. The tested meeting-room pattern writes:

```json
{
  "会议室字段": "会商室",
  "预约日期字段": 1781625600000,
  "预约开始时间字段": "10:30",
  "预约结束时间字段": "11:00",
  "预约状态字段": "1",
  "预约用途字段": "用途",
  "预约人字段": "user-id"
}
```

Then cancel by `PUT /open/applications/{applicationId}/forms/{formModelId}` with the current record `version` and only the status variable set to the canceled option value, such as `"2"`. Treat canceled records as ledger/history rows and exclude them from availability conflict checks.

## Capability boundaries

- The documented OpenAPI list includes form data CRUD and form model/component lookup, not a confirmed public endpoint for creating form models or designing form fields.
- `/open/console/applications/{applicationId}/definitions` is documented as workflow design definitions. Do not describe it as form design.
- If the user asks to create/design forms, first run a capability probe or inspect additional internal management API docs. If no verified endpoint exists, say that form data CRUD is available but form-model design creation is not confirmed.

## Intranet proxy diagnosis

When a `10.*` base URL returns `502` and the remote IP is `127.0.0.1`, the local proxy probably intercepted an intranet URL. Prefer the public `https://e.csg.cn/...` base when the user provides it. If the user specifically needs the intranet path, test direct access before changing code:

```bash
curl --noproxy '*' -m 8 http://<intranet-qiqiao-host>/qiqiao/runtime/api/v1/bpms-integration/
```

For Go tools, use the public base by default. Enable `--use-proxy` only when the target path requires the system proxy; otherwise keep direct transport for intranet addresses.

## UOS Go executable asset

Use `assets/openapi-go-tool/` when the user needs an internal Linux/UOS test binary.

Validation and build:

```bash
cd assets/openapi-go-tool
go test ./...
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o qiqiao-openapi-tool-uos-amd64 .
```

Typical intranet run:

```bash
./qiqiao-openapi-tool-uos-amd64 --qqkf /path/to/qqkf.txt --mode probe
./qiqiao-openapi-tool-uos-amd64 --qqkf /path/to/qqkf.txt --insecure-skip-verify --mode serve --listen 0.0.0.0:8787
```

Modes:
- `probe`: token, current user, form models, form detail, and first-page query summary.
- `schema`: raw form model and component JSON for field discovery.
- `serve`: local web UI for query/create/update/delete through the tool's server-side OpenAPI client.
- `crud-smoke`: create/update/delete a supplied variables JSON payload. Use only on a test form.
- `full-smoke`: broader existing-form suite with auth, users, departments, applications, schema, GET/POST query, single create/get/filter/update, conditional batch-update probe, batch save/update/delete cleanup, optional file upload, workflow definition listing, and form-design boundary reporting.
- `design-probe`: read workflow definition availability without creating anything.

Run a full existing-form test:

```bash
./qiqiao-openapi-tool-uos-amd64 \
  --config /path/to/qiqiao-config.local.json \
  --mode full-smoke \
  --page-size 5 \
  --upload-file /path/to/small-test-file.txt \
  --delete-after=true
```

Private config shape:

```json
{
  "baseUrl": "https://e.csg.cn/qiqiao/runtime/api/v1/bpms-integration",
  "corpId": "...",
  "secret": "...",
  "account": "...",
  "applicationId": "...",
  "formModelId": "...",
  "insecureSkipVerify": true,
  "timeoutSeconds": 30
}
```

If macOS reports `operation not permitted` when reading `qqkf.txt` on `/Volumes`, do not change API code. Copy the private credentials into an accessible local config file outside the repo, set mode `0600`, and pass `--config`.

## Qiqiao-deployed CRUD page asset

Use `assets/openapi-crud-custom-page/` for a custom page deployed inside Qiqiao.

Properties:
- Uses Qiqiao runtime `X-Auth0-Token` from URL query; no durable secret is stored in frontend files.
- Calls same-origin OpenAPI under `/qiqiao/runtime/api/v1/bpms-integration`.
- In preview mode, returns mock diagnostics rather than making OpenAPI calls.
- Requires `applicationId`, `formModelId`, and `loginUserId` inputs unless the target page supplies them through URL parameters or prefilled constants.

Run `scripts/check_qiqiao_page.py assets/openapi-crud-custom-page` before delivery.

## Production form CRUD through server-code

When the page must write real form data and needs CorpID/Secret/account credentials, do not put those values in `index.html`, `index.css`, or `index.js`. Put private credentials only in Qiqiao server-side code/config or a private local config used by test tools. The frontend should call server `API` methods through the custom page execute bridge.

Recommended flow:
- Frontend loads current room/date/state and asks `API.queryReservations(params)` for records.
- Frontend asks `API.resolveUser()` to prefill the reserver. If runtime user lookup fails, the backend may use a configured-account fallback only when the diagnostic response marks that source clearly.
- Frontend posts a normalized payload to `API.createReservation(payload)`.
- Backend resolves only fields that actually require user IDs, then validates required fields, room, date, half-hour start/end boundaries, status values, and overlap conflicts before calling OpenAPI create.
- Backend cancels by updating the status field rather than deleting rows, unless the business explicitly wants destructive deletes.
- Frontend can update the schedule from the returned create/cancel result immediately. On strict-rate deployments, make read-back refresh manual or cooled down instead of automatic.

Backend `diagnose()` should be non-destructive. It may check:
- Qiqiao HTTP client availability.
- Token acquisition with redacted proof only.
- Form model/component schema lookup.
- Query endpoint with a tiny page size.
- Current user resolution path.

Backend `diagnose()` must not return token, access key, secret, raw CorpID/CropID, or admin account values.

If diagnostics show `hasHttpClient: true` and methods such as `sendGet/sendPost/sendPut/sendDel`, but token/schema/query all fail with `Can't find method ... sendGet(org.mozilla.javascript.ConsString,object,org.mozilla.javascript.Undefined)`, fix the adapter argument types before changing base URLs or credentials. The correct first probe is a plain URI, params `HashMap`, and headers `HashMap`.

## Verified existing-form management surface

Against the user's current test form, the existing-form API path has been verified for:

- Token acquisition with the public base URL.
- Current user lookup by account.
- Form component schema retrieval.
- Existing form record query.
- Single-record create, get by ID, filter by ID, update with the current version, and delete.
- Batch save, batch update using `version + 1`, and batch delete cleanup.
- Workflow definition list returns successfully, even when no workflow definitions exist.

The observed test form field types include `textBox`, `textarea`, `fileupload`, `time`, `date`, and `imageUpload`. `full-smoke` populates safe text/time/date fields automatically and skips file/image fields unless `--upload-file` is provided.

Observed non-blocking behaviors:
- Generic `applications.list` and `form_models.list` may return `系统繁忙，请稍后再试`; the specific application/form endpoints can still work.
- File upload can succeed, but may also intermittently return `系统繁忙，请稍后再试`; treat it as optional unless the current task is specifically file handling.
- `PUT /open/applications/{applicationId}/forms/batch_update_by_condition` is listed in `qqkf.txt`, but is not verified on the current deployment. Use the ordinary single-record or batch-update endpoints for production behavior unless this endpoint is separately confirmed.

Do not use OpenAPI to create form models or add fields unless a separate management/build endpoint has been supplied and verified. The public OpenAPI docs list form model/component reads, not form designer writes.
