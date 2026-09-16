function workbenchPairList(pairs) {
  const dl = document.createElement("dl");
  dl.className = "workbench-pairs";
  pairs.forEach(([label, value, code = false]) => {
    if (value === undefined || value === null || value === "") return;
    const dt = document.createElement("dt");
    dt.textContent = label;
    const dd = document.createElement("dd");
    dd.textContent = String(value);
    if (code) dd.className = "workbench-code";
    dl.append(dt, dd);
  });
  return dl;
}

function workbenchRow(title, detail) {
  const row = document.createElement("div");
  row.className = "workbench-row";
  const strong = document.createElement("strong");
  strong.textContent = title;
  const span = document.createElement("span");
  span.textContent = detail || "";
  row.append(strong, span);
  return row;
}

async function workbenchFetch(path, params) {
  const query = new URLSearchParams(params);
  const response = await fetch(`${path}?${query.toString()}`, { cache: "no-store" });
  if (!response.ok) throw new Error((await response.text()).trim() || `request failed (${response.status})`);
  return response.json();
}

function showWorkbenchResult(id, render) {
  const container = byId(id);
  container.replaceChildren();
  container.classList.remove("hidden");
  render(container);
}

function showWorkbenchError(id, error) {
  showWorkbenchResult(id, (container) => container.append(makeEmpty(error.message || "Workbench request failed.")));
}

function renderFileInspection(container, file) {
  const status = file.checksum_status ? `${file.expected_algorithm.toUpperCase()} ${file.checksum_status}` : "not requested";
  container.append(workbenchPairList([
    ["Path", file.path, true],
    ["Size", `${file.size} bytes`],
    ["Permissions", file.permissions, true],
    ["Owner", file.owner ? `${file.owner} (uid ${file.uid})` : `uid ${file.uid}`],
    ["Group", file.group ? `${file.group} (gid ${file.gid})` : `gid ${file.gid}`],
    ["Modified", formatTime(file.modified_at)],
    ["SHA-256", file.sha256, true],
    ["SHA-512", file.sha512, true],
    ["Expected checksum", status],
  ]));
}

function renderFileComparison(container, comparison) {
  const headline = document.createElement("strong");
  headline.textContent = comparison.conclusion || "File comparison complete";
  container.append(headline);
  container.append(workbenchPairList([
    ["Left", comparison.left?.path, true],
    ["Left SHA-256", comparison.left?.sha256, true],
    ["Right", comparison.right?.path, true],
    ["Right SHA-256", comparison.right?.sha256, true],
    ["Same size", comparison.same_size ? "yes" : "no"],
    ["Same SHA-256", comparison.same_sha256 ? "yes" : "no"],
  ]));
}

function renderDNSInspection(container, result) {
  container.append(workbenchPairList([["Query", result.name], ["Resolver", result.resolver]]));
  const records = document.createElement("div");
  records.className = "workbench-list";
  (result.records || []).forEach((record) => records.append(workbenchRow(record.type, record.value)));
  if (!(result.records || []).length) records.append(makeEmpty("No records were returned by the system resolver."));
  container.append(records);

  const unavailable = (result.lookups || []).filter((lookup) => lookup.status !== "pass");
  if (unavailable.length) {
    const notes = document.createElement("div");
    notes.className = "workbench-list";
    unavailable.forEach((lookup) => notes.append(workbenchRow(`${lookup.type} unavailable`, lookup.error || "lookup returned no usable evidence")));
    container.append(notes);
  }
}

function renderHTTPInspection(container, result) {
  const headline = document.createElement("strong");
  headline.textContent = result.conclusion || "HTTP inspection complete";
  container.append(headline);
  const hops = document.createElement("div");
  hops.className = "workbench-list";
  (result.hops || []).forEach((hop, index) => {
    const detail = [hop.url, hop.location ? `Location: ${hop.location}` : "", hop.server ? `Server: ${hop.server}` : "", hop.content_type ? `Content-Type: ${hop.content_type}` : ""].filter(Boolean).join(" · ");
    hops.append(workbenchRow(`${index + 1}. ${hop.status}`, detail));
  });
  container.append(hops);
}

