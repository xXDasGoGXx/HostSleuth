let lastRebootStory = null;

function installRebootStoryUI() {
  if (byId("view-reboot")) return;
  const nav = document.querySelector(".nav-tabs");
  const main = document.querySelector("main");
  if (!nav || !main) return;

  const tab = document.createElement("button");
  tab.className = "nav-tab";
  tab.dataset.view = "reboot";
  tab.textContent = "Reboot";
  tab.addEventListener("click", () => showView("reboot"));
  nav.append(tab);

  const section = document.createElement("section");
  section.className = "view";
  section.id = "view-reboot";
  section.dataset.viewPanel = "reboot";
  section.innerHTML = `
    <section class="panel reboot-story-panel">
      <div class="reboot-story-heading">
        <div>
          <p class="eyebrow">REBOOT STORY</p>
          <h1>What happened around this boot—and what failed to come back?</h1>
          <p class="section-copy">Kernel boot identity, bounded journal evidence, current recovery state, and retained changes near boot. Timing is context, not proof of cause.</p>
        </div>
        <button class="primary-button" id="rebootStoryButton" type="button">Build reboot story</button>
      </div>
    </section>

    <section class="panel diagnosis-result hidden" id="rebootStoryResult" aria-live="polite">
      <div class="diagnosis-heading">
        <div>
          <p class="eyebrow">BOOT / RECOVERY ANSWER</p>
          <h2 id="rebootStoryConclusion">Reboot story complete</h2>
          <p class="muted" id="rebootStoryBootMeta"></p>
          <p class="section-copy" id="rebootStoryCause"></p>
          <p class="context-note" id="rebootStoryContext"></p>
        </div>
        <span class="confidence-badge" id="rebootStoryConfidence"></span>
      </div>

      <div class="result-actions">
        <button class="secondary-button" id="rebootCopyEvidence" type="button">Copy evidence summary</button>
        <span class="copy-status" id="rebootCopyStatus" aria-live="polite"></span>
      </div>

      <div class="reboot-story-grid">
        <div class="context-card">
          <p class="subsection-label">What still has not recovered?</p>
          <div id="rebootRecoveryIssues"></div>
        </div>
        <div class="context-card">
          <p class="subsection-label">Current failed services</p>
          <div id="rebootFailedServices"></div>
        </div>
        <div class="context-card">
          <p class="subsection-label">Previous boot / shutdown evidence</p>
          <div id="rebootPreviousBoot"></div>
        </div>
        <div class="context-card">
          <p class="subsection-label">Current boot journal evidence</p>
          <div id="rebootCurrentBoot"></div>
        </div>
      </div>

      <div class="diagnosis-context-grid reboot-story-events-grid">
        <div>
          <p class="subsection-label">Reboot timeline</p>
          <div id="rebootProblemEvents" class="mini-event-list"></div>
        </div>
        <aside class="context-card">
          <p class="subsection-label">Package / kernel / configuration context</p>
          <p class="context-note">Nearby changes are context only and are not presented as reboot causes.</p>
          <div id="rebootContextEvents" class="mini-event-list"></div>
        </aside>
      </div>
    </section>

    <section class="panel diagnosis-result hidden" id="rebootStoryError" aria-live="polite">
      <p class="eyebrow">REBOOT STORY ERROR</p>
      <h2>HostSleuth could not build the reboot story.</h2>
      <p class="section-copy" id="rebootStoryErrorText"></p>
    </section>
  `;
  main.append(section);
  byId("rebootStoryButton").addEventListener("click", runRebootStory);
  byId("rebootCopyEvidence").addEventListener("click", copyRebootEvidence);
}

function rebootRelativeTime(value, bootStartedAt) {
  const at = Date.parse(value);
  const boot = Date.parse(bootStartedAt);
  if (!Number.isFinite(at) || !Number.isFinite(boot)) return "";
  const seconds = Math.round((at - boot) / 1000);
  if (Math.abs(seconds) < 30) return "at boot time";
  const minutes = Math.max(1, Math.round(Math.abs(seconds) / 60));
  return seconds < 0 ? `${minutes}m before boot` : `${minutes}m after boot`;
}

function appendRebootAnchor(container, bootStartedAt) {
  const row = document.createElement("div");
  row.className = "reboot-anchor";
  const marker = document.createElement("span");
  marker.className = "reboot-anchor-marker";
  const label = document.createElement("div");
  const strong = document.createElement("strong");
  strong.textContent = "Boot started";
  const time = document.createElement("span");
  time.textContent = bootStartedAt ? formatTime(bootStartedAt) : "Boot start unavailable";
  label.append(strong, time);
  row.append(marker, label);
  container.append(row);
}

