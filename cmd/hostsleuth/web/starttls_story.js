(() => {
  let lastStartTLSStory = null;

  function startTLSStoryInit() {
    if (document.querySelector('[data-view="starttls"]')) return;
    const nav = document.querySelector(".nav-tabs");
    const main = document.querySelector("main");
    if (!nav || !main) return;

    const navButton = document.createElement("button");
    navButton.className = "nav-tab";
    navButton.dataset.view = "starttls";
    navButton.textContent = "STARTTLS";
    navButton.addEventListener("click", () => showView("starttls"));
    nav.append(navButton);

    const view = document.createElement("section");
    view.className = "view";
    view.id = "view-starttls";
    view.dataset.viewPanel = "starttls";
    view.innerHTML = `
      <section class="panel section-panel starttls-hero">
        <div class="section-heading">
          <div>
            <p class="eyebrow">STARTTLS / MAIL SERVICE STORY</p>
            <h1>The mail port is open. Did STARTTLS actually negotiate?</h1>
            <p class="section-copy">HostSleuth reads only the greeting and capability metadata, sends the protocol's fixed TLS-upgrade command, and then inspects the negotiated TLS session and served certificate.</p>
          </div>
          <span class="starttls-readonly-chip">No credentials · No mail · Read-only</span>
        </div>

        <form id="startTLSStoryForm" class="starttls-form">
          <label>
            <span>Protocol</span>
            <select id="startTLSProtocol" required>
              <option value="smtp">SMTP STARTTLS</option>
              <option value="imap">IMAP STARTTLS</option>
              <option value="pop3">POP3 STLS</option>
            </select>
          </label>
          <label>
            <span>Mail endpoint</span>
            <input id="startTLSTarget" autocomplete="off" spellcheck="false" placeholder="mail.example.com:25" required>
          </label>
          <div class="starttls-form-actions">
            <button class="primary-button" id="startTLSStoryButton" type="submit">Trace STARTTLS</button>
            <span class="input-hint">Only greeting, capability discovery, STARTTLS/STLS, TLS handshake, and certificate metadata are touched.</span>
          </div>
        </form>
      </section>

      <section class="panel diagnosis-result hidden starttls-result" id="startTLSStoryResult" aria-live="polite">
        <div class="starttls-verdict-row">
          <div>
            <p class="eyebrow">MAIL TLS VERDICT</p>
            <h2 id="startTLSConclusion">STARTTLS evaluated</h2>
            <p class="muted" id="startTLSTargetSummary"></p>
            <p class="starttls-first-problem" id="startTLSFirstProblem"></p>
          </div>
          <span class="confidence-badge starttls-verdict" id="startTLSStatus"></span>
        </div>

        <div class="starttls-flow" id="startTLSFlow"></div>

        <div class="starttls-evidence-grid">
          <section class="starttls-evidence-card">
            <p class="subsection-label">PROTOCOL EVIDENCE</p>
            <dl id="startTLSProtocolEvidence"></dl>
          </section>
          <section class="starttls-evidence-card">
            <p class="subsection-label">TLS SESSION</p>
            <dl id="startTLSSessionEvidence"></dl>
          </section>
        </div>

        <section class="starttls-capability-card hidden" id="startTLSCapabilitiesCard">
          <div class="starttls-card-heading">
            <div>
              <p class="subsection-label">PRE-UPGRADE CAPABILITIES</p>
              <h3>What the server advertised before TLS</h3>
            </div>
            <span>Discarded after upgrade · shown as evidence only</span>
          </div>
          <div class="starttls-capabilities" id="startTLSCapabilities"></div>
        </section>

        <section class="starttls-certificate-card hidden" id="startTLSCertificateCard">
          <div class="starttls-card-heading">
            <div>
              <p class="subsection-label">SERVED CERTIFICATE</p>
              <h3 id="startTLSCertificateSubject">Certificate</h3>
            </div>
            <span id="startTLSCertificateExpiry"></span>
          </div>
          <div class="starttls-certificate-grid" id="startTLSCertificateGrid"></div>
        </section>

        <details class="starttls-scope">
          <summary>Safety and protocol boundary</summary>
          <ul id="startTLSScopeNotes"></ul>
        </details>
      </section>

      <section class="panel diagnosis-result hidden" id="startTLSStoryError" aria-live="polite">
        <p class="eyebrow">STARTTLS STORY ERROR</p>
        <h2>HostSleuth could not evaluate that mail endpoint.</h2>
        <p class="section-copy" id="startTLSStoryErrorText"></p>
      </section>
    `;
    main.append(view);

    byId("startTLSStoryForm").addEventListener("submit", runStartTLSStory);
    if (window.location.hash === "#starttls") showView("starttls");
  }

  async function runStartTLSStory(event) {
    event.preventDefault();
    const button = byId("startTLSStoryButton");
    const result = byId("startTLSStoryResult");
    const error = byId("startTLSStoryError");
    result.classList.add("hidden");
    error.classList.add("hidden");
    button.disabled = true;
    button.textContent = "Tracing…";

    try {
      const protocol = byId("startTLSProtocol").value;
      const target = byId("startTLSTarget").value.trim();
      const params = new URLSearchParams({ protocol, target });
      const response = await fetch(`/api/starttls-story?${params.toString()}`, { cache: "no-store" });
      if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
      lastStartTLSStory = await response.json();
      renderStartTLSStory(lastStartTLSStory);
      if (window.hostSleuthRememberTarget && target) window.hostSleuthRememberTarget(target);
    } catch (err) {
      text(byId("startTLSStoryErrorText"), err instanceof Error ? err.message : String(err));
      error.classList.remove("hidden");
    } finally {
      button.disabled = false;
      button.textContent = "Trace STARTTLS";
    }
  }

  function renderStartTLSStory(story) {
    const status = story.status || "unknown";
    text(byId("startTLSConclusion"), story.conclusion || "STARTTLS evaluated");
    text(byId("startTLSTargetSummary"), `${String(story.protocol || "").toUpperCase()} · ${story.target || "—"}`);
    text(
      byId("startTLSFirstProblem"),
      story.first_problem
        ? `First proven problem: ${startTLSStageTitle(story, story.first_problem)}`
        : "Every tested STARTTLS and certificate gate passed."
    );

    const badge = byId("startTLSStatus");
    badge.textContent = status.toUpperCase();
    badge.className = `confidence-badge starttls-verdict ${status}`;

    renderStartTLSFlow(story.stages || [], story.first_problem || "");
    renderStartTLSProtocolEvidence(story);
    renderStartTLSSessionEvidence(story.tls || null);
    renderStartTLSCapabilities(story.capabilities || []);
    renderStartTLSCertificate(story.tls?.certificate || null);
    renderStartTLSScope(story.scope_notes || []);
    byId("startTLSStoryResult").classList.remove("hidden");
  }

  function renderStartTLSFlow(stages, firstProblem) {
    const flow = byId("startTLSFlow");
    flow.replaceChildren();
    stages.forEach((stage, index) => {
      const wrap = document.createElement("div");
      wrap.className = "starttls-flow-item";

      const card = document.createElement("article");
      card.className = `starttls-stage ${stage.status || "unknown"} ${stage.id === firstProblem ? "first-problem" : ""}`;

      const head = document.createElement("div");
      head.className = "starttls-stage-head";
      const step = document.createElement("span");
      step.className = "starttls-step";
      step.textContent = String(index + 1).padStart(2, "0");
      const title = document.createElement("strong");
      title.textContent = stage.title || stage.id || "Stage";
      const state = document.createElement("span");
      state.className = `starttls-stage-badge ${stage.status || "unknown"}`;
      state.textContent = String(stage.status || "unknown").toUpperCase();
      head.append(step, title, state);

      const summary = document.createElement("p");
      summary.textContent = stage.summary || "No summary.";

      card.append(head, summary);

      if ((stage.evidence || []).length) {
        const details = document.createElement("details");
        if (stage.id === firstProblem || stage.status === "fail") details.open = true;
        const summaryEl = document.createElement("summary");
        summaryEl.textContent = `Evidence (${stage.evidence.length})`;
        const evidence = document.createElement("div");
        evidence.className = "starttls-stage-evidence";
        stage.evidence.forEach((item) => {
          const line = document.createElement("code");
          line.textContent = item;
          evidence.append(line);
        });
        details.append(summaryEl, evidence);
        card.append(details);
      }

      wrap.append(card);
      if (index < stages.length - 1) {
        const arrow = document.createElement("span");
        arrow.className = "starttls-arrow";
        arrow.textContent = "→";
        arrow.setAttribute("aria-hidden", "true");
        wrap.append(arrow);
      }
      flow.append(wrap);
    });
  }

  function renderStartTLSProtocolEvidence(story) {
    const list = byId("startTLSProtocolEvidence");
    list.replaceChildren();
    addStartTLSFact(list, "Greeting", story.greeting || "—");
    addStartTLSFact(list, "Upgrade advertised", story.starttls_advertised ? "yes" : "no");
    addStartTLSFact(list, "Upgrade command", story.upgrade_command || "—");
    addStartTLSFact(list, "Upgrade response", story.upgrade_response || "—");
  }

  function renderStartTLSSessionEvidence(tls) {
    const list = byId("startTLSSessionEvidence");
    list.replaceChildren();
    addStartTLSFact(list, "TLS status", tls?.handshake_status || "—");
    addStartTLSFact(list, "Protocol", tls?.protocol || "—");
    addStartTLSFact(list, "Cipher", tls?.cipher_suite || "—");
    addStartTLSFact(list, "Hostname", tls?.hostname_status || "—");
    addStartTLSFact(list, "Trust", tls?.trust_status || "—");
  }

  function addStartTLSFact(list, label, value) {
    const term = document.createElement("dt");
    term.textContent = label;
    const definition = document.createElement("dd");
    const code = document.createElement("code");
    code.textContent = String(value);
    definition.append(code);
    list.append(term, definition);
  }

  function renderStartTLSCapabilities(capabilities) {
    const card = byId("startTLSCapabilitiesCard");
    const box = byId("startTLSCapabilities");
    box.replaceChildren();
    if (!capabilities.length) {
      card.classList.add("hidden");
      return;
    }
    card.classList.remove("hidden");
    capabilities.forEach((capability) => {
      const chip = document.createElement("code");
      chip.textContent = capability;
      box.append(chip);
    });
  }

  function renderStartTLSCertificate(cert) {
    const card = byId("startTLSCertificateCard");
    const grid = byId("startTLSCertificateGrid");
    grid.replaceChildren();
    if (!cert) {
      card.classList.add("hidden");
      return;
    }
    card.classList.remove("hidden");
    text(byId("startTLSCertificateSubject"), cert.subject || "Certificate");
    text(
      byId("startTLSCertificateExpiry"),
      Number.isFinite(Number(cert.days_remaining)) ? `${cert.days_remaining} day(s) remaining` : ""
    );
    [
      ["Issuer", cert.issuer],
      ["Serial", cert.serial],
      ["Valid from", cert.valid_from],
      ["Valid until", cert.valid_until],
      ["SHA-256", cert.sha256_fingerprint],
      ["SANs", (cert.sans || []).join(", ") || "—"],
    ].forEach(([label, value]) => {
      const item = document.createElement("article");
      const key = document.createElement("span");
      key.textContent = label;
      const val = document.createElement("code");
      val.textContent = value || "—";
      item.append(key, val);
      grid.append(item);
    });
  }

  function renderStartTLSScope(notes) {
    const list = byId("startTLSScopeNotes");
    list.replaceChildren();
    notes.forEach((note) => {
      const item = document.createElement("li");
      item.textContent = note;
      list.append(item);
    });
  }

  function startTLSStageTitle(story, id) {
    const stage = (story.stages || []).find((item) => item.id === id);
    return stage?.title || id;
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", startTLSStoryInit);
  } else {
    startTLSStoryInit();
  }
})();
