# Event and Push Integration Reference

Use this for Qiqiao/七巧/数智彩虹 event push, workflow push, portal todo integration, unified message push, webhook receiver design, and zero-code integration routing. For pull-style OpenAPI calls, read `openapi-integration.md`.

Official docs:
- 事件推送开发指南: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0/%E4%BA%8B%E4%BB%B6%E6%8E%A8%E9%80%81%E5%BC%80%E5%8F%91%E6%8C%87%E5%8D%97.html
- 流程事件推送: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0/%E6%B5%81%E7%A8%8B%E4%BA%8B%E4%BB%B6%E6%8E%A8%E9%80%81.html
- 流程推送开发指南: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0/%E6%B5%81%E7%A8%8B%E6%8E%A8%E9%80%81%E5%BC%80%E5%8F%91%E6%8C%87%E5%8D%97.html
- 门户待办集成: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0/%E9%97%A8%E6%88%B7%E5%BE%85%E5%8A%9E%E9%9B%86%E6%88%90.html
- 统一消息推送: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0/%E7%BB%9F%E4%B8%80%E6%B6%88%E6%81%AF%E6%8E%A8%E9%80%81.html
- 零代码集成方案: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0/%E9%9B%B6%E4%BB%A3%E7%A0%81%E9%9B%86%E6%88%90%E6%96%B9%E6%A1%88.html

## Choose the integration pattern

- **OpenAPI pull/write**: external system or Qiqiao server code requests data or writes data on demand. Use `openapi-integration.md`.
- **Form event push**: Qiqiao pushes form create/update/delete events to a receiver.
- **Workflow event push**: Qiqiao pushes workflow lifecycle events to a receiver.
- **Workflow push / portal todo**: Qiqiao pushes todo/read/done/delete/update data to a unified portal/todo system.
- **Unified message push**: Qiqiao mirrors message cards to a third-party message receiver.
- **Zero-code integration**: route to the configured integration provider when the user's need matches a supported trigger/action and no custom endpoint is needed.

Always name the data owner, trigger, receiver URL, auth secret, retry behavior, idempotency key, and replay/reconciliation path.

## Form event push

Use for realtime form-data integration when records are added, modified, or deleted.

Receiver rules:
- Implement URL verification first. The receiver must handle the `URL_VERIFY` event and return the expected token/signature payload.
- Accept `X-Auth0-DeliverId` and keep it as the delivery/event id for logs and idempotency.
- Respond with a 2xx status within the documented timeout window. Long business processing should be queued after acknowledging.
- Expect retries. Store processed delivery IDs or business record/version keys so repeated events do not duplicate work.
- Do not fail only because the receiver sees a new/unknown `eventType`; log it and return success if the platform expects successful receipt.

Payload handling:
- Expect `variables`, record `id`, `version`, form/process IDs, task ID, author/modifier metadata, timestamps, and `applicationId`.
- Treat field values as platform-shaped values, not display text. Reconcile field IDs/names against form schema when possible.

## Workflow push and portal todo

Use for integrating Qiqiao process data into third-party todo centers.

Rules:
- Configure separate endpoints when required for add, update, and delete todo actions.
- Token/AES-style push secrets belong in receiver-side config, not in frontend code.
- Preserve `workId` as the primary id of a todo/read/done item.
- Track fields such as title, URL, creator, creator name, status, current node, create/update/deal time, process model/instance IDs, task ID, user ID, and application ID.
- Implement idempotent add/update/delete. A delete for a missing work item should be harmless.
- If a third-party portal falls behind, use the documented query/reconciliation OpenAPI where available rather than asking Qiqiao to re-push blindly.

## Unified message push

Use when Qiqiao messages sent to enterprise WeChat need to be mirrored to another message system.

Rules:
- Capture `eventName`, receiver-provided `appid`, `appSecret`, interface URL, homepage URL, and pushed data source.
- Persist the generated `flowId` to distinguish multiple push events.
- Verify request signatures according to the receiver contract.
- Keep message rendering decisions in the receiver. Qiqiao supplies message content; the receiver decides final presentation.

## Receiver implementation checklist

- HTTPS endpoint reachable from Qiqiao deployment.
- URL verification implemented.
- Signature/token verification implemented.
- Raw request body captured before JSON parsing if signature verification needs it.
- `X-Auth0-DeliverId` or equivalent delivery id logged.
- Idempotency store keyed by delivery id plus business id/version when needed.
- Fast 2xx acknowledgement with async downstream processing for slow jobs.
- Retry-safe update/delete behavior.
- Redacted logs. Never log secrets, tokens, raw account credentials, or personal data beyond what debugging requires.
- Manual replay/reconciliation command or documented operator procedure.

## Diagnostics

When a push integration fails, collect:
- Qiqiao push configuration screenshot or redacted config: event type, URL, selected app/form/process, token presence, enabled state.
- Receiver access log with timestamp, method, status, headers minus secrets, and body size.
- Delivery id, event type, business id, version, and response status.
- Whether the receiver returned 2xx within timeout.
- Whether the same delivery id was already processed.

Do not troubleshoot push failures by changing OpenAPI token code unless the receiver also calls OpenAPI during processing.
