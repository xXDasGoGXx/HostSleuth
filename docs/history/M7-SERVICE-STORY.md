# M7 — Service Story

Date: 2026-09-16

## Goal

Answer two common troubleshooting questions from one bounded evidence story:

- why will this systemd service not start?
- why did the endpoint I expected from this service disappear?

M7 deliberately remains evidence-only. It does not add start, stop, restart, reload, enable/disable, arbitrary command execution, or automatic remediation.

## Delivered scope

### Systemd runtime evidence

For one explicitly selected unit, native Linux can read a bounded `systemctl show` property set:

- load state;
- active state;
- sub-state;
- unit-file state;
- result;
- main PID;
- control group;
- main-process exit code/status.

The lookup is bounded to two seconds and 8 KiB. HostSleuth does not collect an arbitrary command line or environment from the unit.

### Bounded journal evidence

HostSleuth reads at most the newest 20 current-boot journal lines for the selected unit, with a two-second timeout and 16 KiB output bound. Existing journal sanitization is reused before evidence is returned.

If journal access is unavailable, that remains `unknown`; it is not converted into a service failure.

### Process/listener ownership

When systemd exposes a main PID/control group, HostSleuth reads the unit's cgroup PID membership from `cgroup.procs` and correlates those PIDs with the existing listener snapshot.

An expected-port collision is only stated when the competing listener has positive PID evidence and that PID does not belong to the selected service. A listener whose owning PID is hidden by local permissions remains `unknown` ownership.

### Endpoint and container correlation

An optional expected `host:port` can be supplied. Service Story then reuses the existing deterministic Diagnose engine for current route/TCP/local-listener/container/firewall/TLS/certificate evidence.

Existing container publication data is also correlated with the expected host port. This is context, not an assumption that the selected systemd service owns a container.

### Related retained changes

Service Story first finds direct retained events for the selected unit and expected listener port. If a direct event exists, it can also show package, configuration, and container events within a bounded +/- 15 minute window around the newest direct event.

That window is explicitly labeled context. M7 does not state that temporal proximity proves causation. M8 remains the broader Incident Lens milestone.

### CLI, API, and Web UI

CLI:

```text
hostsleuth service [-state-dir PATH] [-target HOST:PORT] UNIT
```

API:

```text
GET /api/service-story?service=UNIT&target=HOST:PORT
```

The Web UI keeps the accepted four top-level tabs. Service Story is integrated into **Diagnose** rather than creating a new generic service dashboard. Its JavaScript/CSS are isolated as small embedded assets appended to the existing application assets.

Docker mode remains truthful: native systemd Service Story evidence is unavailable there rather than being simulated from incomplete container visibility.

## Focused tests

M7 tests cover:

- failed service plus a positively identified competing listener;
- active service that owns the expected port and has a reachable endpoint;
- nearby package/configuration context bounded around a direct service event;
- Docker-mode unavailability;
- unit-name normalization;
- regression coverage proving that a listener with a hidden PID does not become a false port-collision claim;
- journal secret redaction reuse.

Full branch validation passed:

- `gofmt` clean;
- `go test ./...`;
- `go vet ./...`;
- native build;
- existing Web JavaScript syntax;
- Service Story JavaScript syntax;
- concatenated served JavaScript syntax.

## Real-host acceptance

Acceptance ran on the real OMV Debian host with a branch binary and isolated state under `/tmp`. The live HostSleuth deployment was not stopped, restarted, reconfigured, or upgraded.

The first SSH acceptance exposed a correctness defect: the unprivileged snapshot could see TCP/22 listening but could not see the SSH listener PID. The initial implementation incorrectly interpreted that missing owner metadata as a competing process.

The fix requires positive listener PID evidence before HostSleuth can claim a port collision. A dedicated regression test was added.

The corrected acceptance for `ssh.service` + `127.0.0.1:22` reported:

- runtime `loaded / active / running`;
- main PID and cgroup membership available;
- journal evidence unavailable because of local permissions, reported as `unknown`;
- TCP/22 listener present but process ownership unavailable, reported as `unknown`;
- expected endpoint reachable;
- final conclusion: `service is active and the expected endpoint is reachable` with high confidence.

A separate isolated Web/API smoke confirmed:

- the served JavaScript contains the embedded Service Story module;
- the served CSS contains the Service Story layout;
- `/api/service-story` returns the accepted SSH story;
- no production state or listener was reused for the test server.

## Boundary after M7

M7 does not authorize any service-control action. The next roadmap milestone remains M8 — Incident Lens.
