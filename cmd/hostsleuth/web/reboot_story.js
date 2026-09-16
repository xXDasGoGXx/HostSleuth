function installRebootStoryUI() {
  const changes = byId("view-changes");
  if (!changes || byId("rebootStoryPanel")) return;

  changes.insertAdjacentHTML("beforeend", `
    <section class="panel reboot-story-panel" id="rebootStoryPanel">
      <div class="reboot-story-heading">
        <div>
          <p class="eyebrow">REBOOT STORY</p>
          <h2>What happened around this boot—and what failed to come back?</h2>
          <p class="section-copy">Correlate kernel boot identity, bounded previous-boot journal evidence when readable, current failed services, and retained changes near boot. HostSleuth does not invent a reboot cause.</p>
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
      <div class="reboot-story-grid">
        <div class="context-card">
          <p class="subsection-label">Current recovery state</p>
          <div id="rebootFailedServices"></div>
        </div>
        <div class="context-card">
          <p class="subsection-label">Previous boot evidence</p>
          <div id="rebootPreviousBoot"></div>
        </div>
      </div>
      <div class="diagnosis-context-grid reboot-story-events-grid">
        <div>
          <p class="subsection-label">Post-boot problems in bounded window</p>
          <div id="rebootProblemEvents" class="mini-event-list"></div>
        </div>
        <aside class="context-card">
          <p class="subsection-label">All retained changes near boot</p>
          <div id="rebootRelatedEvents" class="mini-event-list"></div>
        </aside>
      </div>
    </section>

    <section class="panel diagnosis-result hidden" id="rebootStoryError" aria-live="polite">
      <p class="eyebrow">REBOOT STORY ERROR</p>
      <h2>HostSleuth could not build the reboot story.</h2>
      <p class="section-copy" id="rebootStoryErrorText"></p>
    </section>
  `);

  byId("rebootStoryButton").addEventListener("click", runRebootStory);
}

function renderRebootEventList(container, events, emptyText) {
  container.replaceChildren();
  if (!events?.length) {
    container.append(makeEmpty(emptyText));
    return;
  }
  events.forEach((event) => {
    const item = document.createElement("div");
    item.className = "mini-event";
    const summary = document.createElement("strong");
    summary.textContent = event.summary || "Recorded change";
    const meta = document.createElement("span");
    meta.textContent = `${event.category || "change"} · ${formatTime(event.at)}`;
    item.append(summary, meta);
    container.append(item);
  });
}

function renderRebootStory(story) {
  byId("rebootStoryError").classList.add("hidden");
  byId("rebootStoryResult").classList.remove("hidden");
  text(byId("rebootStoryConclusion"), story.conclusion || "Reboot story complete");
  text(byId("rebootStoryConfidence"), `${story.confidence || "unknown"} confidence`);
  const id = story.boot_id ? ` · boot ${story.boot_id.slice(0, 12)}` : "";
  text(byId("rebootStoryBootMeta"), `${story.boot_started_at ? formatTime(story.boot_started_at) : "boot start unavailable"}${id}`);
  text(byId("rebootStoryCause"), story.cause_assessment || "No reboot cause is claimed.");
  text(byId("rebootStoryContext"), story.context_note || "Boot-time proximity is context, not causation.");

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

  const previous = byId("rebootPreviousBoot");
  previous.replaceChildren();
  const status = document.createElement("strong");
  status.textContent = story.previous_boot?.status === "available" ? "Previous-boot journal available" : "Previous-boot journal unavailable / unknown";
  const assessment = document.createElement("p");
  assessment.className = "context-note";
  assessment.textContent = story.previous_boot?.assessment || "Shutdown classification is unknown.";
  previous.append(status, assessment);
  if (story.previous_boot?.evidence) {
    const details = document.createElement("details");
    const summary = document.createElement("summary");
    summary.textContent = "Show bounded evidence";
    const evidence = document.createElement("pre");
    evidence.className = "reboot-journal-evidence";
    evidence.textContent = story.previous_boot.evidence;
    details.append(summary, evidence);
    previous.append(details);
  }

  renderRebootEventList(byId("rebootProblemEvents"), story.problem_events || [], "No warning service/listener/container events were retained after boot inside the bounded window.");
  renderRebootEventList(byId("rebootRelatedEvents"), story.related_events || [], "No retained host changes were found inside the boot window.");
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