function renderRebootEventList(container, events, emptyText, bootStartedAt = null, showAnchor = false) {
  container.replaceChildren();
  if (!events?.length) {
    if (showAnchor) appendRebootAnchor(container, bootStartedAt);
    container.append(makeEmpty(emptyText));
    return;
  }

  let anchorInserted = false;
  const bootMs = Date.parse(bootStartedAt);
  events.forEach((event) => {
    const eventMs = Date.parse(event.at);
    if (showAnchor && !anchorInserted && Number.isFinite(bootMs) && Number.isFinite(eventMs) && eventMs >= bootMs) {
      appendRebootAnchor(container, bootStartedAt);
      anchorInserted = true;
    }

    const item = document.createElement("div");
    item.className = `reboot-event ${event.severity || ""}`;

    const marker = document.createElement("span");
    marker.className = "reboot-timeline-marker";

    const body = document.createElement("div");
    const summary = document.createElement("strong");
    summary.textContent = event.summary || "Recorded change";
    const meta = document.createElement("span");
    meta.textContent = [
      event.category || "change",
      formatTime(event.at),
      bootStartedAt ? rebootRelativeTime(event.at, bootStartedAt) : "",
    ].filter(Boolean).join(" · ");
    body.append(summary, meta);

    item.append(marker, body);
    container.append(item);
  });

  if (showAnchor && !anchorInserted) appendRebootAnchor(container, bootStartedAt);
}

function renderBootJournalEvidence(container, evidence, previous) {
  container.replaceChildren();
  const status = document.createElement("strong");
  const rawStatus = evidence?.status || "unknown";
  const labels = {
    orderly: "Orderly shutdown evidence present",
    abnormal: "Abnormal termination evidence present",
    available: previous ? "Previous-boot journal available; classification unknown" : "Current-boot journal available",
    unknown: previous ? "Previous-boot journal unavailable / unknown" : "Current-boot journal unavailable / unknown",
  };
  status.textContent = labels[rawStatus] || rawStatus;
  const assessment = document.createElement("p");
  assessment.className = "context-note";
  assessment.textContent = evidence?.assessment || "Classification is unknown.";
  container.append(status, assessment);
  if (evidence?.evidence) {
    const details = document.createElement("details");
    const summary = document.createElement("summary");
    summary.textContent = "Show bounded evidence";
    const body = document.createElement("pre");
    body.className = "reboot-journal-evidence";
    body.textContent = evidence.evidence;
    details.append(summary, body);
    container.append(details);
  }
}

function renderRecoveryIssues(container, issues) {
  container.replaceChildren();
  if (!issues?.length) {
    container.append(makeEmpty("No service, listener, or container is currently confirmed missing/broken by both retained post-boot evidence and the current snapshot."));
    return;
  }
  issues.forEach((issue) => {
    const row = document.createElement("div");
    row.className = "reboot-service-row";
    const left = document.createElement("div");
    const name = document.createElement("strong");
    name.textContent = `${issue.kind || "recovery"} · ${issue.name || "unknown"}`;
    const evidence = document.createElement("div");
    evidence.className = "context-note";
    evidence.textContent = issue.evidence || "";
    left.append(name, evidence);
    const state = document.createElement("span");
    state.textContent = issue.current_state || "unknown";
    row.append(left, state);
    container.append(row);
  });
}

