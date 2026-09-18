# HostSleuth

**A small, local-first Linux flight recorder for two questions: what changed, and why can I not reach this host/service/port?**

HostSleuth records meaningful Linux state changes and performs deterministic, evidence-backed `host:port` diagnosis. It does not require a cloud account, external database, API key, or AI model.

## Quick start — recommended native install

Native Linux is the recommended deployment because it gives HostSleuth the fullest view of the machine.

Requirements: Linux with systemd. Go is **not** required when installing a release. The release installer downloads `SHA256SUMS` and verifies the architecture-specific binary before installing it.

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
sudo ./scripts/install.sh
```

Check that it is running:

```bash
systemctl status hostsleuth --no-pager
hostsleuth version
```

HostSleuth binds its Web UI to loopback by default for safety.

If you are using HostSleuth on the same machine, open:

```text
http://127.0.0.1:8787
```

If HostSleuth is running on a remote server, use an SSH tunnel from your computer:

```bash
ssh -L 8787:127.0.0.1:8787 user@your-server
```

Then open `http://127.0.0.1:8787` locally.

A new install starts by taking a baseline snapshot. An empty change timeline is normal until HostSleuth observes a meaningful change.

## Docker — convenient, reduced host visibility

The supported public image is:

```text
mjmalleo/hostsleuth
```

`latest` represents the newest stable release. Stable releases also receive an explicit numeric tag such as `0.9.0`.

### Docker Compose

The repository includes one supported `compose.yaml` for Linux Docker hosts:

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
docker compose pull
docker compose up -d
```

To pin the current stable release instead of `latest`:

```bash
HOSTSLEUTH_IMAGE=mjmalleo/hostsleuth:0.8.0 docker compose up -d
```

### Docker run

The equivalent direct `docker run` deployment is:

```bash
docker run -d \
  --name hostsleuth \
  --restart unless-stopped \
  --network host \
  --pid host \
  --uts host \
  --read-only \
  --cap-drop ALL \
  --security-opt no-new-privileges:true \
  --tmpfs /tmp:rw,noexec,nosuid,size=16m \
  -v hostsleuth-data:/var/lib/hostsleuth \
  -v /etc/os-release:/host/etc/os-release:ro \
  -v /var/run/docker.sock:/var/run/docker.sock:ro \
  mjmalleo/hostsleuth:latest
```

The Web UI remains on `127.0.0.1:8787`, so use the same local browser or SSH-tunnel workflow described above. Because HostSleuth uses host networking, Docker does not publish a separate port mapping.

### Optional direct LAN access

Loopback-only remains the safe default. If you deliberately want HostSleuth reachable directly from a trusted LAN, override the Compose service command and bind it to one specific host LAN address rather than every interface:

```yaml
services:
  hostsleuth:
    command:
      - serve
      - --state-dir
      - /var/lib/hostsleuth
      - --listen
      - 192.168.1.50:8787
      - --interval
      - 60s
```

Replace `192.168.1.50` with the Linux host address you want HostSleuth to use, redeploy the Compose project, and open `http://192.168.1.50:8787` from a routed device that is allowed to reach that address.

HostSleuth does not currently provide Web UI authentication. Direct LAN access can expose hostnames, IP addresses, listeners, container metadata, and other infrastructure details to any device that can reach the bound address. Prefer a specific trusted LAN address over `0.0.0.0`, and keep routing/firewall policy appropriately restricted. SSH tunneling remains the recommended remote-access default.

The Docker deployment intentionally does **not** use `privileged: true`. It drops all Linux capabilities, uses a read-only container filesystem, and shares only the host namespaces/mounts needed for the evidence it can collect honestly.

Docker mode can observe host networking/listeners and Docker container metadata, but container isolation prevents safe, reliable access to everything the native service can see. In Docker mode:

