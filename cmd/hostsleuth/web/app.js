const state = {
  snapshot: null,
  events: [],
};

const checkNames = {
  target: "Target format",
  port: "Port",
  dns: "DNS resolution",
  route: "Network route",
  tcp: "TCP connection",
  "local-listener": "Local listener",
  docker: "Docker mapping",
  "docker-port": "Docker mapping",
  firewall: "Firewall evidence",
  systemd: "Service evidence",
  "systemd-failures": "Service evidence",
};

function byId(id) {
  return document.getElementById(id);
}

function text(el, value) {
  if (el) el.textContent = value ?? "";
}

function showView(name) {
  document.querySelectorAll("[data-view-panel]").forEach((panel) => {
    panel.classList.toggle("active", panel.dataset.viewPanel === name);
  });
  document.querySelectorAll("[data-view]").forEach((button) => {
    button.classList.toggle("active", button.dataset.view === name);
  });
  window.location.hash = name;
  window.scrollTo({ top: 0, behavior: "smooth" });
}

function formatTime(value) {
  if (!value) return "Unknown time";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString([], {
    year: "numeric",
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  });
}

function formatBytes(value) {
  const bytes = Number(value || 0);
  if (!bytes) return "—";
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1);
  return `${(bytes / Math.pow(1024, index)).toFixed(index > 2 ? 1 : 0)} ${units[index]}`;
}

function relativeSnapshotTime(value) {
  if (!value) return "Snapshot loaded";
  const date = new Date(value);
  const seconds = Math.max(0, Math.round((Date.now() - date.getTime()) / 1000));
  if (seconds < 60) return `Snapshot ${seconds}s ago`;
  const minutes = Math.round(seconds / 60);
  if (minutes < 60) return `Snapshot ${minutes}m ago`;
  return `Snapshot ${formatTime(value)}`;
}

function makeEmpty(message) {
  const el = document.createElement("div");
  el.className = "empty-state";
  el.textContent = message;
  return el;
}

function isRecent(value, hours) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return false;
  return Date.now() - date.getTime() <= hours * 60 * 60 * 1000;
}

function listenerScope(address) {
  const value = String(address || "").toLowerCase();
  if (value.includes("127.0.0.1:") || value.includes("[::1]:") || value.startsWith("::1:")) {
    return { key: "loopback", label: "Loopback" };
  }
  if (value.includes("0.0.0.0:") || value.includes("[::]:") || value.startsWith("*:") || value.startsWith(":::")) {
    return { key: "wildcard", label: "Wildcard" };
  }
  return { key: "specific", label: "Specific" };
}

function listenerTarget(address) {
  const value = String(address || "");
  const portMatch = value.match(/:(\d+)$/);
  if (!portMatch) return "";
  const port = portMatch[1];
  const scope = listenerScope(value).key;
  if (scope === "wildcard" || scope === "loopback") return `127.0.0.1:${port}`;
  return value;
}

function renderEvents(container, events, limit = null) {
  container.replaceChildren();
  const ordered = [...events].reverse();
  const list = limit ? ordered.slice(0, limit) : ordered;
  if (!list.length) {
    container.append(makeEmpty("No meaningful changes recorded yet. HostSleuth has a baseline and will add entries when something actually changes."));
    return;
  }

  list.forEach((event) => {
    const item = document.createElement("article");
    item.className = `timeline-item ${event.severity || ""}`;

    const marker = document.createElement("span");
    marker.className = "timeline-marker";

    const body = document.createElement("div");
    const summary = document.createElement("p");
    summary.className = "timeline-summary";
    summary.textContent = event.summary || "Recorded change";
    const meta = document.createElement("div");
    meta.className = "timeline-meta";
    meta.textContent = event.category || "change";
    body.append(summary, meta);

    const when = document.createElement("time");
    when.className = "timeline-time";
    when.textContent = formatTime(event.at);

    item.append(marker, body, when);
    container.append(item);
  });
}

function renderChangeSummary(events) {
  const container = byId("changeSummary");
  container.replaceChildren();
  if (!events.length) return;

  const counts = new Map();
  events.forEach((event) => {
    const key = event.category || "change";
    counts.set(key, (counts.get(key) || 0) + 1);
  });

  [...counts.entries()]
    .sort((a, b) => b[1] - a[1])
    .forEach(([category, count]) => {
      const chip = document.createElement("span");
      chip.className = "change-chip";
      const strong = document.createElement("strong");
      strong.textContent = count;
      chip.append(strong, document.createTextNode(` ${category}`));
      container.append(chip);
    });
}

