# M18 — Safe Actions II Closeout

Status: source implementation complete and merged on 2026-09-17. Stable/public/live/recovery remain aligned on v0.8.0; publication and rollout are separate gates.

## Owner-approved action

After the required candidate comparison, threat model, and privilege-cost review, the owner explicitly approved exactly one additional Safe Action:

    service.reload

No other M18 candidate was approved.

Security review: `docs/design/M18-SAFE-ACTIONS-II-SECURITY-REVIEW.md`.

## Delivered boundary

M18 preserves the M11 fail-closed action model. `service.reload` requires explicit `--enable-actions`, its own `--allow-reload-service UNIT` allowlist, native deployment mode, conservative `.service` validation, and trusted absolute `/usr/bin/systemctl` or `/bin/systemctl`.

The action uses fixed argv `systemctl reload UNIT.service` with no shell or operator-supplied verb/argv. Preview fails closed unless the unit is loaded, already `ActiveState=active`, and trusted systemd evidence reports `CanReload=yes`. Execution requires the exact `RELOAD UNIT.service` confirmation and a durable requested audit before the command runs.

Execution remains serialized and bounded. There is no retry, `reload-or-restart`, or restart fallback. Success requires the reload command to succeed, post-action runtime evidence to remain available, and `ActiveState=active` afterward. HostSleuth states only that the reload command succeeded and the unit remained active; it does not claim application-specific configuration semantics were verified.

Restart and reload permissions are independent. A unit allowlisted for `service.restart` is not automatically allowed for `service.reload`, and vice versa.

The Actions Admin Console renders only server-provided fixed capabilities and allowlisted targets. State-changing Web/API operations remain loopback-only, bounded JSON, exact-confirmation operations requiring `X-HostSleuth-Action: confirm`. Docker mode exposes neither native systemd action.

## Local validation

The implementation was validated from an isolated `/tmp` working tree on the OMV host with Go 1.24.13. Full `go test ./...`, `go test -race ./internal/core`, `go vet ./...`, native build, Actions JavaScript syntax, focused action tests, and `git diff --check` all passed.

No live HostSleuth service, container, certificate, DNS/proxy configuration, production deployment, or recovery definition was changed.

## GitHub CI acceptance

PR #62 final exact head:

    c46dd72ad806c688970244836ce26e3ca01eec88

CI run:

    35309972305

All jobs passed: format, vet, full Go tests, Web JavaScript syntax, native build, Docker smoke, linux/amd64 image build, linux/arm64 image build, and native Safe Actions smoke.

Docker smoke confirmed both fixed native action capabilities remain disabled and unavailable by default in the supported Docker deployment.

## Real disposable native reload acceptance

The native-actions CI job created a disposable systemd service with `ExecStart=/bin/sleep infinity` and `ExecReload=/bin/true`.

Acceptance proved that restart-only allowlisting could not authorize reload; reload-specific allowlisting plus `CanReload=yes` made the action eligible; preview returned the exact trusted `reload hostsleuth-m11-accept.service` argv and `RELOAD hostsleuth-m11-accept.service` confirmation; the real reload succeeded; MainPID remained unchanged; the unit remained active; requested/completed-success reload audits were present; and the disposable service was removed afterward.

This exercises the real state-changing reload path without modifying any home-lab production service.

## Merge acceptance

PR #62 squash-merged to `main` at:

    cc9c9727fe19786eba03d0b7ef51c7fae7ac8ab1

## Release boundary

M18 source is complete on `main`. Stable/public/live/recovery remain `v0.8.0` / `mjmalleo/hostsleuth:0.8.0`.

M18 did not publish a release, push a public Docker tag, redeploy Arcane, modify disaster recovery, enable Safe Actions in live Docker, add Docker/container mutation, add a generic systemd controller, add arbitrary command execution, add another action family, or change certificates/DNS/proxy/mail/firewall state.

The owner-approved roadmap through M18 is now source-complete. Any new feature milestone, publication, or production/recovery rollout requires a separate owner decision.
