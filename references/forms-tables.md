# Forms, Tables, Page JS, and OpenAPI Reference

Use this for 页面JS事件扩展, 表单字段联动, 表格/表单高级 API, custom components, OpenAPI, and integrations outside the custom page IDE.

Official docs:
- 页面 JS API: https://qiqiao.do1.com.cn/help/develop_manual/%E9%A1%B5%E9%9D%A2JS%E4%BA%8B%E4%BB%B6%E6%89%A9%E5%B1%95/JS-API.html
- 开放平台: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0.html
- 七巧低代码快速入门: https://qiqiao.do1.com.cn/help/user_manual/%E4%B8%83%E5%B7%A7%E4%BD%8E%E4%BB%A3%E7%A0%81%E5%BF%AB%E9%80%9F%E5%85%A5%E9%97%A8.html

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

## OpenAPI and table/form integrations

For OpenAPI or cross-system data flows:
- Confirm that OpenAPI capability is enabled and the needed API permission is available.
- Confirm token acquisition, app/corp/application IDs, form model IDs, process IDs, and endpoint base path from the target deployment.
- Prefer server-side or gateway calls for secrets and internal APIs.
- Log request IDs and normalized errors; never expose raw secrets in UI or exported reports.

For table-like pages, treat schema discovery as mandatory. Ask for or inspect actual fields before building filters, maps, or import/export logic.
