(() => {
  function adminConsoleV2Init() {
    const row = document.querySelector(".command-input-row");
    if (!row || byId("globalProxy")) return;

    const button = document.createElement("button");
    button.className = "command-action proxy-command-action";
    button.id = "globalProxy";
    button.type = "button";
    button.textContent = "Proxy";
    button.addEventListener("click", routeQuickTargetToProxy);
    row.append(button);
  }

  function routeQuickTargetToProxy() {
    const global = byId("globalTarget");
    const input = byId("proxyPublicURL");
    if (!global || !input) return;
    const raw = global.value.trim();
    if (!raw) {
      global.focus();
      return;
    }
    input.value = quickTargetURL(raw);
    if (window.hostSleuthRememberTarget) window.hostSleuthRememberTarget(raw);
    showView("proxy");
    const upstream = byId("proxyUpstreamURL");
    if (upstream && !upstream.value.trim()) upstream.focus();
  }

  function quickTargetURL(raw) {
    if (/^https?:\/\//i.test(raw)) return raw;
    if (raw.startsWith("[")) {
      const match = raw.match(/^\[[^\]]+\]:(\d+)$/);
      if (match) return `${quickScheme(match[1])}://${raw}/`;
      return `https://${raw}/`;
    }
    const match = raw.match(/^(.+):(\d+)$/);
    if (match) return `${quickScheme(match[2])}://${raw}/`;
    if (raw.includes(":")) return `https://[${raw}]/`;
    return `https://${raw}/`;
  }

  function quickScheme(port) {
    return ["80", "8000", "8080", "8787"].includes(String(port)) ? "http" : "https";
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", adminConsoleV2Init);
  } else {
    adminConsoleV2Init();
  }
})();
