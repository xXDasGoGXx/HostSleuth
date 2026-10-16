// Share redacted: turn any evidence summary into text that is safe to paste
// into a forum, an issue or a chat. The server applies the Evidence Bundle
// redaction seeded with this host's own names; nothing leaves the browser
// except to HostSleuth itself.

let shareDialog = null;
let shareLabel = "evidence";

function ensureShareDialog() {
  if (shareDialog) return shareDialog;
  shareDialog = document.createElement("dialog");
  shareDialog.className = "share-dialog";
  shareDialog.setAttribute("aria-labelledby", "shareDialogTitle");
  shareDialog.innerHTML = `
    <form method="dialog" class="share-dialog-body">
      <p class="eyebrow">SHARE REDACTED</p>
      <h2 id="shareDialogTitle">Safe to paste—after a quick read</h2>
      <p class="section-copy">Host names, addresses, container and service names, paths, URLs and secrets are replaced with consistent aliases (<code>host-01</code>, <code>ipv4-01</code>, <code>container-01</code>). Ports, statuses, exit codes and versions stay, because that is what helpers need. Read it once before posting: redaction lowers risk, it cannot promise anonymity.</p>
      <textarea id="shareText" readonly rows="14" aria-label="Redacted evidence"></textarea>
      <div class="share-actions">
        <button class="primary-button" id="shareCopy" type="button">Copy</button>
        <button class="secondary-button" id="shareDownload" type="button">Download .md</button>
        <span class="copy-status" id="shareStatus" aria-live="polite"></span>
        <button class="secondary-button share-close" value="close">Close</button>
      </div>
    </form>`;
  document.body.append(shareDialog);
  byId("shareCopy").addEventListener("click", async () => {
    try {
      await copyBoundedText(byId("shareText").value);
      text(byId("shareStatus"), "Copied.");
    } catch (_) {
      byId("shareText").select();
      text(byId("shareStatus"), "Select all and copy.");
    }
  });
  byId("shareDownload").addEventListener("click", () => {
    const blob = new Blob([byId("shareText").value], { type: "text/markdown;charset=utf-8" });
    const link = document.createElement("a");
    link.href = URL.createObjectURL(blob);
    link.download = `hostsleuth-${shareLabel}-${new Date().toISOString().slice(0, 10)}.md`;
    document.body.append(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(link.href), 4000);
    text(byId("shareStatus"), "Downloaded.");
  });
  return shareDialog;
}

async function shareRedacted(build, label) {
  shareLabel = String(label || "evidence").replace(/[^a-z0-9-]/gi, "") || "evidence";
  let payload = "";
  try {
    payload = String((typeof build === "function" ? build() : build) || "").trim();
  } catch (_) {
    payload = "";
  }
  const dialog = ensureShareDialog();
  const area = byId("shareText");
  text(byId("shareStatus"), "");
  if (!payload) {
    area.value = "Run this check first; there is no result to share yet.";
  } else {
    area.value = "Redacting…";
  }
  if (typeof dialog.showModal === "function") dialog.showModal();
  else dialog.setAttribute("open", "");
  if (!payload) return;

  try {
    const response = await fetch("/api/share", {
      method: "POST",
      headers: { "Content-Type": "text/plain;charset=utf-8" },
      body: payload.slice(0, 60000),
      cache: "no-store",
    });
    if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
    const result = await response.json();
    const version = state?.about?.version ? ` ${state.about.version}` : "";
    area.value = `${result.text}\n\n— Shared from HostSleuth${version}. Identifiers were replaced with aliases.`;
    area.scrollTop = 0;
  } catch (error) {
    area.value = `Redaction failed, so nothing is shown: ${error.message || error}`;
  }
}
