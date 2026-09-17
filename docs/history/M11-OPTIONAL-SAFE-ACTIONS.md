# M11 — Optional Safe Actions — Complete

M11 is complete in source. It is the first HostSleuth milestone to permit a bounded state-changing operation, while preserving read-only behavior as the default product mode.

## Goal

Test whether HostSleuth can offer a surgical administrative action without becoming a generic server-control panel, browser shell, Docker controller, or automatic-remediation engine.

M11 deliberately proves one action only:

`service.restart`

It restarts one explicitly allowlisted native systemd `.service` unit and reports success only when the service is observed active afterward.

## Security model

Actions are fail-closed and disabled by default.

To make `service.restart` eligible, the operator must both:

1. explicitly enable actions with `--enable-actions`; and
2. explicitly allow the target using repeated `--allow-restart-service UNIT` flags.

No target is permitted by default.

Additional boundaries:

- fixed action ID implemented by HostSleuth;
- conservative systemd service-unit validation;
- no arbitrary command, argv, script, or shell field;
- no generic systemd controller;
- no package, firewall, configuration, or generic file-write action;
- no automatic remediation;
- no Docker-container control action;
- no writable Docker socket added for M11;
- Docker deployment mode reports native systemd restart unavailable;
- exact preview before execution;
- exact confirmation value required before execution;
- state-changing Web/API operations are loopback-only;
- Web execution requires JSON plus `X-HostSleuth-Action: confirm`;
- bounded command output and action timeout;
- durable audit write required before the restart command executes;
- final outcome audit after the attempt;
- serialized execution through the action manager;
- success requires an observed `ActiveState=active` postcondition.

## Trusted systemctl boundary

During M11 validation, the state-changing path was tightened so it does not trust `$PATH` to resolve `systemctl`.

Action precondition evidence, postcondition evidence, previewed argv, and restart execution use only a trusted executable found at:

- `/usr/bin/systemctl`, or
- `/bin/systemctl`.

A relative/PATH-resolved executable is not accepted for the action path. The existing read-only evidence collectors remain independent from this stricter state-changing boundary.

## Delivered implementation

### Core

M11 adds:

- `ActionPolicy`;
- `ActionCapability`;
- `ActionPreview`;
- `ActionResult`;
- `ActionAudit`;
- `ActionManager`;
- explicit restart-service allowlisting;
- trusted systemctl resolution;
- bounded native systemd before/after evidence;
- durable `actions.jsonl` audit storage;
- exact confirmation enforcement;
- postcondition verification.

A restart is refused if durable pre-execution audit storage is unavailable.

### CLI

Available commands:

```text
hostsleuth action list
hostsleuth action preview
hostsleuth action run
hostsleuth action audit
```

Action enablement/allowlisting is explicit through:

```text
--enable-actions
--allow-restart-service UNIT
```

### JSON API

Loopback-only action endpoints:

```text
GET  /api/actions
POST /api/actions/preview
POST /api/actions/run
GET  /api/actions/audit
```

POST requests are bounded JSON requests with unknown fields rejected. Execution additionally requires the explicit action header and exact confirmation returned by preview.

### Web UI

M11 adds an **Actions** view that:

- shows whether the capability is disabled, enabled-but-unavailable, or available;
- exposes only server-provided allowlisted targets;
- previews the exact effect and argv;
- displays before-state evidence and eligibility checks;
- enables execution only when the exact confirmation is typed;
- displays bounded before/after action evidence;
- shows recent audit records.

The action API remains loopback-only even when the read-only HostSleuth UI is intentionally bound to a LAN address. Remote action use therefore requires the existing SSH-tunnel/local-browser pattern.

## Tests

Focused regression/security coverage proves:

- unsafe service names are rejected;
- disabled actions do not probe systemd;
- targets must be allowlisted;
- Docker mode cannot expose native service restart;
- exact confirmation is required;
- denied attempts are audited without executing restart;
- successful action results require the postcondition;
- postcondition failure is not reported as success;
- action execution is refused when the audit log cannot be created;
- command failures are reported as failures;
- relative/PATH-provided executables are not trusted by the action path;
- Action API calls are loopback-only;
- execution requires the explicit action header;
- JSON requests reject unknown fields;
- no action becomes available without explicit opt-in.

## Validation

### Exact-branch local validation

The M11 branch was cloned into an isolated `/tmp` directory on the OMV host and validated with the same Go 1.24.13 toolchain used by CI.

Passed:

- `gofmt` cleanliness;
- `go vet ./...`;
- `go test ./...`;
- native build;
- individual JavaScript syntax checks;
- exact concatenated served-JavaScript syntax.

No production HostSleuth process, Compose project, systemd service, or recovery definition was changed by this validation.

### CI run #205

Functional source at commit:

`25318fed095f311221b44d67a427f033b1a46993`

CI run #205 passed all jobs:

- test / formatting / vet / Go tests / JavaScript / native build;
- native-actions-smoke;
- Docker runtime smoke;
- linux/amd64 image build;
- linux/arm64 image build.

Docker smoke confirmed that Optional Safe Actions remain disabled and unavailable by default in the supported Docker deployment.

### Disposable native systemd acceptance

The `native-actions-smoke` job used a disposable systemd unit on the ephemeral GitHub-hosted Ubuntu runner. It did not use the OMV production host.

Acceptance proved:

1. the temporary service was active before testing;
2. preview returned the allowlisted target, trusted absolute systemctl argv, and exact confirmation;
3. an intentionally wrong confirmation returned `denied` and the service PID did not change;
4. the exact confirmation executed the allowlisted restart;
5. the service PID changed after the real restart;
6. HostSleuth reported `success` only after observing `ActiveState=active`;
7. the audit sequence contained denied, requested, and completed-success records;
8. the disposable service was removed afterward.

This provides a real native systemd success-path acceptance without changing any production home-lab service or bypassing HomeCommander safeguards.

## Production and release boundary

M11 completion is a source milestone only.

It did **not**:

- redeploy or restart the production HostSleuth container;
- change the OMV Compose project;
- change the disaster-recovery image pin;
- publish a release;
- move the stable public release beyond `v0.3.0`;
- enable actions in the current Docker deployment.

Production and disaster recovery remain aligned on `mjmalleo/hostsleuth:0.3.0`, which does not contain M7–M11 source work.

## Closeout boundary

Do not add more action families merely because the framework now exists. Any future action must independently justify its privilege cost and preserve the same explicit-schema, allowlist, preview, confirmation, audit, and postcondition model.

A public release/deployment decision is separate from M11 completion. The later Redacted Evidence Bundle remains a separate future milestone and must not start automatically as part of this closeout.
