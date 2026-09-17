# M11 — Optional Safe Actions — Work in Progress

M11 is the first HostSleuth milestone that may perform a bounded state-changing operation. Read-only behavior remains the default and continues to work without action support.

## Security contract

HostSleuth actions are not a general command runner.

An M11 action must:

- have a fixed action ID implemented by HostSleuth;
- target only an explicitly allowlisted object;
- expose a deterministic preview before execution;
- require an explicit confirmation value for execution;
- execute without a shell and without user-controlled command/argument injection;
- capture bounded before/after evidence;
- enforce a timeout;
- report success only when its postcondition is observed;
- append an audit record for every attempted execution;
- fail closed when unavailable, unsupported, disabled, or unauthorized.

M11 must not add:

- arbitrary shell or command execution;
- arbitrary systemd unit control;
- package-manager actions;
- firewall mutation;
- generic file writes/editing;
- writable Docker-socket control;
- automatic remediation.

## Initial scope

The first real action is `service.restart`, available only in native mode for systemd units explicitly listed by the operator.

The allowlist is supplied by repeated `--allow-restart-service UNIT` flags. Unit names are normalized to `.service` and validated against a conservative systemd-unit character set. No service is allowed by default, and the action framework itself must also be explicitly enabled.

Execution uses `systemctl restart UNIT` through direct argv execution. It does not invoke a shell. HostSleuth records the observed systemd state before execution and verifies `ActiveState=active` afterward before reporting success.

The HTTP action endpoints are loopback-only even when the read-only HostSleuth UI is intentionally bound to a LAN address. Remote use therefore requires the existing SSH-tunnel workflow. Docker deployment mode does not receive a writable Docker socket or hidden host-control path; service restart is reported unavailable there.

## M11 delivery order

1. Security contract and action data model.
2. Registry/allowlist and preview.
3. Confirmed `service.restart` execution with bounded audit records.
4. CLI and loopback-only JSON API integration.
5. Focused security/regression tests.
6. Web UI integration after the API/CLI behavior is accepted.
7. Native-host acceptance.
8. Final documentation/closeout.

No release or production deployment is part of M11 implementation unless separately approved.
