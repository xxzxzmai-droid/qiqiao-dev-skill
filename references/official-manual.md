# Official Manual Guidance

Use this when the user asks how to operate Qiqiao/七巧/数智彩虹, asks what the platform supports, or needs a route from business requirements to either configuration or code.

## Contents

- Official entry points
- Coverage map
- User-manual operating model
- App, page, and model routing
- Forms
- Flows
- PC and mobile pages
- Developer-manual routing
- Escalation rules

## Official entry points

- User manual: https://qiqiao.do1.com.cn/help/user_manual/%E4%B8%83%E5%B7%A7%E4%BD%8E%E4%BB%A3%E7%A0%81%E5%BF%AB%E9%80%9F%E5%85%A5%E9%97%A8.html
- Developer manual: https://qiqiao.do1.com.cn/help/develop_manual/%E4%B8%83%E5%B7%A7%E5%BC%80%E5%8F%91%E8%80%85%E6%89%8B%E5%86%8C.html
- App build quick start: https://qiqiao.do1.com.cn/help/user_manual/%E4%B8%83%E5%B7%A7%E4%BD%8E%E4%BB%A3%E7%A0%81%E5%BF%AB%E9%80%9F%E5%85%A5%E9%97%A8/%E6%90%AD%E5%BB%BA%E4%B8%80%E4%B8%AA%E7%AE%80%E5%8D%95%E7%9A%84%E5%BA%94%E7%94%A8.html
- App construction details: https://qiqiao.do1.com.cn/help/user_manual/%E4%B8%83%E5%B7%A7%E4%BD%8E%E4%BB%A3%E7%A0%81%E5%BF%AB%E9%80%9F%E5%85%A5%E9%97%A8/%E6%90%AD%E5%BB%BA%E4%B8%80%E4%B8%AA%E7%AE%80%E5%8D%95%E7%9A%84%E5%BA%94%E7%94%A8/%E4%B8%80%E3%80%81%E5%BA%94%E7%94%A8%E6%90%AD%E5%BB%BA.html
- Form design: https://qiqiao.do1.com.cn/help/user_manual/%E8%A1%A8%E5%8D%95%E8%AE%BE%E8%AE%A1.html
- Flow design: https://qiqiao.do1.com.cn/help/user_manual/%E6%B5%81%E7%A8%8B%E8%AE%BE%E8%AE%A1.html
- PC page design: https://qiqiao.do1.com.cn/help/user_manual/PC%E7%AB%AF%E9%A1%B5%E9%9D%A2%E8%AE%BE%E8%AE%A1.html
- Mobile page design: https://qiqiao.do1.com.cn/help/user_manual/%E7%A7%BB%E5%8A%A8%E7%AB%AF%E9%A1%B5%E9%9D%A2%E8%AE%BE%E8%AE%A1.html
- Developer overview: https://qiqiao.do1.com.cn/help/develop_manual/%E4%B8%83%E5%B7%A7%E5%BC%80%E5%8F%91%E8%AF%B4%E6%98%8E.html
- OpenAPI: https://qiqiao.do1.com.cn/help/develop_manual/%E5%BC%80%E6%94%BE%E5%B9%B3%E5%8F%B0/OpenAPI.html
- Page JS API: https://qiqiao.do1.com.cn/help/develop_manual/%E9%A1%B5%E9%9D%A2JS%E4%BA%8B%E4%BB%B6%E6%89%A9%E5%B1%95/JS-API.html
- Custom page: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2.html
- Custom page tutorial: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2/%E4%BD%BF%E7%94%A8%E6%95%99%E7%A8%8B.html
- Custom form component: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E8%A1%A8%E5%8D%95%E7%BB%84%E4%BB%B6.html
- Custom page component: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2%E7%BB%84%E4%BB%B6.html

If the user asks for exact current UI wording, re-open the relevant official page before answering because product docs may change.

## Coverage map

The official help tree currently covers these broad user-manual areas. Use this as a routing checklist before deciding that code is required:

- 我的应用 / 应用管理 / 应用开发: app creation, page management, model management, application settings, publishing.
- 数据&集成 / 数据工厂: data sources, integration, data processing, import/export-style platform configuration.
- 低码中心 / 低码引擎: script lists, execution logs, online debugging, CSS style extensions, JS script extensions, custom components.
- 系统管理 / 全局 / 通讯录: role and app permissions, process permissions, admin accounts, operation logs, system extensions, SSO, official account/service account/email/unified-todo integration, external users.
- 表单设计: global properties, layouts, base components, relation components, ER view, validation, style, field permissions.
- 流程设计 / 业务编排: approval flows, workflow/business-flow design, nodes, routing, actions, workflow logs, triggering workflows.
- 报表设计 / 聚合表: charts, BI/report pages, aggregation, aggregate query, logs, usage limits.
- PC端页面设计 / 移动端页面设计 / 公共类组件: standard page components, lists, forms, reports, boards, todos, navigation, buttons, print/QR use cases, custom pages.
- 自动化（智能助手） / 智能助手 / 七巧AI: automatic reminders/tasks, assistant enablement/logs, AI-assisted platform features.
- 第三方集成: enterprise WeChat, DingTalk, Feishu and related platform integration.

Developer manual coverage:

- 七巧开发说明 and 自定义对象: platform object models for contact, form, process, validation, and paging.
- 低代码函数库 and 低代码脚本: `$` function libraries, script authoring, logs, online debugging, scenarios.
- 自定义样式: CSS extension scope, runtime element inspection, PC/mobile isolation.
- 页面JS事件扩展: exported functions, context APIs, callback parameters, scenario examples.
- 自定义表单组件 and 自定义页面组件: component packaging, Vue/component code, props/events/settings, publish and permission scope.
- 自定义页面: four-file custom page IDE, preview/debug/publish, server-code API bridge.
- 开放平台: OpenAPI, form/workflow/common APIs, event push, workflow push, portal todo integration, unified message push, and zero-code integration.

