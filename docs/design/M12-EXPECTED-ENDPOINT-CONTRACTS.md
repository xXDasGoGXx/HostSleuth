# M12 — Expected Endpoint Contracts

Status: implemented and locally accepted on the `m12-expected-endpoint-contracts` branch.

## Product question

> What should be true for this endpoint, and which expectation is currently the first proven mismatch?

M12 extends the existing deterministic Diagnose workflow. It does not create a polling loop, uptime monitor, alerting system, configuration manager, or remediation engine.

## Bounded v1 contract

One on-demand contract contains:

- optional human-readable name;
- required `host:port` target;
- optional exact DNS address set;
- TLS expectation:
  - `ignore` — TLS is not part of the contract;
  - `present` — a TLS handshake must succeed;
  - `verified` — handshake, hostname verification, and trust verification must succeed;
  - `forbidden` — a TLS handshake must not succeed;
- optional systemd service expected to be active;
- optional container expected to be running.

TCP reachability is always required.

Expected DNS addresses use normalized exact-set semantics. Order and duplicates do not matter. Extra or missing addresses are a mismatch. This is deliberately explicit because some endpoints use dynamic/CDN answers and should simply omit the exact-address expectation.

## Evaluation model

HostSleuth runs the existing `Diagnose` engine once and builds an ordered Expected-vs-Observed checklist.

Current order:

1. DNS;
2. TCP;
3. TLS when requested;
4. systemd service when requested;
5. container when requested.

Each expectation is one of:

- `pass` — current evidence proves the requested state;
- `fail` — current evidence contradicts the requested state;
- `unknown` — HostSleuth cannot honestly prove or disprove it with the available evidence.

Overall result:

- any failed expectation => `fail`, with the first **proven failure** reported as `first_mismatch`;
- no failures but one or more unknowns => `unknown`, with the first unresolved check reported;
- all requested expectations pass => `pass`.

An earlier unknown never masks a later proven failure.

The complete existing Diagnosis is included in the result so route, listener, Docker-port, firewall, TLS/certificate, and other current evidence remains inspectable without duplicating those engines inside M12.

## Deployment-mode truthfulness

Native systemd expectations remain `unknown` in Docker mode because the supported Docker deployment does not have native systemd inventory/control visibility.

A requested container expectation is `unknown` when Docker inventory is absent or inaccessible. If Docker inventory is available, an absent named container is a failure and a non-running named container is a failure.

M12 does not infer that a named container owns the target port. Host-network containers and other valid deployment patterns make that unsafe to assume in v1.

## Interfaces

CLI:

```text
hostsleuth contract [options] host:port
```

Options:

```text
--name NAME
--expect-ip IP        # repeat for exact DNS set
--tls MODE            # ignore|present|verified|forbidden
--service UNIT
--container NAME
--state-dir PATH
```

HTTP API:

```text
GET /api/contract?target=host%3Aport&ip=...&tls=...&service=...&container=...
```

The Web UI adds an **Expectations** view with the same fields and renders the ordered Expected-vs-Observed checklist plus the first mismatch. A result can be handed directly to the existing full Diagnose view for deeper evidence.

## Security boundary

M12 is read-only.

It adds:

- no state-changing operation;
- no new privilege;
- no arbitrary command execution;
- no arbitrary file read;
- no secret/environment collection;
- no persistent remote-probe scheduler;
- no cloud service;
- no alerts;
- no automatic remediation.

The HTTP surface performs the same class of target probe already available through `/api/diagnose`; M12 reuses that engine rather than introducing a broader probing primitive.

Service and container expectation strings are used only for exact comparison against already-collected snapshot metadata. They are never interpolated into shell commands.

## Explicit non-goals for v1

Not included:

- persistent contract files or a contract database;
- scheduled/continuous contract evaluation;
- notifications or uptime history;
- reverse-proxy configuration parsing;
- automatic discovery of expected state;
- DNS authoritative/delegation comparison;
- container-to-port ownership inference;
- arbitrary file expectations;
- Safe Actions triggered from a contract;
- automatic fixes.

Those may be considered only as separately bounded future milestones.

## Local acceptance

An isolated OMV `/tmp` clone used Go 1.24.13.

Passed:

- `gofmt`;
- `go vet ./...`;
- `go test ./...`;
- native `go build`;
- standalone `node --check` for the M12 UI asset.

Disposable end-to-end server acceptance on `127.0.0.1:18787` confirmed:

- the new API is served;
- a plaintext target with `tls=forbidden` returns overall `pass`;
- the same target with `tls=present` returns `fail` with first mismatch `tls`;
- malformed expected IP input returns HTTP 400;
- the combined served JavaScript contains the Expectations UI/API path and passes JavaScript syntax validation.

The disposable M12 server was stopped after acceptance.