- systemd service/journal evidence is reported as unavailable;
- M7 Service Story is therefore unavailable as a native systemd story rather than being simulated from incomplete Docker evidence;
- host filesystem inventory is reported as unavailable;
- host `apt` / `dpkg` package-history evidence is unavailable in the default Docker deployment because host package logs are not mounted;
- host configuration-fingerprint evidence is unavailable in the default Docker deployment because those host configuration files are not mounted;
- native Certbot lineage/renewal evidence is unavailable in the default Docker deployment because host `/etc/letsencrypt` and systemd state are not mounted;
- firewall evidence may be unavailable without elevated network-administration privileges;
- remote/served TLS certificate evidence remains available because it comes from the diagnosed endpoint itself;
- Incident Lens and Reboot Story can still show whichever retained event categories the Docker deployment actually records, without pretending unavailable native evidence exists;
- M11 native systemd Safe Actions are unavailable in Docker mode;
- native installation remains the recommended choice when full host visibility or optional Safe Actions matter.

The Docker deployment mounts `/var/run/docker.sock` so HostSleuth can inventory Docker containers. Access to the Docker daemon socket is inherently powerful even when its bind path is mounted read-only. HostSleuth uses it only for read-only inventory commands, but only run this deployment on a host where you trust the HostSleuth container and image source.

### Build the container from source

Developers can still build the image locally rather than consuming the public image:

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
docker build -t hostsleuth:dev .
HOSTSLEUTH_IMAGE=hostsleuth:dev docker compose up -d
```

## What it does

### Remember meaningful changes

HostSleuth periodically captures useful local state including:

- host identity, OS, kernel, and memory;
- Linux kernel boot ID and exact boot start when available;
- filesystems and capacity in native mode;
- interfaces and routes;
- listening sockets;
- systemd services in native mode;
- Docker containers, published ports, state, and network names when Docker is readable;
- bounded Debian/Ubuntu package install, update, and removal history from local `dpkg` logs, with `apt` history as a fallback, in native mode;
- SHA-256 fingerprints for a small explicit set of high-value configuration files in native mode, storing path/state/fingerprint/size rather than file contents.

It compares snapshots and records meaningful service/container/listener/package/configuration changes in one local event timeline. A reboot event is recorded only when a previously known kernel boot ID changes to another known boot ID. Known noisy changes such as Docker uptime progression are suppressed. Package history, configuration fingerprints, and boot identity are baselined across schema upgrades so pre-existing or newly introduced evidence is not falsely replayed as a change.

Configuration fingerprinting currently covers `/etc/hosts`, `/etc/fstab`, `/etc/ssh/sshd_config`, `/etc/docker/daemon.json`, and `/etc/nftables.conf`. Missing or unreadable files remain truthful and quiet, and HostSleuth does not recursively crawl `/etc`.

### Diagnose `host:port`

Enter a target such as `192.168.1.20:443` or `example.com:443`. HostSleuth can combine:

- DNS resolution;
- kernel route evidence;
- TCP connectivity;
- target-aware local TCP listener evidence;
- Docker port publication, bind address, and network context;
- bounded nftables candidate evidence when available;
- bounded failed-systemd candidate evidence in native mode;
- TLS handshake result, negotiated protocol, and cipher suite when the endpoint speaks TLS;
- served certificate subject, SANs, issuer, serial, validity window, remaining lifetime, and SHA-256 fingerprint;
- certificate hostname match/mismatch and bounded trust-chain verification;
- on native Linux for local TLS endpoints, bounded read-only Certbot lineage, renewal, timer/service, and local-vs-served certificate evidence when available.

The Web UI presents the answer first and keeps the underlying evidence available for inspection. Successful TCP remains definitive transport evidence; TLS validation is reported separately so a reachable endpoint can still be explained as having a handshake, expiry, hostname, trust, or stale-served-certificate problem. Stronger local bind/listener evidence outranks weaker candidates, and unavailable optional evidence stays `unknown` rather than becoming a false failure.

HostSleuth only calls a local certificate "newer/different than the one this endpoint is serving" when a unique readable Certbot lineage matches the requested host and deterministic validity/fingerprint evidence supports that statement. A fingerprint difference alone is not treated as proof of staleness.

### Expected Endpoint Contracts — stable v0.6.0

M12 adds an on-demand Expected-vs-Observed check for an endpoint. A contract always requires TCP reachability and can optionally require an exact DNS address set, TLS behavior, an active native systemd service, and/or a running container.

Example:

```bash
hostsleuth contract \
  --expect-ip 192.0.2.10 \
  --tls verified \
  --container web \
  example.com:443
