(() => {
  const RECENT_KEY = "hostsleuth.console.recent-targets.v1";
  const DENSITY_KEY = "hostsleuth.console.density.v1";
  const RECENT_LIMIT = 6;

  function adminConsoleInit() {
    const main = document.querySelector("main");
    if (!main || byId("globalTarget")) return;

    const deck = document.createElement("section");
    deck.className = "command-deck panel";
    deck.innerHTML = `
      <div class="command-deck-main">
        <div class="command-label">
          <span class="command-pulse"></span>
          <span>Quick target</span>
          <kbd>/</kbd>
        </div>
        <div class="command-input-row">
          <input id="globalTarget" autocomplete="off" spellcheck="false" placeholder="host-or-ip:port">
          <button class="command-action primary" id="globalDiagnose" type="button">Diagnose</button>
          <button class="command-action" id="globalExpect" type="button">Expect</button>
          <button class="command-action" id="globalDNS" type="button">DNS</button>
        </div>
        <div class="recent-targets" id="recentTargets" aria-label="Recent targets"></div>
      </div>
      <button class="density-toggle" id="densityToggle" type="button" aria-pressed="false" title="Toggle compact evidence density">
        <span>Density</span>
        <strong id="densityLabel">Comfortable</strong>
      </button>
    `;
    main.prepend(deck);

    const input = byId("globalTarget");
    byId("globalDiagnose").addEventListener("click", () => routeGlobalTarget("diagnose"));
    byId("globalExpect").addEventListener("click", () => routeGlobalTarget("contracts"));
    byId("globalDNS").addEventListener("click", () => routeGlobalTarget("dns"));
    input.addEventListener("keydown", (event) => {
      if (event.key === "Enter") {
        event.preventDefault();
        routeGlobalTarget("diagnose");
      }
    });

    byId("densityToggle").addEventListener("click", toggleDensity);
    applyStoredDensity();
    renderRecentTargets();

    window.addEventListener("keydown", (event) => {
      if (event.key !== "/" || event.ctrlKey || event.metaKey || event.altKey) return;
      const active = document.activeElement;
      if (active && ["INPUT", "TEXTAREA", "SELECT"].includes(active.tagName)) return;
      event.preventDefault();
      input.focus();
      input.select();
    });

    window.hostSleuthRememberTarget = rememberTarget;
  }

  function routeGlobalTarget(view) {
    const raw = byId("globalTarget").value.trim();
    if (!raw) {
      byId("globalTarget").focus();
      return;
    }

    if (view === "dns") {
      const host = hostPart(raw);
      const dnsInput = byId("dnsDetectiveName");
      if (dnsInput) dnsInput.value = host;
      showView("dns");
      if (dnsInput) dnsInput.focus();
      rememberTarget(raw);
      return;
    }

    const target = endpointTarget(raw);
    byId("globalTarget").value = target;
    rememberTarget(target);

    if (view === "diagnose") {
      byId("diagnoseTarget").value = target;
      showView("diagnose");
      runDiagnosis(target);
      return;
    }

    const contract = byId("contractTarget");
    if (contract) contract.value = target;
    showView("contracts");
    if (contract) contract.focus();
  }

  function endpointTarget(raw) {
    if (hasPort(raw)) return raw;
    if (raw.includes(":") && !raw.startsWith("[")) return `[${raw}]:443`;
    return `${raw}:443`;
  }

  function hasPort(raw) {
    if (raw.startsWith("[")) return /^\[[^\]]+\]:\d+$/.test(raw);
    return /^[^:]+:\d+$/.test(raw);
  }

  function hostPart(raw) {
    if (raw.startsWith("[")) {
      const end = raw.indexOf("]");
      return end > 1 ? raw.slice(1, end) : raw;
    }
    if (hasPort(raw)) return raw.slice(0, raw.lastIndexOf(":"));
    return raw;
  }

  function rememberTarget(target) {
    target = String(target || "").trim();
    if (!target) return;
    const recent = readRecentTargets().filter((value) => value !== target);
    recent.unshift(target);
    try {
      localStorage.setItem(RECENT_KEY, JSON.stringify(recent.slice(0, RECENT_LIMIT)));
    } catch (_) {
      // Browser-local convenience must never block troubleshooting.
    }
    renderRecentTargets();
  }

  function readRecentTargets() {
    try {
      const parsed = JSON.parse(localStorage.getItem(RECENT_KEY) || "[]");
      return Array.isArray(parsed) ? parsed.map(String).filter(Boolean).slice(0, RECENT_LIMIT) : [];
    } catch (_) {
      return [];
    }
  }

  function renderRecentTargets() {
    const box = byId("recentTargets");
    if (!box) return;
    box.replaceChildren();
    const recent = readRecentTargets();
    if (!recent.length) {
      const hint = document.createElement("span");
      hint.className = "recent-target-hint";
      hint.textContent = "Recent targets stay in this browser only.";
      box.append(hint);
      return;
    }

    recent.forEach((target) => {
      const button = document.createElement("button");
      button.type = "button";
      button.className = "recent-target";
      button.textContent = target;
      button.title = `Use ${target}`;
      button.addEventListener("click", () => {
        byId("globalTarget").value = target;
        byId("globalTarget").focus();
      });
      box.append(button);
    });
  }

  function applyStoredDensity() {
    let density = "comfortable";
    try {
      density = localStorage.getItem(DENSITY_KEY) || "comfortable";
    } catch (_) {}
    if (!["comfortable", "compact"].includes(density)) density = "comfortable";
    applyDensity(density);
  }

  function toggleDensity() {
    const current = document.documentElement.dataset.density || "comfortable";
    const next = current === "compact" ? "comfortable" : "compact";
    try {
      localStorage.setItem(DENSITY_KEY, next);
    } catch (_) {}
    applyDensity(next);
  }

  function applyDensity(density) {
    document.documentElement.dataset.density = density;
    const toggle = byId("densityToggle");
    const label = byId("densityLabel");
    if (toggle) toggle.setAttribute("aria-pressed", String(density === "compact"));
    if (label) label.textContent = density === "compact" ? "Compact" : "Comfortable";
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", adminConsoleInit);
  } else {
    adminConsoleInit();
  }
})();
