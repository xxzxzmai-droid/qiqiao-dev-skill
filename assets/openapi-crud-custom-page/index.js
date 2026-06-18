(function () {
  "use strict";

  var DEFAULTS = {
    applicationId: "",
    formModelId: "",
    openApiBase: "/qiqiao/runtime/api/v1/bpms-integration"
  };

  function init() {
    var app = document.getElementById("app");
    if (!app || app.dataset.ready === "1") return;
    app.dataset.ready = "1";

    fillRuntimeConfig();
    bind("btnProbe", probe);
    bind("btnSchema", loadSchema);
    bind("btnQuery", queryRecords);
    bind("btnCreate", createRecord);
    bind("btnUpdate", updateRecord);
    bind("btnDelete", deleteRecord);
  }

  function bind(id, handler) {
    var node = document.getElementById(id);
    if (node) node.addEventListener("click", handler);
  }

  function fillRuntimeConfig() {
    setValue("applicationId", readParam("applicationId") || DEFAULTS.applicationId);
    setValue("formModelId", readParam("formModelId") || DEFAULTS.formModelId);
    var token = getToken();
    var mode = token ? "运行态" : "预览态";
    setStatus(mode + " / " + (token ? "已读取运行 token" : "未发现 token，接口返回模拟数据"));
  }

  function getConfig() {
    return {
      applicationId: value("applicationId"),
      formModelId: value("formModelId"),
      loginUserId: value("loginUserId"),
      recordId: value("recordId"),
      version: Number(value("version") || 0),
      openApiBase: readParam("openApiBase") || DEFAULTS.openApiBase,
      token: getToken()
    };
  }

  function getToken() {
    return readParam("X-Auth0-Token") || readParam("token") || "";
  }

  function readParam(name) {
    var params = new URLSearchParams(window.location.search || "");
    return params.get(name) || "";
  }

  function value(id) {
    var node = document.getElementById(id);
    return node ? node.value.trim() : "";
  }

  function setValue(id, text) {
    var node = document.getElementById(id);
    if (node && text) node.value = text;
  }

  function variables() {
    var text = value("variablesJson");
    if (!text) return {};
    return JSON.parse(text);
  }

  function cleanVariables(raw) {
    var next = {};
    Object.keys(raw || {}).forEach(function (key) {
      if (key !== "id" && key !== "loginUserId" && key !== "version") {
        next[key] = raw[key];
      }
    });
    return next;
  }

  function savePayload() {
    var cfg = getConfig();
    return {
      variables: cleanVariables(variables()),
      id: cfg.recordId || randomId(),
      loginUserId: cfg.loginUserId
    };
  }

  function updatePayload() {
    var cfg = getConfig();
    return {
      variables: cleanVariables(variables()),
      id: cfg.recordId,
      version: cfg.version,
      loginUserId: cfg.loginUserId
    };
  }

  function randomId() {
    var bytes = new Uint8Array(16);
    if (window.crypto && window.crypto.getRandomValues) {
      window.crypto.getRandomValues(bytes);
      return Array.prototype.map.call(bytes, function (b) {
        return ("0" + b.toString(16)).slice(-2);
      }).join("");
    }
    return String(Date.now()) + String(Math.random()).slice(2, 20);
  }

  function endpoint(path) {
    var cfg = getConfig();
    return cfg.openApiBase.replace(/\/+$/, "") + path;
  }

  function requireConfig(requireLogin) {
    var cfg = getConfig();
    if (!cfg.applicationId || !cfg.formModelId) {
      throw new Error("请填写应用 ID 和表单模型 ID");
    }
    if (requireLogin && !cfg.loginUserId) {
      throw new Error("新增/修改需要填写登录用户 ID");
    }
    return cfg;
  }

  function api(method, path, body) {
    var cfg = getConfig();
    if (!cfg.token) {
      return Promise.resolve({
        code: 0,
        msg: "preview / 预览模式模拟返回",
        data: {
          method: method,
          path: path,
          body: body || null
        }
      });
    }
    return fetch(endpoint(path), {
      method: method,
      headers: {
        "Content-Type": "application/json",
        "X-Auth0-Token": cfg.token
      },
      body: body === undefined ? undefined : JSON.stringify(body)
    }).then(function (res) {
      return res.text().then(function (text) {
        var data = parseApiResponse(text, res);
        if (!res.ok) {
          var error = new Error((data && data.message) || ("HTTP " + res.status + ": " + text.slice(0, 300)));
          error.detail = data;
          throw error;
        }
        return data;
      });
    });
  }

  function parseApiResponse(text, res) {
    var contentType = res && res.headers ? (res.headers.get("content-type") || "") : "";
    var trimmed = (text || "").trim();
    if (!trimmed) return {};
    if (trimmed.charAt(0) === "<" || contentType.indexOf("text/html") >= 0) {
      return {
        ok: false,
        nonJson: true,
        message: "OpenAPI 返回 HTML/非 JSON，可能登录态失效、路径不匹配或网关返回错误页。",
        httpStatus: res ? res.status : 0,
        contentType: contentType,
        snippet: trimmed.slice(0, 500)
      };
    }
    try {
      return JSON.parse(trimmed);
    } catch (err) {
      return {
        ok: false,
        nonJson: true,
        message: err.message,
        httpStatus: res ? res.status : 0,
        contentType: contentType,
        snippet: trimmed.slice(0, 500)
      };
    }
  }

  function probe() {
    try {
      var cfg = requireConfig(false);
      run("连通探测", api("GET", "/open/applications/" + encodeURIComponent(cfg.applicationId) + "/form_models"));
    } catch (err) {
      show("连通探测", err);
    }
  }

  function loadSchema() {
    try {
      var cfg = requireConfig(false);
      run("读取结构", api("GET", "/open/applications/" + encodeURIComponent(cfg.applicationId) + "/form_models/" + encodeURIComponent(cfg.formModelId)));
    } catch (err) {
      show("读取结构", err);
    }
  }

  function queryRecords() {
    try {
      var cfg = requireConfig(false);
      var path = "/open/applications/" + encodeURIComponent(cfg.applicationId) + "/forms/" + encodeURIComponent(cfg.formModelId) + "/query?page=1&pageSize=20";
      run("查询", api("POST", path, []), renderRecords);
    } catch (err) {
      show("查询", err);
    }
  }

  function createRecord() {
    try {
      var cfg = requireConfig(true);
      var path = "/open/applications/" + encodeURIComponent(cfg.applicationId) + "/forms/" + encodeURIComponent(cfg.formModelId);
      run("新增", api("POST", path, savePayload()));
    } catch (err) {
      show("新增", err);
    }
  }

  function updateRecord() {
    try {
      var cfg = requireConfig(true);
      if (!cfg.recordId) throw new Error("修改需要填写记录 ID");
      var path = "/open/applications/" + encodeURIComponent(cfg.applicationId) + "/forms/" + encodeURIComponent(cfg.formModelId);
      run("修改", api("PUT", path, updatePayload()));
    } catch (err) {
      show("修改", err);
    }
  }

  function deleteRecord() {
    try {
      var cfg = requireConfig(false);
      if (!cfg.recordId) throw new Error("删除需要填写记录 ID");
      var path = "/open/applications/" + encodeURIComponent(cfg.applicationId) + "/forms/" + encodeURIComponent(cfg.formModelId) + "/" + encodeURIComponent(cfg.recordId);
      run("删除", api("DELETE", path));
    } catch (err) {
      show("删除", err);
    }
  }

  function run(action, promise, after) {
    setMessage("请求中...");
    promise.then(function (data) {
      show(action, data);
      if (after) after(data);
    }).catch(function (err) {
      show(action, err);
    });
  }

  function show(action, data) {
    var output = document.getElementById("output");
    var lastAction = document.getElementById("lastAction");
    if (lastAction) lastAction.textContent = action;
    if (output) {
      output.textContent = data instanceof Error ? data.message : JSON.stringify(data, null, 2);
    }
    setMessage(data instanceof Error ? "失败" : "完成");
  }

  function setStatus(text) {
    var node = document.getElementById("runtimeStatus");
    if (node) node.textContent = text;
  }

  function setMessage(text) {
    var node = document.getElementById("message");
    if (node) node.textContent = text;
  }

  function renderRecords(data) {
    var host = document.getElementById("records");
    if (!host) return;
    var list = data && data.data && (data.data.list || data.data.rows);
    if (!Array.isArray(list)) {
      host.innerHTML = "";
      return;
    }
    host.innerHTML = "<table><thead><tr><th>ID</th><th>version</th><th>variables</th></tr></thead><tbody>" +
      list.map(function (item) {
        return "<tr><td><code>" + escapeHTML(item.id || "") + "</code></td><td>" + escapeHTML(String(item.version || "")) + "</td><td><code>" +
          escapeHTML(JSON.stringify(item.variables || item)) + "</code></td></tr>";
      }).join("") +
      "</tbody></table>";
  }

  function escapeHTML(text) {
    return String(text).replace(/[&<>"']/g, function (c) {
      return { "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c];
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init, { once: true });
  } else {
    init();
  }
})();