```

The result reports each expectation as `pass`, `fail`, or `unknown`, identifies the first proven mismatch, and includes the existing full Diagnosis evidence for deeper inspection. Exact DNS expectations are order-independent exact sets; omit them for endpoints whose addresses are intentionally dynamic.

The Web UI exposes the same workflow under **Expectations** and can hand the target to the existing full Diagnose view.

M12 is read-only. It does not add polling, alerts, persistent contract storage, automatic discovery of expected state, new privileges, or remediation.

Full design: `docs/design/M12-EXPECTED-ENDPOINT-CONTRACTS.md`.

### DNS Detective — stable v0.7.0

M13 adds an on-demand resolver comparison story for one DNS name or IP. HostSleuth always shows the system resolver view and can compare up to four resolver IPs that you explicitly supply.

Example:

```bash
hostsleuth dns \
  --resolver local-a=192.0.2.53 \
  --resolver public=1.1.1.1 \
  app.example.com
```

For names, DNS Detective compares normalized A/AAAA answers plus canonical CNAME evidence. For IP inputs, it compares PTR answers. Each resolver view includes lookup duration, bounded error evidence, and address scope (loopback/private/link-local/global/other).

Results are deterministic:

- `agree` — compared resolver views match;
- `diverge` — successful resolver views disagree;
- `partial` — one or more resolver lookups failed without a proven disagreement;
- `single` — only the system resolver view is available.

When resolver views differ between private/local and global addresses, HostSleuth may say the pattern is **consistent with split-view DNS or resolver-specific overrides**. It does not claim split-horizon configuration exists without direct configuration evidence.

Custom Web/API resolvers are IP addresses on DNS port 53 only. HostSleuth never silently sends a hostname to a public resolver; a custom resolver is queried only because the operator explicitly supplied it.

M13 also introduces Admin Console v1: desktop navigation rail, global Quick Target bar, browser-local recent targets and density preference, keyboard `/` focus, and responsive fallback navigation.

Full design: `docs/design/M13-DNS-DETECTIVE-ADMIN-CONSOLE.md`.

### Reverse Proxy / Upstream Story — stable v0.8.0

M14 adds a bounded read-only request-path story for the admin question: **the public endpoint answers, so where does the proxy/upstream path actually break?**

Give HostSleuth the public URL and the upstream you expect behind it:

```bash
hostsleuth proxy \
  --upstream http://192.0.2.40:8080/ \
  https://app.example.com/
