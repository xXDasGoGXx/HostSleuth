# M18 — Safe Actions II Security Review

Status: owner-approved source implementation complete and merged. See `docs/history/M18-SAFE-ACTIONS-II.md` for acceptance details.

## Goal

M18 may add exactly one additional Safe Action only after a fresh privilege and threat-model gate.

The existing M11 action is:

    service.restart

M18 must not turn HostSleuth into a generic command runner, generic systemd controller, Docker controller, or automatic-remediation engine.

## Existing M11 boundary

The current native Safe Action pattern already requires:

- actions disabled by default;
- explicit global opt-in with --enable-actions;
- explicit per-service allowlisting;
- fixed action IDs and command verbs in code;
- conservative .service unit validation;
- trusted absolute /usr/bin/systemctl or /bin/systemctl;
- no shell or arbitrary argv;
- Docker mode unavailable for native systemd actions;
- deterministic preview;
- exact typed confirmation;
- loopback-only state-changing Web/API access;
- bounded JSON plus X-HostSleuth-Action: confirm;
- durable requested audit before execution;
- serialized execution;
- bounded command runtime/output;
- bounded before/after evidence;
- deterministic postcondition verification.

The native HostSleuth systemd unit already runs as root. M18 therefore must not add sudoers rules, a generic privilege broker, extra Linux capabilities, an unrestricted shell, or a writable Docker control path. The relevant privilege cost is expansion of HostSleuth's exposed state-changing surface, not a need for new operating-system credentials.

## Candidate comparison

### service.reload — preferred candidate

Effect:

    systemctl reload UNIT.service

Value:

- useful after supported configuration or certificate changes;
- can apply changes without the process replacement/downtime of a restart;
- complements existing configuration, TLS, STARTTLS, certificate-rollout, and service stories;
- remains targetable to exactly one explicitly named systemd service.

Privilege and blast radius:

- same native systemd command family and target shape as M11;
- no Docker control;
- no manager-wide operation;
- no arbitrary argv;
- service-defined ExecReload may itself perform privileged work, so reload requires its own independent allowlist and must never inherit the restart allowlist.

Security fit:

- strongest fit with the existing M11 architecture;
- can reuse trusted systemctl resolution, confirmation, durable audit, serialization, timeout/output limits, and before/after evidence;
- preview should require the unit to be loaded, active, and reported reload-capable;
- reload must never fall back to restart.

Operational caveat:

A successful systemd reload proves that systemd accepted and completed the unit's reload operation and that HostSleuth's postcondition passed. It does not prove every application-specific configuration setting was semantically applied. HostSleuth must not overclaim that.

### service.start

Effect:

    systemctl start UNIT.service

Value:

- useful for recovering a stopped or failed service.

Privilege and blast radius:

- can activate a deliberately dormant service;
- can open listeners, start dependencies, or trigger application side effects;
- changes a service from inactive to running.

Assessment:

Higher operational impact than reload. Technically compatible with a unit allowlist, but not preferred for the single M18 addition.

### service.stop

Effect:

    systemctl stop UNIT.service

Value:

- useful for maintenance or emergency containment.

Privilege and blast radius:

- intentionally causes unavailability;
- can interrupt users/workloads;
- dependency relationships can amplify impact;
- a mistaken allowlist entry becomes a direct denial-of-service control.

Assessment:

Reject for M18.

### service.reset-failed

Effect:

    systemctl reset-failed UNIT.service

Value:

- clears systemd failure state/counters.

Privilege and blast radius:

- relatively narrow mutation;
- does not restore service function;
- can remove useful failure-state evidence while leaving the problem untouched.

Assessment:

Low privilege cost but weak remediation value and poor fit for a diagnostic flight recorder. Reject for M18.

### container.restart

Value:

- potentially useful for container workloads.

Privilege and blast radius:

- intentionally turns Docker-daemon access into a state-changing capability;
- Docker API access is effectively host-powerful even when the socket path is mounted read-only;
- conflicts with the M11 boundary that Docker mode exposes no host action;
- pressures HostSleuth toward a generic container controller.

Assessment:

Reject. M18 must not weaken Docker security.

### systemd daemon-reload

Effect:

    systemctl daemon-reload

Value:

- makes systemd re-read unit definitions.

Privilege and blast radius:

- manager-wide rather than target-bound;
- affects all changed unit definitions;
- has no clean per-unit allowlist or postcondition.

Assessment:

Reject. Scope is too broad.

### service.reload-or-restart

Value:

- convenient when reload support varies.

Privilege and blast radius:

- effect is not one fixed primitive;
- a reload request may silently become restart;
- preview/confirmation no longer describe one deterministic mutation.

Assessment:

Reject. Convenience is not worth weakening exact-effect semantics.

## Approved candidate

The owner explicitly approved exactly one M18 action:

    service.reload

Implementation must remain inside this document's security contract. No additional M18 action family is approved.

## Required implementation contract

If approved, service.reload must satisfy all of the following.

### Fixed action identity

Action ID:

    service.reload

No generic operation or verb parameter.

### Separate allowlist

Add a dedicated reload allowlist such as:

    --allow-reload-service UNIT

Restart permission must not imply reload permission, and reload permission must not imply restart permission.

### Existing global opt-in

The existing --enable-actions gate remains required.

No target is available by default.

### Native mode only

service.reload remains unavailable in Docker deployment mode.

Do not add Docker-socket mutation or a host-systemd bridge.

### Conservative target validation

Reuse the existing conservative .service unit-name validator.

No arbitrary unit type, command, option, environment variable, path, or argv field.

### Trusted executable

Use only the existing trusted absolute systemctl resolver:

