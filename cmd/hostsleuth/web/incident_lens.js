let lastIncidentLens = null;

function incidentLocalValue(value) {
  const date = value ? new Date(value) : new Date();
  if (Number.isNaN(date.getTime())) return "";
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 16);
}

function installIncidentLensUI() {
  const changes = byId("view-changes");
  if (!changes || byId("incidentLensForm")) return;

  changes.insertAdjacentHTML("beforeend", `
    <section class="panel incident-lens-panel">
      <p class="eyebrow">INCIDENT LENS</p>
      <h2>What changed around the time this broke?</h2>
      <p class="section-copy">Anchor a ±15 minute evidence window on an exact time, recorded change, or completed diagnosis. Nearby events are context only—not proof that they caused the incident.</p>
      <form id="incidentLensForm" class="diagnose-form">
        <div class="incident-fields">
          <div>
            <label for="incidentAnchor">Incident time</label>
            <input id="incidentAnchor" name="at" type="datetime-local" required>
          </div>
          <div>
            <label for="incidentTarget">Endpoint to check now <span class="muted">(optional)</span></label>
            <input id="incidentTarget" name="target" autocomplete="off" spellcheck="false" placeholder="mail.example.com:443">
          </div>
        </div>
        <div class="incident-submit">
          <button class="primary-button" id="incidentLensButton" type="submit">Inspect incident window</button>
          <p class="input-hint">Endpoint evidence is freshly captured and shown separately from historical events.</p>
        </div>
      </form>
    </section>

    <section class="panel diagnosis-result hidden" id="incidentLensResult" aria-live="polite">
      <div class="diagnosis-heading">
        <div>
          <p class="eyebrow">INCIDENT CONTEXT</p>
          <h2 id="incidentConclusion">Incident window</h2>
          <p class="muted incident-window-meta" id="incidentWindowMeta"></p>
          <p class="section-copy" id="incidentContextNote"></p>
        </div>
      </div>
      <div class="result-actions">
        <button class="secondary-button" id="incidentCopyEvidence" type="button">Copy evidence summary</button>
        <span class="copy-status" id="incidentCopyStatus" aria-live="polite"></span>
      </div>
      <div class="diagnosis-context-grid">
        <div>
          <p class="subsection-label">Incident timeline</p>
          <div id="incidentEvents" class="incident-event-list"></div>
        </div>
        <aside class="context-card" id="incidentEndpointCard">
          <p class="subsection-label">Current endpoint evidence</p>
          <div id="incidentEndpoint"></div>
        </aside>
      </div>
    </section>

    <section class="panel diagnosis-result hidden" id="incidentLensError" aria-live="polite">
      <p class="eyebrow">INCIDENT LENS ERROR</p>
      <h2>HostSleuth could not build that incident window.</h2>
      <p class="section-copy" id="incidentLensErrorText"></p>
    </section>
  `);

  const latest = [...(state.events || [])].filter((event) => event.at).sort((a, b) => new Date(b.at) - new Date(a.at))[0];
  byId("incidentAnchor").value = incidentLocalValue(latest?.at || new Date());

  byId("incidentLensForm").addEventListener("submit", (event) => {
    event.preventDefault();
    runIncidentLens(byId("incidentAnchor").value, byId("incidentTarget").value.trim());
  });
  byId("incidentCopyEvidence").addEventListener("click", copyIncidentEvidence);
}

function openIncidentLensAt(value, target = "") {
  installIncidentLensUI();
  byId("incidentAnchor").value = incidentLocalValue(value);
  byId("incidentTarget").value = target || "";
  showView("changes");
  byId("incidentLensForm").scrollIntoView({ behavior: "smooth", block: "start" });
}

