(function () {
  "use strict";

  var state = {
    startedAt: new Date().toISOString(),
    checks: {},
    events: [],
    env: {}
  };

  function $(selector) {
    return document.querySelector(selector);
  }

  function safeString(value) {
    if (value == null) return "";
    try {
      if (typeof value === "string") return value;
      return JSON.stringify(value);
    } catch (err) {
      return String(value);
    }
  }

  function log(message, data) {
    var line = "[" + new Date().toLocaleTimeString() + "] " + message + (data === undefined ? "" : " " + safeString(data));
    state.events.push(line);
    var box = $("#logBox");
    if (box) box.textContent = state.events.join("\n");
  }

  function setCheck(key, status, detail) {
    state.checks[key] = { status: status, detail: detail || "" };
    var row = document.querySelector('[data-check="' + key + '"]');
    if (!row) return;
    row.classList.remove("pass", "fail", "warn");
    row.classList.add(status);
    var badge = row.querySelector("span");
    var em = row.querySelector("em");
    if (badge) badge.textContent = status === "pass" ? "PASS" : status === "fail" ? "FAIL" : "INFO";
    if (em && detail) em.textContent = detail;
    refreshOverall();
  }

  function refreshOverall() {
    var values = Object.keys(state.checks).map(function (key) { return state.checks[key].status; });
    var fail = values.filter(function (v) { return v === "fail"; }).length;
    var pass = values.filter(function (v) { return v === "pass"; }).length;
    var status = $("#overallStatus");
    if (!status) return;
    if (fail > 0) {
      status.textContent = "发现 " + fail + " 个失败项";
      status.style.color = "#b91c1c";
      status.style.background = "#fee2e2";
    } else if (pass >= 8) {
      status.textContent = "前端链路正常";
      status.style.color = "#15803d";
      status.style.background = "#dcfce7";
    } else {
      status.textContent = "检测中";
      status.style.color = "#1d4ed8";
      status.style.background = "#dbeafe";
    }
  }

  function parseRuntimeEnv() {
    var search = new URLSearchParams(location.search);
    var pathMatch = location.pathname.match(/\/business\/([^/]+)\/([^/]+)\/custompage\//);
    var applicationId = search.get("applicationId") || search.get("appId") || (pathMatch && pathMatch[1]) || "";
    var businessId = search.get("businessId") || search.get("bpmsId") || (pathMatch && pathMatch[2]) || "";
    var token = search.get("X-Auth0-Token") || search.get("x-auth0-token") || search.get("token") || "";
    var basePath = search.get("path") || "/qiqiao/runtime";
    var executeUrl = location.origin + "/dev-runtime/api/v1/runtime/business/" + applicationId + "/" + businessId + "/custompage/code/execute";

    state.env = {
      href: location.href,
      origin: location.origin,
      pathname: location.pathname,
      applicationId: applicationId,
      businessId: businessId,
      tokenPresent: !!token,
      basePath: basePath,
      executeUrl: executeUrl,
      iframe: window.self !== window.top,
      hasParentTriggerSocket: !!(window.parent && window.parent.triggerSocket),
      isPreviewPage: location.pathname.indexOf("/preview.html") >= 0,
      mode: location.pathname.indexOf("/preview.html") >= 0
        ? "preview"
        : (window.self !== window.top && window.parent && window.parent.triggerSocket)
          ? "debug"
          : "runtime"
    };
    return { applicationId: applicationId, businessId: businessId, token: token, executeUrl: executeUrl };
  }

  function runFrontendCheck() {
    var env = parseRuntimeEnv();
    setCheck("html", "pass", "HTML 已渲染。title=" + document.title);

    var cssLoaded = getComputedStyle(document.documentElement).getPropertyValue("--qq-css-loaded").trim() === "yes";
    setCheck("css", cssLoaded ? "pass" : "fail", cssLoaded ? "检测到 --qq-css-loaded=yes。" : "未检测到 index.css 注入变量。");

    setCheck("js", "pass", "index.js 已执行。startedAt=" + state.startedAt);
    setCheck("event", "pass", "按钮事件绑定正常：" + new Date().toLocaleTimeString());

    try {
      var s = document.createElement("script");
      s.type = "module";
      s.textContent = "window.__QQ_MODULE_OK__=true;document.dispatchEvent(new CustomEvent('qq-module-ok'));";
      document.head.appendChild(s);
      window.setTimeout(function () {
        setCheck("module", window.__QQ_MODULE_OK__ ? "pass" : "warn", window.__QQ_MODULE_OK__ ? "动态 module 可执行。" : "动态 module 未回报，可能被 CSP 限制。");
      }, 80);
    } catch (err) {
      setCheck("module", "fail", "动态 module 创建失败：" + err.message);
    }

    setCheck("ids", env.applicationId && env.businessId ? "pass" : "warn", "applicationId=" + (env.applicationId || "未取到") + "，businessId=" + (env.businessId || "未取到"));
    setCheck("token", env.token ? "pass" : "warn", env.token ? "URL 中存在 X-Auth0-Token/token。" : "URL 中未发现 token；运行态可能由平台自行鉴权，也可能导致 execute 接口 401。");
    if (state.env.mode === "preview") {
      setCheck("api-url", "warn", "当前是预览页。七巧文档说明预览不会调用服务端代码，因此不会请求：" + env.executeUrl);
      setCheck("mock", "pass", "当前是七巧预览模式，后端按钮将返回预览模拟结果，不再请求 execute 接口。");
    } else if (state.env.mode === "debug") {
      setCheck("api-url", "pass", "当前是调试 iframe，将通过 parent.triggerSocket 触发服务端调试。");
      setCheck("mock", "warn", "调试模式不走本地模拟，但返回值通常在七巧调试器面板查看。");
    } else {
      setCheck("api-url", env.applicationId && env.businessId ? "pass" : "warn", env.executeUrl);
      setCheck("mock", (env.applicationId && env.businessId) ? "warn" : "pass", (env.applicationId && env.businessId) ? "当前像发布运行环境，优先真实调用。" : "当前不是七巧运行环境，后端调用将返回本地模拟结果。");
    }

    log("前端自检完成", state.env);
  }

  function serializeArg(arg) {
    return JSON.stringify(arg);
  }

  function buildCode(methodName, args) {
    var list = (args || []).map(serializeArg).join(",");
    return "REST.API." + methodName + "(" + list + ")";
  }

  function mockBackend(methodName, args, reason) {
    return Promise.resolve({
      mocked: true,
      ok: true,
      method: methodName,
      args: args || [],
      reason: reason,
      time: new Date().toISOString()
    });
  }

  function callQiqiaoApi(methodName, args, options) {
    var env = parseRuntimeEnv();
    var force = options && options.force;
    var code = buildCode(methodName, args);

    if (state.env.mode === "preview") {
      return mockBackend(methodName, args, "七巧预览页不会调用服务端代码，按开发手册返回预览模拟结果。");
    }

    if (!force && (!env.applicationId || !env.businessId)) {
      return mockBackend(methodName, args, "缺少 applicationId/businessId，当前按本地或预览环境处理。");
    }

    if (state.env.mode === "debug" && window.parent && window.parent.triggerSocket) {
      try {
        window.parent.triggerSocket(code, "");
        return Promise.resolve({
          debugSocket: true,
          ok: true,
          method: methodName,
          code: code,
          note: "已调用 parent.triggerSocket，通常用于七巧 IDE 调试态；真实返回请看平台调试面板。"
        });
      } catch (err) {
        return Promise.reject(err);
      }
    }

    var headers = { "Content-Type": "application/json" };
    if (env.token) headers["X-Auth0-Token"] = env.token;

    var payload = {
      code: code,
      methodName: methodName,
      applicationId: env.applicationId,
      businessId: env.businessId
    };

    return fetch(env.executeUrl, {
      method: "POST",
      headers: headers,
      credentials: "include",
      body: JSON.stringify(payload)
    }).then(function (response) {
      return response.text().then(function (text) {
        var parsed;
        try { parsed = JSON.parse(text); } catch (err) { parsed = { rawText: text }; }
        if (!response.ok) {
          var error = new Error("HTTP " + response.status + " " + response.statusText);
          error.detail = parsed;
          error.request = payload;
          throw error;
        }
        return {
          ok: true,
          httpStatus: response.status,
          request: payload,
          response: parsed
        };
      });
    });
  }

  function renderResult(data) {
    $("#resultBox").textContent = JSON.stringify(data, null, 2);
  }

  function callAndRender(methodName, args) {
    setCheck("backend", "warn", "正在调用 REST.API." + methodName + "...");
    renderResult({ loading: true, method: methodName, args: args });
    log("开始后端调用", { method: methodName, args: args });

    callQiqiaoApi(methodName, args, { force: false })
      .then(function (data) {
        setCheck("backend", "pass", data.mocked ? "本地模拟调用成功。" : data.debugSocket ? "已触发七巧调试调用。" : "真实 execute 接口调用成功。");
        renderResult(data);
        log("后端调用成功", data);
      })
      .catch(function (err) {
        setCheck("backend", "fail", err.message || String(err));
        renderResult({
          ok: false,
          message: err.message || String(err),
          detail: err.detail || null,
          env: state.env
        });
        log("后端调用失败", { message: err.message || String(err), detail: err.detail || null });
      });
  }

  function bindEvents() {
    $("#runFrontend").addEventListener("click", runFrontendCheck);
    $("#callHealth").addEventListener("click", function () { callAndRender("health", []); });
    $("#callHello").addEventListener("click", function () { callAndRender("hello", ["七巧"]); });
    $("#callAdd").addEventListener("click", function () { callAndRender("add", [7, 35]); });
    $("#callEcho").addEventListener("click", function () {
      callAndRender("echo", [{ message: "来自前端 index.js", url: location.href, at: new Date().toISOString() }]);
    });
    $("#copyReport").addEventListener("click", function () {
      var report = JSON.stringify({ checks: state.checks, env: state.env, events: state.events }, null, 2);
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(report).then(function () { log("报告已复制"); });
      } else {
        renderResult({ report: report, note: "当前环境无 clipboard API，请手动复制。" });
      }
    });
  }

  function init() {
    bindEvents();
    runFrontendCheck();
    log("index.js 注入并初始化完成");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init, { once: true });
  } else {
    init();
  }

  window.addEventListener("error", function (event) {
    log("window.onerror", { message: event.message, source: event.filename, line: event.lineno });
  });
  window.addEventListener("unhandledrejection", function (event) {
    log("unhandledrejection", { reason: safeString(event.reason) });
  });
})();
