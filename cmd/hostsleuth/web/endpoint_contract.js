(() => {
  function contractInit() {
    if (document.querySelector('[data-view="contracts"]')) return;

    const nav = document.querySelector(".nav-tabs");
    const main = document.querySelector("main");
    if (!nav || !main) return;

    const navButton = document.createElement("button");
    navButton.className = "nav-tab";
    navButton.dataset.view = "contracts";
    navButton.textContent = "Expectations";
    navButton.addEventListener("click", () => showView("contracts"));
    nav.append(navButton);

    const view = document.createElement("section");
    view.className = "view";
    view.id = "view-contracts";
    view.dataset.viewPanel = "contracts";
    view.innerHTML = `
      <section class="panel section-panel contract-intro">
        <p class="eyebrow">EXPECTED ENDPOINT CONTRACT</p>
        <h1>What should be true?</h1>
        <p class="section-copy">Describe the endpoint you expect. HostSleuth compares that contract with current deterministic evidence and identifies the first mismatch. This is an on-demand check, not continuous monitoring.</p>

        <form id="contractForm" class="contract-form">
          <div class="contract-grid">
            <label>
              <span>Target</span>
              <input id="contractTarget" autocomplete="off" spellcheck="false" placeholder="example.com:443" required>
            </label>
            <label>
              <span>Contract name <small>optional</small></span>
              <input id="contractName" autocomplete="off" placeholder="photos">
            </label>
            <label>
              <span>Exact DNS addresses <small>optional, comma-separated</small></span>
              <input id="contractIPs" autocomplete="off" spellcheck="false" placeholder="192.0.2.10, 2001:db8::10">
            </label>
            <label>
              <span>TLS expectation</span>
              <select id="contractTLS">
                <option value="ignore">Do not require TLS</option>
                <option value="present">TLS handshake must succeed</option>
                <option value="verified">TLS + hostname + trust must verify</option>
                <option value="forbidden">TLS handshake must not succeed</option>
              </select>
            </label>
            <label>
              <span>Systemd service <small>optional, native mode</small></span>
              <input id="contractService" autocomplete="off" spellcheck="false" placeholder="nginx.service">
            </label>
            <label>
              <span>Container <small>optional</small></span>
              <input id="contractContainer" autocomplete="off" spellcheck="false" placeholder="photos">
            </label>
          </div>
          <div class="contract-actions">
            <button class="primary-button" id="contractButton" type="submit">Check expectations</button>
            <span class="input-hint">TCP reachability is always required. Expected DNS addresses use exact-set matching.</span>
          </div>
        </form>
      </section>

      <section class="panel diagnosis-result hidden" id="contractResult" aria-live="polite">
        <div class="diagnosis-heading">
          <div>
            <p class="eyebrow">EXPECTED VS OBSERVED</p>
            <h2 id="contractConclusion">Contract evaluated</h2>
            <p class="muted" id="contractResultTarget"></p>
            <p class="section-copy" id="contractFirstMismatch"></p>
          </div>
          <span class="confidence-badge" id="contractStatus"></span>
        </div>
        <div class="contract-checks" id="contractChecks"></div>
        <div class="contract-result-actions">
          <button class="secondary-button" id="contractDiagnose" type="button">Open full diagnosis</button>
        </div>
      </section>

      <section class="panel diagnosis-result hidden" id="contractError" aria-live="polite">
        <p class="eyebrow">CONTRACT ERROR</p>
        <h2>HostSleuth could not evaluate that contract.</h2>
        <p class="section-copy" id="contractErrorText"></p>
      </section>
    `;
    main.append(view);

    const form = byId("contractForm");
    const result = byId("contractResult");
    const error = byId("contractError");
    const submit = byId("contractButton");

    form.addEventListener("submit", async (event) => {
      event.preventDefault();
      result.classList.add("hidden");
      error.classList.add("hidden");
      submit.disabled = true;
      submit.textContent = "Checking…";

      try {
        const params = new URLSearchParams();
        const target = byId("contractTarget").value.trim();
        params.set("target", target);

        const name = byId("contractName").value.trim();
        if (name) params.set("name", name);

        const tls = byId("contractTLS").value;
        if (tls) params.set("tls", tls);

        const service = byId("contractService").value.trim();
        if (service) params.set("service", service);

        const container = byId("contractContainer").value.trim();
        if (container) params.set("container", container);

        byId("contractIPs").value
          .split(",")
          .map((value) => value.trim())
          .filter(Boolean)
          .forEach((value) => params.append("ip", value));

        const response = await fetch(`/api/contract?${params.toString()}`, { cache: "no-store" });
        if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
        renderContractEvaluation(await response.json());
      } catch (err) {
        byId("contractErrorText").textContent = err instanceof Error ? err.message : String(err);
        error.classList.remove("hidden");
      } finally {
        submit.disabled = false;
        submit.textContent = "Check expectations";
      }
    });

    byId("contractDiagnose").addEventListener("click", () => {
      const target = byId("contractTarget").value.trim();
      if (!target) return;
      byId("diagnoseTarget").value = target;
      showView("diagnose");
      byId("diagnoseTarget").focus();
    });

    if (window.location.hash === "#contracts") showView("contracts");
  }

  function renderContractEvaluation(evaluation) {
    const result = byId("contractResult");
    const error = byId("contractError");
    error.classList.add("hidden");

    text(byId("contractConclusion"), evaluation.conclusion || "Contract evaluated");
    text(byId("contractResultTarget"), evaluation.contract?.target || "");
    text(
      byId("contractFirstMismatch"),
      evaluation.first_mismatch
        ? `First unresolved expectation: ${evaluation.first_mismatch}`
        : "Every requested expectation is currently satisfied."
    );

    const status = byId("contractStatus");
    status.textContent = String(evaluation.status || "unknown").toUpperCase();
    status.className = `confidence-badge contract-status ${evaluation.status || "unknown"}`;

    const checks = byId("contractChecks");
    checks.replaceChildren();
    (evaluation.checks || []).forEach((check) => {
      const row = document.createElement("article");
      row.className = `contract-check ${check.status || "unknown"}`;

      const head = document.createElement("div");
      head.className = "contract-check-head";

      const name = document.createElement("strong");
      name.textContent = check.name || "expectation";

      const badge = document.createElement("span");
      badge.className = `contract-check-status ${check.status || "unknown"}`;
      badge.textContent = String(check.status || "unknown").toUpperCase();
      head.append(name, badge);

      const expected = document.createElement("p");
      expected.className = "contract-expected";
      expected.textContent = `Expected: ${check.expectation || "—"}`;

      const observed = document.createElement("p");
      observed.className = "contract-observed";
      observed.textContent = `Observed: ${check.observed || "—"}`;

      row.append(head, expected, observed);
      checks.append(row);
    });

    if (!(evaluation.checks || []).length) {
      checks.append(makeEmpty("No contract checks were produced."));
    }
    result.classList.remove("hidden");
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", contractInit);
  } else {
    contractInit();
  }
})();