function renderListenerList(container, listeners, limit = null) {
  container.replaceChildren();
  const ordered = [...listeners].sort((a, b) => String(a.address || "").localeCompare(String(b.address || "")));
  const list = limit ? ordered.slice(0, limit) : ordered;

  if (!list.length) {
    container.append(makeEmpty("No listening sockets are present in the current snapshot."));
    return;
  }

  list.forEach((listener) => {
    const scope = listenerScope(listener.address);
    const item = document.createElement("article");
    item.className = "surface-item";

    const badge = document.createElement("span");
    badge.className = `surface-scope ${scope.key}`;
    badge.textContent = scope.label;

    const main = document.createElement("div");
    main.className = "surface-main";
    const address = document.createElement("span");
    address.className = "surface-address";
    address.textContent = listener.address || "Unknown address";
    const meta = document.createElement("span");
    meta.className = "surface-meta";
    meta.textContent = [listener.protocol, listener.process].filter(Boolean).join(" · ") || "Listener evidence";
    main.append(address, meta);

    const action = document.createElement("button");
    action.className = "surface-action";
    action.type = "button";
    action.textContent = "Diagnose";
    const target = listenerTarget(listener.address);
    action.disabled = !target;
    action.addEventListener("click", () => {
      if (!target) return;
      byId("diagnoseTarget").value = target;
      showView("diagnose");
      byId("diagnoseTarget").focus();
    });

    item.append(badge, main, action);
    container.append(item);
  });

  if (limit && ordered.length > limit) {
    const more = document.createElement("div");
    more.className = "empty-state";
    more.textContent = `${ordered.length - limit} more listening sockets are available under Host evidence.`;
    container.append(more);
  }
}

function renderInterfaces(interfaces) {
  const container = byId("interfaceList");
  container.replaceChildren();
  if (!interfaces.length) {
    container.append(makeEmpty("No interface information is available."));
    return;
  }
  interfaces.forEach((item) => {
    const row = document.createElement("article");
    row.className = "stack-item";
    const head = document.createElement("div");
    head.className = "stack-item-head";
    const title = document.createElement("span");
    title.className = "stack-item-title";
    title.textContent = item.name || "Interface";
    const meta = document.createElement("span");
    meta.className = "stack-item-meta";
    meta.textContent = item.state || "unknown";
    head.append(title, meta);
    const body = document.createElement("div");
    body.className = "stack-item-body";
    body.textContent = (item.addresses || []).join(" · ") || "No addresses reported";
    row.append(head, body);
    container.append(row);
  });
}

function renderRoutes(routes) {
  const container = byId("routeList");
  container.replaceChildren();
  if (!routes.length) {
    container.append(makeEmpty("No route information is available."));
    return;
  }
  routes.forEach((route, index) => {
    const row = document.createElement("article");
    row.className = "stack-item";
    const head = document.createElement("div");
    head.className = "stack-item-head";
    const title = document.createElement("span");
    title.className = "stack-item-title";
    title.textContent = index === 0 ? "Routing table" : "Route";
    head.append(title);
    const body = document.createElement("div");
    body.className = "stack-item-body";
    body.textContent = route;
    row.append(head, body);
    container.append(row);
  });
}

function renderContainers(containers) {
  const container = byId("containerList");
  container.replaceChildren();
  if (!containers.length) {
    container.append(makeEmpty("No Docker containers are visible in the current snapshot."));
    return;
  }
  containers.forEach((item) => {
    const row = document.createElement("article");
    row.className = "stack-item";
    const head = document.createElement("div");
    head.className = "stack-item-head";
    const title = document.createElement("span");
    title.className = "stack-item-title";
    title.textContent = item.name || item.id || "Container";
    const meta = document.createElement("span");
    meta.className = "stack-item-meta";
    meta.textContent = item.status || "unknown";
    head.append(title, meta);
    const body = document.createElement("div");
    body.className = "stack-item-body";
    body.textContent = [item.image, item.ports ? `ports ${item.ports}` : "", item.networks ? `networks ${item.networks}` : ""].filter(Boolean).join(" · ") || "Container evidence";
    row.append(head, body);
    container.append(row);
  });
}

