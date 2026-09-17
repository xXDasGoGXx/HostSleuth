# M12 — Expected Endpoint Contracts Closeout

Status: source implementation complete; publication remains a separate step.

## Delivered

M12 adds one bounded, read-only Expected Endpoint Contract evaluation path.

A user can state a target plus optional DNS/TLS/service/container expectations. HostSleuth evaluates the contract against current deterministic evidence and returns:

- normalized contract;
- ordered Expected-vs-Observed checks;
- `pass`, `fail`, or `unknown` per expectation;
- overall status;
- the first proven mismatch, or first unresolved check when nothing is proven false;
- the existing full Diagnosis evidence for the same target.

Interfaces delivered:

- `hostsleuth contract` CLI;
- `GET /api/contract`;
- Web UI **Expectations** view;
- one-click handoff from a contract result to the existing full Diagnose view.

## Contract fields

- target: required `host:port`;
- expected DNS IPs: optional, normalized exact-set semantics;
- TLS: `ignore`, `present`, `verified`, or `forbidden`;
- systemd service: optional active-state expectation;
- container: optional running-state expectation.

TCP success is always required.

## Deterministic status rules

A proven failure wins over unknown evidence. If multiple expectations fail, the first failed check in contract order is reported as the first mismatch.

If nothing fails but one or more checks are unavailable, overall state is `unknown`; HostSleuth does not turn missing Docker/systemd evidence into a false failure.

## Security / product boundary

M12 remains read-only and adds no privileges, command surface, file-content access, scheduler, monitoring loop, notification system, or remediation path.

The new API reuses the existing Diagnose probe class. Service/container values are snapshot comparisons only.

## Validation

Local Go 1.24.13 validation passed:

- format;
- vet;
- full Go tests;
- native build;
- M12 JavaScript syntax.

Disposable server acceptance confirmed pass/fail behavior, first-mismatch reporting, HTTP 400 validation, and delivery of the combined UI asset.

## Release boundary

Stable/public/live/recovery remain v0.5.0 until M12 is independently published and rolled out through the normal release process.
