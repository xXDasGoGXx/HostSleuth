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
  el.textContent = value ?? "";
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

function renderEvents(container, events, limit = null) {
  container.replaceChildren();
  const ordered = [...events].reverse();
  const list = limit ? ordered.slice(0, limit) : ordered;
  if (!list.length) {
    container.append(makeEmpty("No meaningful changes recorded yet. On a new install, this is expected—HostSleuth has a baseline and will add entries when something changes."));
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

function renderHost(snapshot) {
  const host = snapshot.host || {};
  const dockerMode = snapshot.mode === "docker";
  const services = snapshot.services || [];
  const containers = snapshot.containers || [];
  const listeners = snapshot.listeners || [];
  const failedServices = services.filter((service) => service.active === "failed").length;
  const runningContainers = containers.filter((container) => String(container.status || "").toLowerCase().startsWith("up")).length;

  text(byId("hostName"), host.hostname || "This host");
  text(byId("hostDetailName"), host.hostname || "Host details");
  const summaryParts = [host.os, host.kernel ? `kernel ${host.kernel}` : ""];
  if (dockerMode) summaryParts.push("Docker deployment");
  text(byId("hostSummary"), summaryParts.filter(Boolean).join(" · ") || "Host snapshot loaded.");
  text(byId("serviceCount"), dockerMode ? "—" : services.length);
  text(byId("containerCount"), containers.length);
  text(byId("listenerCount"), listeners.length);
  text(byId("changeCount"), state.events.length);

  const serviceNote = byId("serviceNote");
  serviceNote.classList.remove("good", "problem");
  if (dockerMode) {
    serviceNote.textContent = "Unavailable in Docker mode";
  } else if (failedServices > 0) {
    serviceNote.textContent = `${failedServices} failed`;
    serviceNote.classList.add("problem");
  } else {
    serviceNote.textContent = "No failed services";
    serviceNote.classList.add("good");
  }

  const containerNote = byId("containerNote");
  containerNote.textContent = containers.length ? `${runningContainers} running` : "No containers found";

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

  renderInterfaces(snapshot.interfaces || []);
  renderFilesystems(snapshot.filesystems || [], dockerMode);
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

function renderDiagnosis(diagnosis) {
  byId("diagnosisError").classList.add("hidden");
  const result = byId("diagnosisResult");
  result.classList.remove("hidden");

  text(byId("diagnosisConclusion"), diagnosisTitle(diagnosis));
  text(byId("diagnosisTarget"), diagnosis.target || "");
  text(byId("diagnosisExplanation"), diagnosis.conclusion || "HostSleuth completed the requested checks.");
  text(byId("diagnosisConfidence"), `${diagnosis.confidence || "unknown"} confidence`);

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
    // Build information is helpful but must never block the dashboard.
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
  } catch (error) {
    snapshotState.classList.add("bad");
    snapshotState.lastElementChild.textContent = "Could not load host snapshot";
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
