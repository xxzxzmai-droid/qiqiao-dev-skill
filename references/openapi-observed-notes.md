# OpenAPI Observed Notes

Use this only after `openapi-integration.md` when a task matches these observed symptoms. These notes are deployment observations, not universal platform guarantees.

## Verified existing-form surface

One prior test deployment verified:

- Token acquisition through a public HTTPS base.
- Current user lookup by account.
- Form component schema retrieval.
- Existing form record query.
- Single-record create, get by ID, filter by ID, update with the current version, and delete.
- Batch save, batch update using `version + 1`, and batch delete cleanup.
- Workflow definition list returning successfully, including when no workflow definitions exist.

## Non-blocking responses seen in practice

- Generic `applications.list` and `form_models.list` can return `系统繁忙，请稍后再试` while specific application/form endpoints still work.
- File upload can succeed and still intermittently return `系统繁忙，请稍后再试`; treat upload as optional unless the task is specifically file handling.
- A captured `qqkf.txt` can list `PUT /open/applications/{applicationId}/forms/batch_update_by_condition`, but that does not prove the current deployment accepts the payload shape. Prefer ordinary single-record or documented batch endpoints unless this endpoint is separately confirmed.
- A bare HTTP client can receive an HTML or gateway response from an HTTPS base even when credentials are correct. Retry with browser-like headers before changing credentials.

## Reusable lessons

- Single-record `PUT /forms/{formModelId}` sends the current `version`; do not increment it locally.
- Batch update may require `version + 1`; keep single-record and batch semantics separate.
- If a response body is complete JSON and the transport ends with `unexpected EOF`, treat EOF as a read-tail condition rather than proof that the JSON is invalid.
- If a probe result looks like API failure, first separate proxy/certificate/gateway behavior from service business errors.