function renderCertificateInspection(container, result) {
  const comparison = result.local ? result : null;
  const local = comparison ? comparison.local : result;
  if (comparison) {
    const headline = document.createElement("strong");
    headline.textContent = comparison.conclusion || "Certificate comparison complete";
    headline.className = comparison.status === "match" ? "workbench-status-match" : comparison.status === "mismatch" ? "workbench-status-mismatch" : "";
    container.append(headline);
  }
  const cert = local?.certificate || {};
  container.append(workbenchPairList([
    ["File", local?.path, true],
    ["Subject", cert.subject],
    ["Issuer", cert.issuer],
    ["Serial", cert.serial, true],
    ["Valid from", cert.valid_from ? formatTime(cert.valid_from) : ""],
    ["Valid until", cert.valid_until ? formatTime(cert.valid_until) : ""],
    ["Days remaining", cert.days_remaining],
    ["SHA-256 fingerprint", cert.sha256_fingerprint, true],
    ["Public key", local?.public_key_algorithm],
    ["Signature", local?.signature_algorithm],
    ["CA certificate", local?.is_ca ? "yes" : "no"],
  ]));
  if (cert.sans?.length) {
    const sans = document.createElement("div");
    sans.className = "workbench-list";
    cert.sans.forEach((san) => sans.append(workbenchRow("SAN", san)));
    container.append(sans);
  }
  if (comparison?.served?.certificate) {
    container.append(workbenchPairList([
      ["Endpoint", comparison.target, true],
      ["Served SHA-256", comparison.served.certificate.sha256_fingerprint, true],
      ["Hostname validation", comparison.served.hostname_status],
      ["Trust validation", comparison.served.trust_status],
    ]));
  }
}

