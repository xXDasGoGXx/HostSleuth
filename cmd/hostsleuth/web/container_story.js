// Container Story: which containers keep restarting, why (Docker's own
// exit evidence), what changed (image updates) and which listeners moved
// at the same time. Read-only; built from the snapshot and the timeline.

(() => {
  let lastContainerList = null;
  let lastContainerStory = null;

  const ATTENTION_TEXT = {
    unhealthy: "Unhealthy",
    restarting: "Restarting",
    "crash-looping": "Crash-looping",
    stopped: "Stopped",
    unstable: "Unstable",
  };

  function installContainerStoryUI() {
    if (byId("view-containers")) return;
    const nav = document.querySelector(".nav-tabs");
    const main = document.querySelector("main");
    if (!nav || !main) return;

    const tab = document.createElement("button");
    tab.className = "nav-tab";
    tab.dataset.view = "containers";
    tab.textContent = "Containers";
    tab.addEventListener("click", () => showView("containers"));
    const reboot = nav.querySelector('[data-view="reboot"]');
    if (reboot) reboot.before(tab);
    else nav.append(tab);

    const section = document.createElement("section");
    section.className = "view";
    section.id = "view-containers";
    section.dataset.viewPanel = "containers";
    section.innerHTML = `
      <section class="panel section-panel container-story-panel">
        <div class="reboot-story-heading">
          <div>
            <p class="eyebrow">CONTAINER STORY</p>
            <h1>Which containers keep restarting—and what changed?</h1>
            <p class="section-copy">Restarts with Docker's own exit reason, out-of-memory kills, health changes and image updates from the last 7 days. Listener changes at the same moment are context, not proof.</p>
          </div>
          <button class="secondary-button" id="containerListRefresh" type="button">Refresh</button>
        </div>
        <div class="container-grid" id="containerStoryList" aria-live="polite"></div>
      </section>

      <section class="panel diagnosis-result hidden" id="containerStoryResult" aria-live="polite">
        <div class="diagnosis-heading">
          <div>
            <p class="eyebrow">CONTAINER ANSWER</p>
            <h2 id="containerStoryConclusion">Container story</h2>
            <p class="muted" id="containerStoryMeta"></p>
          </div>
          <span class="confidence-badge" id="containerStoryStatus"></span>
        </div>
        <div class="result-actions">
          <button class="secondary-button" id="containerCopyEvidence" type="button">Copy evidence summary</button>
          <button class="secondary-button share-button" id="containerCopyEvidenceShare" type="button">Share redacted</button>
          <span class="copy-status" id="containerCopyStatus" aria-live="polite"></span>
        </div>
        <div class="reboot-story-grid">
          <div class="context-card">
            <p class="subsection-label">What happened to it (newest first)</p>
            <div id="containerStoryTimeline"></div>
          </div>
          <div class="context-card">
            <p class="subsection-label">Listeners that changed at the same time</p>
            <div id="containerStoryNearby"></div>
          </div>
        </div>
        <ul class="container-story-notes" id="containerStoryNotes"></ul>
      </section>`;
    main.append(section);

    byId("containerListRefresh").addEventListener("click", loadContainerList);
    byId("containerCopyEvidence").addEventListener("click", copyContainerEvidence);
    byId("containerCopyEvidenceShare").addEventListener("click", () => shareRedacted(() => containerEvidenceSummary(lastContainerStory), "container"));
    document.addEventListener("hostsleuth:view", (event) => {
      if (event.detail === "containers") loadContainerList();
    });
    document.addEventListener("hostsleuth:snapshot", enrichOverviewContainers);
    if (activeViewName() === "containers") loadContainerList();
  }

  async function fetchContainerList() {
    const response = await fetch("/api/containers", { cache: "no-store" });
    if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
    return response.json();
  }

  async function loadContainerList() {
    const list = byId("containerStoryList");
    if (!lastContainerList) list.replaceChildren(makeEmpty("Reading containers…"));
    try {
      lastContainerList = await fetchContainerList();
      renderContainerList(lastContainerList);
    } catch (error) {
      list.replaceChildren(makeEmpty(`HostSleuth could not load containers: ${error.message || error}`));
    }
  }

  function countLine(summary) {
    const parts = [];
    if (summary.restarts_24h) parts.push(`${summary.restarts_24h} restart${summary.restarts_24h === 1 ? "" : "s"} today`);
    if (summary.restarts_7d && summary.restarts_7d !== summary.restarts_24h) parts.push(`${summary.restarts_7d} in 7 days`);
    if (summary.oom_kills_7d) parts.push(`${summary.oom_kills_7d} out-of-memory`);
    if (summary.unhealthy_7d) parts.push(`unhealthy ${summary.unhealthy_7d}×`);
    return parts.join(" · ") || "No restarts or health problems in 7 days";
  }

  function renderContainerList(data) {
    const list = byId("containerStoryList");
    list.replaceChildren();
    const containers = data.containers || [];
    if (!containers.length) {
      list.append(makeEmpty(data.note || "No containers in the current snapshot."));
      return;
    }
    containers.forEach((summary) => {
      const card = document.createElement("button");
      card.type = "button";
      card.className = `container-card ${summary.attention ? `attention-${summary.attention}` : ""}`;
      card.setAttribute("aria-label", `Container story for ${summary.name}`);

      const head = document.createElement("span");
      head.className = "container-card-head";
      const name = document.createElement("strong");
      name.textContent = summary.name;
      head.append(name);
      if (summary.attention) {
        const badge = document.createElement("span");
        badge.className = `container-badge ${summary.attention}`;
        badge.textContent = ATTENTION_TEXT[summary.attention] || summary.attention;
        head.append(badge);
      }

      const image = document.createElement("span");
      image.className = "container-card-image";
      image.textContent = [summary.image, summary.image_version ? `v${String(summary.image_version).replace(/^v/i, "")}` : ""].filter(Boolean).join(" · ");

      const counts = document.createElement("span");
      counts.className = "container-card-counts";
      counts.textContent = countLine(summary);

      card.append(head, image, counts);
      if (summary.last_update_at) {
        const update = document.createElement("span");
        update.className = "container-card-update";
        update.textContent = `Updated ${formatTime(summary.last_update_at)}`;
        card.append(update);
      }
      card.addEventListener("click", () => loadContainerStory(summary.name));
      list.append(card);
    });
  }

  async function loadContainerStory(name) {
    const result = byId("containerStoryResult");
    text(byId("containerStoryConclusion"), `Reading ${name}…`);
    result.classList.remove("hidden");
    try {
      const response = await fetch(`/api/container-story?name=${encodeURIComponent(name)}`, { cache: "no-store" });
      if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
      lastContainerStory = await response.json();
      renderContainerStory(lastContainerStory);
      result.scrollIntoView({ behavior: window.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches ? "auto" : "smooth", block: "start" });
    } catch (error) {
      text(byId("containerStoryConclusion"), `HostSleuth could not build this story: ${error.message || error}`);
    }
  }

  function eventRow(event) {
    const row = document.createElement("div");
    row.className = `container-event ${event.severity === "warning" ? "warning" : ""}`;
    const when = document.createElement("span");
    when.className = "container-event-time";
    when.textContent = formatTime(event.at);
    const what = document.createElement("span");
    what.textContent = event.summary;
    row.append(when, what);
    return row;
  }

  function renderContainerStory(story) {
    text(byId("containerStoryConclusion"), story.conclusion || "Container story");
    const badge = byId("containerStoryStatus");
    badge.textContent = genericStatusLabel(story.status);
    badge.className = `confidence-badge container-verdict ${story.status || "unknown"}`;

    const current = story.current || {};
    const meta = [];
    if (current.image) meta.push(current.image + (current.image_version ? ` · version ${current.image_version}` : ""));
    if (current.started_at) meta.push(`running since ${formatTime(current.started_at)}`);
    if (current.restart_count) meta.push(`Docker restart count ${current.restart_count}`);
    text(byId("containerStoryMeta"), meta.join(" · "));

    const timeline = byId("containerStoryTimeline");
    timeline.replaceChildren();
    const events = [...(story.timeline || [])].reverse();
    if (!events.length) timeline.append(makeEmpty("No restarts, health changes or updates in the last 7 days."));
    events.forEach((event) => timeline.append(eventRow(event)));

    const nearby = byId("containerStoryNearby");
    nearby.replaceChildren();
    const near = [...(story.nearby_listener_changes || [])].reverse();
    if (!near.length) nearby.append(makeEmpty("No listener changed within two minutes of this container's changes."));
    near.forEach((event) => nearby.append(eventRow(event)));

    const notes = byId("containerStoryNotes");
    notes.replaceChildren();
    (story.notes || []).forEach((note) => {
      const li = document.createElement("li");
      li.textContent = note;
      notes.append(li);
    });
  }

  function containerEvidenceSummary(story) {
    if (!story) return "";
    const lines = [
      `HostSleuth Container Story: ${story.name}`,
      `Verdict: ${genericStatusLabel(story.status)} — ${story.conclusion || ""}`,
    ];
    const current = story.current || {};
    if (current.image) lines.push(`Image: ${current.image}${current.image_version ? ` (version ${current.image_version})` : ""}`);
    if (current.status) lines.push(`Docker status: ${current.status}`);
    if (current.restart_count) lines.push(`Docker restart count: ${current.restart_count}`);
    lines.push("", "Changes (newest first):");
    const events = [...(story.timeline || [])].reverse();
    if (!events.length) lines.push("  none in 7 days");
    events.forEach((event) => lines.push(`  ${event.at}  ${event.summary}`));
    const near = story.nearby_listener_changes || [];
    if (near.length) {
      lines.push("", "Listener changes within two minutes (context, not proof):");
      [...near].reverse().forEach((event) => lines.push(`  ${event.at}  ${event.summary}`));
    }
    return lines.join("\n");
  }

  async function copyContainerEvidence() {
    const status = byId("containerCopyStatus");
    try {
      await copyBoundedText(containerEvidenceSummary(lastContainerStory));
      text(status, "Copied.");
    } catch (_) {
      text(status, "Clipboard unavailable.");
    }
    window.setTimeout(() => text(status, ""), 2500);
  }

  // Overview's Containers tile counts what Docker reports right now; add
  // what the timeline knows about today's restarts.
  async function enrichOverviewContainers() {
    const note = byId("containerNote");
    if (!note || !(state.snapshot?.containers || []).length) return;
    try {
      const data = await fetchContainerList();
      lastContainerList = data;
      const restarted = (data.containers || []).filter((item) => item.restarts_24h > 0);
      if (!restarted.length) return;
      const total = restarted.reduce((sum, item) => sum + item.restarts_24h, 0);
      const base = note.textContent.replace(/ · \d+ restarts? today.*$/, "");
      note.textContent = `${base} · ${total} restart${total === 1 ? "" : "s"} today`;
      note.classList.add("problem");
    } catch (_) {
      // Overview stays as it was.
    }
  }

  installContainerStoryUI();
})();