function addIncidentAnchors(container, events, limit = null) {
  if (!container) return;
  const ordered = [...events].reverse();
  const list = limit ? ordered.slice(0, limit) : ordered;
  const rows = [...container.querySelectorAll(".timeline-item")];
  rows.forEach((row, index) => {
    const event = list[index];
    if (!event?.at || row.querySelector(".incident-event-anchor")) return;
    const button = document.createElement("button");
    button.type = "button";
    button.className = "surface-action incident-event-anchor";
    button.textContent = "Inspect window";
    button.addEventListener("click", () => openIncidentLensAt(event.at));
    row.append(button);
  });
}

const renderEventsWithoutIncidentLens = renderEvents;
renderEvents = function renderEventsWithIncidentLens(container, events, limit = null) {
  renderEventsWithoutIncidentLens(container, events, limit);
  addIncidentAnchors(container, events, limit);
};

const renderDiagnosisWithoutIncidentLens = renderDiagnosis;
renderDiagnosis = function renderDiagnosisWithIncidentLens(diagnosis) {
  renderDiagnosisWithoutIncidentLens(diagnosis);
  const result = byId("diagnosisResult");
  const heading = result?.querySelector(".diagnosis-heading");
  if (!heading) return;
  const previous = heading.querySelector(".incident-diagnosis-anchor");
  if (previous) previous.remove();
  const button = document.createElement("button");
  button.type = "button";
  button.className = "surface-action incident-diagnosis-anchor";
  button.textContent = "Inspect changes around this diagnosis";
  button.addEventListener("click", () => openIncidentLensAt(diagnosis.started_at || new Date(), diagnosis.target || ""));
  heading.append(button);
};

function incidentRelativeTime(value, anchor) {
  const at = Date.parse(value);
  const anchorAt = Date.parse(anchor);
  if (!Number.isFinite(at) || !Number.isFinite(anchorAt)) return "";
  const seconds = Math.round((at - anchorAt) / 1000);
  if (Math.abs(seconds) < 30) return "at incident time";
  const minutes = Math.max(1, Math.round(Math.abs(seconds) / 60));
  return seconds < 0 ? `${minutes}m before` : `${minutes}m after`;
}

function appendIncidentAnchor(container, anchor) {
  const row = document.createElement("div");
  row.className = "incident-anchor";
  const marker = document.createElement("span");
  marker.className = "incident-anchor-marker";
  const label = document.createElement("div");
  const strong = document.createElement("strong");
  strong.textContent = "Incident time";
  const time = document.createElement("span");
  time.textContent = formatTime(anchor);
  label.append(strong, time);
  row.append(marker, label);
  container.append(row);
}

function renderIncidentLens(lens) {
  lastIncidentLens = lens;
  byId("incidentLensError").classList.add("hidden");
  byId("incidentLensResult").classList.remove("hidden");
  text(byId("incidentConclusion"), lens.conclusion || "Incident window");
  text(byId("incidentContextNote"), lens.context_note || "Temporal proximity is context, not causation.");
  text(byId("incidentWindowMeta"), `${formatTime(lens.window_start)} → ${formatTime(lens.window_end)} · anchor ${formatTime(lens.anchor_at)}`);

  const events = byId("incidentEvents");
  events.replaceChildren();
  const retained = lens.events || [];
  if (!retained.length) {
    appendIncidentAnchor(events, lens.anchor_at);
    events.append(makeEmpty("No retained host changes were recorded inside this bounded window."));
  } else {
    let anchorInserted = false;
    const anchorMs = Date.parse(lens.anchor_at);
    retained.forEach((event) => {
      const eventMs = Date.parse(event.at);
      if (!anchorInserted && Number.isFinite(anchorMs) && Number.isFinite(eventMs) && eventMs >= anchorMs) {
        appendIncidentAnchor(events, lens.anchor_at);
        anchorInserted = true;
      }

      const row = document.createElement("article");
      row.className = `incident-event ${event.severity || ""}`;

      const marker = document.createElement("span");
      marker.className = "incident-timeline-marker";

      const body = document.createElement("div");
      const summary = document.createElement("strong");
      summary.textContent = event.summary || "Recorded change";
      const meta = document.createElement("span");
      meta.textContent = [
        event.category || "change",
        formatTime(event.at),
        incidentRelativeTime(event.at, lens.anchor_at),
      ].filter(Boolean).join(" · ");
      body.append(summary, meta);
      row.append(marker, body);
      events.append(row);
    });
    if (!anchorInserted) appendIncidentAnchor(events, lens.anchor_at);
  }

  const endpoint = byId("incidentEndpoint");
  endpoint.replaceChildren();
  if (!lens.current_endpoint) {
    endpoint.append(makeEmpty("No endpoint was supplied. Incident Lens did not perform a network re-probe."));
    return;
  }
  const note = document.createElement("p");
  note.className = "context-note";
  note.textContent = `Captured now (${formatTime(lens.endpoint_captured_at)}), not reconstructed from the incident time.`;
  const conclusion = document.createElement("strong");
  conclusion.textContent = lens.current_endpoint.conclusion || "Current endpoint check complete";
  const confidence = document.createElement("p");
  confidence.className = "muted";
  confidence.textContent = `${lens.current_endpoint.confidence || "unknown"} confidence · ${lens.target || "endpoint"}`;
  endpoint.append(note, conclusion, confidence);
}