## User-manual operating model

Guide ordinary configuration work before proposing code:

1. Prepare: clarify users, pain points, roles, business objects, functional list, and business flow. Keep boundaries concrete.
2. Build: create or import the app, then configure page management, model management, and flow/automation.
3. Run: publish, test with a small group, verify the main business flow, then expand usage and monitor data.

When advising a user, answer with:

```text
操作入口 -> 配置步骤 -> 发布/启用 -> 验证点 -> 何时需要代码扩展
```

Do not jump to custom code if a standard form, list page, report, flow, automation, or page component can satisfy the requirement.

## App, page, and model routing

- App creation supports blank creation, template creation, and importing an existing app. Capture app name, icon, theme color, and group if the user asks for setup guidance.
- Page management is where users choose the visible business surface. Route by page type:
  - 普通表单页面: simple data collection and independent form links.
  - 流程表单页面: approval or business circulation.
  - 报表页面: data display and BI-style analysis.
  - 基础页面: standard component-driven data management.
  - 高级页面: finer custom capability, external embeds, and IDE-coded custom pages.
- Model management centralizes forms and reports so complex pages can reuse them.
- Flow & automation covers approvals, workflows, and automatic actions.

For coded delivery, identify the target surface first: PC page, mobile page, custom page component, custom page IDE, page JS extension, or server custom function.

## Forms

- Treat a form as the lowest-level data model for collecting, storing, and managing business data.
- Configuration path: new form -> choose components -> configure component properties -> adjust layout -> save.
- Before coding against a form, collect exact form model ID, field names, field IDs, component types, value shapes, required rules, option values, permissions, and whether the form is used in add/edit/detail/process contexts.
- Common components include layout, text, number, single/multi choice, date/time, upload, user/department selectors, address/cascade, generated code, rating, location, summary, signature, online editing, and invoice components. Do not guess OpenAPI payload shapes from display labels alone.

## Flows

- Build flows from nodes and connecting lines. Artificial/task nodes need handler settings, circulation rules, node actions, and form-field read/write/hidden permissions.
- Multiple outgoing lines from a human/system/subprocess node behave as an exclusive gateway: branch order and conditions matter, and the first matching branch wins before the default branch.
- Configure launch/visibility scope when enabling a flow. Unenabled flows cannot be started from the runtime entry.
- Flow design updates affect newly started data after save/publish/version behavior; do not assume old running instances change automatically.
- Remember separate web and mobile field permission configuration when the workflow is used on both ends.

## PC and mobile pages

- PC page path: create page -> design page -> bind form/data source where needed -> add operation buttons -> configure properties -> test and publish.
- PC pages can use lists for CRUD, reports, rankings, summaries, custom pages, workflow/task views, and other components.
- Mobile page path: add page/group -> add component -> bind form where needed -> configure properties -> save and publish.
- Mobile pages support data display, search/filter, add/edit/delete, quick form entry, navigation icons, bottom menus/navigation, statistics, summaries, charts, boards, and todo lists.
- For mobile coded pages, keep helper/step modules compact and prioritize the visible selection/content area. Avoid large persistent instructional panels.

## Developer-manual routing

Use developer features only when standard configuration is insufficient:

- Low-code scripts: JavaScript syntax, platform APIs, Java class access when available, save-effective editing, code completion, formatting, and syntax validation. Prefer closure/IIFE patterns when writing scripts in platform editors.
- Page JS event extension: exported functions run with `this.context`, `this.utils`, `this.business`, and event callback parameters. Use runtime/F12 inspection for exact parameters. Read `forms-tables.md` for patterns.
- Custom styles: use for CSS-only presentation changes across form, flow, app page, or global page surfaces. Inspect runtime DOM/classes first, scope selectors tightly, and separate PC/mobile rules.
- Custom page IDE: four files only: `index.html`, `index.css`, `index.js`, and one backend custom function file. Preview is frontend-only; debug can hit backend breakpoints/logs; publish creates the runtime version users see. Read `custom-page.md` and `backend-api.md`.
- Custom page bridge: frontend calls backend `API` methods through `applyApi`/execute with `applicationId`, `businessId`, and `X-Auth0-Token`. Keep guarded parsing and deployment-prefix diagnostics from `backend-api.md`.
- OpenAPI: documented categories include contact/user/department, form, workflow, message, auth, material/file, and related common APIs. Read `openapi-integration.md` before implementing auth, form/workflow CRUD, or UOS tooling.
- Push integrations: use event push for form-data changes, workflow push/portal todo for todo/read/done data, and unified message push for message-card synchronization. Read `integration-events.md`.
- Custom form components: use when the standard form component set cannot provide the UI/data behavior. Expect richer UI, combined elements, data-set usage, and independent iteration.
- Custom page components: use when reusable JS/CSS/Vue page components are needed across PC/H5/mini-program surfaces, including component-to-component data passing, linkage, dynamic visibility, and property association.

## Escalation rules

- If the user asks "怎么操作", prefer a short official-manual-aligned procedure and avoid code unless necessary.
- If the user asks for a working artifact, implement and validate using the relevant development references.
- If user-supplied field names, IDs, permissions, or deployment URLs are missing, inspect the target app/export/API when possible; otherwise state the exact missing platform facts.
- If a feature crosses system boundaries, secrets, or intranet access, keep secrets server-side and use OpenAPI/event/bridge diagnostics instead of frontend-only calls.
- If a task involves an official feature that is configuration-only in the manual, provide exact setup guidance and verification points instead of manufacturing a code artifact.
