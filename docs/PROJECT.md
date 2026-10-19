# Project records

For people working on HostSleuth itself. Users want the [README](../README.md).

## Product boundaries

HostSleuth is intentionally:

- local-first;
- single-host first;
- read-only by default;
- deterministic before explanatory;
- loopback-only by default for sensitive Web/API operations;
- explicit rather than automatic about its narrowly fixed native Safe Actions.

HostSleuth does **not** automatically restart or reload services, modify firewall rules, repair containers, install/remove/update packages, edit configuration, renew/install certificates, reboot/shut down the host, manage ACME accounts, handle private keys, or reconfigure the host. Native mode exposes only two fixed state-changing capabilities: independently allowlisted `service.restart` and `service.reload`. Docker mode exposes neither action.

## Current stage

All roadmap milestones through M18 are complete and published, together with the v1.0 readiness/hardening cycle and the Redacted Evidence Bundle.

The newest stable release, with native amd64/arm64 binaries, `SHA256SUMS` and notes, is always on the [Releases page](https://github.com/xXDasGoGXx/HostSleuth/releases/latest); Docker `mjmalleo/hostsleuth:latest` tracks it. The exact published source, verification and live deployment record for the current release are in [`CURRENT-HANDOFF.md`](../CURRENT-HANDOFF.md).

Since v1.0: v1.1.0 made the change timeline meaningful on busy container hosts (routine socket churn is not recorded) and restored the build revision in Docker images; v1.2.0 collapses listeners that go away and come back (reboots, restarts) into one line, surfaces unhealthy containers, and reworks the console for tablets and phones; v1.3.0 adds Container Story (restarts with Docker's exit reason, out-of-memory kills, image updates), Share redacted on every result, and a rewritten README with screenshots from demo data; v1.4.0 makes Replay the home screen: the containers drawn as a blueprint, rewindable through the last seven days with a plain-sentence account of each moment; v1.5.0 explains why a container is unhealthy from Docker's own health check.

The owner-approved fixed native `service.reload` and `service.restart` actions remain native-only and independently allowlisted. Docker mode exposes neither.

See [`docs/design/V1.0-READINESS.md`](design/V1.0-READINESS.md) for the readiness gates, [`docs/design/V1.0-RELEASE-CHECKLIST.md`](design/V1.0-RELEASE-CHECKLIST.md) for publication/rollback ordering, and [`docs/ROADMAP.md`](ROADMAP.md) for product guardrails.

## Project files

- `README.md` — product definition, usage, and current scope.
- `CURRENT-HANDOFF.md` — current project state and next task.
- `docs/OPERATIONS.md` — operational architecture, credentials-by-name, deployment, and rebuild/recovery map.
- `ops/project.yaml` — machine-readable project/integration inventory.
- `ops/README.md` — reusable operations-registry convention for this and future repositories.
- `docker/DOCKERHUB.md` — canonical Docker Hub Overview source.
- `docker/DOCKERHUB-SHORT.txt` — canonical Docker Hub short-description source.
- `TO-DO.md` — active checklist and product decisions.
- `docs/ROADMAP.md` — owner-approved ordered product roadmap.
- `docs/design/V1.0-READINESS.md` — feature-free v1.0 hardening gates and release-readiness contract.
- `docs/design/V1.0-RELEASE-CHECKLIST.md` — exact v1.0 publication, live acceptance, recovery-alignment, and rollback checklist.
- `docs/history/V1.0-READINESS.md` — v1.0 technical readiness acceptance closeout.
- `docs/history/V1.0.1-PUBLICATION.md` — v1.0.1 publication and independent verification record.
- `docs/history/V1.0.1-PRODUCTION-ALIGNMENT.md` — v1.0.1 live/recovery rollout and acceptance record.
- `docs/history/V1.0.0-PUBLICATION.md` — v1.0.0 publication and independent verification record.
- `docs/history/V1.0.0-PRODUCTION-ALIGNMENT.md` — v1.0.0 live/recovery rollout and acceptance record.
- `docs/history/DEVELOPMENT-HISTORY.md` — milestone and validation history.
- `docs/history/M10-REBOOT-STORY.md` — completed M10 implementation and acceptance record.
- `docs/history/M11-OPTIONAL-SAFE-ACTIONS.md` — completed M11 security model and acceptance record.
- `docs/history/V0.4.0-PUBLICATION.md` — v0.4.0 publication and verification record.
- `docs/design/REDACTED-EVIDENCE-BUNDLE.md` — accepted bundle threat model and redaction contract.
- `docs/history/REDACTED-EVIDENCE-BUNDLE.md` — bundle implementation and acceptance closeout.
- `docs/history/V0.5.0-PUBLICATION.md` — v0.5.0 publication and verification record.
- `docs/history/V0.5.0-PRODUCTION-ALIGNMENT.md` — v0.5.0 live/recovery rollout and acceptance record.
- `docs/design/M12-EXPECTED-ENDPOINT-CONTRACTS.md` — M12 contract semantics and security/product boundary.
- `docs/history/M12-EXPECTED-ENDPOINT-CONTRACTS.md` — M12 implementation and acceptance closeout.
- `docs/history/V0.6.0-PUBLICATION.md` — v0.6.0 publication and independent verification record.
- `docs/history/V0.6.0-PRODUCTION-ALIGNMENT.md` — v0.6.0 live/recovery rollout and acceptance record.
- `docs/design/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md` — M13 DNS Detective and Admin Console v1 design.
- `docs/history/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md` — M13 implementation and acceptance closeout.
- `docs/history/V0.7.0-PUBLICATION.md` — v0.7.0 publication and independent verification record.
- `docs/history/V0.7.0-PRODUCTION-ALIGNMENT.md` — v0.7.0 live/recovery rollout and acceptance record.
- `docs/design/M14-REVERSE-PROXY-UPSTREAM-STORY.md` — M14 request-path semantics and HTTP/security boundary.
- `docs/history/M14-REVERSE-PROXY-UPSTREAM-STORY.md` — M14 implementation and local acceptance closeout.
- `docs/design/M15-DEPLOYMENT-PERMISSIONS-STORY.md` — M15 permission semantics and privacy/security boundary.
- `docs/history/M15-DEPLOYMENT-PERMISSIONS-STORY.md` — M15 implementation and acceptance closeout.
- `docs/design/M16-STARTTLS-MAIL-SERVICE-STORY.md` — M16 mail protocol/TLS semantics and safety boundary.
- `docs/history/M16-STARTTLS-MAIL-SERVICE-STORY.md` — M16 implementation and acceptance closeout.
- `docs/design/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md` — M17 rollout-comparison semantics and no-private-key-read boundary.
- `docs/history/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md` — M17 implementation and acceptance closeout.
- `docs/design/M18-SAFE-ACTIONS-II-SECURITY-REVIEW.md` — M18 candidate comparison, accepted threat model, and privilege-cost contract.
- `docs/history/M18-SAFE-ACTIONS-II.md` — M18 implementation and real native reload acceptance closeout.
- `docs/history/V0.8.0-PUBLICATION.md` — v0.8.0 publication and independent verification record.
- `docs/history/V0.8.0-PRODUCTION-ALIGNMENT.md` — v0.8.0 live/recovery rollout and acceptance record.
- `docs/history/V0.9.0-PUBLICATION.md` — v0.9.0 publication and independent verification record.
- `docs/history/V0.9.0-PRODUCTION-ALIGNMENT.md` — v0.9.0 live/recovery rollout and acceptance record.
- `docs/history/V1.0.0-PUBLICATION.md` — v1.0.0 publication and independent verification record.
- `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md` — sourced research input for differentiated future workflows; not active scope by itself.