function incidentEvidenceSummary(lens) {
  const lines = [
    "HostSleuth Incident Lens",
    `Window: ${formatTime(lens?.window_start)} -> ${formatTime(lens?.window_end)}`,
    `Anchor: ${formatTime(lens?.anchor_at)}`,
    `Conclusion: ${lens?.conclusion || "unknown"}`,
    lens?.context_note || "Temporal proximity is context, not causation.",
  ];

  if (lens?.target) {
    lines.push(
      "",
      `Current endpoint: ${lens.target}`,
      `Captured: ${formatTime(lens.endpoint_captured_at)}`,
      `Endpoint conclusion: ${lens.current_endpoint?.conclusion || "unavailable"}`,
      `Endpoint confidence: ${lens.current_endpoint?.confidence || "unknown"}`,
    );
  }

  lines.push("", "Retained changes:");
  const events = lens?.events || [];
  if (!events.length) {
    lines.push("- none in bounded window");
  } else {
    events.forEach((event) => {
      lines.push(`- [${String(event.severity || "info").toUpperCase()}] ${formatTime(event.at)} · ${event.category || "change"} · ${event.summary || "Recorded change"}`);
    });
  }
  return lines.join("\n");
}

async function copyIncidentEvidence() {
  const status = byId("incidentCopyStatus");
  try {
    await copyBoundedText(incidentEvidenceSummary(lastIncidentLens));
    text(status, "Copied.");
  } catch (_) {
    text(status, "Clipboard unavailable.");
  }
  window.setTimeout(() => text(status, ""), 2500);
}

async function runIncidentLens(localAnchor, target) {
  const button = byId("incidentLensButton");
  const result = byId("incidentLensResult");
  const errorPanel = byId("incidentLensError");
  const date = new Date(localAnchor);
  if (Number.isNaN(date.getTime())) {
    text(byId("incidentLensErrorText"), "Choose a valid incident time.");
    errorPanel.classList.remove("hidden");
    return;
  }

  button.disabled = true;
  button.textContent = "Inspecting…";
  result.classList.add("hidden");
  errorPanel.classList.add("hidden");
  try {
    const params = new URLSearchParams({ at: date.toISOString() });
    if (target) params.set("target", target);
    const response = await fetch(`/api/incident-lens?${params.toString()}`, { cache: "no-store" });
    if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
    renderIncidentLens(await response.json());
  } catch (error) {
    text(byId("incidentLensErrorText"), error.message || "Unknown error");
    errorPanel.classList.remove("hidden");
  } finally {
    button.disabled = false;
    button.textContent = "Inspect incident window";
  }
}

installIncidentLensUI();
