/*
 * 七巧 IDE 服务端代码测试 Demo
 *
 * 使用方式：
 * 1. 把 index.html / index.css / index.js 分别粘到自定义页面三个前端文件。
 * 2. 把本文件完整粘到“服务端代码”区域。
 * 3. 保存后打开运行页，点击 health / hello / add / echo 按钮测试前后端链路。
 *
 * 说明：
 * - 七巧开发手册要求服务端代码声明 var API = { ... }。
 * - 前端会通过 REST.API.health() / REST.API.hello("七巧") 等形式调用。
 * - 本文件尽量使用 ES5 写法，避免服务端 JS 运行环境不支持新语法。
 */

var FUNC = {
  now: function () {
    return new Date().toISOString();
  },

  log: function (message, data) {
    try {
      if (typeof $ !== "undefined" && $.log && $.log.info) {
        $.log.info("[七巧后端Demo] " + message + (data === undefined ? "" : " " + JSON.stringify(data)));
      }
    } catch (e) {
      // 日志失败不影响业务返回
    }
  },

  getUserId: function () {
    try {
      if (typeof $ !== "undefined" && $.context && $.context.getCurrentUserId) {
        return $.context.getCurrentUserId();
      }
    } catch (e) {
      return "context-error: " + e.message;
    }
    return "context-unavailable";
  },

  wrap: function (name, data) {
    return {
      ok: true,
      api: name,
      serverTime: FUNC.now(),
      userId: FUNC.getUserId(),
      data: data
    };
  }
};

var API = {
  health: function () {
    FUNC.log("health called");
    return FUNC.wrap("health", {
      message: "服务端 API 可用",
      runtime: "qiqiao-server-code"
    });
  },

  hello: function (name) {
    FUNC.log("hello called", { name: name });
    return FUNC.wrap("hello", {
      input: name,
      text: "你好，" + String(name || "七巧")
    });
  },

  add: function (a, b) {
    var x = Number(a);
    var y = Number(b);
    FUNC.log("add called", { a: a, b: b });
    return FUNC.wrap("add", {
      a: x,
      b: y,
      result: x + y
    });
  },

  echo: function (payload) {
    FUNC.log("echo called", payload);
    return FUNC.wrap("echo", {
      payloadType: typeof payload,
      payload: payload
    });
  }
};
