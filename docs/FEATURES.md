# What HostSleuth does

Back to the [README](../README.md). Every story is read-only unless it says otherwise.


### Replay (home screen)

Question answered:

> What happened to this box this week—and what does it look like right now?

Replay draws the containers as an engineering sheet: one box per container, grouped by compose project (the `com.docker.compose.project` label), each with a connector showing how it is reachable—bound to every address, one address, this host only, the host's network, or no network. Addresses are never sent to the browser, only the class and the port numbers.

A tape below covers the last seven days of recorded changes. Drag it, step with **Previous/Next change**, or press **Play the week**, and the drawing redraws for that moment:

- down or restarting containers are dashed with a cut connector;
- containers failing their health check are hatched;
- anything that changed in the 24 hours before the moment gets a revision cloud and a numbered Δ, listed in the title block;
- the top line says what was happening in one plain sentence, built by fixed rules from the recorded events, never guessed.

The rules recognise a host reboot (anchored at the boot time, with the kernel change when one was recorded), an all-container restart that fits a Docker restart (at least four containers and 60% of the health-checked ones entering *starting* in the same minute, away from a reboot), crashes with exit codes, clean stops, health failures and recoveries, containers that took over 30 minutes to become healthy, image updates and recreations, and package and configuration changes. When an autoheal container runs on the host, a quick restart after a health failure says so.

When the latest snapshot has an unhealthy container, the line under the headline says why (see *Why it is unhealthy* below), and so does its panel.

Tap a container to see its week; **Open its Container Story** goes to the full Container Story. Tap a connector to highlight everything bound the same way. While on the latest moment, Replay refreshes every minute.

API: `GET /api/replay`. The drawing uses the B612 typeface (SIL Open Font License, `/assets/fonts/OFL.txt`), served by HostSleuth itself; nothing loads from outside.

![Replay](images/replay.png)

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

It compares snapshots and records meaningful service/container/listener/package/configuration changes in one local event timeline. A reboot event is recorded only when a previously known kernel boot ID changes to another known boot ID. Known noisy changes such as Docker uptime progression are suppressed. For containers it also records image updates ("container updated: jellyfin: jellyfin/jellyfin:latest, version 10.10.2 → 10.10.3", including a new build behind the same tag), recreations, and restarts that happen between snapshots with Docker's own reason ("killed for running out of memory", "crashed, exit code 1", "stopped by SIGTERM, exit code 143"). Listener churn that follows interfaces or client activity rather than services is kept as current evidence but not recorded as a change: link-local and multicast sockets, discovery-protocol binds on one specific interface address (mDNS, LLMNR, SSDP, WS-Discovery), and UDP sockets in the Linux ephemeral port range (except fixed services such as WireGuard 51820 and Tailscale 41641). On a Docker host this is the difference between a timeline of real events and hundreds of per-veth socket lines after every reboot; churn recorded by older versions is skipped when the timeline is read. TCP listeners are always recorded. When listeners go away and come back within 15 minutes, as they do across a reboot or a Docker restart, the timeline shows one line per snapshot ("34 listeners went away after the host reboot and were back within 61s: …") instead of a disappeared/appeared pair for each; a listener that does not come back keeps its own warning. After a recorded reboot, a listener on a dynamically assigned port (the ephemeral range, or the 512–1023 range RPC helpers bind) that is replaced by a new port on the same protocol and address is shown as one "replaced by new ports" line, so services such as NFS mountd/statd that pick a new port at every boot are no longer reported by Reboot Story as never recovered; fixed service ports are never paired this way. This is a view over the recorded events, so Service Story still sees each one. The timeline file is compacted automatically past 4 MiB, keeping the newest 5,000 meaningful events. Package history, configuration fingerprints, and boot identity are baselined across schema upgrades so pre-existing or newly introduced evidence is not falsely replayed as a change.

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

### Expected Endpoint Contracts

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

### DNS Detective

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

### Reverse Proxy / Upstream Story

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

### Deployment / Permissions Story

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

### STARTTLS / Mail Service Story

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

### Certificate Rollout Verification

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

### Service Story

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

### Container Story

Question answered:

> Why does this container keep restarting—and what changed?

