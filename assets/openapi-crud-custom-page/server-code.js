var FUNC = {
  log: function (message, data) {
    try {
      if (typeof $ !== "undefined" && $.log && $.log.info) {
        $.log.info(message + (data === undefined ? "" : " " + JSON.stringify(data)));
      }
    } catch (e) {}
  }
};

var API = {
  health: function () {
    FUNC.log("openapi-crud-custom-page health");
    return {
      ok: true,
      message: "server-code is available",
      time: new Date().toISOString()
    };
  }
};