function renderFilesystems(filesystems, dockerMode = false) {
  const container = byId("filesystemList");
  container.replaceChildren();
  if (!filesystems.length) {
    const message = dockerMode
      ? "Host filesystem inventory is unavailable in Docker mode. Native installation provides full filesystem visibility."
      : "No filesystem information is available.";
    container.append(makeEmpty(message));
    return;
  }
  filesystems.forEach((item) => {
    const row = document.createElement("article");
    row.className = "stack-item";
    const head = document.createElement("div");
    head.className = "stack-item-head";
    const title = document.createElement("span");
    title.className = "stack-item-title";
    title.textContent = item.mount_point || "Filesystem";
    const meta = document.createElement("span");
    meta.className = "stack-item-meta";
    meta.textContent = item.filesystem_type || "unknown";
    head.append(title, meta);

    const body = document.createElement("div");
    body.className = "stack-item-body";
    const total = Number(item.total_bytes || 0);
    const available = Number(item.available_bytes || 0);
    const used = total > 0 ? Math.max(0, total - available) : 0;
    const capacity = total > 0 ? `${formatBytes(used)} used of ${formatBytes(total)}` : "Capacity unavailable";
    body.textContent = item.source ? `${capacity} · ${item.source}` : capacity;

    row.append(head, body);
    container.append(row);
  });
}

function renderAttention({ dockerMode, failedServices, exposedListeners, recentChanges }) {
  const strip = byId("attentionStrip");
  strip.classList.remove("problem", "caution");

  if (!dockerMode && failedServices > 0) {
    strip.classList.add("problem");
    text(byId("attentionTitle"), `${failedServices} failed ${failedServices === 1 ? "service" : "services"} need attention`);
    text(byId("attentionDetail"), "HostSleuth can use failed-service evidence during local reachability diagnosis.");
    return;
  }

  if (recentChanges > 0 && exposedListeners > 0) {
    strip.classList.add("caution");
    text(byId("attentionTitle"), `${recentChanges} meaningful ${recentChanges === 1 ? "change" : "changes"} in 24h while ${exposedListeners} sockets listen beyond loopback`);
    text(byId("attentionDetail"), "This is context, not an alarm. The timeline and reachability surface show exactly what HostSleuth observed.");
    return;
  }

  if (recentChanges > 0) {
    strip.classList.add("caution");
    text(byId("attentionTitle"), `${recentChanges} meaningful ${recentChanges === 1 ? "change" : "changes"} recorded in the last 24 hours`);
    text(byId("attentionDetail"), "Open Changes to see the newest differences first; routine Docker uptime churn is suppressed.");
    return;
  }

  if (exposedListeners > 0) {
    text(byId("attentionTitle"), `${exposedListeners} listening ${exposedListeners === 1 ? "socket is" : "sockets are"} bound beyond loopback`);
    text(byId("attentionDetail"), "That is not proof of remote reachability. Routing and firewall state still matter; diagnose any endpoint to test the path.");
    return;
  }

  text(byId("attentionTitle"), "No obvious local issue stands out in the current evidence");
  text(byId("attentionDetail"), "The latest snapshot has no failed native services, recent changes, or listeners beyond loopback that HostSleuth can currently see.");
}

