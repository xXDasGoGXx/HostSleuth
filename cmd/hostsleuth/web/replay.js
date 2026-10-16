// Replay: the host's containers drawn as an engineering sheet that rewinds
// through the last week. Every narrated line comes from /api/replay, which
// derives it from recorded events by fixed rules.

(() => {
  const H = 3600e3;
  const ALL_DOWN = new Set(["reboot", "docker"]);
  let D = null, START = 0, NOW = 0, PARTS = [], EVENTS = [], byId = {};
  let t = 0, sel = "", hlNet = "", loadedAt = 0;
  let playing = false, raf = 0, last = 0, holdUntil = 0;
  let layout = {}, modules = [], TB = {}, LEG = {}, W = 1290, HEIGHT = 900, tapeStatic = "";
  const COLX = [24, 444, 864], MODW = 402, PITCH = 60, HEAD = 46, GAP = 16, TOP = 24;
  const reduced = window.matchMedia?.("(prefers-reduced-motion: reduce)")?.matches;

  const $ = (id) => document.getElementById(id);
  const esc = (s) => String(s ?? "").replace(/[&<>"]/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;" }[c]));
  const cap = (s) => s.replace(/^./, (c) => c.toUpperCase());
  const clip = (s, n) => (s.length > n ? s.slice(0, n - 1) + "…" : s);
  const fmtWhen = (v) => new Date(v).toLocaleString([], { weekday: "short", day: "numeric", month: "short", hour: "numeric", minute: "2-digit" });
  const fmtShort = (v) => new Date(v).toLocaleString([], { weekday: "short", hour: "numeric", minute: "2-digit" });
  function ago(msv) {
    const m = Math.round(msv / 60e3);
    if (m < 1) return "just now";
    if (m < 60) return `${m} min ago`;
    const h = Math.round(m / 60);
    return h === 1 ? "an hour ago" : `${h} hours ago`;
  }
  function dur(msv) {
    const m = Math.round(msv / 60e3);
    if (m < 2) return "a minute";
    if (m < 90) return `${m} min`;
    const h = Math.round(msv / H);
    return h < 48 ? `${h} hours` : `${Math.round(h / 24)} days`;
  }

  // ---- state at a moment ------------------------------------------------
  function healthAt(p, at) {
    const tl = D.trans[p.id];
    if (!tl || !tl.length) return { s: "up", label: /\(healthy\)/.test(p.status) ? "healthy" : "running" };
    let state = tl[0][1], since = null;
    for (const [tt, , to] of tl) {
      if (tt > at) break;
      if (to !== state) since = tt;
      state = to;
    }
    if (state === "running (unhealthy)") {
      const s0 = since ?? START;
      return { s: "unhealthy", label: `unhealthy for ${dur(at - s0)}`, since: s0 };
    }
    if (state === "running (health: starting)") {
      const s0 = since ?? START;
      return at - s0 > 30 * 60e3 ? { s: "starting", label: `not healthy yet, ${dur(at - s0)}`, slow: true } : { s: "starting", label: "health check starting" };
    }
    if (state.startsWith("restarting")) return { s: "down", label: `restarting · exit ${state.match(/\d+/)?.[0] ?? "?"}` };
    if (state.startsWith("exited") || state.startsWith("dead")) return { s: "down", label: state };
    return { s: "up", label: "healthy" };
  }
  function stateAt(p, at) {
    for (const e of EVENTS) {
      if (!e.down || at < e.down[0] || at >= e.down[1]) continue;
      if (ALL_DOWN.has(e.k) || e.c === p.id) {
        const why = e.k === "reboot" ? "host rebooting" : e.k === "docker" ? "Docker restarting" : e.k === "update" ? "updating" : "restarting";
        return { s: "down", label: why };
      }
    }
    return healthAt(p, at);
  }
  function versionAt(p, at) {
    const v = D.versions[p.id];
    return v && at < v.t ? v.from : p.ver;
  }

  // ---- layout -----------------------------------------------------------
  function computeLayout() {
    layout = {}; modules = [];
    const colY = [TOP, TOP, TOP];
    const order = [...D.stacks].sort((a, b) => b.parts.length - a.parts.length || a.id.localeCompare(b.id));
    order.forEach((s) => {
      const col = colY.indexOf(Math.min(...colY));
      const m = { id: s.id, x: COLX[col], y: colY[col], h: HEAD + s.parts.length * PITCH - 6 };
      s.parts.forEach((id, i) => { layout[id] = { x: m.x + 16, y: m.y + HEAD + i * PITCH, w: 196, h: 46 }; });
      modules.push(m);
      colY[col] += m.h + GAP;
    });
    const byHeight = [0, 1, 2].sort((a, b) => colY[a] - colY[b]);
    TB = { x: COLX[byHeight[0]], y: colY[byHeight[0]] + 4, w: MODW, h: 290 };
    colY[byHeight[0]] += TB.h + GAP + 4;
    const lc = [0, 1, 2].sort((a, b) => colY[a] - colY[b])[0];
    LEG = { x: COLX[lc], y: colY[lc] + 4, w: MODW, h: 272 };
    colY[lc] += LEG.h + GAP + 4;
    HEIGHT = Math.max(...colY) + 12;
    $("rpDrawing").setAttribute("viewBox", `0 0 ${W} ${HEIGHT}`);
  }

  function cloudPath(x, y, w, h, r = 6) {
    const seg = (len) => { const n = Math.max(2, Math.round(len / (2 * r))); return { n, s: len / n }; };
    const a = seg(w), b = seg(h);
    let d = `M${x},${y}`;
    for (let i = 0; i < a.n; i++) d += ` a${r},${r} 0 0 1 ${a.s},0`;
    for (let i = 0; i < b.n; i++) d += ` a${r},${r} 0 0 1 0,${b.s}`;
    for (let i = 0; i < a.n; i++) d += ` a${r},${r} 0 0 1 ${-a.s},0`;
    for (let i = 0; i < b.n; i++) d += ` a${r},${r} 0 0 1 0,${-b.s}`;
    return d + " Z";
  }
  const flagShape = (x, y, w, hh) => `${x},${y - hh} ${x + w - 9},${y - hh} ${x + w},${y} ${x + w - 9},${y + hh} ${x},${y + hh}`;
  const flagWidth = (label) => Math.min(168, Math.max(96, label.length * 8.1 + 22));

  function drawSheet(at) {
    const svg = $("rpDrawing");
    const recent = new Map();
    EVENTS.forEach((e) => { if (e.c && e.t <= at && at - e.t <= 24 * H) recent.set(e.c, e); });
    const stamp = EVENTS.find((e) => ALL_DOWN.has(e.k) && e.down && at >= e.down[0] && at < e.down[1] + 60e3);
    const out = [];
    out.push(`<defs><pattern id="rpGrid" width="20" height="20" patternUnits="userSpaceOnUse"><path class="gridline" d="M20 0H0V20" fill="none"/></pattern>
      <pattern id="rpHatch" width="9" height="9" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><line class="hatch-line" x1="0" y1="0" x2="0" y2="9"/></pattern></defs>`);
    out.push(`<rect x="0" y="0" width="${W}" height="${HEIGHT}" fill="url(#rpGrid)"/><rect class="frame" x="10" y="10" width="${W - 20}" height="${HEIGHT - 20}"/>`);
    modules.forEach((m) => out.push(`<rect class="module" x="${m.x}" y="${m.y}" width="${MODW}" height="${m.h}" rx="6"/><text class="module-tag" x="${m.x + 16}" y="${m.y + 27}">${esc(m.id)}</text>`));

    PARTS.forEach((p) => {
      const L = layout[p.id];
      if (!L) return;
      const st = stateAt(p, at), g = [], rev = recent.get(p.id);
      if (rev) g.push(`<path class="cloud" d="${cloudPath(L.x - 8, L.y - 6, L.w + 16, L.h + 12)}"/>`);
      g.push(`<rect class="selmark" x="${L.x - 13}" y="${L.y - 10}" width="${L.w + 26}" height="${L.h + 20}" rx="4"/>`);
      g.push(`<rect class="body" x="${L.x}" y="${L.y}" width="${L.w}" height="${L.h}" rx="3"/>`);
      g.push(`<text class="name" x="${L.x + 11}" y="${L.y + 20}">${esc(clip(p.id, 20))}</text>`);
      const sub = st.s === "up" ? [p.base, versionAt(p, at)].filter(Boolean).join(" ") : st.label;
      g.push(`<text class="sub" x="${L.x + 11}" y="${L.y + 37}">${esc(clip(sub, 25))}</text>`);
      const n = Math.min(2, p.flags.length);
      p.flags.slice(0, 2).forEach((f, i) => {
        const py = n === 1 ? L.y + L.h / 2 : L.y + 13 + i * 20;
        const x0 = L.x + L.w, x1 = x0 + 14;
        g.push(`<circle class="pin" cx="${x0}" cy="${py}" r="3"/>`);
        g.push(st.s === "down"
          ? `<line class="stub" x1="${x0}" y1="${py}" x2="${x0 + 4}" y2="${py}"/><line class="stub" x1="${x0 + 10}" y1="${py}" x2="${x1}" y2="${py}"/><line class="cut" x1="${x0 + 4}" y1="${py - 6}" x2="${x0 + 10}" y2="${py + 6}"/>`
          : `<line class="stub" x1="${x0}" y1="${py}" x2="${x1}" y2="${py}"/>`);
        const label = clip(f.label, 18);
        g.push(`<g class="flag ${esc(f.net)}" data-net="${esc(f.net)}"><polygon points="${flagShape(x1, py, flagWidth(label), n === 1 ? 10 : 8.5)}"/><text x="${x1 + 8}" y="${py + 4.3}">${esc(label)}</text></g>`);
      });
      if (rev) {
        const tx = L.x - 8, ty = L.y - 8;
        g.push(`<g class="delta"><polygon points="${tx},${ty - 11} ${tx + 12},${ty + 10} ${tx - 12},${ty + 10}"/><text x="${tx}" y="${ty + 6}" text-anchor="middle">${rev.n}</text></g>`);
      }
      out.push(`<g class="part ${st.s} ${sel === p.id ? "selected" : ""}" data-c="${esc(p.id)}" tabindex="0" role="button" aria-label="${esc(`${p.id}: ${st.label}`)}">${g.join("")}</g>`);
    });

    const tb = [`<rect x="${TB.x}" y="${TB.y}" width="${TB.w}" height="${TB.h}"/>`,
      `<text class="big" x="${TB.x + 14}" y="${TB.y + 30}">${esc(clip(D.host.name || "this host", 26))}</text>`,
      `<text class="small" x="${TB.x + 14}" y="${TB.y + 50}">${esc(clip(`${(D.host.os || "").replace(/ GNU\/Linux/, "")} · ${PARTS.length} containers · drawn by HostSleuth`, 58))}</text>`,
      `<text class="small" x="${TB.x + 14}" y="${TB.y + 68}">As of ${esc(fmtWhen(at))}</text>`,
      `<line x1="${TB.x}" y1="${TB.y + 80}" x2="${TB.x + TB.w}" y2="${TB.y + 80}"/>`,
      `<text class="small" x="${TB.x + 14}" y="${TB.y + 98}">Rev</text><text class="small" x="${TB.x + 60}" y="${TB.y + 98}">When</text><text class="small" x="${TB.x + 162}" y="${TB.y + 98}">What changed</text>`];
    const rows = Math.floor((TB.h - 110) / 19);
    const list = EVENTS.filter((e) => e.t <= at).slice(-rows).reverse();
    list.forEach((e, i) => {
      const y = TB.y + 118 + i * 19, f = e.fault ? " fault" : "";
      tb.push(`<text class="row${f}" x="${TB.x + 14}" y="${y}">Δ${e.n}</text><text class="row${f}" x="${TB.x + 60}" y="${y}">${esc(fmtShort(e.t))}</text><text class="row${f}" x="${TB.x + 162}" y="${y}">${esc(clip(e.short, 30))}</text>`);
    });
    if (!list.length) tb.push(`<text class="row" x="${TB.x + 14}" y="${TB.y + 118}">No changes recorded yet this week.</text>`);
    out.push(`<g class="tb">${tb.join("")}</g>`);

    const leg = [`<text class="module-tag" x="${LEG.x + 16}" y="${LEG.y + 26}">how to read the drawing</text>`];
    [["all", "all :3000", "Bound to every address on the host."], ["lan", "LAN :8080", "Bound to one address only."], ["local", "local :8088", "This host only (loopback)."], ["host", "host network", "Shares the host's network."], ["none", "no network", "No network at all."]].forEach(([net, label, note], i) => {
      const x = LEG.x + 16, y = LEG.y + 54 + i * 30;
      leg.push(`<g class="flag ${net}" data-net="${net}"><polygon points="${flagShape(x, y, flagWidth(label), 10)}"/><text x="${x + 8}" y="${y + 4.3}">${label}</text></g><text class="note" x="${x + 136}" y="${y + 4.5}">${note}</text>`);
    });
    const ly = LEG.y + 54 + 5 * 30;
    leg.push(`<path class="cloud" d="${cloudPath(LEG.x + 16, ly - 8, 34, 20, 5)}"/><text class="note" x="${LEG.x + 64}" y="${ly + 7}">Changed in the 24 hours before this moment.</text>`);
    leg.push(`<rect class="legend-hatch" x="${LEG.x + 16}" y="${ly + 24}" width="34" height="20" fill="url(#rpHatch)"/><text class="note" x="${LEG.x + 64}" y="${ly + 39}">Running, but failing its health check.</text>`);
    leg.push(`<line class="stub" x1="${LEG.x + 16}" y1="${ly + 66}" x2="${LEG.x + 26}" y2="${ly + 66}"/><line class="stub" x1="${LEG.x + 34}" y1="${ly + 66}" x2="${LEG.x + 50}" y2="${ly + 66}"/><line class="cut" x1="${LEG.x + 26}" y1="${ly + 60}" x2="${LEG.x + 34}" y2="${ly + 72}"/><text class="note" x="${LEG.x + 64}" y="${ly + 71}">Cut wire: stopped or restarting.</text>`);
    out.push(`<g class="legend">${leg.join("")}</g>`);

    if (stamp) {
      const cx = W / 2, cy = TOP + 150;
      out.push(`<g class="stamp" transform="rotate(-6 ${cx} ${cy})"><rect x="${cx - 190}" y="${cy - 32}" width="380" height="58" rx="6"/><text x="${cx}" y="${cy + 6}" text-anchor="middle">${stamp.k === "reboot" ? "HOST REBOOTING" : "DOCKER RESTARTING"}</text></g>`);
    }
    svg.innerHTML = out.join("");
    if (hlNet) applyHighlight();
  }

  // ---- narration ----------------------------------------------------------
  function narrate(at) {
    const live = NOW - at < 60e3;
    $("rpWhen").textContent = fmtWhen(at);
    $("rpLive").classList.toggle("on", live);
    $("rpLive").textContent = live ? "Latest snapshot" : "Replay";
    const states = PARTS.map((p) => [p, stateAt(p, at)]);
    const sick = states.filter(([, s]) => s.s === "unhealthy").sort((a, b) => a[1].since - b[1].since);
    const lastEv = [...EVENTS].reverse().find((e) => e.t <= at);
    const fresh = lastEv && at - lastEv.t <= 2 * H;
    let say, more = "";
    if (!PARTS.length) say = D.notes?.[0] || "Nothing to draw yet.";
    else if (fresh && at - lastEv.t < 5 * 60e3) say = lastEv.text;
    else if (sick.length) {
      say = `${sick[0][0].id} has been failing its health check for ${dur(at - sick[0][1].since)}.`;
      if (lastEv) more = `Last change ${ago(at - lastEv.t)}: ${lastEv.text}`;
    } else if (fresh) say = `${cap(ago(at - lastEv.t))}: ${lastEv.text}`;
    else say = lastEv ? `All quiet. Last change ${ago(at - lastEv.t)}: ${lastEv.text}` : "All quiet this week.";
    $("rpSay").textContent = say;
    $("rpMore").textContent = more;
    const running = states.filter(([, s]) => s.s !== "down").length;
    const bits = [`<b>${running}</b> of ${PARTS.length} running`];
    if (sick.length) bits.push(`<span class="bad">${sick.length} unhealthy</span>`);
    if (PARTS.length - running) bits.push(`<span class="bad">${PARTS.length - running} stopped</span>`);
    const slow = states.filter(([, s]) => s.slow).length;
    if (slow) bits.push(`<span class="bad">${slow} stuck starting</span>`);
    if (PARTS.length && !sick.length && running === PARTS.length && !slow) bits.push("all healthy");
    $("rpCounts").innerHTML = bits.join(" · ");
  }

  // ---- tape -------------------------------------------------------------
  const TX0 = 12, TX1 = 988;
  const tToX = (v) => TX0 + (v - START) / (NOW - START) * (TX1 - TX0);
  const xToT = (x) => START + Math.min(1, Math.max(0, (x - TX0) / (TX1 - TX0))) * (NOW - START);
  function buildTapeStatic() {
    const s = [];
    const d0 = new Date(START); d0.setHours(24, 0, 0, 0);
    for (let d = d0.getTime(); d < NOW; d += 24 * H) {
      const x = tToX(d);
      s.push(`<line class="day" x1="${x}" y1="20" x2="${x}" y2="58"/><text class="daylabel" x="${x + 4}" y="74">${esc(new Date(d).toLocaleDateString([], { weekday: "short", day: "numeric" }))}</text>`);
    }
    const N = 480;
    for (let i = 0; i < N; i++) {
      const at = START + (i + .5) * (NOW - START) / N;
      if (PARTS.some((p) => { const st = stateAt(p, at); return st.s === "unhealthy" || st.s === "down" || st.slow; })) {
        s.push(`<rect class="attn" x="${tToX(START + i * (NOW - START) / N)}" y="50" width="${(TX1 - TX0) / N + .5}" height="5"/>`);
      }
    }
    s.push(`<line class="base" x1="${TX0}" y1="40" x2="${TX1}" y2="40"/>`);
    EVENTS.forEach((e) => {
      const x = tToX(e.t);
      if (e.k === "reboot") s.push(`<line class="ev-reboot" x1="${x}" y1="12" x2="${x}" y2="48"/><text class="reboot-label" x="${x + 4}" y="14">reboot</text>`);
      else if (e.k === "docker") s.push(`<line class="ev-info" x1="${x}" y1="20" x2="${x}" y2="40"/>`);
      else if (e.fault) s.push(`<line class="ev-fault" x1="${x}" y1="26" x2="${x}" y2="40"/>`);
      else if (e.k === "update") s.push(`<rect class="ev-rev" x="${x - 5}" y="28" width="10" height="10" transform="rotate(45 ${x} 33)"/>`);
      else s.push(`<line class="ev-info" x1="${x}" y1="32" x2="${x}" y2="40"/>`);
    });
    tapeStatic = s.join("");
  }
  function drawTape(at) {
    const x = tToX(at), tape = $("rpTape");
    tape.innerHTML = `<rect class="past" x="${x}" y="18" width="${Math.max(0, TX1 - x)}" height="40"/>${tapeStatic}<line class="head" x1="${x}" y1="6" x2="${x}" y2="60"/><circle class="knob" cx="${x}" cy="6" r="6"/>`;
    tape.setAttribute("aria-valuenow", ((at - START) / H).toFixed(1));
    tape.setAttribute("aria-valuetext", fmtWhen(at));
  }

  // ---- detail -----------------------------------------------------------
  function drawDetail(at) {
    const box = $("rpDetail"), p = byId[sel];
    if (!p) { box.innerHTML = `<p class="rp-empty">Tap a container on the drawing to see its week.</p>`; return; }
    const st = stateAt(p, at);
    const own = EVENTS.filter((e) => e.c === p.id);
    const count = (...k) => own.filter((e) => k.includes(e.k)).length;
    const bits = [];
    const sickN = count("unhealthy"), crashN = count("crash", "oom"), restartN = count("restart", "clean"), slowN = count("slow"), updN = count("update", "recreate");
    if (sickN) bits.push(`unhealthy ${sickN === 1 ? "once" : `${sickN} times`}`);
    if (crashN) bits.push(`${crashN} ${crashN === 1 ? "crash" : "crashes"}`);
    if (restartN) bits.push(`${restartN} restart${restartN === 1 ? "" : "s"} on its own`);
    if (slowN) bits.push(`${slowN} slow start${slowN === 1 ? "" : "s"}`);
    if (updN) bits.push(`${updN} update${updN === 1 ? "" : "s"}`);
    let week = bits.length ? `This week: ${bits.join(", ")}.` : "Nothing of its own this week; it only restarted with the host or Docker.";
    if (st.s === "unhealthy") week = `Failing its health check for ${dur(at - st.since)}${NOW - at < 60e3 && p.streak ? `, ${p.streak} checks in a row` : ""}. ` + week;
    if (!p.hc) week += " It has no health check, so HostSleuth can only see whether it is running.";
    const mine = EVENTS.filter((e) => e.c === p.id || ALL_DOWN.has(e.k));
    const re = new RegExp("^" + p.id.replace(/[.*+?^${}()|[\]\\-]/g, "\\$&") + " ");
    const items = mine.map((e) => {
      const text = ALL_DOWN.has(e.k) ? e.text.split(". ")[0] + "." : e.text.replace(re, "");
      return `<li class="${e.t > at ? "later" : ""} ${e.fault ? "fault" : ""}"><span class="d">Δ${e.n}</span><span class="t">${esc(fmtShort(e.t))}</span><span class="w">${esc(text)}</span></li>`;
    }).reverse().join("");
    box.innerHTML = `<div>
        <h2>${esc(p.id)}</h2>
        <div class="rp-img">${esc([p.base, versionAt(p, at)].filter(Boolean).join(" "))} · ${esc(p.stack)} stack</div>
        <p>At this moment: <span class="rp-state ${st.s === "up" ? "" : "bad"}">${esc(st.label)}</span></p>
        <p class="rp-week">${esc(week)}</p>
        <p><button type="button" id="rpOpenStory">Open its Container Story</button></p>
      </div>
      <ol aria-label="${esc(p.id)} this week">${items || `<li><span></span><span></span><span class="w">Nothing this week.</span></li>`}</ol>`;
    $("rpOpenStory").addEventListener("click", () => window.hostSleuthOpenContainer?.(p.id));
  }

  // ---- interaction --------------------------------------------------------
  function render() {
    if (!D) return;
    drawSheet(t); narrate(t); drawTape(t); drawDetail(t);
    $("rpLatest").setAttribute("aria-pressed", String(NOW - t < 60e3));
  }
  function setT(v) { t = Math.min(NOW, Math.max(START, v)); render(); }
  function applyHighlight() {
    const svg = $("rpDrawing");
    svg.classList.toggle("hl", !!hlNet);
    svg.querySelectorAll(".flag").forEach((f) => f.classList.toggle("hit", f.dataset.net === hlNet));
  }
  const landing = (e) => (e.down ? e.down[0] + 4e3 : e.t + 3e3);
  function jump(dir) {
    stop();
    const list = dir < 0 ? [...EVENTS].reverse().filter((e) => landing(e) < t - 1e3) : EVENTS.filter((e) => landing(e) > t + 1e3);
    const e = list[0];
    if (e) { if (e.c && byId[e.c]) sel = e.c; setT(landing(e)); }
  }
  function stop() { playing = false; cancelAnimationFrame(raf); const b = $("rpPlay"); if (b) b.textContent = "Play the week"; }
  function tick(now) {
    if (!playing) return;
    if (now < holdUntil) { raf = requestAnimationFrame(tick); return; }
    const dt = Math.min(100, now - last); last = now;
    const nextEv = EVENTS.find((e) => landing(e) > t + 1e3);
    const nt = t + dt * 4.2 * H / 1000;
    if (nextEv && nt >= landing(nextEv)) {
      if (nextEv.c && byId[nextEv.c]) sel = nextEv.c;
      setT(landing(nextEv));
      holdUntil = now + (reduced ? 2600 : 2000);
      last = holdUntil;
      raf = requestAnimationFrame(tick);
      return;
    }
    if (nt >= NOW) { setT(NOW); stop(); return; }
    setT(nt);
    raf = requestAnimationFrame(tick);
  }

  function defaultSelection() {
    const sick = PARTS.map((p) => [p, stateAt(p, NOW)]).filter(([, s]) => s.s === "unhealthy").sort((a, b) => a[1].since - b[1].since);
    if (sick.length) return sick[0][0].id;
    const recentEv = [...EVENTS].reverse().find((e) => e.c && byId[e.c]);
    return recentEv ? recentEv.c : PARTS[0]?.id || "";
  }

  async function load() {
    try {
      const response = await fetch("/api/replay", { cache: "no-store" });
      if (!response.ok) throw new Error((await response.text()) || `request failed (${response.status})`);
      const data = await response.json();
      const wasLive = !D || NOW - t < 60e3;
      D = data; NOW = D.now; START = D.start; PARTS = D.parts; EVENTS = D.events;
      byId = Object.fromEntries(PARTS.map((p) => [p.id, p]));
      if (!byId[sel]) sel = defaultSelection();
      loadedAt = Date.now();
      computeLayout(); buildTapeStatic();
      t = wasLive ? NOW : Math.min(NOW, Math.max(START, t));
      render();
    } catch (error) {
      $("rpSay").textContent = `HostSleuth could not build the replay: ${error.message || error}`;
    }
  }

  function install() {
    if ($("view-replay")) return;
    const nav = document.querySelector(".nav-tabs");
    const main = document.querySelector("main");
    if (!nav || !main) return;
    const tab = document.createElement("button");
    tab.className = "nav-tab";
    tab.dataset.view = "replay";
    tab.textContent = "Replay";
    tab.addEventListener("click", () => showView("replay"));
    nav.prepend(tab);

    const section = document.createElement("section");
    section.className = "view";
    section.id = "view-replay";
    section.dataset.viewPanel = "replay";
    section.innerHTML = `
      <div class="rp-sheet">
        <div class="rp-top">
          <div class="rp-mark">HostSleuth <span>Replay</span></div>
          <div class="rp-note">Your containers as a drawing, rewindable through the last 7 days of recorded changes.</div>
        </div>
        <div class="rp-narration" aria-live="polite">
          <div class="rp-when"><span id="rpWhen"></span><span class="rp-live" id="rpLive">Latest snapshot</span></div>
          <h1 class="rp-say" id="rpSay">Reading the week…</h1>
          <div class="rp-more" id="rpMore"></div>
          <div class="rp-counts" id="rpCounts"></div>
        </div>
        <div class="rp-drawing-wrap"><svg id="rpDrawing" viewBox="0 0 1290 900" role="img" aria-label="Drawing of the containers at the selected moment"></svg></div>
        <svg id="rpTape" viewBox="0 0 1000 78" tabindex="0" role="slider" aria-label="Moment shown" aria-valuemin="0" aria-valuemax="168"></svg>
        <div class="rp-controls">
          <button type="button" id="rpPrev">Previous change</button>
          <button type="button" id="rpPlay" class="primary">Play the week</button>
          <button type="button" id="rpNext">Next change</button>
          <button type="button" id="rpLatest" aria-pressed="true">Latest</button>
          <span class="rp-hint">Drag the tape to rewind. Tap a container for its week. Tap a connector to see everything bound the same way.</span>
        </div>
        <section class="rp-detail" id="rpDetail" aria-live="polite"></section>
      </div>`;
    const first = main.querySelector("[data-view-panel]");
    if (first) first.before(section); else main.append(section);

    const svg = $("rpDrawing"), tape = $("rpTape");
    svg.addEventListener("click", (ev) => {
      const flag = ev.target.closest(".flag");
      if (flag) { hlNet = hlNet === flag.dataset.net ? "" : flag.dataset.net; applyHighlight(); return; }
      const part = ev.target.closest(".part");
      if (part) { sel = part.dataset.c; render(); return; }
      if (hlNet) { hlNet = ""; applyHighlight(); }
    });
    svg.addEventListener("keydown", (ev) => {
      const part = ev.target.closest?.(".part");
      if (part && (ev.key === "Enter" || ev.key === " ")) {
        ev.preventDefault(); sel = part.dataset.c; render();
        svg.querySelector(`.part[data-c="${CSS.escape(sel)}"]`)?.focus();
      }
    });
    let dragging = false;
    const tapeT = (ev) => { const r = tape.getBoundingClientRect(); return xToT((ev.clientX - r.left) / r.width * 1000); };
    tape.addEventListener("pointerdown", (ev) => { if (!D) return; dragging = true; stop(); tape.setPointerCapture(ev.pointerId); setT(tapeT(ev)); });
    tape.addEventListener("pointermove", (ev) => { if (dragging) setT(tapeT(ev)); });
    tape.addEventListener("pointerup", () => { dragging = false; });
    tape.addEventListener("pointercancel", () => { dragging = false; });
    tape.addEventListener("keydown", (ev) => {
      if (!D) return;
      const step = ev.shiftKey ? 24 * H : H;
      const keys = { ArrowLeft: () => setT(t - step), ArrowRight: () => setT(t + step), Home: () => setT(START), End: () => setT(NOW), PageUp: () => jump(-1), PageDown: () => jump(1) };
      if (keys[ev.key]) { ev.preventDefault(); stop(); keys[ev.key](); }
    });
    $("rpPrev").addEventListener("click", () => D && jump(-1));
    $("rpNext").addEventListener("click", () => D && jump(1));
    $("rpLatest").addEventListener("click", () => { if (!D) return; stop(); load(); });
    $("rpPlay").addEventListener("click", () => {
      if (!D) return;
      if (playing) { stop(); return; }
      if (NOW - t < 60e3) setT(START);
      playing = true; $("rpPlay").textContent = "Pause";
      last = performance.now(); holdUntil = 0;
      raf = requestAnimationFrame(tick);
    });

    document.addEventListener("hostsleuth:view", (event) => {
      if (event.detail === "replay") { if (!D || Date.now() - loadedAt > 60e3) load(); }
      else stop();
    });
    // Keep the latest view current while it is open.
    window.setInterval(() => {
      if (D && !playing && activeViewName() === "replay" && NOW - t < 60e3 && !document.hidden) load();
    }, 60e3);

    if (!window.location.hash || window.location.hash === "#replay") showView("replay");
    load();
  }

  install();
})();
