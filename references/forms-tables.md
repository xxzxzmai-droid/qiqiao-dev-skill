# Forms, Tables, Page JS, and OpenAPI Reference

Use this for low-code scripts, custom CSS styles, 页面JS事件扩展, 表单字段联动, 表格/表单高级 API, custom form/page components, OpenAPI-adjacent form logic, and integrations outside the custom page IDE. For 数智彩虹 intranet token flows, UOS test binaries, and form data CRUD tooling, also read `openapi-integration.md`.

## Contents

- Low-code scripts
- Custom styles
- Page JS event extension
- Form context
- Custom form components
- Custom page components
- OpenAPI and table/form integrations

Official docs:
- 低代码脚本: https://qiqiao.do1.com.cn/help/develop_manual/%E4%BD%8E%E4%BB%A3%E7%A0%81%E8%84%9A%E6%9C%AC.html
- 低代码函数库: https://qiqiao.do1.com.cn/help/develop_manual/%E4%BD%8E%E4%BB%A3%E7%A0%81%E5%87%BD%E6%95%B0%E5%BA%93.html
- 自定义样式: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E6%A0%B7%E5%BC%8F.html
- 页面 JS API: https://qiqiao.do1.com.cn/help/develop_manual/%E9%A1%B5%E9%9D%A2JS%E4%BA%8B%E4%BB%B6%E6%89%A9%E5%B1%95/JS-API.html
- 自定义表单组件: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E8%A1%A8%E5%8D%95%E7%BB%84%E4%BB%B6.html
- 自定义页面组件: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2%E7%BB%84%E4%BB%B6.html
- 开放平台: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0.html
- 七巧低代码快速入门: https://qiqiao.do1.com.cn/help/user_manual/%E4%B8%83%E5%B7%A7%E4%BD%8E%E4%BB%A3%E7%A0%81%E5%BF%AB%E9%80%9F%E5%85%A5%E9%97%A8.html

## Low-code scripts

Use low-code scripts when the logic belongs inside standard Qiqiao surfaces rather than a custom page:

- First identify where the script runs: flow, business modeling, form event, page action, global script, or backend custom function.
- Clarify timing and mode: before/after submit, synchronous/asynchronous, field-level validation, workflow routing, data write-back, or external API call.
- Prefer the platform `$` function libraries before making external HTTP/OpenAPI calls. This avoids extra token load and usually has better runtime context.
- Debug with execution logs, online monitoring, `$.log`, and temporary station-message output when needed. Remove noisy debug outputs before delivery.
- Treat platform-returned DTOs/Java objects as runtime objects. Convert them to plain JSON-safe data before returning them to frontend or external systems.

Useful function-library groups:

- `$.contact`: users, departments, user groups, supervisors, account/user lookup.
- `$.context`: current application, current user, current form/process/page context.
- `$.form`: form model/data operations, child forms, document lookup and write helpers.
- `$.process` / workflow-related libraries: process instance/task context and workflow operations when available.
- `$.date`, `$.json`, `$.message`, `$.log`: date conversion, JSON conversion, station messages, logs.

## Custom styles

Use custom styles for presentation-only changes to standard Qiqiao pages/components.

Rules:
- Define the exact scope first: form design, flow page, app page, global page, PC, mobile, add/edit/detail, or runtime-only.
- Inspect the runtime DOM/classes in browser devtools, test CSS there, then paste the minimal selector into Qiqiao.
- Scope selectors with the documented root/isolation class or page/form-specific container; do not write broad global CSS that changes unrelated apps.
- Keep PC and mobile CSS separate when selectors or layout differ.
- Do not use custom CSS to hide required errors, permissions, workflow controls, or audit information unless the business owner explicitly approves.

## Page JS event extension

Do not use custom-page globals blindly in 页面JS事件扩展. That surface uses exported functions and contextual APIs:

```js
export function onClick() {
  const params = this.params;
  this.utils.Message({ message: "已触发" });
}
```

Rules:
- Export functions that must appear in the action panel.
- Keep function names unique.
- Treat `this.context.*` data as read-only unless using the provided setter API.
- Use `this.utils.dayjs`, `this.utils.axios`, `this.utils.Message`, `this.utils.MessageBox`, and `this.utils.loadScript` when available.
- Use `this.business.executeServiceAPI` / global service API patterns for custom function calls when the task is in page JS extension rather than custom page frontend code.
- Inspect callback parameters in the real runtime/F12 when the event surface is unclear. Form field `onChange` and PC-list button `onClick` pass different shapes.
- Keep event handlers short. Put repeated business logic in a service API or backend method rather than duplicating it across page buttons.

## Form context

Typical patterns:

```js
export function mounted() {
  const formData = this.context.form.getFormData();
  const formInfo = this.context.form.getFormInfo();
  const inputs = this.context.inputs;
}
```

To update form data:

```js
export function onChange() {
  this.context.form.setFormData({
    字段名: "新值"
  });
}
```

Before writing form logic, collect:
- form/model ID
- exact field names and field IDs
- field types and expected value shapes
- trigger event type
- whether the logic runs in add/edit/detail/process context
- read/write/hidden permissions for PC and mobile if the form participates in a workflow

## Custom form components

Use custom form components only when standard form components plus validation/linkage cannot meet the requirement.

Rules:
- Decide whether the component is an entity field or a non-entity/display component before implementation.
- Implement separate PC/mobile behavior when required by the component configuration.
- Respect platform props such as bound value, config, form data, validation rules, disabled/readonly display, child-table context, process IDs, task IDs, and component IDs.
- Update the bound `value` in the platform-supported shape; do not maintain a disconnected visual state that fails to save.
- Package/export components with fake sample data only. Never include real tenant IDs, accounts, secrets, or production field names in public examples.

## Custom page components

Use custom page components when a reusable Vue-based component should be dragged into page design.

Rules:
- Current official docs describe Vue 2 single-component SPA-style development. Verify the target platform before using newer Vue/runtime assumptions.
- Expect component resources such as `index.vue`, `setting.js`, and `event.json` when the component needs configurable properties or exposed events.
- Publish/enable the component and configure component permission scope before expecting page designers to use it.
- Use page JS event-extension APIs for component events when needed; keep component-to-component communication explicit and documented.

## OpenAPI and table/form integrations

For OpenAPI or cross-system data flows:
- Confirm that OpenAPI capability is enabled and the needed API permission is available.
- Confirm token acquisition, app/corp/application IDs, form model IDs, process IDs, and endpoint base path from the target deployment.
- Prefer server-side or gateway calls for secrets and internal APIs.
- Log request IDs and normalized errors; never expose raw secrets in UI or exported reports.
- For Qiqiao-deployed custom pages, runtime token calls may use same-origin OpenAPI when no durable secret is needed; still keep `CorpID`/`Secret` out of frontend files.
- Treat form-model creation/design as unconfirmed unless a management endpoint has been verified in the target deployment.

For table-like pages, treat schema discovery as mandatory. Ask for or inspect actual fields before building filters, maps, or import/export logic.