function renderHost(snapshot) {
  const host = snapshot.host || {};
  const dockerMode = snapshot.mode === "docker";
  const services = snapshot.services || [];
  const containers = snapshot.containers || [];
  const listeners = snapshot.listeners || [];
  const interfaces = snapshot.interfaces || [];
  const failedServices = services.filter((service) => service.active === "failed").length;
  const runningContainers = containers.filter((container) => String(container.status || "").toLowerCase().startsWith("up")).length;
  const exposedListeners = listeners.filter((listener) => listenerScope(listener.address).key !== "loopback").length;
  const recentChanges = state.events.filter((event) => isRecent(event.at, 24)).length;

  text(byId("hostName"), host.hostname || "This host");
  text(byId("hostDetailName"), host.hostname || "Host details");
  text(byId("deploymentBadge"), dockerMode ? "Docker view" : "Native view");

  const summaryParts = [host.os, host.kernel ? `kernel ${host.kernel}` : ""];
  text(byId("hostSummary"), summaryParts.filter(Boolean).join(" · ") || "Host snapshot loaded.");

  const storyParts = [
    `${listeners.length} ${listeners.length === 1 ? "listener" : "listeners"}`,
    `${interfaces.length} ${interfaces.length === 1 ? "interface" : "interfaces"}`,
    containers.length ? `${runningContainers}/${containers.length} containers running` : "no visible containers",
    recentChanges ? `${recentChanges} meaningful changes in 24h` : "no meaningful changes in 24h",
  ];
  text(byId("hostStoryLine"), `Right now HostSleuth sees ${storyParts.join(" · ")}.`);

  text(byId("exposedListenerCount"), exposedListeners);
  text(byId("exposedListenerNote"), exposedListeners ? `${listeners.length - exposedListeners} loopback-only` : "Loopback-only surface");
  text(byId("failedServiceCount"), dockerMode ? "—" : failedServices);
  text(byId("failedServiceNote"), dockerMode ? "Unavailable in Docker mode" : failedServices ? `${failedServices} need attention` : "No failed services");
  byId("failedServiceNote").className = `signal-note ${!dockerMode && failedServices ? "problem" : !dockerMode ? "good" : ""}`;
  text(byId("containerCount"), containers.length);
  text(byId("containerNote"), containers.length ? `${runningContainers} running` : "No containers found");
  text(byId("change24hCount"), recentChanges);
  text(byId("change24hNote"), recentChanges ? `${state.events.length} retained total` : "No recent meaningful changes");
  byId("change24hNote").className = `signal-note ${recentChanges ? "warning" : "good"}`;

  renderAttention({ dockerMode, failedServices, exposedListeners, recentChanges });

  const snapshotState = byId("snapshotState");
  snapshotState.classList.remove("bad");
  snapshotState.classList.add("good");
  snapshotState.lastElementChild.textContent = relativeSnapshotTime(snapshot.captured_at);

  const facts = [
    ["Deployment", dockerMode ? "Docker · reduced host visibility" : "Native Linux"],
    ["Operating system", host.os || "—"],
    ["Kernel", host.kernel || "—"],
    ["Architecture", host.architecture || "—"],
    ["CPU", host.cpu_count ? `${host.cpu_count} logical CPUs` : "—"],
    ["Memory", host.memory_total || "—"],
    ["Uptime", host.uptime || "—"],
    ["Snapshot", formatTime(snapshot.captured_at)],
  ];
  const hostFacts = byId("hostFacts");
  hostFacts.replaceChildren();
  facts.forEach(([label, value]) => {
    const item = document.createElement("div");
    item.className = "detail-item";
    const key = document.createElement("span");
    key.className = "detail-label";
    key.textContent = label;
    const val = document.createElement("div");
    val.className = "detail-value";
    val.textContent = value;
    item.append(key, val);
    hostFacts.append(item);
  });

  renderListenerList(byId("overviewListeners"), listeners, 6);
  renderListenerList(byId("listenerList"), listeners);
  renderInterfaces(interfaces);
  renderRoutes(snapshot.routes || []);
  renderContainers(containers);
  renderFilesystems(snapshot.filesystems || [], dockerMode);
}

function diagnosisTitle(diagnosis) {
  const conclusion = String(diagnosis.conclusion || "").toLowerCase();
  if (conclusion === "target is reachable") return "Connection succeeded";
  if (conclusion === "name resolution failed") return "Name resolution failed";
  if (conclusion === "invalid target") return "Enter a host and port";
  if (conclusion === "invalid port") return "Port must be a number";
  if (conclusion.includes("no process is listening") || conclusion.includes("no process appears to be listening")) return "Nothing is listening there";
  if (conclusion.includes("kernel route lookup reports the destination unreachable")) return "No usable route to the target";
  if (conclusion.includes("snapshot shows a listener")) return "A listener exists, but the connection failed";
  if (conclusion.includes("remote tcp connection failed")) return "Connection failed";
  return "Diagnosis complete";
}

function stageStatus(checks, names) {
  const relevant = checks.filter((check) => names.includes(check.name));
  if (!relevant.length) return { key: "unknown", label: "not needed" };
  if (relevant.some((check) => check.status === "fail")) return { key: "fail", label: "failed" };
  if (relevant.some((check) => check.status === "pass")) return { key: "pass", label: "passed" };
  return { key: "unknown", label: "unknown" };
}

function renderDiagnosisPath(diagnosis) {
  const checks = diagnosis.checks || [];
  const stages = [
    ["Name", ["target", "port", "dns"]],
    ["Route", ["route"]],
    ["TCP", ["tcp"]],
    ["Local bind", ["local-listener", "docker", "docker-port"]],
    ["Candidates", ["firewall", "systemd", "systemd-failures"]],
  ];

  const container = byId("diagnosisPath");
  container.replaceChildren();
  stages.forEach(([label, names]) => {
    const status = stageStatus(checks, names);
    const stage = document.createElement("div");
    stage.className = `path-stage ${status.key}`;
    const name = document.createElement("span");
    name.textContent = label;
    const value = document.createElement("strong");
    value.textContent = status.label;
    stage.append(name, value);
    container.append(stage);
  });
}

