async function actionRequest(path, options = {}) {
  const response = await fetch(path, { cache: "no-store", ...options });
  const contentType = response.headers.get("content-type") || "";
  const payload = contentType.includes("application/json") ? await response.json() : await response.text();
  if (!response.ok) {
    const message = typeof payload === "string" ? payload.trim() : payload?.error || payload?.result?.summary || `request failed (${response.status})`;
    const error = new Error(message || `request failed (${response.status})`);
    error.payload = payload;
    throw error;
  }
  return payload;
}

function actionPairList(pairs) {
  const dl = document.createElement("dl");
  dl.className = "action-pairs";
  pairs.forEach(([label, value, code = false]) => {
    if (value === undefined || value === null || value === "") return;
    const dt = document.createElement("dt");
    dt.textContent = label;
    const dd = document.createElement("dd");
    dd.textContent = String(value);
    if (code) dd.className = "action-code";
    dl.append(dt, dd);
  });
  return dl;
}

function renderActionChecks(container, checks) {
  const list = document.createElement("div");
  list.className = "action-check-list";
  (checks || []).forEach((check) => {
    const row = document.createElement("div");
    row.className = `action-check ${check.status || "unknown"}`;
    const title = document.createElement("strong");
    title.textContent = check.name || "check";
    const detail = document.createElement("span");
    detail.textContent = check.evidence || "";
    row.append(title, detail);
    list.append(row);
  });
  container.append(list);
}

function renderActionPreview(container, preview) {
  container.replaceChildren();
  const headline = document.createElement("strong");
  headline.textContent = preview.available ? "Action is eligible for explicit confirmation." : "Action is not eligible to run.";
  container.append(headline);
  container.append(actionPairList([
    ["Action", preview.action_id, true],
    ["Target", preview.target, true],
    ["Effect", preview.effect],
    ["Command", (preview.command || []).join(" "), true],
    ["Confirmation", preview.confirmation, true],
    ["Before", preview.before ? `load=${preview.before.load_state || "unknown"} active=${preview.before.active_state || "unknown"} sub=${preview.before.sub_state || "unknown"}` : ""],
  ]));
  renderActionChecks(container, preview.checks || []);
}

function renderActionResult(container, result) {
  container.replaceChildren();
  const headline = document.createElement("strong");
  headline.textContent = `${String(result.status || "unknown").toUpperCase()}: ${result.summary || "Action finished"}`;
  headline.className = `action-result-${result.status || "unknown"}`;
  container.append(headline);
  container.append(actionPairList([
    ["Action", result.action_id, true],
    ["Target", result.target, true],
    ["Started", formatTime(result.started_at)],
    ["Finished", formatTime(result.finished_at)],
    ["Before", result.before ? `active=${result.before.active_state || "unknown"} sub=${result.before.sub_state || "unknown"}` : ""],
    ["After", result.after ? `active=${result.after.active_state || "unknown"} sub=${result.after.sub_state || "unknown"}` : ""],
    ["Bounded command output", result.command_output, true],
  ]));
}

function renderActionAudit(container, audits) {
  container.replaceChildren();
  if (!(audits || []).length) {
    container.append(makeEmpty("No action attempts have been recorded."));
    return;
  }
  [...audits].reverse().forEach((audit) => {
    const row = document.createElement("article");
    row.className = "action-audit-row";
    const title = document.createElement("strong");
    title.textContent = `${audit.action_id || "action"} · ${audit.target || "unknown target"}`;
    const detail = document.createElement("span");
    detail.textContent = `${audit.phase || "event"} · ${audit.status || "unknown"} · ${formatTime(audit.at)}`;
    const summary = document.createElement("p");
    summary.textContent = audit.summary || "";
    row.append(title, detail, summary);
    container.append(row);
  });
}

function renderActionCapabilities(container, capabilities) {
  container.replaceChildren();
  (capabilities || []).forEach((capability) => {
    const card = document.createElement("article");
    card.className = "action-capability-item";
    const title = document.createElement("strong");
    title.textContent = capability.title || capability.action_id || "Safe Action";
    card.append(title);
    card.append(actionPairList([
      ["Action", capability.action_id, true],
      ["State", capability.available ? "available" : capability.enabled ? "enabled but unavailable" : "disabled"],
      ["Reason", capability.reason],
      ["Allowed targets", (capability.allowed_targets || []).join(", "), true],
    ]));
    container.append(card);
  });
}

