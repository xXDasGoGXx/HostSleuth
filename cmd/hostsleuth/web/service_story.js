Object.assign(checkNames, {
  "service-runtime": "Systemd runtime",
  "service-journal": "Recent service journal",
  "service-listeners": "Service-owned listeners",
  "service-port": "Expected port ownership",
  "service-container": "Container port overlap",
  "service-endpoint": "Expected endpoint",
  "service-target": "Expected endpoint format",
  "service-unit": "Service unit",
});

function installServiceStoryUI() {
  const diagnose = byId("view-diagnose");
  if (!diagnose || byId("serviceStoryForm")) return;

  diagnose.insertAdjacentHTML("beforeend", `
    <section class="panel diagnose-panel service-story-panel">
      <p class="eyebrow">SERVICE STORY</p>
      <h2>Why did this service fail—or why did its endpoint disappear?</h2>
      <p class="section-copy">Inspect one systemd service with bounded journal, process/listener ownership, expected-port collisions, endpoint/TLS evidence, and nearby recorded changes. HostSleuth reports evidence, not an invented cause.</p>
      <form id="serviceStoryForm" class="diagnose-form">
        <div class="service-story-fields">
          <div>
            <label for="serviceStoryUnit">Systemd service</label>
            <input id="serviceStoryUnit" name="service" list="serviceOptions" autocomplete="off" spellcheck="false" placeholder="postfix.service" required>
            <datalist id="serviceOptions"></datalist>
          </div>
          <div>
            <label for="serviceStoryTarget">Expected endpoint <span class="muted">(optional)</span></label>
            <input id="serviceStoryTarget" name="target" autocomplete="off" spellcheck="false" placeholder="127.0.0.1:25">
          </div>
        </div>
        <div class="service-story-submit">
          <button class="primary-button" id="serviceStoryButton" type="submit">Build service story</button>
          <p class="input-hint">Examples: <code>postfix.service</code>, or <code>dovecot.service</code> with <code>127.0.0.1:993</code>.</p>
        </div>
      </form>
    </section>

    <section class="panel diagnosis-result hidden" id="serviceStoryResult" aria-live="polite">
      <div class="diagnosis-heading">
        <div>
          <p class="eyebrow">SERVICE ANSWER</p>
          <h2 id="serviceStoryConclusion">Service story complete</h2>
          <p class="muted" id="serviceStorySubject"></p>
          <p class="section-copy" id="serviceStoryExplanation"></p>
        </div>
        <span class="confidence-badge" id="serviceStoryConfidence"></span>
      </div>
      <div class="diagnosis-context-grid service-story-grid">
        <div>
          <p class="subsection-label">Evidence</p>
          <div id="serviceStoryChecks" class="check-list"></div>
        </div>
        <aside class="context-card">
          <p class="subsection-label">Related recorded context</p>
          <p class="context-note">Direct service/listener history plus package/config/container changes within 15 minutes of the newest direct event. Proximity is not causation.</p>
          <div id="serviceStoryEvents" class="mini-event-list"></div>
        </aside>
      </div>
    </section>

    <section class="panel diagnosis-result hidden" id="serviceStoryError" aria-live="polite">
      <p class="eyebrow">SERVICE STORY ERROR</p>
      <h2>HostSleuth could not complete that service story.</h2>
      <p class="section-copy" id="serviceStoryErrorText"></p>
    </section>
  `);

  byId("serviceStoryForm").addEventListener("submit", (event) => {
    event.preventDefault();
    const unit = byId("serviceStoryUnit").value.trim();
    const target = byId("serviceStoryTarget").value.trim();
    if (unit) runServiceStory(unit, target);
  });

  loadServiceOptions();
}

async function loadServiceOptions() {
  const options = byId("serviceOptions");
  if (!options) return;
  try {
    const response = await fetch("/api/snapshot", { cache: "no-store" });
    if (!response.ok) return;
    const snapshot = await response.json();
    options.replaceChildren();
    [...(snapshot.services || [])]
      .sort((a, b) => String(a.name || "").localeCompare(String(b.name || "")))
      .forEach((item) => {
        const option = document.createElement("option");
        option.value = item.name || "";
        options.append(option);
      });
  } catch (_) {
    // Service names are a convenience only; manual entry remains available.
  }
}

function renderServiceStory(story) {
  byId("serviceStoryError").classList.add("hidden");
  const result = byId("serviceStoryResult");
  result.classList.remove("hidden");

  text(byId("serviceStoryConclusion"), story.conclusion || "Service story complete");
  text(byId("serviceStorySubject"), [story.service || "", story.target || ""].filter(Boolean).join(" · "));
  text(byId("serviceStoryExplanation"), "HostSleuth correlated current systemd, listener/process ownership, endpoint evidence, and retained host changes without claiming nearby events caused the problem.");
  text(byId("serviceStoryConfidence"), `${story.confidence || "unknown"} confidence`);

  const checks = byId("serviceStoryChecks");
  checks.replaceChildren();
  (story.checks || []).forEach((check) => {
    const status = ["pass", "fail", "unknown"].includes(check.status) ? check.status : "unknown";
    const details = document.createElement("details");
    details.className = `check-item ${status}`;
    if (status === "fail") details.open = true;

    const summary = document.createElement("summary");
    summary.className = "check-summary";
    const icon = document.createElement("span");
    icon.className = "check-icon";
    icon.textContent = status === "pass" ? "✓" : status === "fail" ? "×" : "?";
    const name = document.createElement("span");
    name.className = "check-name";
    name.textContent = checkNames[check.name] || check.name || "Check";
    const statusText = document.createElement("span");
    statusText.className = "check-status";
    statusText.textContent = status;
    summary.append(icon, name, statusText);

    const evidence = document.createElement("p");
    evidence.className = "check-evidence";
    evidence.textContent = check.evidence || "No additional evidence was returned.";
    details.append(summary, evidence);
    checks.append(details);
  });

  const events = byId("serviceStoryEvents");
  events.replaceChildren();
  const related = story.related_events || [];
  if (!related.length) {
    events.append(makeEmpty("No directly related retained service/listener changes were found, so HostSleuth did not invent nearby context."));
    return;
  }
  related.forEach((event) => {
    const item = document.createElement("div");
    item.className = "mini-event";
    const summary = document.createElement("strong");
    summary.textContent = event.summary || "Recorded change";
    const meta = document.createElement("span");
    meta.textContent = `${event.category || "change"} · ${formatTime(event.at)}`;
    item.append(summary, meta);
    events.append(item);
  });
}

async function runServiceStory(unit, target) {
  const button = byId("serviceStoryButton");
  const result = byId("serviceStoryResult");
  const errorPanel = byId("serviceStoryError");
  button.disabled = true;
  button.textContent = "Building story…";
  result.classList.add("hidden");
  errorPanel.classList.add("hidden");

  try {
    const params = new URLSearchParams({ service: unit });
    if (target) params.set("target", target);
    const response = await fetch(`/api/service-story?${params.toString()}`, { cache: "no-store" });
    if (!response.ok) {
      const message = await response.text();
      throw new Error(message || `request failed (${response.status})`);
    }
    renderServiceStory(await response.json());
  } catch (error) {
    text(byId("serviceStoryErrorText"), error.message || "Unknown error");
    errorPanel.classList.remove("hidden");
  } finally {
    button.disabled = false;
    button.textContent = "Build service story";
  }
}

installServiceStoryUI();
