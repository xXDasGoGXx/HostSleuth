(() => {
  let lastProxyStory = null;

  function proxyStoryInit() {
    if (document.querySelector('[data-view="proxy"]')) return;

    const nav = document.querySelector(".nav-tabs");
    const main = document.querySelector("main");
    if (!nav || !main) return;

    const navButton = document.createElement("button");
    navButton.className = "nav-tab";
    navButton.dataset.view = "proxy";
    navButton.textContent = "Proxy Path";
    navButton.addEventListener("click", () => showView("proxy"));
    nav.append(navButton);

    const view = document.createElement("section");
    view.className = "view";
    view.id = "view-proxy";
    view.dataset.viewPanel = "proxy";
    view.innerHTML = `
      <section class="panel section-panel proxy-hero">
        <div class="section-heading">
          <div>
            <p class="eyebrow">REVERSE PROXY / UPSTREAM STORY</p>
            <h1>Where does the request path actually break?</h1>
            <p class="section-copy">Give HostSleuth the public URL and the upstream you expect behind it. It checks the front door, local proxy evidence, and the upstream path separately—then compares native upstream behavior with the public Host/SNI when those identities differ.</p>
          </div>
          <span class="proxy-readonly-chip">HEAD only · No bodies · Read-only</span>
        </div>

        <form id="proxyStoryForm" class="proxy-form">
          <label>
            <span>Public URL</span>
            <input id="proxyPublicURL" autocomplete="off" spellcheck="false" placeholder="https://app.example.com/" required>
          </label>
          <div class="proxy-direction" aria-hidden="true"><span>front door</span><strong>→</strong><span>expected upstream</span></div>
          <label>
            <span>Expected upstream URL</span>
            <input id="proxyUpstreamURL" autocomplete="off" spellcheck="false" placeholder="http://192.0.2.40:8080/" required>
          </label>
          <div class="proxy-form-actions">
            <button class="primary-button" id="proxyStoryButton" type="submit">Trace request path</button>
            <span class="input-hint">No credentials, cookies, request bodies, custom headers, proxy config reads, or cross-host redirect following.</span>
          </div>
        </form>
      </section>

      <section class="panel diagnosis-result hidden proxy-result" id="proxyStoryResult" aria-live="polite">
        <div class="proxy-verdict-row">
          <div>
            <p class="eyebrow">REQUEST PATH VERDICT</p>
            <h2 id="proxyStoryConclusion">Request path evaluated</h2>
            <p class="muted" id="proxyStoryTargets"></p>
            <p class="proxy-first-problem" id="proxyStoryFirstProblem"></p>
          </div>
          <span class="confidence-badge proxy-verdict" id="proxyStoryStatus"></span>
        </div>

        <div class="proxy-group-labels" aria-hidden="true">
          <span>FRONT DOOR</span><span>PROXY CONTEXT</span><span>UPSTREAM</span>
        </div>
        <div class="proxy-flow" id="proxyStoryFlow"></div>

        <section class="proxy-identity-compare hidden" id="proxyIdentityCompare">
          <div class="proxy-compare-heading">
            <div>
              <p class="subsection-label">HOST / SNI COMPARISON</p>
              <h3>Does the upstream behave differently with the public identity?</h3>
            </div>
            <span class="proxy-identity-note">Same upstream address · Different request identity</span>
          </div>
          <div class="proxy-compare-grid" id="proxyCompareGrid"></div>
        </section>

        <div class="proxy-result-actions">
          <button class="secondary-button" id="proxyCopySummary" type="button">Copy evidence summary</button>
          <button class="secondary-button" id="proxyDiagnosePublic" type="button">Diagnose public endpoint</button>
          <button class="secondary-button" id="proxyDiagnoseUpstream" type="button">Diagnose upstream</button>
          <button class="secondary-button" id="proxyDNSPublic" type="button">DNS public host</button>
          <span class="proxy-copy-status" id="proxyCopyStatus" aria-live="polite"></span>
        </div>
      </section>

      <section class="panel diagnosis-result hidden" id="proxyStoryError" aria-live="polite">
        <p class="eyebrow">PROXY PATH ERROR</p>
        <h2>HostSleuth could not evaluate that request path.</h2>
        <p class="section-copy" id="proxyStoryErrorText"></p>
      </section>
    `;
    main.append(view);

    byId("proxyStoryForm").addEventListener("submit", runProxyStory);
    byId("proxyCopySummary").addEventListener("click", copyProxySummary);
    byId("proxyDiagnosePublic").addEventListener("click", () => proxyDiagnose("public_target"));
    byId("proxyDiagnoseUpstream").addEventListener("click", () => proxyDiagnose("upstream_target"));
    byId("proxyDNSPublic").addEventListener("click", proxyDNSPublic);

    if (window.location.hash === "#proxy") showView("proxy");
  }

  async function runProxyStory(event) {
    event.preventDefault();
    const button = byId("proxyStoryButton");
    const result = byId("proxyStoryResult");
    const error = byId("proxyStoryError");
    result.classList.add("hidden");
    error.classList.add("hidden");
    button.disabled = true;
    button.textContent = "Tracing…";

    try {
      const publicURL = byId("proxyPublicURL").value.trim();
      const upstreamURL = byId("proxyUpstreamURL").value.trim();
      const params = new URLSearchParams({ public: publicURL, upstream: upstreamURL });
      const response = await fetch(`/api/proxy-story?${params.toString()}`, { cache: "no-store" });
      if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
      lastProxyStory = await response.json();
      renderProxyStory(lastProxyStory);
      if (window.hostSleuthRememberTarget && lastProxyStory.public_target) {
        window.hostSleuthRememberTarget(lastProxyStory.public_target);
      }
    } catch (err) {
      text(byId("proxyStoryErrorText"), err instanceof Error ? err.message : String(err));
      error.classList.remove("hidden");
    } finally {
      button.disabled = false;
      button.textContent = "Trace request path";
    }
  }

  function renderProxyStory(story) {
    text(byId("proxyStoryConclusion"), story.conclusion || "Request path evaluated");
    text(byId("proxyStoryTargets"), `${story.public_url || "public"} → ${story.upstream_url || "upstream"}`);
    text(
      byId("proxyStoryFirstProblem"),
      story.first_problem
        ? `First item needing attention: ${proxyStageTitle(story, story.first_problem)}`
        : "Every tested request-path gate passed."
    );

    const status = byId("proxyStoryStatus");
    const value = story.status || "unknown";
    status.textContent = value.toUpperCase();
    status.className = `confidence-badge proxy-verdict ${value}`;

    renderProxyFlow(story.stages || [], story.first_problem || "");
    renderProxyIdentityCompare(story);
    byId("proxyStoryResult").classList.remove("hidden");
  }

  function renderProxyFlow(stages, firstProblem) {
    const flow = byId("proxyStoryFlow");
    flow.replaceChildren();

    stages.forEach((stage, index) => {
      const wrap = document.createElement("div");
      wrap.className = "proxy-flow-item";

      const card = document.createElement("article");
      card.className = `proxy-stage ${stage.status || "unknown"} ${stage.id === firstProblem ? "first-problem" : ""}`;

      const head = document.createElement("div");
      head.className = "proxy-stage-head";
      const step = document.createElement("span");
      step.className = "proxy-step";
      step.textContent = String(index + 1).padStart(2, "0");
      const title = document.createElement("strong");
      title.textContent = stage.title || stage.id || "Stage";
      const badge = document.createElement("span");
      badge.className = `proxy-stage-badge ${stage.status || "unknown"}`;
      badge.textContent = String(stage.status || "unknown").toUpperCase();
      head.append(step, title, badge);

      const summary = document.createElement("p");
      summary.className = "proxy-stage-summary";
      summary.textContent = stage.summary || "No summary.";

      card.append(head, summary);

      if ((stage.evidence || []).length) {
        const details = document.createElement("details");
        details.className = "proxy-stage-evidence";
        if (stage.id === firstProblem || stage.status === "fail") details.open = true;
        const summaryEl = document.createElement("summary");
        summaryEl.textContent = `Evidence (${stage.evidence.length})`;
        const list = document.createElement("div");
        list.className = "proxy-evidence-list";
        stage.evidence.forEach((item) => {
          const line = document.createElement("code");
          line.textContent = item;
          list.append(line);
        });
        details.append(summaryEl, list);
        card.append(details);
      }

      wrap.append(card);
      if (index < stages.length - 1) {
        const connector = document.createElement("span");
        connector.className = "proxy-connector";
        connector.textContent = "→";
        connector.setAttribute("aria-hidden", "true");
        wrap.append(connector);
      }
      flow.append(wrap);
    });
  }

  function renderProxyIdentityCompare(story) {
    const section = byId("proxyIdentityCompare");
    const grid = byId("proxyCompareGrid");
    grid.replaceChildren();

    const nativeProbe = story.upstream_native_http;
    const publicProbe = story.upstream_public_host_http;
    if (!publicProbe) {
      section.classList.add("hidden");
      return;
    }

    section.classList.remove("hidden");
    grid.append(
      proxyProbeCard("Native upstream identity", nativeProbe),
      proxyProbeCard("Public Host / SNI", publicProbe)
    );
  }

  function proxyProbeCard(label, probe) {
    const card = document.createElement("article");
    const classification = proxyProbeClass(probe);
    card.className = `proxy-compare-card ${classification}`;

    const head = document.createElement("div");
    head.className = "proxy-compare-card-head";
    const title = document.createElement("strong");
    title.textContent = label;
    const badge = document.createElement("span");
    badge.textContent = classification.toUpperCase();
    badge.className = `proxy-stage-badge ${classification}`;
    head.append(title, badge);

    const identity = document.createElement("div");
    identity.className = "proxy-identity-lines";
    [["Dial", probe?.dial_target], ["Host", probe?.host_header], ["SNI", probe?.tls_server_name || "—"], ["HTTP", probe?.http_status || probe?.error || "—"]].forEach(([key, value]) => {
      const row = document.createElement("div");
      const k = document.createElement("span");
      k.textContent = key;
      const v = document.createElement("code");
      v.textContent = value || "—";
      row.append(k, v);
      identity.append(row);
    });

    card.append(head, identity);
    return card;
  }

  function proxyProbeClass(probe) {
    if (!probe || probe.error || probe.status !== "pass") return "fail";
    const code = Number(probe.status_code || 0);
    if (code >= 500) return "fail";
    if (code >= 400) return "warn";
    if (code >= 100) return "pass";
    return "unknown";
  }

  function proxyStageTitle(story, id) {
    const stage = (story.stages || []).find((item) => item.id === id);
    return stage?.title || id;
  }

  function proxyDiagnose(field) {
    if (!lastProxyStory?.[field]) return;
    byId("diagnoseTarget").value = lastProxyStory[field];
    showView("diagnose");
    runDiagnosis(lastProxyStory[field]);
  }

  function proxyDNSPublic() {
    if (!lastProxyStory?.public_url) return;
    try {
      const host = new URL(lastProxyStory.public_url.replace("?[redacted]", "")).hostname;
      const input = byId("dnsDetectiveName");
      if (input) input.value = host;
      showView("dns");
      if (input) input.focus();
    } catch (_) {}
  }

  async function copyProxySummary() {
    if (!lastProxyStory) return;
    const lines = [
      "HostSleuth Reverse Proxy / Upstream Story",
      `Public: ${lastProxyStory.public_url || "—"}`,
      `Upstream: ${lastProxyStory.upstream_url || "—"}`,
      `Verdict: ${String(lastProxyStory.status || "unknown").toUpperCase()} — ${lastProxyStory.conclusion || ""}`,
      "",
    ];
    (lastProxyStory.stages || []).forEach((stage) => {
      lines.push(`[${String(stage.status || "unknown").toUpperCase()}] ${stage.title}: ${stage.summary || ""}`);
      (stage.evidence || []).forEach((evidence) => lines.push(`  - ${evidence}`));
    });
    const payload = lines.join("\n").slice(0, 12000);
    const status = byId("proxyCopyStatus");
    try {
      await navigator.clipboard.writeText(payload);
      text(status, "Copied.");
    } catch (_) {
      text(status, "Clipboard unavailable in this browser context.");
    }
    window.setTimeout(() => text(status, ""), 2500);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", proxyStoryInit);
  } else {
    proxyStoryInit();
  }
})();
