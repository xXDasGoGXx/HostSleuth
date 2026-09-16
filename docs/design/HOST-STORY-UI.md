# HostSleuth — Host Story UI exploration

Status: isolated prototype on `explore/host-story-ui`.

This is a product-experience exploration, not a new HostSleuth subsystem and not part of M3.4 container publication unless it is deliberately accepted later.

## Problem

The accepted M3 UI is clean and usable, but the Overview primarily presents counts and lists. It does not immediately communicate why HostSleuth is different from a generic host monitor.

HostSleuth's two jobs remain:

1. remember meaningful host changes;
2. deterministically explain why a host/service/port is or is not reachable.

The UI should make those jobs obvious without becoming a metrics dashboard.

## Presentation research

Useful patterns from mature tools:

- Cockpit: make a server discoverable and understandable at a glance rather than requiring command recall.
- Uptime Kuma: keep the primary workflow immediately obvious and visually approachable.
- Netdata: bring related troubleshooting evidence into one view instead of making users jump between disconnected tools.

HostSleuth should borrow those presentation principles, not their feature sets or architecture.

## Prototype principles

- Evidence over decorative charts.
- Answer questions, do not merely display counters.
- Keep the same Overview / Diagnose / Changes / Host navigation.
- Do not add alerting, remediation, AI explanation, multi-host management, agents, SNMP, time-series metrics, or configuration frameworks in this exploration.
- Do not infer causality from nearby events. Recent changes shown beside a diagnosis are explicitly labeled as context only.
- Do not treat a listening socket as proof of end-to-end reachability; routing and firewall state still matter.
- Continue to report unavailable Docker-mode evidence honestly.
- Keep the interface useful on a phone and on a large desktop display.

## What the prototype changes

### Overview becomes a Host Story

The top of the UI summarizes what HostSleuth can currently prove about the machine rather than showing four generic counters.

It surfaces:

- listeners bound beyond loopback;
- failed native services when visible;
- visible Docker containers and running count;
- meaningful changes in the last 24 hours;
- a plain-language current-host sentence built only from snapshot/event facts.

### Reachability Surface

Listeners become an actionable surface rather than a hidden count.

Each socket is classified as:

- loopback;
- wildcard;
- specific-address.

Each visible listener can pre-fill the existing deterministic Diagnose workflow. Classification is descriptive only; it does not claim that a wildcard or non-loopback listener is unsafe or remotely reachable.

### Diagnose becomes an evidence path

The existing deterministic checks are grouped visually into an ordered path:

1. name/target;
2. route;
3. TCP;
4. local bind/container evidence;
5. candidate firewall/service evidence.

The exact existing checks and evidence remain available below the path.

The latest retained host changes are shown as nearby context and are explicitly marked as non-causal.

### Host evidence is less hidden

The Host view now exposes data HostSleuth already collects but the M3 interface underuses:

- routes;
- full listener surface;
- Docker container image/status/ports/networks;
- interfaces;
- filesystems;
- host facts.

### Changes gains a compact category summary

The existing event timeline remains authoritative. A small category count summary helps users understand what kind of change history is present without adding filtering/configuration complexity.

## Non-goals

This exploration does not add:

- resource graphs;
- polling dashboards;
- SNMP/network-device monitoring;
- notifications;
- uptime monitoring;
- automatic correlation claims;
- new APIs or storage;
- backend collectors;
- new settings.

Those would require separate product decisions if ever pursued.

## Acceptance question

The prototype succeeds if a new user can open HostSleuth and quickly answer:

- What machine am I looking at?
- Where is this host listening, and on what scope?
- Has anything meaningful changed recently?
- Is there an obvious local condition worth inspecting?
- Where do I click to investigate a host:port problem?
- What evidence led HostSleuth to its diagnosis?

without mistaking HostSleuth for a generic monitoring product.