- /usr/bin/systemctl
- /bin/systemctl

Never resolve the state-changing executable through PATH.

### Precondition

Preview must establish bounded systemd evidence showing at minimum:

- unit is installed/loaded;
- ActiveState=active;
- systemd reports the unit as reload-capable.

Implementation may add action-specific trusted systemctl-show capability evidence. If reliable reload-capability evidence is unavailable, preview fails closed.

### Preview and confirmation

Preview shows:

- exact action ID;
- exact target;
- exact command argv;
- bounded before state;
- reload-capability precondition;
- exact effect;
- exact confirmation.

Expected command shape:

    /usr/bin/systemctl reload UNIT.service

or the trusted /bin/systemctl equivalent.

Expected confirmation:

    RELOAD UNIT.service

### Execution-time revalidation

Run must re-run policy and preconditions inside the serialized action path before executing.

A stale browser preview cannot bypass current policy/preconditions.

### Durable pre-execution audit

A requested audit record must be durably written before systemctl reload executes.

If audit storage is unavailable, execution does not occur.

### Bounded execution

Reuse bounded timeout/output limits unless focused testing justifies tighter limits.

No shell and no generic argv.

### No fallback

If reload is unsupported or fails:

- report failure;
- retain evidence;
- do not restart;
- do not retry with another state-changing action.

### Postcondition

Success requires:

- systemctl reload returned success;
- post-action runtime evidence is available;
- ActiveState remains active.

Result wording must say only that the allowlisted reload command succeeded and the service remained active. It must not claim arbitrary application configuration was semantically applied.

### Final audit

Write a completed audit record with status and bounded before/after evidence.

A final-audit write failure remains visible even if reload already occurred; HostSleuth cannot undo a completed reload.

### Web/API boundary

Keep the existing boundary:

- loopback-only;
- bounded JSON;
- unknown fields rejected;
- X-HostSleuth-Action: confirm required on execution;
- exact confirmation required.

A LAN-bound read-only UI must not gain remote state-changing access.

## Threat model

### Arbitrary command/argument injection

Controls:

- fixed service.reload action ID;
- fixed reload verb;
- conservative .service regex;
- server-side explicit reload allowlist;
- trusted absolute executable;
- no shell;
- no arbitrary argv fields.

### Restart/reload action confusion

Controls:

- separate allowlists;
- distinct action IDs;
- distinct preview text;
- confirmation must begin RELOAD and contain the exact unit.

### Remote/LAN-triggered mutation

Controls:

- loopback-only action API;
- local browser/SSH tunnel or CLI workflow;
- explicit execution header;
- exact confirmation.

### Unsupported reload silently becoming restart

Controls:

- require reload-capability evidence;
- execute only systemctl reload;
- no reload-or-restart helper;
- no fallback.

### Powerful or surprising ExecReload behavior

Risk:

A systemd unit reload handler may itself execute privileged commands or have application-specific side effects.

Controls:

- independent reload allowlist;
- no default targets;
- exact target/effect preview;
- operator remains responsible for which installed unit definitions are trusted for reload.

HostSleuth must not execute or interpret arbitrary ExecReload content itself.

### Preview-to-run staleness

Controls:

- run re-evaluates preview/policy/preconditions under the ActionManager serialization lock;
- exact confirmation remains required.

An independent root administrator can still change a unit definition outside HostSleuth; M18 must not expand into generic unit-file locking/configuration management.

### Unaudited mutation

Control:

- requested audit must persist before command execution;
- fail closed if it cannot.

### False success

Controls:

- require command success;
- require post-action runtime evidence;
- require ActiveState=active;
- record bounded command output;
- state only the systemd-level postcondition.

### Concurrent/repeated mutation

Controls:

- retain ActionManager mutex;
- one bounded action at a time;
- no automatic retries.

### Docker privilege expansion

Controls:

- native mode only;
- no Docker write path;
- no new socket mode or bridge.

## Privilege-cost conclusion

service.reload expands HostSleuth's exposed state-changing surface, but requires no new OS privilege mechanism in the native installation because HostSleuth already runs as root.

The acceptable design depends on keeping the software boundary narrow:

- exact action;
- exact target;
- independent allowlist;
- trusted executable;
- fail-closed preconditions;
- preview/confirmation;
- pre-execution audit;
- bounded execution;
- deterministic postcondition;
- native-only availability.

Among the candidates reviewed, service.reload provides the strongest operational value while preserving the M11 architecture and avoiding Docker, manager-wide, availability-control, or generic-command expansion.

## Decision gate

Owner decision:

    APPROVED: service.reload

Approved implementation scope:

- exactly service.reload;
- independent reload allowlist;
- native mode only;
- loaded + active + CanReload=yes precondition;
- fixed trusted systemctl reload argv;
- exact RELOAD confirmation;
- pre-execution audit;
- bounded serialized execution;
- no restart fallback;
- ActiveState=active postcondition;
- existing loopback-only Web/API boundary.

No other candidate from this review is approved for implementation.

Source acceptance completed under this contract.

PR #62 exact final head:

`c46dd72ad806c688970244836ce26e3ca01eec88`

CI run:

`35309972305` — all jobs successful.

Real disposable native systemd acceptance proved independent reload allowlisting, `CanReload=yes` gating, exact reload argv/confirmation, unchanged MainPID across reload, active postcondition, durable audit records, and cleanup.

PR #62 squash-merged to `main` at:

`cc9c9727fe19786eba03d0b7ef51c7fae7ac8ab1`

Full closeout: `docs/history/M18-SAFE-ACTIONS-II.md`.

Stable/public/live/recovery remain on v0.8.0 unless a separate publication/rollout decision is made.