The Containers view lists every container with the troubled ones first (unhealthy, restarting, crash-looping, stopped), and for each one shows:

- restarts in the last day and week, how many were crashes, and how many were out-of-memory kills;
- every restart with Docker's reason, from Docker's own event log (`die` and `oom` events, exit codes), so a container that crashes and comes back between two snapshots is still caught;
- health changes and the current failing-check streak;
- image updates and recreations;
- listener changes on the host ports the container publishes, within two minutes of its restarts, updates or failures. For host-network containers, where Docker reports no published ports, nearby listener changes are shown as context only.

Evidence comes from a field-limited `docker inspect` (image ID, version label, restart count, start/finish time, exit code, OOM flag, health streak) and `docker events` limited to `die`/`oom` with name and exit code. HostSleuth never reads container environment variables, commands, mounts or host configuration, which commonly hold secrets.

CLI: `hostsleuth containers` (all) or `hostsleuth containers NAME`. API: `GET /api/containers`, `GET /api/container-story?name=NAME`.

![Container Story](images/container-story.png)

### Why it is unhealthy

Question answered:

> Docker says *unhealthy*. Why?

Docker runs each container's own health check and keeps its last few results, but only shows the word *unhealthy*. HostSleuth reads the check's command and its newest result (exit code and what it printed), masks anything that looks like a password or token, and explains the failure with fixed rules:

- **What the check printed:** nothing listening (connection refused), a timeout, a name (DNS) failure, an HTTP error code, or the check's own tool missing after an image update.
- **What the check requires**, when it only tests files and prints nothing: "/staging/READY exists", "UNCLASSIFIED.txt is empty or missing". If a condition is a file a program writes to report trouble (names with *failed*, *error*, *unclassified*, *lock* and similar), HostSleuth says that is the most likely cause.
- **Waiting to catch up:** a check that compares two files (for example "the newest backup has been uploaded") is reported as catching up, with how long it took to clear before.
- **What changed just before:** the container itself, another container in the same compose project, or the host (reboot, Docker restart, packages, configuration), within ten minutes.
- **Whether it has done this before** this week, and how long it took to clear.

Every suggestion is worded as *most likely* or *probably*; HostSleuth never runs anything inside the container to confirm it. It appears in Replay, in Container Story (and its Copy/Share summary), and in `hostsleuth containers NAME` and `GET /api/container-story` as `why`.

### Share redacted

Every result view (Diagnose, Containers, Reboot, Incident Lens, Expectations, DNS, Proxy Path, Permissions, STARTTLS, Cert Rollout) has **Share redacted** next to **Copy evidence summary**. It sends the summary to HostSleuth's own `POST /api/share`, which applies the Evidence Bundle redaction seeded with this host's names and addresses, and shows the result to copy or download as Markdown:

```text
HostSleuth Container Story: container-09
Verdict: WARNING — container-09 is running now. It restarted 5 times in 7 days (4 crashes, 3 out-of-memory kills).
Image: host-02.invalid/linuxserver/container-09:latest (version 5.0.4)
```

Host names, addresses, container and service names, paths, URLs, e-mail addresses and secrets become consistent aliases; ports, statuses, exit codes and versions stay. Nothing is stored or sent anywhere else. Redaction lowers risk; read the text once before posting.

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

### Optional Safe Actions

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

### Redacted Evidence Bundle

The Redacted Evidence Bundle creates a bounded local support package from selected HostSleuth evidence. The first version exports only the current snapshot, bounded recent events, and bounded Safe Action audit records through typed redactors. It pseudonymizes host/infrastructure identifiers, scrubs credential/private-key patterns, excludes raw journal/command/config/file content, and includes a manifest plus SHA-256 checksums.

Preview and export are separate operations: preview writes no archive, while export creates an owner-only local ZIP and refuses to overwrite an existing destination. Redaction reduces disclosure risk but cannot guarantee anonymity; review the preview and bundle before sharing.

Full security rules are documented in `docs/design/REDACTED-EVIDENCE-BUNDLE.md` and `SECURITY.md`.

![HostSleuth Diagnose view](images/diagnose.png)

_Real public-safe Diagnose view captured from the supported Docker Compose deployment during M3 acceptance._