function renderRebootStory(story) {
  lastRebootStory = story;
  byId("rebootStoryError").classList.add("hidden");
  byId("rebootStoryResult").classList.remove("hidden");
  text(byId("rebootStoryConclusion"), story.conclusion || "Reboot story complete");
  text(byId("rebootStoryConfidence"), `${story.confidence || "unknown"} confidence`);
  const id = story.boot_id ? ` · boot ${story.boot_id.slice(0, 12)}` : "";
  text(byId("rebootStoryBootMeta"), `${story.boot_started_at ? formatTime(story.boot_started_at) : "boot start unavailable"}${id}`);
  text(byId("rebootStoryCause"), story.cause_assessment || "No reboot cause is claimed.");
  text(byId("rebootStoryContext"), story.context_note || "Boot-time proximity is context, not causation.");

  renderRecoveryIssues(byId("rebootRecoveryIssues"), story.recovery_issues || []);

  const failed = byId("rebootFailedServices");
  failed.replaceChildren();
  if (!story.failed_services?.length) {
    failed.append(makeEmpty(story.mode === "docker" ? "Native systemd service inventory is unavailable in Docker mode." : "No currently failed systemd services are present in the snapshot."));
  } else {
    story.failed_services.forEach((service) => {
      const item = document.createElement("div");
      item.className = "reboot-service-row";
      const name = document.createElement("strong");
      name.textContent = service.name || "service";
      const state = document.createElement("span");
      state.textContent = `${service.active || "unknown"}/${service.sub || "unknown"}`;
      item.append(name, state);
      failed.append(item);
    });
  }

  renderBootJournalEvidence(byId("rebootPreviousBoot"), story.previous_boot, true);
  renderBootJournalEvidence(byId("rebootCurrentBoot"), story.current_boot, false);
  const timelineEvents = [...(story.problem_events || [])].sort((a, b) => new Date(a.at) - new Date(b.at));
  const contextEvents = [...(story.context_events || [])].sort((a, b) => new Date(a.at) - new Date(b.at));
  renderRebootEventList(byId("rebootProblemEvents"), timelineEvents, "No warning service/listener/container events were retained after boot inside the bounded window.", story.boot_started_at, true);
  renderRebootEventList(byId("rebootContextEvents"), contextEvents, "No retained package, kernel/system, or configuration changes were found inside the boot window.", story.boot_started_at, false);
}

function rebootEvidenceSummary(story) {
  const lines = [
    "HostSleuth Reboot Story",
    `Boot started: ${story?.boot_started_at ? formatTime(story.boot_started_at) : "unavailable"}`,
    story?.boot_id ? `Boot ID: ${String(story.boot_id).slice(0, 12)}` : "",
    `Conclusion: ${story?.conclusion || "unknown"}`,
    `Confidence: ${story?.confidence || "unknown"}`,
    `Assessment: ${story?.cause_assessment || "No reboot cause is claimed."}`,
    story?.context_note || "Boot-time proximity is context, not causation.",
    "",
    `Previous boot: ${story?.previous_boot?.status || "unknown"} · ${story?.previous_boot?.assessment || ""}`,
    `Current boot: ${story?.current_boot?.status || "unknown"} · ${story?.current_boot?.assessment || ""}`,
    "",
    "Unrecovered items:",
  ].filter(Boolean);

  const issues = story?.recovery_issues || [];
  if (!issues.length) {
    lines.push("- none confirmed");
  } else {
    issues.forEach((issue) => {
      lines.push(`- ${issue.kind || "recovery"} · ${issue.name || "unknown"} · ${issue.current_state || "unknown"}`);
    });
  }

  lines.push("", "Failed services:");
  const failed = story?.failed_services || [];
  if (!failed.length) {
    lines.push("- none observed");
  } else {
    failed.forEach((service) => {
      lines.push(`- ${service.name || "service"} · ${service.active || "unknown"}/${service.sub || "unknown"}`);
    });
  }

  lines.push("", "Post-boot problem events:");
  const problems = story?.problem_events || [];
  if (!problems.length) {
    lines.push("- none retained");
  } else {
    problems.forEach((event) => {
      lines.push(`- ${formatTime(event.at)} · ${event.category || "change"} · ${event.summary || "Recorded change"}`);
    });
  }

  lines.push("", "Nearby package/kernel/configuration context:");
  const context = story?.context_events || [];
  if (!context.length) {
    lines.push("- none retained");
  } else {
    context.forEach((event) => {
      lines.push(`- ${formatTime(event.at)} · ${event.category || "change"} · ${event.summary || "Recorded change"}`);
    });
  }

  return lines.join("\n");
}

async function copyRebootEvidence() {
  const status = byId("rebootCopyStatus");
  try {
    await copyBoundedText(rebootEvidenceSummary(lastRebootStory));
    text(status, "Copied.");
  } catch (_) {
    text(status, "Clipboard unavailable.");
  }
  window.setTimeout(() => text(status, ""), 2500);
}

async function runRebootStory() {
  const button = byId("rebootStoryButton");
  const result = byId("rebootStoryResult");
  const errorPanel = byId("rebootStoryError");
  button.disabled = true;
  button.textContent = "Building story…";
  result.classList.add("hidden");
  errorPanel.classList.add("hidden");
  try {
    const response = await fetch("/api/reboot-story", { cache: "no-store" });
    if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
    renderRebootStory(await response.json());
  } catch (error) {
    text(byId("rebootStoryErrorText"), error.message || "Unknown error");
    errorPanel.classList.remove("hidden");
  } finally {
    button.disabled = false;
    button.textContent = "Build reboot story";
  }
}

installRebootStoryUI();
