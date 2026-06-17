# Qiqiao Custom Page Reference

Use this for 七巧自定义页面, custompage code, 三文件 IDE pages, and upload/paste delivery.

Official docs:
- 自定义页面概述: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2.html
- 使用教程: https://qiqiao.do1.com.cn/help/develop_manual/%E8%87%AA%E5%AE%9A%E4%B9%89%E9%A1%B5%E9%9D%A2/%E4%BD%BF%E7%94%A8%E6%95%99%E7%A8%8B.html

## Platform facts

- The custom page IDE has a frontend group and a backend group.
- The frontend group is fixed to `index.html`, `index.css`, and `index.js`.
- The current custom page editor supports only these three frontend files plus one backend custom function file.
- Qiqiao injects CSS and JS into the HTML; the HTML does not need explicit references.
- Preview renders frontend only. Server code is not called during preview.
- Debugging and published runtime are separate from preview. Published users see the published version, not the in-development version.

## Frontend structure

Use this shape:

```html
<!doctype html>
<html lang="zh-CN">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>页面标题</title>
</head>
<body>
  <main id="app">...</main>
</body>
</html>
```

Do not include:

```html
<link rel="stylesheet" href="./index.css">
<script src="./index.js"></script>
<script type="module" src="/assets/app.js"></script>
```

## JavaScript pattern

Use an IIFE and one guarded init:

```js
(function () {
  "use strict";

  function init() {
    var app = document.getElementById("app");
    if (!app || app.dataset.ready === "1") return;
    app.dataset.ready = "1";
    // bind events here
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init, { once: true });
  } else {
    init();
  }
})();
```

## Assets and dependencies

- Prefer CSS, SVG, or data URLs for small local assets.
- Use the Qiqiao material library links for images or larger files when available.
- For third-party libraries, prefer a no-chunk IIFE bundle copied into `index.js`, or a platform-approved CDN/asset URL. Avoid ESM chunks unless Qiqiao is verified to load them.
- If an existing Vite/React app must be moved into Qiqiao, build a single IIFE/no-split bundle or produce a single-file fallback; do not paste Vite's normal `dist/index.html` directly.

## Template

Copy `assets/custom-page-injection/` when starting a custom page that needs frontend/backend verification. It intentionally omits manual CSS/JS tags and includes runtime diagnostics.