```

HostSleuth evaluates the public and upstream paths separately across DNS, route/TCP, TLS when applicable, and HTTP response metadata. When the public hostname differs from the upstream hostname, it also probes the same explicit upstream address using the public HTTP Host and TLS SNI. That comparison can expose virtual-host/SNI-sensitive behavior without reading proxy configuration.

M14's HTTP probes are deliberately narrow:

- HEAD only;
- no request or response bodies;
- no credentials, cookies, Authorization, or arbitrary custom headers;
- no environment proxy;
- bounded same-host redirects on the public URL;
- cross-host redirects are recorded but not followed;
- upstream redirects are recorded but never followed;
- query strings are redacted from returned URL/Location evidence.

HTTP 4xx is reported as a warning because the path answered but the application/policy rejected the request; HTTP 5xx is a failure. A 405 explicitly explains the HEAD-only boundary rather than silently retrying with GET.

The Web UI adds a visual **Proxy Path** view with first-problem highlighting, expandable stage evidence, native-vs-public Host/SNI comparison, bounded evidence copy, and Admin Console v2 Quick Target routing.

Full design: `docs/design/M14-REVERSE-PROXY-UPSTREAM-STORY.md`.

### Deployment / Permissions Story — stable v0.9.0

M15 adds a bounded read-only permission story for the admin question: **the process is running; why can it not use this path or Unix-domain socket?**

```bash
hostsleuth permissions --service nginx.service /srv/app/data
```

For one native systemd service and one explicit absolute path, HostSleuth records the observed running process identity (PID, effective UID/GID, supplementary groups, working directory, executable, and effective capability mask), resolves the exact path chain, and evaluates the owner/group/other mode class at each parent and the target. Parent directories require traversal permission; the target receives deterministic read/write/execute/traverse decisions, while a Unix-domain socket receives a connect decision using the socket write bit.

The story is deliberately bounded: no file contents, no sibling enumeration, no recursive crawl, no chmod/chown/setfacl, no service/container mutation, and no automatic remediation. POSIX ACLs, Linux Security Modules, namespaces, mount flags, and application policy remain explicit reasoning limits rather than being silently inferred.

When native Docker inventory and the Docker CLI are already available, M15 may add only bind-mount metadata relevant to the explicit requested path. Docker deployment mode does not cross the container boundary to manufacture native systemd identity; it reports that evidence as unavailable.

The Web UI adds a dedicated **Permissions** view with process-identity cards, a visual permission chain, searchable evidence rows, and relevant bind-mount context.

Full design: `docs/design/M15-DEPLOYMENT-PERMISSIONS-STORY.md`.

Closeout: `docs/history/M15-DEPLOYMENT-PERMISSIONS-STORY.md`.

### STARTTLS / Mail Service Story — stable v0.9.0

M16 adds a bounded read-only pre-authentication mail TLS story for the question: **the mail port is open; did STARTTLS actually negotiate correctly?**

```bash
hostsleuth starttls --protocol smtp mail.example.com:25
hostsleuth starttls --protocol imap mail.example.com:143
hostsleuth starttls --protocol pop3 mail.example.com:110
```

HostSleuth supports SMTP STARTTLS, IMAP STARTTLS, and POP3 STLS using one deterministic stage model: TCP, greeting, capability advertisement, protocol upgrade, TLS negotiation, certificate validity, hostname, and trust. Protocol reads are bounded and the TLS/certificate interpretation is shared with the existing TLS engine.

The protocol surface is deliberately fixed and pre-authentication only. HostSleuth does not accept or send usernames, passwords, OAuth tokens, arbitrary mail commands, mail submission commands, mailbox selection/retrieval commands, message bodies, or attachments. It does not edit mail configuration, restart mail services, renew/install certificates, or read private keys.

The Web UI adds a dedicated **STARTTLS** view with protocol selection, visual upgrade/TLS stages, pre-upgrade capability evidence, TLS version/cipher, and served-certificate metadata.

Full design: `docs/design/M16-STARTTLS-MAIL-SERVICE-STORY.md`.

Closeout: `docs/history/M16-STARTTLS-MAIL-SERVICE-STORY.md`.

### Certificate Rollout Verification — stable v0.9.0

M17 adds a bounded read-only certificate rollout story for the question: **I renewed or replaced a certificate; which explicit endpoint is still serving a different one?**

```bash
hostsleuth cert-rollout \
  --fingerprint SHA256 \
  --endpoint edge-a.example.com:443 \
  --endpoint edge-b.example.com:443
