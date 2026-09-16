# M8 — Incident Lens

Status: complete in source after merge acceptance.

## Goal

Answer **"what changed around the time this broke?"** by placing retained HostSleuth events into one bounded evidence window without turning temporal proximity into an invented causal conclusion.

## Delivered scope

M8 adds a read-only Incident Lens with an explicit anchor time and a fixed initial window of +/- 15 minutes.

The lens:

- includes all existing retained event categories inside the bounded window, including package, configuration, service, container, listener, and system events when present;
- orders retained events chronologically with deterministic ordering for equal timestamps;
- caps returned incident events at 100;
- can optionally run the existing endpoint Diagnose engine for a supplied `host:port`;
- labels that endpoint result as **current evidence captured now**, not reconstructed historical state from the incident time;
- states explicitly that temporal proximity does not prove causation;
- reuses HostSleuth's existing JSONL event model rather than adding a time-series database.

Interfaces:

- CLI: `hostsleuth incident --at <RFC3339> [--target host:port]`;
- API: `/api/incident-lens?at=<RFC3339>[&target=host:port]`;
- Web UI: Incident Lens inside the existing **Changes** view;
- retained timeline entries gain an **Inspect window** action that anchors the lens directly on that event.

## Truthfulness boundary

Historical evidence and current evidence are deliberately separated.

HostSleuth can truthfully show which retained changes it recorded around an old incident time. It cannot reconstruct a TLS handshake or endpoint state that was never retained at that historical moment. Therefore an optional endpoint/TLS check is stamped with its current capture time and presented separately.

No nearby event is labeled as a cause merely because it falls inside the incident window.

## Not added

M8 did not add:

- causal inference;
- alerts or uptime monitoring;
- a time-series database;
- historical packet/TLS reconstruction;
- reboot-cause analysis reserved for M10;
- service controls;
- remediation;
- arbitrary commands;
- any live OMV deployment change.

## Focused regression coverage

Tests cover:

- inclusive +/- 15 minute boundaries;
- exclusion of events outside the window;
- preservation of all event categories;
- deterministic ordering when timestamps match;
- bounded result limits;
- explicit non-causal wording;
- absence of endpoint evidence/timestamp when no target was supplied;
- current endpoint evidence and timestamp remaining separate from an older historical anchor.

## Real-host acceptance

Acceptance ran on the actual OMV Debian host entirely under `/tmp/hostsleuth-m8-accept` using the existing user-space Go toolchain and an isolated HostSleuth state directory.

Validation confirmed:

- `go test ./...` passed;
- native build passed;
- base, Service Story, and Incident Lens JavaScript syntax passed;
- a synthetic listener-change event written only to isolated acceptance state was returned inside a 15-minute Incident Lens window;
- the output explicitly stated that temporal proximity does not prove causation;
- optional current diagnosis of `127.0.0.1:22` reported the endpoint reachable and emitted a separate endpoint-capture timestamp;
- without a target, `endpoint_captured_at` is omitted rather than serialized as a misleading zero timestamp;
- isolated Web/API smoke returned `window_minutes=15`, served the normal HostSleuth UI, and included the Incident Lens client asset.

The first API smoke exposed the zero-time serialization issue described above; M8 changed the endpoint capture timestamp to an optional value and added regression coverage before closeout.

No live HostSleuth process, production state, Compose file, Arcane project, Docker image pin, or recovery definition was stopped, restarted, reconfigured, or upgraded.

## Product/research boundary

The separate consumer-opportunity research remains design input only. M8 did not pull STARTTLS tools, certificate deployment recipes, ACME actions, or other later ideas forward.

The next approved milestone is M9 — HostSleuth Workbench.