async function installActionsUI() {
  if (byId("view-actions")) return;
  const nav = document.querySelector(".nav-tabs");
  const main = document.querySelector("main");
  if (!nav || !main) return;

  const tab = document.createElement("button");
  tab.className = "nav-tab";
  tab.dataset.view = "actions";
  tab.textContent = "Actions";
  tab.addEventListener("click", () => showView("actions"));
  nav.append(tab);

  const section = document.createElement("section");
  section.className = "view";
  section.id = "view-actions";
  section.dataset.viewPanel = "actions";
  section.innerHTML = `
    <section class="panel section-panel action-intro">
      <p class="eyebrow">OPTIONAL SAFE ACTIONS</p>
      <h1>Small, allowlisted remedies with evidence before and after.</h1>
      <p class="section-copy">Actions are disabled by default. This UI is loopback-only, accepts no arbitrary command, and requires a matching preview confirmation before a state change. Restart and reload permissions are independently allowlisted.</p>
      <div id="actionCapability" class="action-capability"></div>
    </section>
    <div class="action-grid">
      <section class="panel section-panel action-card">
        <p class="eyebrow">SAFE ACTION</p>
        <h2 id="actionHeading">Select an explicitly allowlisted systemd action.</h2>
        <p class="section-copy" id="actionDescription">HostSleuth will preview the exact target and command before any state change.</p>
        <form id="actionPreviewForm" class="action-form">
          <label>Action<select id="actionID" required></select></label>
          <label>Allowlisted service<select id="actionTarget" required></select></label>
          <button id="actionPreviewButton" class="primary-button" type="submit">Preview action</button>
        </form>
        <div id="actionPreviewResult" class="action-result hidden"></div>
        <div id="actionConfirmPanel" class="action-confirm hidden">
          <label>Type the exact confirmation shown in the preview<input id="actionConfirmation" autocomplete="off" spellcheck="false"></label>
          <button id="actionRunButton" class="primary-button" type="button" disabled>Run confirmed action</button>
        </div>
        <div id="actionRunResult" class="action-result hidden"></div>
      </section>
      <section class="panel section-panel action-card">
        <p class="eyebrow">AUDIT</p>
        <h2>Recent action attempts.</h2>
        <p class="section-copy">Requested, denied, failed, and successful action phases are written to a dedicated JSONL audit log.</p>
        <div id="actionAudit" class="action-audit"></div>
      </section>
    </div>`;
  main.append(section);

  const capabilityBox = byId("actionCapability");
  const actionSelect = byId("actionID");
  const targetSelect = byId("actionTarget");
  const previewButton = byId("actionPreviewButton");
  const previewResult = byId("actionPreviewResult");
  const confirmPanel = byId("actionConfirmPanel");
  const confirmationInput = byId("actionConfirmation");
  const runButton = byId("actionRunButton");
  const runResult = byId("actionRunResult");
  const auditBox = byId("actionAudit");
  let capabilities = [];
  let activePreview = null;

  const loadAudit = async () => {
    try {
      const audits = await actionRequest("/api/actions/audit");
      renderActionAudit(auditBox, audits);
    } catch (error) {
      auditBox.replaceChildren(makeEmpty(error.message || "Action audit unavailable."));
    }
  };

  const selectedCapability = () => capabilities.find((item) => item.action_id === actionSelect.value);

  const refreshActionSelection = () => {
    const capability = selectedCapability();
    targetSelect.replaceChildren();
    activePreview = null;
    previewResult.classList.add("hidden");
    confirmPanel.classList.add("hidden");
    runResult.classList.add("hidden");
    confirmationInput.value = "";
    runButton.disabled = true;

    if (!capability) {
      text(byId("actionHeading"), "No Safe Action capability is available.");
      text(byId("actionDescription"), "HostSleuth returned no fixed action definition.");
      previewButton.disabled = true;
      return;
    }

    text(byId("actionHeading"), capability.title || capability.action_id);
    text(byId("actionDescription"), capability.description || "Preview the exact effect before execution.");

    (capability.allowed_targets || []).forEach((target) => {
      const option = document.createElement("option");
      option.value = target;
      option.textContent = target;
      targetSelect.append(option);
    });
    if (!(capability.allowed_targets || []).length) {
      const option = document.createElement("option");
      option.value = "";
      option.textContent = "No services allowlisted for this action";
      targetSelect.append(option);
    }
    previewButton.disabled = !capability.available;
  };

  try {
    capabilities = await actionRequest("/api/actions");
    renderActionCapabilities(capabilityBox, capabilities);
    actionSelect.replaceChildren();
    capabilities.forEach((capability) => {
      const option = document.createElement("option");
      option.value = capability.action_id;
      option.textContent = capability.action_id;
      actionSelect.append(option);
    });
    if (!capabilities.length) throw new Error("No action capabilities were returned.");
    actionSelect.addEventListener("change", refreshActionSelection);
    refreshActionSelection();
  } catch (error) {
    capabilityBox.replaceChildren(makeEmpty(error.message || "Action API unavailable. If this UI is being accessed over the LAN, use an SSH tunnel for Optional Safe Actions."));
    previewButton.disabled = true;
    actionSelect.replaceChildren();
    targetSelect.replaceChildren();
  }

  byId("actionPreviewForm").addEventListener("submit", async (event) => {
    event.preventDefault();
    activePreview = null;
    confirmPanel.classList.add("hidden");
    runResult.classList.add("hidden");
    try {
      const preview = await actionRequest("/api/actions/preview", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ action_id: actionSelect.value, target: targetSelect.value }),
      });
      activePreview = preview;
      previewResult.classList.remove("hidden");
      renderActionPreview(previewResult, preview);
      confirmationInput.value = "";
      runButton.disabled = true;
      if (preview.available) confirmPanel.classList.remove("hidden");
    } catch (error) {
      previewResult.classList.remove("hidden");
      previewResult.replaceChildren(makeEmpty(error.message || "Action preview failed."));
    }
  });

  confirmationInput.addEventListener("input", () => {
    runButton.disabled = !activePreview?.available || confirmationInput.value !== activePreview.confirmation;
  });

  runButton.addEventListener("click", async () => {
    if (!activePreview?.available) return;
    runButton.disabled = true;
    try {
      const result = await actionRequest("/api/actions/run", {
        method: "POST",
        headers: { "Content-Type": "application/json", "X-HostSleuth-Action": "confirm" },
        body: JSON.stringify({ action_id: activePreview.action_id, target: activePreview.target, confirmation: confirmationInput.value }),
      });
      runResult.classList.remove("hidden");
      renderActionResult(runResult, result);
      await loadAudit();
    } catch (error) {
      runResult.classList.remove("hidden");
      if (error.payload?.result) renderActionResult(runResult, error.payload.result);
      else runResult.replaceChildren(makeEmpty(error.message || "Action failed."));
      await loadAudit();
    } finally {
      confirmationInput.value = "";
      runButton.disabled = true;
      activePreview = null;
      confirmPanel.classList.add("hidden");
    }
  });

  await loadAudit();
}

installActionsUI();