```

You can also use one explicit direct-TLS reference endpoint as the expected source instead of typing a fingerprint. HostSleuth compares up to 16 explicit direct-TLS endpoints and reports rollout identity as `match`, `mismatch`, or `unknown`, while keeping certificate validity, hostname, and trust health separate.

Arbitrary certificate-file paths are intentionally not exposed: HostSleuth cannot prove an arbitrary path is not a private key before opening it. M17 therefore preserves the no-private-key-read boundary while still supporting deterministic rollout verification through fingerprint/reference sources.

The Web UI adds a dedicated **Cert Rollout** matrix with expected-certificate context, rollout summary, subject, expiry, hostname, trust, and served SHA-256 fingerprint evidence.

Full design: `docs/design/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md`.

Closeout: `docs/history/M17-CERTIFICATE-ROLLOUT-VERIFICATION.md`.

### Service Story — native Linux

M7 adds a read-only service story inside the existing Diagnose workflow. Select a systemd unit and optionally the `host:port` you expect it to provide. HostSleuth can correlate:

- bounded systemd runtime/result evidence;
- bounded, sanitized current-boot journal evidence;
- the service main PID and cgroup process membership;
- listener ownership when local permissions expose listener PIDs;
- deterministic expected-port collisions only when competing PID evidence is actually visible;
- expected-port presence when ownership is hidden;
- related container host-port publication context;
- the existing endpoint Diagnose/TLS/certificate story;
- retained service/listener changes plus bounded nearby package/configuration/container context.

Nearby retained changes are context, not proof of causation. If local permissions hide a journal or listener owner, HostSleuth reports `unknown` rather than inventing an answer.

Service Story itself does not start, stop, restart, reload, enable, disable, or otherwise modify a service.

### Incident Lens

M8 adds a read-only Incident Lens inside the existing Changes workflow. Anchor the lens on a recorded change or an exact time and HostSleuth shows all retained changes in a bounded +/- 15 minute window.

Incident Lens:

- preserves existing event categories instead of creating a second history store;
- orders events deterministically and caps the result at 100 events;
- labels temporal proximity as context, never proof of causation;
- optionally runs the existing Diagnose/TLS engine for a supplied endpoint;
- labels that endpoint evidence with its current capture time so it is not confused with historical state that HostSleuth never recorded.

Incident Lens does not create a time-series database, reconstruct historical packets/TLS sessions, alert on uptime, infer causes, or modify the host.

### Reboot Story

M10 adds a read-only Reboot Story for answering what happened around the current boot and which observed services/listeners/containers failed to recover.

Reboot Story:

- uses kernel boot identity rather than human-readable uptime to detect a new boot;
- anchors a bounded +/- 15 minute retained-event window on the exact boot start when available;
- reads bounded previous/current boot journal evidence where the running account has access;
- reports inaccessible or unavailable journal history as `unknown` instead of inventing evidence;
- classifies orderly or abnormal shutdown only when direct bounded evidence supports that distinction;
- shows current failed systemd services in native mode;
- calls a service/listener/container a recovery issue only when retained post-boot evidence and the current snapshot agree that the problem remains;
- shows package/kernel/system/configuration changes near boot as context only;
- explicitly does not claim reboot cause from temporal proximity.

Reboot Story does not reboot, shut down, restart, reload, repair, or otherwise modify the host.

### Optional Safe Actions — native Linux

M11 introduced one deliberately narrow state-changing action, `service.restart`. M18 adds exactly one more under a separate security review: `service.reload`.

Safe Actions are **disabled by default**. Enabling them requires explicit global opt-in plus a separate per-action service allowlist. Restart permission does not imply reload permission, and reload permission does not imply restart permission. There is no arbitrary command, script, shell, generic service-control field, reload-or-restart fallback, or automatic remediation path.

Restart preview:

```bash
hostsleuth action preview \
  --enable-actions \
  --allow-restart-service nginx \
  --id service.restart \
  --target nginx.service
```

Reload preview:

```bash
hostsleuth action preview \
  --enable-actions \
  --allow-reload-service nginx \
  --id service.reload \
  --target nginx.service
