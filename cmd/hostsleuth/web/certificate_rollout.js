(() => {
  let lastCertificateRollout = null;

  function certificateRolloutInit() {
    if (document.querySelector('[data-view="cert-rollout"]')) return;
    const nav = document.querySelector(".nav-tabs");
    const main = document.querySelector("main");
    if (!nav || !main) return;

    const navButton = document.createElement("button");
    navButton.className = "nav-tab";
    navButton.dataset.view = "cert-rollout";
    navButton.textContent = "Cert Rollout";
    navButton.addEventListener("click", () => showView("cert-rollout"));
    nav.append(navButton);

    const view = document.createElement("section");
    view.className = "view";
    view.id = "view-cert-rollout";
    view.dataset.viewPanel = "cert-rollout";
    view.innerHTML = `
      <section class="panel section-panel cert-rollout-hero">
        <div class="section-heading">
          <div>
            <p class="eyebrow">CERTIFICATE ROLLOUT VERIFICATION</p>
            <h1>Which endpoint is still serving the old certificate?</h1>
            <p class="section-copy">Choose one expected SHA-256 fingerprint or one reference TLS endpoint, then compare up to 16 explicit direct-TLS endpoints. Fingerprint rollout status stays separate from validity, hostname, and trust health.</p>
          </div>
          <span class="cert-rollout-readonly-chip">Read-only · No renewal · No private keys</span>
        </div>

        <form id="certificateRolloutForm" class="cert-rollout-form">
          <label>
            <span>Expected source</span>
            <select id="certificateRolloutSource">
              <option value="fingerprint">SHA-256 fingerprint</option>
              <option value="reference">Reference TLS endpoint</option>
            </select>
          </label>
          <label class="cert-rollout-source-value">
            <span id="certificateRolloutSourceLabel">Expected SHA-256 fingerprint</span>
            <input id="certificateRolloutExpected" autocomplete="off" spellcheck="false" placeholder="64 hex characters; colons allowed" required>
          </label>
          <label class="cert-rollout-endpoints">
            <span>Endpoints · one host:port per line · maximum 16</span>
            <textarea id="certificateRolloutEndpoints" rows="6" spellcheck="false" placeholder="edge-a.example.com:443&#10;edge-b.example.com:443&#10;edge-c.example.com:443" required></textarea>
          </label>
          <div class="cert-rollout-form-actions">
            <button class="primary-button" id="certificateRolloutButton" type="submit">Verify rollout</button>
            <span class="input-hint">Direct TLS only. HostSleuth does not open certificate files, renew certificates, reload services, or touch ACME/private keys.</span>
          </div>
        </form>
      </section>

      <section class="panel diagnosis-result hidden cert-rollout-result" id="certificateRolloutResult" aria-live="polite">
        <div class="cert-rollout-verdict-row">
          <div>
            <p class="eyebrow">ROLLOUT VERDICT</p>
            <h2 id="certificateRolloutConclusion">Certificate rollout evaluated</h2>
            <p class="cert-rollout-first-problem" id="certificateRolloutFirstProblem"></p>
          </div>
          <span class="confidence-badge cert-rollout-verdict" id="certificateRolloutStatus"></span>
        </div>

        <div class="cert-rollout-summary" id="certificateRolloutSummary"></div>

        <section class="cert-rollout-expected-card">
          <div class="cert-rollout-card-heading">
            <div>
              <p class="subsection-label">EXPECTED CERTIFICATE</p>
              <h3 id="certificateRolloutExpectedTitle">Expected fingerprint</h3>
            </div>
            <span id="certificateRolloutExpectedStatus"></span>
          </div>
          <div class="cert-rollout-expected-grid" id="certificateRolloutExpectedGrid"></div>
        </section>

        <div class="cert-rollout-matrix-head">
          <div>
            <p class="subsection-label">ENDPOINT MATRIX</p>
            <h3>Rollout match and certificate health</h3>
          </div>
        </div>
        <div class="cert-rollout-table-wrap">
          <table class="cert-rollout-table">
            <thead>
              <tr>
                <th>Endpoint</th>
                <th>Rollout</th>
                <th>Health</th>
                <th>Subject</th>
                <th>Expires</th>
                <th>Hostname</th>
                <th>Trust</th>
                <th>SHA-256 fingerprint</th>
              </tr>
            </thead>
            <tbody id="certificateRolloutRows"></tbody>
          </table>
        </div>

        <details class="cert-rollout-scope">
          <summary>Safety and comparison boundary</summary>
          <ul id="certificateRolloutScopeNotes"></ul>
        </details>
      </section>

      <section class="panel diagnosis-result hidden" id="certificateRolloutError" aria-live="polite">
        <p class="eyebrow">CERTIFICATE ROLLOUT ERROR</p>
        <h2>HostSleuth could not evaluate that rollout request.</h2>
        <p class="section-copy" id="certificateRolloutErrorText"></p>
      </section>
    `;
    main.append(view);

    byId("certificateRolloutSource").addEventListener("change", updateCertificateRolloutSourceUI);
    byId("certificateRolloutForm").addEventListener("submit", runCertificateRollout);
    updateCertificateRolloutSourceUI();
    if (window.location.hash === "#cert-rollout") showView("cert-rollout");
  }

  function updateCertificateRolloutSourceUI() {
    const source = byId("certificateRolloutSource")?.value || "fingerprint";
    const label = byId("certificateRolloutSourceLabel");
    const input = byId("certificateRolloutExpected");
    if (!label || !input) return;
    if (source === "reference") {
      label.textContent = "Reference direct-TLS endpoint";
      input.placeholder = "reference.example.com:443";
    } else {
      label.textContent = "Expected SHA-256 fingerprint";
      input.placeholder = "64 hex characters; colons allowed";
    }
  }

  async function runCertificateRollout(event) {
    event.preventDefault();
    const button = byId("certificateRolloutButton");
    const result = byId("certificateRolloutResult");
    const error = byId("certificateRolloutError");
    result.classList.add("hidden");
    error.classList.add("hidden");
    button.disabled = true;
    button.textContent = "Verifying…";

    try {
      const source = byId("certificateRolloutSource").value;
      const expected = byId("certificateRolloutExpected").value.trim();
      const endpoints = byId("certificateRolloutEndpoints").value
        .split(/\r?\n/)
        .map((value) => value.trim())
        .filter(Boolean);

      if (!endpoints.length) throw new Error("Enter at least one endpoint.");
      if (endpoints.length > 16) throw new Error("At most 16 endpoints are allowed.");

      const params = new URLSearchParams();
      params.set(source === "reference" ? "reference" : "fingerprint", expected);
      endpoints.forEach((endpoint) => params.append("endpoint", endpoint));

      const response = await fetch(`/api/certificate-rollout?${params.toString()}`, { cache: "no-store" });
      if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
      lastCertificateRollout = await response.json();
      renderCertificateRollout(lastCertificateRollout);
    } catch (err) {
      text(byId("certificateRolloutErrorText"), err instanceof Error ? err.message : String(err));
      error.classList.remove("hidden");
    } finally {
      button.disabled = false;
      button.textContent = "Verify rollout";
    }
  }

  function renderCertificateRollout(story) {
    const status = story.status || "unknown";
    text(byId("certificateRolloutConclusion"), story.conclusion || "Certificate rollout evaluated");
    text(
      byId("certificateRolloutFirstProblem"),
      story.first_problem ? `First problem: ${story.first_problem}` : "Every endpoint matched the expected certificate and passed certificate health checks."
    );

    const badge = byId("certificateRolloutStatus");
    badge.textContent = status.toUpperCase();
    badge.className = `confidence-badge cert-rollout-verdict ${status}`;

    renderCertificateRolloutSummary(story.summary || {});
    renderCertificateRolloutExpected(story.expected || {});
    renderCertificateRolloutRows(story.endpoints || [], story.first_problem || "");
    renderCertificateRolloutScope(story.scope_notes || []);
    byId("certificateRolloutResult").classList.remove("hidden");
  }

  function renderCertificateRolloutSummary(summary) {
    const box = byId("certificateRolloutSummary");
    box.replaceChildren();
    [
      ["Endpoints", summary.total ?? 0],
      ["Matched", summary.matched ?? 0],
      ["Mismatched", summary.mismatched ?? 0],
      ["Unknown", summary.unknown ?? 0],
      ["Healthy", summary.healthy ?? 0],
      ["Unhealthy", summary.unhealthy ?? 0],
    ].forEach(([label, value]) => {
      const card = document.createElement("article");
      const key = document.createElement("span");
      key.textContent = label;
      const val = document.createElement("strong");
      val.textContent = String(value);
      card.append(key, val);
      box.append(card);
    });
  }

  function renderCertificateRolloutExpected(expected) {
    text(
      byId("certificateRolloutExpectedTitle"),
      expected.source === "reference" ? `Reference: ${expected.value || "—"}` : "Expected SHA-256 fingerprint"
    );
    const status = byId("certificateRolloutExpectedStatus");
    status.textContent = String(expected.status || "unknown").toUpperCase();
    status.className = `cert-rollout-source-status ${expected.status || "unknown"}`;

    const grid = byId("certificateRolloutExpectedGrid");
    grid.replaceChildren();
    [
      ["Source", expected.source || "—"],
      ["Fingerprint", expected.fingerprint || "—"],
      ["Subject", expected.certificate?.subject || "—"],
      ["Valid until", expected.certificate?.valid_until || "—"],
      ["Hostname", expected.tls?.hostname_status || "—"],
      ["Trust", expected.tls?.trust_status || "—"],
    ].forEach(([label, value]) => {
      const item = document.createElement("article");
      const key = document.createElement("span");
      key.textContent = label;
      const val = document.createElement("code");
      val.textContent = String(value);
      item.append(key, val);
      grid.append(item);
    });

    if (expected.problem) {
      const problem = document.createElement("p");
      problem.className = "cert-rollout-source-problem";
      problem.textContent = expected.problem;
      grid.append(problem);
    }
  }

  function renderCertificateRolloutRows(rows, firstProblem) {
    const body = byId("certificateRolloutRows");
    body.replaceChildren();

    rows.forEach((row) => {
      const tr = document.createElement("tr");
      if (row.target === firstProblem) tr.classList.add("first-problem");

      const cert = row.certificate || {};
      const tls = row.tls || {};
      const values = [
        row.target || "—",
        String(row.match_status || "unknown").toUpperCase(),
        String(row.status || "unknown").toUpperCase(),
        cert.subject || "—",
        cert.valid_until || "—",
        tls.hostname_status || "—",
        tls.trust_status || "—",
        cert.sha256_fingerprint || "—",
      ];

      values.forEach((value, index) => {
        const td = document.createElement("td");
        if (index === 0 || index === 3 || index === 7) {
          const code = document.createElement("code");
          code.textContent = value;
          td.append(code);
        } else {
          td.textContent = value;
        }
        if (index === 1) td.className = `cert-rollout-match ${row.match_status || "unknown"}`;
        if (index === 2) td.className = `cert-rollout-health ${row.status || "unknown"}`;
        tr.append(td);
      });

      if (row.problem) {
        const problemRow = document.createElement("tr");
        problemRow.className = "cert-rollout-problem-row";
        const td = document.createElement("td");
        td.colSpan = 8;
        td.textContent = row.problem;
        problemRow.append(td);
        body.append(tr, problemRow);
      } else {
        body.append(tr);
      }
    });

    if (!rows.length) {
      const tr = document.createElement("tr");
      const td = document.createElement("td");
      td.colSpan = 8;
      td.className = "cert-rollout-empty";
      td.textContent = "No endpoint rows were returned.";
      tr.append(td);
      body.append(tr);
    }
  }

  function renderCertificateRolloutScope(notes) {
    const list = byId("certificateRolloutScopeNotes");
    list.replaceChildren();
    notes.forEach((note) => {
      const item = document.createElement("li");
      item.textContent = note;
      list.append(item);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", certificateRolloutInit);
  } else {
    certificateRolloutInit();
  }
})();