function renderDiagnosisContext() {
  const container = byId("diagnosisRecentEvents");
  container.replaceChildren();
  const events = [...state.events].reverse().slice(0, 3);
  if (!events.length) {
    container.append(makeEmpty("No meaningful host changes are currently retained."));
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

function renderDiagnosis(diagnosis) {
  byId("diagnosisError").classList.add("hidden");
  const result = byId("diagnosisResult");
  result.classList.remove("hidden");

  text(byId("diagnosisConclusion"), diagnosisTitle(diagnosis));
  text(byId("diagnosisTarget"), diagnosis.target || "");
  text(byId("diagnosisExplanation"), diagnosis.conclusion || "HostSleuth completed the requested checks.");
  text(byId("diagnosisConfidence"), `${diagnosis.confidence || "unknown"} confidence`);
  renderDiagnosisPath(diagnosis);
  renderDiagnosisContext();

  const list = byId("diagnosisChecks");
  list.replaceChildren();
  (diagnosis.checks || []).forEach((check) => {
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
    list.append(details);
  });
}

async function loadAbout() {
  try {
    const response = await fetch("/api/about", { cache: "no-store" });
    if (!response.ok) return;
    const about = await response.json();
    const version = about.version || "";
    const revision = about.revision || "";
    text(byId("appVersion"), [version, revision ? `(${revision})` : ""].filter(Boolean).join(" "));
  } catch (_) {
    // Build information is useful but must never block the dashboard.
  }
}

async function loadDashboard() {
  const snapshotState = byId("snapshotState");
  try {
    const [snapshotResponse, eventsResponse] = await Promise.all([
      fetch("/api/snapshot", { cache: "no-store" }),
      fetch("/api/events", { cache: "no-store" }),
    ]);
    if (!snapshotResponse.ok) throw new Error(`snapshot request failed (${snapshotResponse.status})`);
    if (!eventsResponse.ok) throw new Error(`events request failed (${eventsResponse.status})`);

    state.snapshot = await snapshotResponse.json();
    state.events = await eventsResponse.json();
    renderHost(state.snapshot);
    renderEvents(byId("overviewEvents"), state.events, 5);
    renderEvents(byId("allEvents"), state.events);
    renderChangeSummary(state.events);
  } catch (error) {
    snapshotState.classList.add("bad");
    snapshotState.lastElementChild.textContent = "Could not load host snapshot";
    byId("overviewListeners").replaceChildren(makeEmpty("HostSleuth could not load the reachability surface."));
    byId("overviewEvents").replaceChildren(makeEmpty("HostSleuth could not load recent changes."));
    byId("allEvents").replaceChildren(makeEmpty("HostSleuth could not load recent changes."));
  }
}

async function runDiagnosis(target) {
  const button = byId("diagnoseButton");
  const result = byId("diagnosisResult");
  const errorPanel = byId("diagnosisError");
  button.disabled = true;
  button.textContent = "Diagnosing…";
  result.classList.add("hidden");
  errorPanel.classList.add("hidden");

  try {
    const response = await fetch(`/api/diagnose?target=${encodeURIComponent(target)}`, { cache: "no-store" });
    if (!response.ok) {
      const message = await response.text();
      throw new Error(message || `request failed (${response.status})`);
    }
    renderDiagnosis(await response.json());
  } catch (error) {
    text(byId("diagnosisErrorText"), error.message || "Unknown error");
    errorPanel.classList.remove("hidden");
  } finally {
    button.disabled = false;
    button.textContent = "Diagnose";
  }
}

document.querySelectorAll("[data-view]").forEach((button) => {
  button.addEventListener("click", () => showView(button.dataset.view));
});

document.querySelectorAll("[data-open-view]").forEach((button) => {
  button.addEventListener("click", () => showView(button.dataset.openView));
});

byId("diagnoseForm").addEventListener("submit", (event) => {
  event.preventDefault();
  const target = byId("diagnoseTarget").value.trim();
  if (target) runDiagnosis(target);
});

const initialView = window.location.hash.replace("#", "");
if (["overview", "diagnose", "changes", "host"].includes(initialView)) {
  showView(initialView);
}

loadAbout();
loadDashboard();