```

A reload is eligible only when the explicit target is loaded, currently active, independently reload-allowlisted, and systemd reports `CanReload=yes`. Its exact confirmation is `RELOAD UNIT.service`, and HostSleuth executes only the trusted absolute `systemctl reload UNIT.service` argv. If reload is unsupported or fails, HostSleuth reports failure and never falls back to restart.

Both fixed actions require an exact preview confirmation, durable audit evidence before execution, bounded command time/output, serialized execution, and bounded before/after systemd evidence. Success requires an observed `ActiveState=active` postcondition. For `service.reload`, success means only that the reload command succeeded and the service remained active; it does not claim application-specific configuration semantics were verified.

The Action Web/API surface remains loopback-only. State-changing Web requests require JSON and `X-HostSleuth-Action: confirm`. Docker mode reports both native systemd actions unavailable, so the supported Docker deployment gains no host service-control capability.

M18 security contract: `docs/design/M18-SAFE-ACTIONS-II-SECURITY-REVIEW.md`.

### Redacted Evidence Bundle — stable v0.5.0

The Redacted Evidence Bundle creates a bounded local support package from selected HostSleuth evidence. The first version exports only the current snapshot, bounded recent events, and bounded Safe Action audit records through typed redactors. It pseudonymizes host/infrastructure identifiers, scrubs credential/private-key patterns, excludes raw journal/command/config/file content, and includes a manifest plus SHA-256 checksums.

Preview and export are separate operations: preview writes no archive, while export creates an owner-only local ZIP and refuses to overwrite an existing destination. Redaction reduces disclosure risk but cannot guarantee anonymity; review the preview and bundle before sharing.

Full security rules are documented in `docs/design/REDACTED-EVIDENCE-BUNDLE.md` and `SECURITY.md`.

![HostSleuth Diagnose view](docs/images/hostsleuth-diagnose.png)

_Real public-safe Diagnose view captured from the supported Docker Compose deployment during M3 acceptance._

## CLI basics

Capture a snapshot:

```bash
hostsleuth snapshot
```

Diagnose a target:

```bash
hostsleuth diagnose example.com:443
```

Check an expected endpoint contract:

```bash
hostsleuth contract --tls verified --expect-ip 192.0.2.10 example.com:443
```

Compare DNS resolver views:

```bash
hostsleuth dns --resolver local=192.0.2.53 --resolver public=1.1.1.1 example.com
```

Trace a public endpoint to its expected upstream:

```bash
hostsleuth proxy \
  --upstream http://192.0.2.40:8080/ \
  https://app.example.com/
