const state = {
  snapshot: null,
  events: [],
  about: null,
  lastDiagnosis: null,
  changeFilter: { category: "", warningsOnly: false },
  showRoutineListeners: false,
  listenerQuery: "",
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

function genericStatusLabel(value) {
  const status = String(value || "unknown").toLowerCase();
  if (status === "pass") return "PASS";
  if (status === "warn" || status === "warning") return "WARNING";
  if (status === "fail" || status === "failed") return "FAILED";
  if (status === "unknown") return "UNKNOWN";
  return String(value || "unknown").toUpperCase();
}

async function copyBoundedText(payload) {
  const value = String(payload || "").slice(0, 12000);
  if (!value) throw new Error("nothing to copy");

  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value);
      return;
    } catch (_) {
      // LAN HTTP may not expose the secure Clipboard API; use the local fallback.
    }
  }

  const textarea = document.createElement("textarea");
  textarea.value = value;
  textarea.setAttribute("readonly", "");
  textarea.style.position = "fixed";
  textarea.style.left = "-9999px";
  textarea.style.top = "0";
  document.body.append(textarea);
  textarea.select();
  textarea.setSelectionRange(0, textarea.value.length);
  const copied = document.execCommand("copy");
  textarea.remove();
  if (!copied) throw new Error("copy command was refused");
}

function activeViewName() {
  const active = document.querySelector("[data-view-panel].active");
  return active?.dataset.viewPanel || "overview";
}

function syncViewAccessibility(name = activeViewName()) {
  document.querySelectorAll("[data-view]").forEach((button) => {
    const view = button.dataset.view;
    const active = view === name;
    if (!button.id) button.id = `nav-${view}`;
    button.setAttribute("role", "tab");
    button.setAttribute("aria-controls", `view-${view}`);
    button.setAttribute("aria-selected", String(active));
    button.tabIndex = active ? 0 : -1;
  });

  document.querySelectorAll("[data-view-panel]").forEach((panel) => {
    const view = panel.dataset.viewPanel;
    const active = view === name;
    if (!panel.id) panel.id = `view-${view}`;
    panel.setAttribute("role", "tabpanel");
    panel.setAttribute("aria-labelledby", `nav-${view}`);
    panel.setAttribute("aria-hidden", String(!active));
    panel.hidden = !active;
  });
}

function showView(name) {
  if (!document.querySelector(`[data-view-panel="${CSS.escape(name)}"]`)) name = "overview";
  document.querySelectorAll("[data-view-panel]").forEach((panel) => {
    panel.classList.toggle("active", panel.dataset.viewPanel === name);
  });
  document.querySelectorAll("[data-view]").forEach((button) => {
    button.classList.toggle("active", button.dataset.view === name);
  });
  syncViewAccessibility(name);
  if (window.location.hash !== `#${name}`) window.location.hash = name;
  const reducedMotion = window.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches;
  window.scrollTo({ top: 0, behavior: reducedMotion ? "auto" : "smooth" });
  document.dispatchEvent(new CustomEvent("hostsleuth:view", { detail: name }));
}

// Every view, including ones added by later scripts, is reachable by link,
// and the browser back/forward buttons move between views.
function routeFromHash() {
  const name = decodeURIComponent(window.location.hash.replace("#", ""));
  if (!name) return;
  if (name === activeViewName() && document.querySelector(`[data-view-panel="${CSS.escape(name)}"].active`)) return;
  showView(name);
}

