(() => {
  let lastPermissionStory = null;

  function permissionStoryInit() {
    if (document.querySelector('[data-view="permissions"]')) return;
    const nav = document.querySelector(".nav-tabs");
    const main = document.querySelector("main");
    if (!nav || !main) return;

    const navButton = document.createElement("button");
    navButton.className = "nav-tab";
    navButton.dataset.view = "permissions";
    navButton.textContent = "Permissions";
    navButton.addEventListener("click", () => showView("permissions"));
    nav.append(navButton);

    const view = document.createElement("section");
    view.className = "view";
    view.id = "view-permissions";
    view.dataset.viewPanel = "permissions";
    view.innerHTML = `
      <section class="panel section-panel permission-hero">
        <div class="section-heading">
          <div>
            <p class="eyebrow">DEPLOYMENT / PERMISSIONS STORY</p>
            <h1>The process is running. Why can it not use this path?</h1>
            <p class="section-copy">Inspect one native systemd service and one explicit absolute path. HostSleuth compares the observed process UID/GID/groups with ownership and mode bits across every parent directory and the target.</p>
          </div>
          <span class="permission-readonly-chip">Metadata only · No contents · Read-only</span>
        </div>
        <form id="permissionStoryForm" class="permission-form">
          <label>
            <span>Native service</span>
            <input id="permissionService" autocomplete="off" spellcheck="false" placeholder="nginx.service" required>
          </label>
          <label>
            <span>Absolute path / socket</span>
            <input id="permissionPath" autocomplete="off" spellcheck="false" placeholder="/srv/app/data" required>
          </label>
          <div class="permission-form-actions">
            <button class="primary-button" id="permissionStoryButton" type="submit">Trace permission chain</button>
            <span class="input-hint">No recursive crawl, file reads, chmod/chown, ACL changes, or service changes.</span>
          </div>
        </form>
      </section>

      <section class="panel diagnosis-result hidden permission-result" id="permissionStoryResult" aria-live="polite">
        <div class="permission-verdict-row">
          <div>
            <p class="eyebrow">PERMISSION VERDICT</p>
            <h2 id="permissionConclusion">Permission chain evaluated</h2>
            <p class="muted" id="permissionTarget"></p>
          </div>
          <span class="confidence-badge permission-verdict" id="permissionStatus"></span>
        </div>

        <div class="permission-identity-grid" id="permissionIdentity"></div>

        <div class="permission-section-head">
          <div>
            <p class="subsection-label">VISUAL PERMISSION CHAIN</p>
            <h3>Where does traversal or target access break?</h3>
          </div>
        </div>
        <div class="permission-chain" id="permissionChain"></div>

        <div class="permission-section-head evidence-head">
          <div>
            <p class="subsection-label">SEARCHABLE EVIDENCE</p>
            <h3>Ownership, mode, relation, and decision</h3>
          </div>
          <input id="permissionSearch" class="permission-search" type="search" placeholder="Filter path, operation, status…">
        </div>
        <div class="permission-table-wrap">
          <table class="permission-table">
            <thead><tr><th>Path</th><th>Operation</th><th>Status</th><th>Class</th><th>Evidence</th></tr></thead>
            <tbody id="permissionEvidenceRows"></tbody>
          </table>
        </div>

        <section class="permission-mounts hidden" id="permissionMountSection">
          <p class="subsection-label">RELEVANT DOCKER BIND MOUNTS</p>
          <div id="permissionMounts"></div>
        </section>

        <details class="permission-scope">
          <summary>Reasoning boundary</summary>
          <ul id="permissionScopeNotes"></ul>
        </details>
      </section>

      <section class="panel diagnosis-result hidden" id="permissionStoryError" aria-live="polite">
        <p class="eyebrow">PERMISSION STORY ERROR</p>
        <h2>HostSleuth could not evaluate that permission chain.</h2>
        <p class="section-copy" id="permissionStoryErrorText"></p>
      </section>
    `;
    main.append(view);

    byId("permissionStoryForm").addEventListener("submit", runPermissionStory);
    byId("permissionSearch").addEventListener("input", () => renderPermissionEvidence(lastPermissionStory));
    if (window.location.hash === "#permissions") showView("permissions");
  }

  async function runPermissionStory(event) {
    event.preventDefault();
    const button = byId("permissionStoryButton");
    const result = byId("permissionStoryResult");
    const error = byId("permissionStoryError");
    result.classList.add("hidden");
    error.classList.add("hidden");
    button.disabled = true;
    button.textContent = "Tracing…";

    try {
      const service = byId("permissionService").value.trim();
      const path = byId("permissionPath").value.trim();
      const params = new URLSearchParams({ service, path });
      const response = await fetch(`/api/permissions-story?${params.toString()}`, { cache: "no-store" });
      if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
      lastPermissionStory = await response.json();
      renderPermissionStory(lastPermissionStory);
    } catch (err) {
      text(byId("permissionStoryErrorText"), err instanceof Error ? err.message : String(err));
      error.classList.remove("hidden");
    } finally {
      button.disabled = false;
      button.textContent = "Trace permission chain";
    }
  }

  function renderPermissionStory(story) {
    const status = story.status || "unknown";
    text(byId("permissionConclusion"), story.conclusion || "Permission chain evaluated");
    text(byId("permissionTarget"), `${story.service || "service"} · ${story.requested_path || "path"}${story.resolved_path && story.resolved_path !== story.requested_path ? ` → ${story.resolved_path}` : ""}`);
    const badge = byId("permissionStatus");
    badge.textContent = status.toUpperCase();
    badge.className = `confidence-badge permission-verdict ${status}`;

    renderPermissionIdentity(story.identity || {});
    renderPermissionChain(story);
    renderPermissionEvidence(story);
    renderPermissionMounts(story.docker_bind_mounts || []);
    renderPermissionScope(story.scope_notes || []);
    byId("permissionStoryResult").classList.remove("hidden");
  }

  function renderPermissionIdentity(identity) {
    const grid = byId("permissionIdentity");
    grid.replaceChildren();
    const facts = [
      ["PID", identity.pid || "—"],
      ["UID / GID", identity.pid ? `${identity.uid} / ${identity.gid}` : "—"],
      ["Supplementary GIDs", (identity.supplementary_gids || []).join(", ") || "—"],
      ["Configured identity", [identity.configured_user, identity.configured_group].filter(Boolean).join(" : ") || "—"],
      ["Working directory", identity.working_directory || "—"],
      ["Executable", identity.executable || "—"],
    ];
    facts.forEach(([label, value]) => {
      const card = document.createElement("article");
      card.className = "permission-identity-card";
      const key = document.createElement("span");
      key.textContent = label;
      const val = document.createElement("code");
      val.textContent = String(value);
      card.append(key, val);
      grid.append(card);
    });
  }

  function renderPermissionChain(story) {
    const chain = byId("permissionChain");
    chain.replaceChildren();
    const decisions = story.decisions || [];
    (story.chain || []).forEach((node, index, nodes) => {
      const related = decisions.filter((item) => item.path === node.path);
      const state = node.stat_error
        ? "fail"
        : related.some((item) => item.status === "fail")
          ? "fail"
          : related.some((item) => item.status === "unknown")
            ? "unknown"
            : "pass";
      const card = document.createElement("article");
      card.className = `permission-node ${state} ${node.is_target ? "target" : ""}`;
      const top = document.createElement("div");
      top.className = "permission-node-head";
      const path = document.createElement("code");
      path.textContent = node.path;
      const badge = document.createElement("span");
      badge.className = `permission-node-badge ${state}`;
      badge.textContent = state.toUpperCase();
      top.append(path, badge);

      const meta = document.createElement("p");
      meta.textContent = node.stat_error
        ? node.stat_error
        : `${node.kind || "node"} · mode ${node.mode || "—"} · uid ${node.uid} · gid ${node.gid} · ${node.relation || "unknown"} class`;

      const ops = document.createElement("div");
      ops.className = "permission-node-ops";
      related.forEach((decision) => {
        const chip = document.createElement("span");
        chip.className = `permission-op ${decision.status || "unknown"}`;
        chip.textContent = `${decision.operation}: ${decision.status}`;
        ops.append(chip);
      });
      card.append(top, meta, ops);
      chain.append(card);

      if (index < nodes.length - 1) {
        const arrow = document.createElement("span");
        arrow.className = "permission-chain-arrow";
        arrow.textContent = "›";
        arrow.setAttribute("aria-hidden", "true");
        chain.append(arrow);
      }
    });
  }

  function renderPermissionEvidence(story) {
    const body = byId("permissionEvidenceRows");
    if (!body) return;
    body.replaceChildren();
    const needle = (byId("permissionSearch")?.value || "").trim().toLowerCase();
    const rows = (story?.decisions || []).filter((item) => {
      const haystack = [item.path, item.operation, item.status, item.relation, item.required, item.evidence].join(" ").toLowerCase();
      return !needle || haystack.includes(needle);
    });
    rows.forEach((item) => {
      const row = document.createElement("tr");
      const values = [item.path, item.operation, item.status, item.relation || "—", item.evidence || "—"];
      values.forEach((value, index) => {
        const cell = document.createElement("td");
        if (index === 0 || index === 4) {
          const code = document.createElement("code");
          code.textContent = value;
          cell.append(code);
        } else {
          cell.textContent = value;
        }
        if (index === 2) cell.className = `permission-status-cell ${item.status || "unknown"}`;
        row.append(cell);
      });
      body.append(row);
    });
    if (!rows.length) {
      const row = document.createElement("tr");
      const cell = document.createElement("td");
      cell.colSpan = 5;
      cell.className = "permission-empty";
      cell.textContent = "No evidence rows match this filter.";
      row.append(cell);
      body.append(row);
    }
  }

  function renderPermissionMounts(mounts) {
    const section = byId("permissionMountSection");
    const box = byId("permissionMounts");
    box.replaceChildren();
    if (!mounts.length) {
      section.classList.add("hidden");
      return;
    }
    section.classList.remove("hidden");
    mounts.forEach((mount) => {
      const card = document.createElement("article");
      card.className = "permission-mount-card";
      const title = document.createElement("strong");
      title.textContent = mount.container || "container";
      const line = document.createElement("code");
      line.textContent = `${mount.source} → ${mount.destination} · ${mount.read_write ? "rw" : "ro"} · ${mount.match_side}`;
      card.append(title, line);
      box.append(card);
    });
  }

  function renderPermissionScope(notes) {
    const list = byId("permissionScopeNotes");
    list.replaceChildren();
    notes.forEach((note) => {
      const item = document.createElement("li");
      item.textContent = note;
      list.append(item);
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", permissionStoryInit);
  } else {
    permissionStoryInit();
  }
})();