```

Build a native systemd service story, optionally with its expected endpoint:

```bash
hostsleuth service -target 127.0.0.1:443 nginx.service
```

Inspect a bounded incident window, optionally with a fresh endpoint check:

```bash
hostsleuth incident --at 2026-09-16T20:00:00Z --target example.com:443
```

Build the current Reboot Story:

```bash
hostsleuth reboot
```

Inspect Safe Action capabilities:

```bash
hostsleuth action list
```

Preview a redacted evidence bundle without writing an archive:

```bash
hostsleuth evidence preview
```

Export the bounded redacted bundle to a local ZIP:

```bash
hostsleuth evidence export --output hostsleuth-evidence.zip
```

Show recent events:

```bash
hostsleuth events
```

Show the installed version:

```bash
hostsleuth version
```

## Install a specific release

```bash
sudo HOSTSLEUTH_VERSION=v0.9.0 ./scripts/install.sh
```

The installer fetches both the selected native binary and that release's `SHA256SUMS`. Installation stops before replacing `/usr/local/bin/hostsleuth` if the expected architecture checksum is missing or does not verify.

## Build from source

Requirements: Linux and Go 1.24+.

```bash
git clone https://github.com/xXDasGoGXx/HostSleuth.git
cd HostSleuth
go test ./...
go build -o hostsleuth ./cmd/hostsleuth
./hostsleuth serve
```

Developers who intentionally want the installer to build on the target host can use:

```bash
sudo ./scripts/install-source.sh
```

## Uninstall

Remove the native binary and service while preserving recorded HostSleuth state:

```bash
sudo ./scripts/uninstall.sh
```

For Docker Compose:

```bash
docker compose down
```

For direct Docker:

```bash
docker rm -f hostsleuth
```

The named `hostsleuth-data` volume is retained unless you explicitly remove it.

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

M0 repository foundation, M1 deployable single-host MVP, M2 deeper deterministic diagnosis, M3 Product Experience, M3.4 Public Container Distribution, M4 package-change timeline, M5 configuration fingerprinting, M6 Certificate Story / TLS Detective, M7 Service Story, M8 Incident Lens, M9 HostSleuth Workbench, M10 Reboot Story, M11 Optional Safe Actions, and the Redacted Evidence Bundle are complete.

Stable `v0.9.0` is the current published native/Docker release, published from exact source `b518ed901e2d3f4e95a9bb74ade37d7b3a156540` and independently verified.

Docker `mjmalleo/hostsleuth:0.9.0` and `latest` resolve to OCI index `sha256:c99f417419b864756d246232602a8f0fb31067fc11324430a59baec65d62dc5b`.

The live OMV Arcane/Docker deployment and `OMV-Docker-Rebuild` recovery definition are aligned on `mjmalleo/hostsleuth:0.9.0`. Live acceptance confirmed v0.9.0, schema 4 / Docker mode, the running 0.9.0 image, all 100 retained events, real M16 STARTTLS and M17 certificate-rollout execution, M15's Docker boundary, LAN Action API rejection, and both fixed native actions disabled/unavailable in Docker mode.

M12 Expected Endpoint Contracts, M13 DNS Detective + Admin Console v1, and M14 Reverse Proxy / Upstream Story + Admin Console v2 remain published and live from earlier releases.

M15 Deployment / Permissions Story, M16 STARTTLS / Mail Service Story, M17 Certificate Rollout Verification, and M18 Safe Actions II are now included in published stable v0.9.0.

The owner-approved fixed native `service.reload` action remains native-only and independently allowlisted. Docker mode exposes neither native systemd action.

A feature-free **v1.0 Readiness / Hardening** cycle is now active. The accepted first slice adds persisted-state durability regression coverage, Admin Console tab/tabpanel keyboard/accessibility semantics, reduced-motion/focus handling, responsive browser acceptance, and explicit served-asset budgets. The second slice hardens native release installation with SHA-256 verification and adds restrictive browser security headers plus disposable native-install CI acceptance. Stable/public/live/recovery remain v0.9.0 until readiness is accepted and a separate v1.0.0 publication decision is made.

See [`docs/design/V1.0-READINESS.md`](docs/design/V1.0-READINESS.md) for the readiness gates and [`docs/ROADMAP.md`](docs/ROADMAP.md) for product guardrails.

## Security and privacy

HostSleuth can collect hostnames, IP addresses, mount paths, service names, listener addresses, container metadata, package names/versions, configuration paths/fingerprints, certificate metadata/fingerprints, kernel boot identity/start time, bounded service runtime properties, sanitized journal evidence, retained incident-window context, bounded boot/recovery context, and Safe Action audit metadata.

It does not store configuration file contents as part of M5 fingerprinting, M6 does not read private keys, M8 does not infer causal relationships from nearby timestamps, M10 does not infer reboot cause from temporal proximity or broaden privileges to obtain inaccessible journal history, and M11 provides no arbitrary command surface or automatic remediation. The Redacted Evidence Bundle uses a separate fail-closed export policy with deterministic pseudonymization and explicit omissions, but redaction cannot guarantee anonymity. Treat snapshots, event logs, diagnostic output, action audit logs, and exported bundles as potentially sensitive. See [`SECURITY.md`](SECURITY.md) for the current security posture and vulnerability-reporting guidance.

## Project files

- `README.md` — product definition, usage, and current scope.
- `CURRENT-HANDOFF.md` — current project state and next task.
- `TO-DO.md` — active checklist and product decisions.
- `docs/ROADMAP.md` — owner-approved ordered product roadmap.
- `docs/design/V1.0-READINESS.md` — feature-free v1.0 hardening gates and release-readiness contract.
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
- `docs/research/CONSUMER-OPPORTUNITY-LANDSCAPE.md` — sourced research input for differentiated future workflows; not active scope by itself.

## License

HostSleuth is released under the MIT License. See `LICENSE`.