function installNavigationAccessibility() {
  const nav = document.querySelector(".nav-tabs");
  if (!nav) return;

  nav.setAttribute("role", "tablist");
  nav.addEventListener("keydown", (event) => {
    if (!["ArrowLeft", "ArrowRight", "Home", "End"].includes(event.key)) return;
    const tabs = [...nav.querySelectorAll("[data-view]")].filter((button) => !button.disabled);
    if (!tabs.length) return;

    const current = Math.max(0, tabs.indexOf(document.activeElement));
    let next = current;
    if (event.key === "ArrowLeft") next = (current - 1 + tabs.length) % tabs.length;
    if (event.key === "ArrowRight") next = (current + 1) % tabs.length;
    if (event.key === "Home") next = 0;
    if (event.key === "End") next = tabs.length - 1;

    event.preventDefault();
    const target = tabs[next];
    showView(target.dataset.view);
    target.focus();
  });

  const observer = new MutationObserver(() => syncViewAccessibility(activeViewName()));
  observer.observe(nav, { childList: true });
  syncViewAccessibility(activeViewName());
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

function formatClock(value) {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return formatTime(value);
  return date.toLocaleTimeString([], { hour: "numeric", minute: "2-digit" });
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

function endpointParts(raw) {
  const value = String(raw || "").trim();
  if (!value) return { host: "", port: "" };
  if (value.startsWith("[")) {
    const end = value.indexOf("]");
    if (end > 1) {
      return {
        host: value.slice(1, end),
        port: value.slice(end + 1).replace(/^:/, ""),
      };
    }
  }
  const split = value.lastIndexOf(":");
  if (split > 0 && value.indexOf(":") === split) {
    return { host: value.slice(0, split), port: value.slice(split + 1) };
  }
  return { host: value, port: "" };
}

function openDiagnosisForTarget(target) {
  if (!target) return;
  const input = byId("diagnoseTarget");
  if (input) input.value = target;
  if (window.hostSleuthRememberTarget) window.hostSleuthRememberTarget(target);
  showView("diagnose");
  runDiagnosis(target);
}

function openDNSForTarget(target) {
  if (!target) return;
  const parts = endpointParts(target);
  const name = byId("dnsDetectiveName");
  const port = byId("dnsDetectivePort");
  if (name) name.value = parts.host;
  if (port && parts.port) port.value = parts.port;
  if (window.hostSleuthRememberTarget) window.hostSleuthRememberTarget(target);
  showView("dns");
  if (name) name.focus();
}

function openContractForTarget(target) {
  if (!target) return;
  const input = byId("contractTarget");
  if (input) input.value = target;
  if (window.hostSleuthRememberTarget) window.hostSleuthRememberTarget(target);
  showView("contracts");
  if (input) input.focus();
}

function listenerScope(address) {
  const value = String(address || "").toLowerCase();
  const parts = splitListenerAddress(value);
  const host = parts?.host ?? "";
  if (host === "::1" || /^127\./.test(host) || parts?.zone === "lo") {
    return { key: "loopback", label: "Loopback" };
  }
  if (host === "0.0.0.0" || host === "::" || host === "*" || value.startsWith(":::")) {
    return { key: "wildcard", label: "Wildcard" };
  }
  return { key: "specific", label: "Specific" };
}

const ROUTINE_DISCOVERY_PORTS = { udp: [1900, 3702, 5353, 5355], tcp: [5355, 5357] };
const FIXED_HIGH_UDP_PORTS = [41641, 51820];

function splitListenerAddress(address) {
  const value = String(address || "").trim();
  const split = value.lastIndexOf(":");
  if (split <= 0) return null;
  const port = Number(value.slice(split + 1));
  if (!Number.isInteger(port)) return null;
  let host = value.slice(0, split);
  let zone = "";
  const pct = host.indexOf("%");
  if (pct >= 0) {
    zone = host.slice(pct + 1).replace(/[[\]]/g, "");
    host = host.slice(0, pct);
  }
  host = host.replace(/^\[/, "").replace(/\]$/, "").toLowerCase();
  return { host, zone, port };
}

// Interface and client churn: link-local and multicast sockets, discovery
// binds on one interface address, and ephemeral UDP client sockets. They
// are still listed under Host evidence on request.
function isRoutineListener(listener) {
  const protocol = String(listener?.protocol || "").toLowerCase();
  const parts = splitListenerAddress(listener?.address);
  if (!parts) return false;
  const { host, zone, port } = parts;
  const wildcard = host === "*" || host === "0.0.0.0" || host === "::";
  if (!wildcard && !host.includes(":") && !/^\d{1,3}(\.\d{1,3}){3}$/.test(host)) return false;
  const linkLocalOrMulticast =
    /^fe[89ab][0-9a-f]:/.test(host) || /^ff[0-9a-f]{2}:/.test(host) || /^169\.254\./.test(host) ||
    /^(22[4-9]|23[0-9])\./.test(host);
  if (!wildcard && linkLocalOrMulticast) return true;
  const loopback = host === "::1" || /^127\./.test(host);
  if (!wildcard && (ROUTINE_DISCOVERY_PORTS[protocol] || []).includes(port) && (zone || !loopback)) return true;
  return protocol === "udp" && port >= 32768 && port <= 60999 && !FIXED_HIGH_UDP_PORTS.includes(port);
}

function listenerPort(listener) {
  return splitListenerAddress(listener?.address)?.port ?? 0;
}

// Most useful first: reachable beyond loopback, TCP before UDP, low ports.
function compareListeners(a, b) {
  const order = { wildcard: 0, specific: 1, loopback: 2 };
  const rank = (listener) => order[listenerScope(listener.address).key];
  return rank(a) - rank(b) ||
    String(a.protocol).localeCompare(String(b.protocol)) ||
    listenerPort(a) - listenerPort(b) ||
    String(a.address || "").localeCompare(String(b.address || ""));
}

// A hint, not evidence: the conventional service for a well-known port.
const WELL_KNOWN_PORTS = {
  tcp: { 21: "FTP", 22: "SSH", 25: "SMTP", 53: "DNS", 80: "HTTP", 110: "POP3", 111: "RPC portmapper", 139: "NetBIOS/SMB",
    143: "IMAP", 443: "HTTPS", 445: "SMB", 465: "SMTPS", 587: "SMTP submission", 631: "IPP printing", 993: "IMAPS",
    995: "POP3S", 2049: "NFS", 3306: "MySQL", 5432: "PostgreSQL", 5900: "VNC", 6379: "Redis", 8080: "HTTP alt", 8443: "HTTPS alt" },
  udp: { 53: "DNS", 67: "DHCP", 68: "DHCP client", 69: "TFTP", 111: "RPC portmapper", 123: "NTP", 137: "NetBIOS", 161: "SNMP",
    323: "chrony", 500: "IKE", 514: "syslog", 1900: "SSDP", 2049: "NFS", 3702: "WS-Discovery", 4500: "IPsec NAT-T",
    5353: "mDNS", 5355: "LLMNR", 41641: "Tailscale", 51820: "WireGuard" },
};

function wellKnownService(protocol, port) {
  return WELL_KNOWN_PORTS[String(protocol || "").toLowerCase()]?.[port] || "";
}

// One row per dual-stack wildcard service instead of 0.0.0.0:P and [::]:P.
function mergeDualStack(listeners) {
  const byKey = new Map();
  const out = [];
  listeners.forEach((listener) => {
    const parts = splitListenerAddress(listener.address);
    const wildcard = parts && (parts.host === "0.0.0.0" || parts.host === "::" || parts.host === "*");
    if (!wildcard) {
      out.push(listener);
      return;
    }
    const key = `${listener.protocol} ${parts.port}`;
    const existing = byKey.get(key);
    if (existing) {
      existing.dualStack = true;
      existing.address = `*:${parts.port}`;
      if (!existing.process && listener.process) existing.process = listener.process;
      return;
    }
    const copy = { ...listener };
    byKey.set(key, copy);
    out.push(copy);
  });
  return out;
}

function containerHealth(container) {
  const status = String(container?.status || "").toLowerCase();
  if (status.includes("(unhealthy)")) return "unhealthy";
  if (status.startsWith("restarting")) return "restarting";
  if (status.startsWith("exited") || status.startsWith("dead")) return "stopped";
  if (status.includes("(paused)") || status.startsWith("paused")) return "paused";
  if (status.startsWith("up")) return "running";
  return "other";
}

function formatMemory(value) {
  const match = String(value || "").match(/^(\d+)\s*kB$/i);
  if (!match) return value || "—";
  return `${(Number(match[1]) / 1048576).toFixed(1)} GiB`;
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

  let lastDay = "";
  list.forEach((event) => {
    if (!limit) {
      const date = new Date(event.at);
      const day = Number.isNaN(date.getTime()) ? "" : date.toLocaleDateString([], { weekday: "long", month: "short", day: "numeric" });
      if (day && day !== lastDay) {
        const heading = document.createElement("h3");
        heading.className = "timeline-day";
        heading.textContent = day;
        container.append(heading);
        lastDay = day;
      }
    }
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
    when.textContent = limit ? formatTime(event.at) : formatClock(event.at);
    when.title = formatTime(event.at);

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
  const warnings = events.filter((event) => event.severity === "warning").length;
  const filter = state.changeFilter;

  const chip = (label, count, active, onClick) => {
    const button = document.createElement("button");
    button.type = "button";
    button.className = `change-chip${active ? " active" : ""}`;
    button.setAttribute("aria-pressed", String(active));
    const strong = document.createElement("strong");
    strong.textContent = count;
    button.append(strong, document.createTextNode(` ${label}`));
    button.addEventListener("click", onClick);
    return button;
  };

  container.append(chip("all", events.length, !filter.category && !filter.warningsOnly, () => {
    state.changeFilter = { category: "", warningsOnly: false };
    renderChangesView();
  }));
  [...counts.entries()]
    .sort((a, b) => b[1] - a[1])
    .forEach(([category, count]) => {
      container.append(chip(category, count, filter.category === category, () => {
        state.changeFilter.category = filter.category === category ? "" : category;
        renderChangesView();
      }));
    });
  if (warnings) {
    const warn = chip(warnings === 1 ? "warning" : "warnings", warnings, filter.warningsOnly, () => {
      state.changeFilter.warningsOnly = !filter.warningsOnly;
      renderChangesView();
    });
    warn.classList.add("warning-chip");
    container.append(warn);
  }
}

function renderChangesView() {
  const filter = state.changeFilter;
  const events = state.events.filter((event) =>
    (!filter.category || (event.category || "change") === filter.category) &&
    (!filter.warningsOnly || event.severity === "warning"));
  renderChangeSummary(state.events);
  renderEvents(byId("allEvents"), events);
}

function renderListenerList(container, listeners, limit = null) {
  container.replaceChildren();
  const ordered = mergeDualStack(listeners).sort(compareListeners);
  const list = limit ? ordered.slice(0, limit) : ordered;

  if (!list.length) {
    container.append(makeEmpty("No listening sockets are present in the current snapshot."));
    return;
  }

  list.forEach((listener) => container.append(listenerItem(listener)));

  if (limit && ordered.length > limit) {
    const more = document.createElement("div");
    more.className = "empty-state";
    more.textContent = `${ordered.length - limit} more listening sockets are under Host evidence.`;
    container.append(more);
  }
}

function listenerItem(listener) {
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
  const hint = wellKnownService(listener.protocol, listenerPort(listener));
  meta.textContent = [
    listener.protocol,
    listener.dualStack ? "IPv4 + IPv6" : "",
    listener.process || (hint ? `usually ${hint}` : ""),
  ].filter(Boolean).join(" · ") || "Listener evidence";
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
  return item;
}

// Host evidence groups sockets by who can reach them. Loopback-only
// sockets cannot be reached from other machines, so they start folded.
const LISTENER_GROUPS = [
  { key: "wildcard", title: "Every interface", note: "Bound to all addresses: reachable on any network this host is on, unless a firewall says otherwise." },
  { key: "specific", title: "One address", note: "Reachable only on the address shown." },
  { key: "loopback", title: "This host only", note: "Loopback: not reachable from other machines.", fold: true },
];

function renderGroupedListeners(container, listeners, expandAll) {
  container.replaceChildren();
  const ordered = mergeDualStack(listeners).sort(compareListeners);
  if (!ordered.length) {
    container.append(makeEmpty("No listening sockets match."));
    return;
  }
  LISTENER_GROUPS.forEach((group) => {
    const members = ordered.filter((listener) => listenerScope(listener.address).key === group.key);
    if (!members.length) return;
    const wrap = document.createElement(group.fold ? "details" : "section");
    wrap.className = "surface-group";
    if (group.fold && expandAll) wrap.open = true;
    const head = document.createElement(group.fold ? "summary" : "div");
    head.className = "surface-group-head";
    const title = document.createElement("strong");
    title.textContent = `${group.title} · ${members.length}`;
    const note = document.createElement("span");
    note.textContent = group.note;
    head.append(title, note);
    const body = document.createElement("div");
    body.className = "surface-list full-surface";
    members.forEach((listener) => body.append(listenerItem(listener)));
    wrap.append(head, body);
    container.append(wrap);
  });
}

function isVirtualInterface(name) {
  return /^(veth|vnet|tap|cali|flannel|cni|lxc[^b]|virbr\d+-nic)/.test(String(name || ""));
}

function interfaceRow(item) {
  const row = document.createElement("article");
  row.className = "stack-item";
  const head = document.createElement("div");
  head.className = "stack-item-head";
  const title = document.createElement("span");
  title.className = "stack-item-title";
  title.textContent = item.name || "Interface";
  const meta = document.createElement("span");
  meta.className = `stack-item-meta iface-state ${item.state === "up" ? "up" : "down"}`;
  meta.textContent = item.state || "unknown";
  head.append(title, meta);
  const body = document.createElement("div");
  body.className = "stack-item-body";
  body.textContent = (item.addresses || []).join(" · ") || "No addresses reported";
  row.append(head, body);
  return row;
}

function renderHostListeners() {
  const container = byId("listenerList");
  if (!container) return;
  const all = state.snapshot?.listeners || [];
  const routine = all.filter(isRoutineListener).length;
  const query = state.listenerQuery.trim().toLowerCase();
  const visible = all.filter((listener) =>
    (state.showRoutineListeners || !isRoutineListener(listener)) &&
    (!query || `${listener.protocol} ${listener.address} ${listener.process || ""}`.toLowerCase().includes(query)));

  let tools = byId("listenerTools");
  if (!tools) {
    tools = document.createElement("div");
    tools.id = "listenerTools";
    tools.className = "listener-tools";
    const search = document.createElement("input");
    search.type = "search";
    search.id = "listenerSearch";
    search.placeholder = "Filter by port, address or process";
    search.setAttribute("aria-label", "Filter listening sockets");
    search.addEventListener("input", () => {
      state.listenerQuery = search.value;
      renderHostListeners();
    });
    const toggle = document.createElement("label");
    toggle.className = "listener-routine-toggle";
    const box = document.createElement("input");
    box.type = "checkbox";
    box.id = "listenerShowRoutine";
    box.addEventListener("change", () => {
      state.showRoutineListeners = box.checked;
      renderHostListeners();
    });
    toggle.append(box, document.createTextNode(" Show routine sockets"));
    const count = document.createElement("span");
    count.id = "listenerCount";
    count.className = "listener-count";
    tools.append(search, toggle, count);
    container.before(tools);
  }
  text(byId("listenerCount"), `${visible.length} shown · ${routine} routine ${state.showRoutineListeners ? "included" : "hidden"}`);
  byId("listenerShowRoutine").title = "Link-local and multicast sockets, discovery binds on single interfaces, and ephemeral UDP client sockets";
  container.classList.remove("surface-list", "full-surface");
  renderGroupedListeners(container, visible, Boolean(query));
}

function renderInterfaces(interfaces) {
  const container = byId("interfaceList");
  container.replaceChildren();
  if (!interfaces.length) {
    container.append(makeEmpty("No interface information is available."));
    return;
  }
  const byName = (a, b) => String(a.name).localeCompare(String(b.name), undefined, { numeric: true });
  const addressed = (item) => (item.addresses || []).some((address) => !/^fe80:/i.test(address));
  const networks = interfaces.filter((item) => isContainerNetwork(item.name)).sort(byName);
  const links = interfaces.filter((item) => isVirtualInterface(item.name)).sort(byName);
  const rest = interfaces.filter((item) => !isContainerNetwork(item.name) && !isVirtualInterface(item.name));
  const primary = rest.filter(addressed).sort((a, b) => (a.name === "lo") - (b.name === "lo") || byName(a, b));
  const bare = rest.filter((item) => !addressed(item)).sort(byName);
  primary.forEach((item) => container.append(interfaceRow(item)));
  const down = bare.filter((item) => item.state !== "up").length;
  appendFold(container, bare, `${bare.length} without an IP address (bridge ports, VLANs${down ? `, ${down} down` : ""})`, interfaceRow);
  appendFold(container, networks, `${networks.length} container and VM ${networks.length === 1 ? "network" : "networks"} (Docker, Podman, LXC, libvirt)`, interfaceRow);
  appendFold(container, links, `${links.length} container and VM link ${links.length === 1 ? "interface" : "interfaces"} (veth, vnet, tap)`, interfaceRow);
}

function isContainerNetwork(name) {
  return /^(br-[0-9a-f]{12}$|docker\d+$|podman\d+$|lxcbr\d+$|virbr\d+$|cni\d*$|flannel|cali|cilium|kube-)/.test(String(name || ""));
}

function appendFold(container, items, label, makeRow) {
  if (!items.length) return;
  const details = document.createElement("details");
  details.className = "stack-fold";
  const summary = document.createElement("summary");
  summary.textContent = label;
  details.append(summary);
  items.forEach((item) => details.append(makeRow(item)));
  container.append(details);
}

function renderRoutes(routes) {
  const container = byId("routeList");
  container.replaceChildren();
  if (!routes.length) {
    container.append(makeEmpty("No route information is available."));
    return;
  }
  const device = (route) => (String(route).match(/\bdev\s+(\S+)/) || [])[1] || "";
  const containerRoutes = routes.filter((route) => isContainerNetwork(device(route)));
  routes.filter((route) => !isContainerNetwork(device(route))).forEach((route) => container.append(routeRow(route)));
  appendFold(container, containerRoutes, `${containerRoutes.length} container and VM network ${containerRoutes.length === 1 ? "route" : "routes"}`, routeRow);
}

function routeRow(route) {
  const row = document.createElement("article");
  row.className = "stack-item";
  const head = document.createElement("div");
  head.className = "stack-item-head";
  const title = document.createElement("span");
  title.className = "stack-item-title";
  const [destination, ...rest] = String(route || "").trim().split(/\s+/);
  title.textContent = destination === "default" ? "Default route" : destination || "Route";
  head.append(title);
  const body = document.createElement("div");
  body.className = "stack-item-body";
  body.textContent = rest.join(" ") || route;
  row.append(head, body);
  return row;
}

function renderContainers(containers) {
  const container = byId("containerList");
  container.replaceChildren();
  if (!containers.length) {
    container.append(makeEmpty("No Docker containers are visible in the current snapshot."));
    return;
  }
  const healthRank = { unhealthy: 0, restarting: 0, stopped: 1, paused: 1, other: 2, running: 3 };
  const ordered = [...containers].sort((a, b) =>
    healthRank[containerHealth(a)] - healthRank[containerHealth(b)] || String(a.name).localeCompare(String(b.name)));
  ordered.forEach((item) => {
    const health = containerHealth(item);
    const row = document.createElement("article");
    row.className = `stack-item container-${health}`;
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

function renderAttention({ dockerMode, failedServices, exposedListeners, recentChanges, troubledContainers = [] }) {
  const strip = byId("attentionStrip");
  const action = byId("attentionAction");
  strip.classList.remove("problem", "caution");

  const setAction = (label, view) => {
    if (!action) return;
    action.textContent = label;
    action.dataset.targetView = view;
  };

  if (!dockerMode && failedServices > 0) {
    strip.classList.add("problem");
    text(byId("attentionTitle"), `${failedServices} failed ${failedServices === 1 ? "service" : "services"} need attention`);
    text(byId("attentionDetail"), "HostSleuth can use failed-service evidence during local reachability diagnosis.");
    setAction("Inspect services", "diagnose");
    return;
  }

  if (troubledContainers.length) {
    strip.classList.add("problem");
    const names = troubledContainers.map((item) => `${item.name || item.id} (${containerHealth(item)})`);
    const shown = names.slice(0, 3).join(", ") + (names.length > 3 ? ` and ${names.length - 3} more` : "");
    text(byId("attentionTitle"), `${troubledContainers.length} ${troubledContainers.length === 1 ? "container is" : "containers are"} not healthy: ${shown}`);
    text(byId("attentionDetail"), "Docker reports this right now. Changes shows when it started; Host evidence lists every container with problems first.");
    setAction("Inspect containers", "containers");
    return;
  }

  if (recentChanges > 0 && exposedListeners > 0) {
    strip.classList.add("caution");
    text(byId("attentionTitle"), `${recentChanges} meaningful ${recentChanges === 1 ? "change" : "changes"} in 24h · ${exposedListeners} service ${exposedListeners === 1 ? "socket listens" : "sockets listen"} beyond loopback`);
    text(byId("attentionDetail"), "This is context, not an alarm. The timeline and reachability surface show exactly what HostSleuth observed.");
    setAction("Review changes", "changes");
    return;
  }

  if (recentChanges > 0) {
    strip.classList.add("caution");
    text(byId("attentionTitle"), `${recentChanges} meaningful ${recentChanges === 1 ? "change" : "changes"} recorded in the last 24 hours`);
    text(byId("attentionDetail"), "Open Changes to see the newest differences first; routine Docker uptime churn is suppressed.");
    setAction("Open changes", "changes");
    return;
  }

  if (exposedListeners > 0) {
    text(byId("attentionTitle"), `${exposedListeners} listening ${exposedListeners === 1 ? "socket is" : "sockets are"} bound beyond loopback`);
    text(byId("attentionDetail"), "That is not proof of remote reachability. Routing and firewall state still matter; diagnose any endpoint to test the path.");
    setAction("Inspect listeners", "host");
    return;
  }

  text(byId("attentionTitle"), "No obvious local issue stands out in the current evidence");
  text(byId("attentionDetail"), "The latest snapshot has no failed native services, recent changes, or listeners beyond loopback that HostSleuth can currently see.");
  setAction("Diagnose a target", "diagnose");
}

function renderHostIdentity(snapshot) {
  if (!snapshot) return;
  const host = snapshot.host || {};
  const dockerMode = snapshot.mode === "docker";
  text(byId("identityHost"), host.hostname || "Unknown");
  text(byId("identityMode"), dockerMode ? "Docker" : "Native Linux");
  text(byId("identityOS"), host.os || "Unknown");
  text(byId("identityBootAge"), host.uptime || (host.boot_started_at ? `Since ${formatTime(host.boot_started_at)}` : "Unknown"));
  text(byId("identitySnapshot"), relativeSnapshotTime(snapshot.captured_at).replace(/^Snapshot\s+/i, ""));
  const version = state.about?.version || "";
  const revision = state.about?.revision || "";
  text(byId("identityVersion"), [version || "Unknown", revision ? revision.slice(0, 8) : ""].filter(Boolean).join(" · "));

  const boot = byId("identityBootAge");
  if (boot && host.boot_started_at) boot.title = `Boot started ${formatTime(host.boot_started_at)}`;
  const snap = byId("identitySnapshot");
  if (snap && snapshot.captured_at) snap.title = `Captured ${formatTime(snapshot.captured_at)}`;
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
  const troubledContainers = containers.filter((container) => ["unhealthy", "restarting", "stopped"].includes(containerHealth(container)));
  const serviceListeners = listeners.filter((listener) => !isRoutineListener(listener));
  const routineListeners = listeners.length - serviceListeners.length;
  const exposedListeners = serviceListeners.filter((listener) => listenerScope(listener.address).key !== "loopback").length;
  const recentChanges = state.events.filter((event) => isRecent(event.at, 24)).length;

  text(byId("hostName"), host.hostname || "This host");
  text(byId("hostDetailName"), host.hostname || "Host details");
  text(byId("deploymentBadge"), dockerMode ? "Docker view" : "Native view");
  renderHostIdentity(snapshot);

  const summaryParts = [host.os, host.kernel ? `kernel ${host.kernel}` : ""];
  text(byId("hostSummary"), summaryParts.filter(Boolean).join(" · ") || "Host snapshot loaded.");

  const storyParts = [
    `${serviceListeners.length} service ${serviceListeners.length === 1 ? "listener" : "listeners"}${routineListeners ? ` (+${routineListeners} routine)` : ""}`,
    `${interfaces.length} ${interfaces.length === 1 ? "interface" : "interfaces"}`,
    containers.length
      ? `${runningContainers}/${containers.length} containers running${troubledContainers.length ? `, ${troubledContainers.length} not healthy` : ""}`
      : "no visible containers",
    recentChanges ? `${recentChanges} meaningful changes in 24h` : "no meaningful changes in 24h",
  ];
  text(byId("hostStoryLine"), `Right now HostSleuth sees ${storyParts.join(" · ")}.`);

  text(byId("exposedListenerCount"), exposedListeners);
  text(byId("exposedListenerNote"), exposedListeners
    ? `${serviceListeners.length - exposedListeners} loopback-only${routineListeners ? ` · ${routineListeners} routine hidden` : ""}`
    : "Loopback-only surface");
  text(byId("failedServiceCount"), dockerMode ? "—" : failedServices);
  text(byId("failedServiceNote"), dockerMode ? "Unavailable in Docker mode" : failedServices ? `${failedServices} need attention` : "No failed services");
  byId("failedServiceNote").className = `signal-note ${!dockerMode && failedServices ? "problem" : !dockerMode ? "good" : ""}`;
  text(byId("containerCount"), containers.length);
  text(byId("containerNote"), !containers.length
    ? "No containers found"
    : troubledContainers.length
      ? `${troubledContainers.length} not healthy · ${runningContainers} running`
      : `${runningContainers} running, all healthy`);
  byId("containerNote").className = `signal-note ${troubledContainers.length ? "problem" : containers.length ? "good" : ""}`;
  text(byId("change24hCount"), recentChanges);
  text(byId("change24hNote"), recentChanges ? `${state.events.length} retained total` : "No recent meaningful changes");
  byId("change24hNote").className = `signal-note ${recentChanges ? "warning" : "good"}`;

  renderAttention({ dockerMode, failedServices, exposedListeners, recentChanges, troubledContainers });

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
    ["Memory", formatMemory(host.memory_total)],
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

  renderListenerList(byId("overviewListeners"), serviceListeners, 6);
  renderHostListeners();
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
  if (!relevant.length) return { key: "unknown", label: "NOT NEEDED" };
  if (relevant.some((check) => check.status === "fail")) return { key: "fail", label: genericStatusLabel("fail") };
  if (relevant.some((check) => check.status === "pass")) return { key: "pass", label: genericStatusLabel("pass") };
  return { key: "unknown", label: genericStatusLabel("unknown") };
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
  state.lastDiagnosis = diagnosis;
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
    statusText.textContent = genericStatusLabel(status);
    summary.append(icon, name, statusText);

    const evidence = document.createElement("p");
    evidence.className = "check-evidence";
    evidence.textContent = check.evidence || "No additional evidence was returned.";

    details.append(summary, evidence);
    list.append(details);
  });
}

function diagnosisEvidenceSummary(diagnosis) {
  const lines = [
    "HostSleuth Diagnose",
    `Target: ${diagnosis?.target || "unknown"}`,
    `Conclusion: ${diagnosis?.conclusion || "unknown"}`,
    `Confidence: ${diagnosis?.confidence || "unknown"}`,
    "",
  ];
  (diagnosis?.checks || []).forEach((check) => {
    lines.push(`[${genericStatusLabel(check.status)}] ${checkNames[check.name] || check.name || "Check"}: ${check.evidence || ""}`);
  });
  return lines.join("\n");
}

async function copyDiagnosisEvidence() {
  const status = byId("diagnosisCopyStatus");
  try {
    await copyBoundedText(diagnosisEvidenceSummary(state.lastDiagnosis));
    text(status, "Copied.");
  } catch (_) {
    text(status, "Clipboard unavailable.");
  }
  window.setTimeout(() => text(status, ""), 2500);
}

async function loadAbout() {
  try {
    const response = await fetch("/api/about", { cache: "no-store" });
    if (!response.ok) return;
    const about = await response.json();
    state.about = about;
    const version = about.version || "";
    const revision = about.revision || "";
    text(byId("appVersion"), [version, revision ? `(${revision})` : ""].filter(Boolean).join(" "));
    if (state.snapshot) renderHostIdentity(state.snapshot);
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
    renderChangesView();
    document.dispatchEvent(new CustomEvent("hostsleuth:snapshot"));
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

byId("attentionAction")?.addEventListener("click", () => {
  const view = byId("attentionAction")?.dataset.targetView;
  if (view) showView(view);
});

byId("diagnosisCopyEvidence")?.addEventListener("click", copyDiagnosisEvidence);
byId("diagnosisCopyEvidenceShare")?.addEventListener("click", () => shareRedacted(() => diagnosisEvidenceSummary(state.lastDiagnosis), "diagnosis"));
byId("diagnosisToDNS")?.addEventListener("click", () => openDNSForTarget(state.lastDiagnosis?.target));
byId("diagnosisToContract")?.addEventListener("click", () => openContractForTarget(state.lastDiagnosis?.target));

byId("diagnoseForm").addEventListener("submit", (event) => {
  event.preventDefault();
  const target = byId("diagnoseTarget").value.trim();
  if (target) runDiagnosis(target);
});

installNavigationAccessibility();

syncViewAccessibility("overview");
window.addEventListener("hashchange", routeFromHash);
if (document.readyState === "complete") routeFromHash();
else window.addEventListener("load", routeFromHash);

loadAbout();
loadDashboard();
