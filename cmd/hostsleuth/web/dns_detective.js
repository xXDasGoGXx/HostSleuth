(() => {
  const DNS_RESOLVER_STORAGE = "hostsleuth.dns.resolvers.v1";

  function dnsInit() {
    if (document.querySelector('[data-view="dns"]')) return;

    const nav = document.querySelector(".nav-tabs");
    const main = document.querySelector("main");
    if (!nav || !main) return;

    const navButton = document.createElement("button");
    navButton.className = "nav-tab";
    navButton.dataset.view = "dns";
    navButton.textContent = "DNS Detective";
    navButton.addEventListener("click", () => showView("dns"));
    nav.append(navButton);

    const view = document.createElement("section");
    view.className = "view";
    view.id = "view-dns";
    view.dataset.viewPanel = "dns";
    view.innerHTML = `
      <section class="panel section-panel dns-hero">
        <div class="section-heading">
          <div>
            <p class="eyebrow">DNS DETECTIVE</p>
            <h1>Who is giving you that answer?</h1>
            <p class="section-copy">Compare the resolver view HostSleuth is using with DNS servers you explicitly choose. Nothing is sent to a public resolver unless you enter it.</p>
          </div>
          <span class="dns-privacy-chip">On-demand · Read-only</span>
        </div>

        <form id="dnsDetectiveForm" class="dns-form">
          <div class="dns-primary-grid">
            <label>
              <span>DNS name or IP</span>
              <input id="dnsDetectiveName" autocomplete="off" spellcheck="false" placeholder="app.example.com" required>
            </label>
            <label>
              <span>Endpoint port <small>for handoff</small></span>
              <input id="dnsDetectivePort" inputmode="numeric" autocomplete="off" value="443" placeholder="443">
            </label>
          </div>

          <div class="dns-resolver-heading">
            <div>
              <strong>Compare resolver views</strong>
              <p>The system resolver is always included. Add up to four DNS server IPs; custom resolvers are restricted to port 53.</p>
            </div>
            <button class="secondary-button dns-add-resolver" id="dnsAddResolver" type="button">+ Add resolver</button>
          </div>

          <div id="dnsResolverRows" class="dns-resolver-rows"></div>

          <div class="dns-form-actions">
            <button class="primary-button" id="dnsDetectiveButton" type="submit">Compare DNS views</button>
            <span class="input-hint">Examples: <code>pihole=192.168.2.5</code> · <code>public=1.1.1.1</code></span>
          </div>
        </form>
      </section>

      <section class="panel diagnosis-result hidden dns-result" id="dnsDetectiveResult" aria-live="polite">
        <div class="diagnosis-heading">
          <div>
            <p class="eyebrow">RESOLVER VERDICT</p>
            <h2 id="dnsDetectiveConclusion">DNS comparison complete</h2>
            <p class="muted" id="dnsDetectiveTarget"></p>
            <p class="section-copy hidden" id="dnsSplitHint"></p>
          </div>
          <span class="confidence-badge dns-verdict" id="dnsDetectiveStatus"></span>
        </div>

        <div id="dnsDifferences" class="dns-differences hidden"></div>
        <div id="dnsResolverCards" class="dns-resolver-cards"></div>

        <section class="dns-runtime-card">
          <div>
            <p class="subsection-label">Runtime resolver context</p>
            <p id="dnsRuntimeNote" class="context-note"></p>
          </div>
          <div id="dnsRuntimeFacts" class="dns-runtime-facts"></div>
        </section>

        <div class="dns-handoff-actions">
          <button class="secondary-button" id="dnsToDiagnose" type="button">Diagnose endpoint</button>
          <button class="secondary-button" id="dnsToContract" type="button">Build expectation</button>
        </div>
      </section>

      <section class="panel diagnosis-result hidden" id="dnsDetectiveError" aria-live="polite">
        <p class="eyebrow">DNS DETECTIVE ERROR</p>
        <h2>HostSleuth could not compare those resolver views.</h2>
        <p class="section-copy" id="dnsDetectiveErrorText"></p>
      </section>
    `;
    main.append(view);

    restoreDNSResolvers();
    if (!document.querySelector(".dns-resolver-row")) addDNSResolverRow("");

    byId("dnsAddResolver").addEventListener("click", () => addDNSResolverRow(""));
    byId("dnsDetectiveForm").addEventListener("submit", runDNSDetective);
    byId("dnsToDiagnose").addEventListener("click", () => dnsHandoff("diagnose"));
    byId("dnsToContract").addEventListener("click", () => dnsHandoff("contracts"));

    if (window.location.hash === "#dns") showView("dns");
  }

  function addDNSResolverRow(value) {
    const rows = byId("dnsResolverRows");
    if (!rows || rows.children.length >= 4) return;

    const row = document.createElement("div");
    row.className = "dns-resolver-row";

    const input = document.createElement("input");
    input.className = "dns-resolver-input";
    input.autocomplete = "off";
    input.spellcheck = false;
    input.placeholder = "label=192.168.2.5";
    input.value = value || "";

    const remove = document.createElement("button");
    remove.type = "button";
    remove.className = "dns-resolver-remove";
    remove.setAttribute("aria-label", "Remove resolver");
    remove.textContent = "×";
    remove.addEventListener("click", () => {
      row.remove();
      persistDNSResolvers();
      updateDNSAddButton();
    });

    input.addEventListener("change", persistDNSResolvers);
    row.append(input, remove);
    rows.append(row);
    updateDNSAddButton();
  }

  function updateDNSAddButton() {
    const button = byId("dnsAddResolver");
    const rows = byId("dnsResolverRows");
    if (button && rows) button.disabled = rows.children.length >= 4;
  }

  function currentDNSResolvers() {
    return [...document.querySelectorAll(".dns-resolver-input")]
      .map((input) => input.value.trim())
      .filter(Boolean);
  }

  function persistDNSResolvers() {
    try {
      localStorage.setItem(DNS_RESOLVER_STORAGE, JSON.stringify(currentDNSResolvers()));
    } catch (_) {
      // Browser-local convenience must never block DNS inspection.
    }
  }

  function restoreDNSResolvers() {
    try {
      const values = JSON.parse(localStorage.getItem(DNS_RESOLVER_STORAGE) || "[]");
      if (Array.isArray(values)) values.slice(0, 4).forEach((value) => addDNSResolverRow(String(value)));
    } catch (_) {
      // Ignore malformed or unavailable browser-local state.
    }
  }

  async function runDNSDetective(event) {
    event.preventDefault();
    const button = byId("dnsDetectiveButton");
    const result = byId("dnsDetectiveResult");
    const error = byId("dnsDetectiveError");
    result.classList.add("hidden");
    error.classList.add("hidden");
    button.disabled = true;
    button.textContent = "Comparing…";

    try {
      const name = byId("dnsDetectiveName").value.trim();
      const params = new URLSearchParams({ name });
      currentDNSResolvers().forEach((resolver) => params.append("resolver", resolver));
      persistDNSResolvers();

      const response = await fetch(`/api/dns-detective?${params.toString()}`, { cache: "no-store" });
      if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
      renderDNSDetective(await response.json());
    } catch (err) {
      text(byId("dnsDetectiveErrorText"), err instanceof Error ? err.message : String(err));
      error.classList.remove("hidden");
    } finally {
      button.disabled = false;
      button.textContent = "Compare DNS views";
    }
  }

  function renderDNSDetective(data) {
    text(byId("dnsDetectiveConclusion"), data.conclusion || "DNS comparison complete");
    text(byId("dnsDetectiveTarget"), data.name || "");

    const status = byId("dnsDetectiveStatus");
    const value = data.status || "partial";
    status.textContent = value.toUpperCase();
    status.className = `confidence-badge dns-verdict ${value}`;

    const hint = byId("dnsSplitHint");
    if (data.split_view_hint) {
      hint.textContent = "Private/local and global answers differ. That pattern is consistent with split-view DNS or resolver-specific overrides; HostSleuth is not claiming intent without configuration evidence.";
      hint.classList.remove("hidden");
    } else {
      hint.classList.add("hidden");
    }

    renderDNSDifferences(data.differences || []);
    renderDNSResolverCards(data.resolvers || []);
    renderDNSRuntime(data.runtime || {});
    byId("dnsDetectiveResult").classList.remove("hidden");
  }

  function renderDNSDifferences(differences) {
    const box = byId("dnsDifferences");
    box.replaceChildren();
    if (!differences.length) {
      box.classList.add("hidden");
      return;
    }
    box.classList.remove("hidden");
    const label = document.createElement("p");
    label.className = "subsection-label";
    label.textContent = "Where resolver views differ";
    box.append(label);

    differences.forEach((difference) => {
      const item = document.createElement("div");
      item.className = "dns-difference";
      const title = document.createElement("strong");
      title.textContent = `${difference.resolver || "resolver"} · ${difference.kind || "answer"}`;
      const values = document.createElement("p");
      values.textContent = `Baseline: ${difference.baseline || "—"} · Observed: ${difference.observed || "—"}`;
      item.append(title, values);
      box.append(item);
    });
  }

  function renderDNSResolverCards(resolvers) {
    const container = byId("dnsResolverCards");
    container.replaceChildren();

    resolvers.forEach((resolver) => {
      const card = document.createElement("article");
      card.className = `dns-resolver-card ${resolver.status || "fail"}`;

      const head = document.createElement("div");
      head.className = "dns-resolver-card-head";
      const title = document.createElement("div");
      const name = document.createElement("strong");
      name.textContent = resolver.label || "resolver";
      const server = document.createElement("span");
      server.textContent = resolver.server || "";
      title.append(name, server);
      const timing = document.createElement("span");
      timing.className = "dns-timing";
      timing.textContent = Number.isFinite(resolver.duration_ms) ? `${resolver.duration_ms} ms` : "—";
      head.append(title, timing);
      card.append(head);

      if (resolver.error) {
        const error = document.createElement("p");
        error.className = "dns-resolver-error";
        error.textContent = resolver.error;
        card.append(error);
      }

      const answers = document.createElement("div");
      answers.className = "dns-answer-list";
      (resolver.addresses || []).forEach((address) => {
        const row = document.createElement("div");
        row.className = "dns-answer";
        const value = document.createElement("code");
        value.textContent = address.address || "";
        const badges = document.createElement("span");
        badges.className = "dns-answer-badges";
        badges.append(dnsBadge(address.family || "IP"), dnsBadge(address.scope || "other"));
        row.append(value, badges);
        answers.append(row);
      });

      (resolver.ptr || []).forEach((ptr) => {
        const row = document.createElement("div");
        row.className = "dns-answer";
        const value = document.createElement("code");
        value.textContent = ptr;
        row.append(value, dnsBadge("PTR"));
        answers.append(row);
      });

      if (resolver.cname) {
        const row = document.createElement("div");
        row.className = "dns-answer cname";
        const value = document.createElement("code");
        value.textContent = resolver.cname;
        row.append(value, dnsBadge("CNAME"));
        answers.append(row);
      }

      if (!answers.children.length && !resolver.error) {
        answers.append(makeEmpty("Resolver returned no displayable answer."));
      }
      card.append(answers);
      container.append(card);
    });
  }

  function dnsBadge(label) {
    const badge = document.createElement("span");
    badge.className = `dns-answer-badge ${String(label).toLowerCase().replace(/[^a-z0-9-]/g, "-")}`;
    badge.textContent = label;
    return badge;
  }

  function renderDNSRuntime(runtime) {
    text(byId("dnsRuntimeNote"), runtime.note || "");
    const facts = byId("dnsRuntimeFacts");
    facts.replaceChildren();

    [
      ["Nameservers", runtime.nameservers || []],
      ["Search", runtime.search || []],
      ["Options", runtime.options || []],
    ].forEach(([label, values]) => {
      const item = document.createElement("div");
      item.className = "dns-runtime-fact";
      const key = document.createElement("span");
      key.textContent = label;
      const value = document.createElement("strong");
      value.textContent = values.length ? values.join(", ") : "—";
      item.append(key, value);
      facts.append(item);
    });
  }

  function dnsHandoff(view) {
    const name = byId("dnsDetectiveName").value.trim();
    const port = byId("dnsDetectivePort").value.trim() || "443";
    if (!name) return;
    const endpoint = dnsEndpoint(name, port);
    if (window.hostSleuthRememberTarget) window.hostSleuthRememberTarget(endpoint);

    if (view === "diagnose") {
      byId("diagnoseTarget").value = endpoint;
      showView("diagnose");
      runDiagnosis(endpoint);
      return;
    }

    const contract = byId("contractTarget");
    if (contract) contract.value = endpoint;
    showView("contracts");
    if (contract) contract.focus();
  }

  function dnsEndpoint(name, port) {
    if (name.startsWith("[") && name.includes("]")) return `${name}:${port}`;
    if (name.includes(":") && !name.includes(".")) return `[${name}]:${port}`;
    return `${name}:${port}`;
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", dnsInit);
  } else {
    dnsInit();
  }
})();