function installWorkbenchUI() {
  if (byId("view-workbench")) return;
  const nav = document.querySelector(".nav-tabs");
  const main = document.querySelector("main");
  if (!nav || !main) return;

  const tab = document.createElement("button");
  tab.className = "nav-tab";
  tab.dataset.view = "workbench";
  tab.textContent = "Workbench";
  tab.addEventListener("click", () => showView("workbench"));
  nav.append(tab);

  const section = document.createElement("section");
  section.className = "view";
  section.id = "view-workbench";
  section.dataset.viewPanel = "workbench";
  section.innerHTML = `
    <section class="panel section-panel workbench-intro">
      <p class="eyebrow">HOSTSLEUTH WORKBENCH</p>
      <h1>Small tools for concrete troubleshooting questions.</h1>
      <p class="section-copy">Read-only helpers for file identity, DNS answers, HTTP redirects, and certificate identity. No shell, file editor, custom request headers, credentials, or hidden commands.</p>
    </section>
    <div class="workbench-grid">
      <section class="panel section-panel workbench-card">
        <p class="eyebrow">FILE IDENTITY</p><h2>Is this the exact file I expect?</h2>
        <p class="section-copy">Inspect metadata plus SHA-256/SHA-512, and optionally verify an expected checksum. File contents are not returned or stored.</p>
        <form id="wbFileForm" class="workbench-form"><div class="workbench-fields">
          <label>Local file path<input id="wbFilePath" required placeholder="/etc/hosts"></label>
          <label>Expected checksum <span class="muted">(optional)</span><input id="wbFileExpected" placeholder="SHA-256 or SHA-512"></label>
        </div><button class="primary-button" type="submit">Inspect file</button></form>
        <div id="wbFileResult" class="workbench-result hidden"></div>
      </section>

      <section class="panel section-panel workbench-card">
        <p class="eyebrow">FILE COMPARISON</p><h2>Are these really the same bytes?</h2>
        <p class="section-copy">Compare two readable regular files by SHA-256 instead of filename, path, or timestamp.</p>
        <form id="wbCompareForm" class="workbench-form"><div class="workbench-fields">
          <label>First file<input id="wbCompareLeft" required placeholder="/path/source"></label>
          <label>Second file<input id="wbCompareRight" required placeholder="/path/destination"></label>
        </div><button class="primary-button" type="submit">Compare files</button></form>
        <div id="wbCompareResult" class="workbench-result hidden"></div>
      </section>

      <section class="panel section-panel workbench-card">
        <p class="eyebrow">DNS</p><h2>What does this host resolve right now?</h2>
        <p class="section-copy">Query the host's system resolver for common A/AAAA, CNAME, MX, NS, TXT, or PTR evidence with deterministic ordering.</p>
        <form id="wbDNSForm" class="workbench-form"><label>Name or IP<input id="wbDNSName" required placeholder="example.com"></label><button class="primary-button" type="submit">Inspect DNS</button></form>
        <div id="wbDNSResult" class="workbench-result hidden"></div>
      </section>

      <section class="panel section-panel workbench-card">
        <p class="eyebrow">HTTP</p><h2>Where does this URL actually lead?</h2>
        <p class="section-copy">Issue a direct HEAD request and show a bounded redirect chain. HostSleuth sends no body, credentials, cookies, or custom headers.</p>
        <form id="wbHTTPForm" class="workbench-form"><label>HTTP or HTTPS URL<input id="wbHTTPURL" required placeholder="https://example.com/"></label><button class="primary-button" type="submit">Inspect redirects</button></form>
        <div id="wbHTTPResult" class="workbench-result hidden"></div>
      </section>

      <section class="panel section-panel workbench-card workbench-wide">
        <p class="eyebrow">CERTIFICATE IDENTITY</p><h2>What certificate is this file—and is the endpoint serving it?</h2>
        <p class="section-copy">Inspect a public certificate PEM and optionally compare its exact SHA-256 fingerprint with a direct-TLS endpoint. Private-key PEM blocks are not accepted.</p>
        <form id="wbCertForm" class="workbench-form"><div class="workbench-fields">
          <label>Certificate PEM path<input id="wbCertPath" required placeholder="/etc/letsencrypt/live/example.com/cert.pem"></label>
          <label>Endpoint <span class="muted">(optional)</span><input id="wbCertTarget" placeholder="example.com:443"></label>
        </div><button class="primary-button" type="submit">Inspect certificate</button></form>
        <div id="wbCertResult" class="workbench-result hidden"></div>
      </section>
    </div>`;
  main.append(section);

  byId("wbFileForm").addEventListener("submit", async (event) => {
    event.preventDefault();
    try {
      const data = await workbenchFetch("/api/workbench/file", { path: byId("wbFilePath").value.trim(), expected: byId("wbFileExpected").value.trim() });
      showWorkbenchResult("wbFileResult", (container) => renderFileInspection(container, data));
    } catch (error) { showWorkbenchError("wbFileResult", error); }
  });
  byId("wbCompareForm").addEventListener("submit", async (event) => {
    event.preventDefault();
    try {
      const data = await workbenchFetch("/api/workbench/compare", { left: byId("wbCompareLeft").value.trim(), right: byId("wbCompareRight").value.trim() });
      showWorkbenchResult("wbCompareResult", (container) => renderFileComparison(container, data));
    } catch (error) { showWorkbenchError("wbCompareResult", error); }
  });
  byId("wbDNSForm").addEventListener("submit", async (event) => {
    event.preventDefault();
    try {
      const data = await workbenchFetch("/api/workbench/dns", { name: byId("wbDNSName").value.trim() });
      showWorkbenchResult("wbDNSResult", (container) => renderDNSInspection(container, data));
    } catch (error) { showWorkbenchError("wbDNSResult", error); }
  });
  byId("wbHTTPForm").addEventListener("submit", async (event) => {
    event.preventDefault();
    try {
      const data = await workbenchFetch("/api/workbench/http", { url: byId("wbHTTPURL").value.trim() });
      showWorkbenchResult("wbHTTPResult", (container) => renderHTTPInspection(container, data));
    } catch (error) { showWorkbenchError("wbHTTPResult", error); }
  });
  byId("wbCertForm").addEventListener("submit", async (event) => {
    event.preventDefault();
    try {
      const data = await workbenchFetch("/api/workbench/cert", { path: byId("wbCertPath").value.trim(), target: byId("wbCertTarget").value.trim() });
      showWorkbenchResult("wbCertResult", (container) => renderCertificateInspection(container, data));
    } catch (error) { showWorkbenchError("wbCertResult", error); }
  });
}

installWorkbenchUI();
