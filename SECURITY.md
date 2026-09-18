# Security Policy

HostSleuth is early-stage software that inventories local system state. Treat its output as potentially sensitive because it can contain hostnames, IP addresses, mount paths, service names, listener addresses, container metadata, file metadata/fingerprints, DNS answers, HTTP endpoint metadata, certificate metadata/fingerprints, and optional-action audit evidence.

## Current security posture

- HostSleuth remains read-only by default. Optional Safe Actions require explicit enablement and an explicit target allowlist.
- The web interface binds to `127.0.0.1:8787` by default.
- Non-loopback exposure is not recommended until authentication is implemented.
- Snapshot, event, and action-audit files are created with owner-only permissions where supported.
- Automatic repair/remediation remains out of scope. M11/M18 actions are explicit operator requests, never automatic responses to a diagnosis.
- The supported Docker Compose deployment does not use unrestricted privileged mode, drops all Linux capabilities, enables `no-new-privileges`, and uses a read-only container filesystem.
- Docker mode deliberately reports systemd and host-filesystem evidence as unavailable instead of mounting broad host control surfaces to recreate native visibility.

## M11 / M18 Optional Safe Actions boundary

Development source exposes exactly two fixed native systemd actions:

- `service.restart` from M11;
- `service.reload` from M18.

The action framework is disabled unless HostSleuth is started or invoked with `--enable-actions`. Restart and reload use independent explicit allowlists: `--allow-restart-service UNIT` and `--allow-reload-service UNIT`. Permission for one action never grants permission for the other. No service is permitted by default. Allowlisted service names are normalized to `.service` and must pass the same conservative unit-name validator.

Both actions require a deterministic preview, exact confirmation, a durable pre-execution audit record, trusted absolute `systemctl` argv without a shell, bounded execution/output, bounded before/after systemd evidence, serialized execution, and an observed `ActiveState=active` postcondition before HostSleuth reports success.

`service.reload` has additional fail-closed requirements: the unit must already be active and trusted systemd evidence must report `CanReload=yes`. The command is exactly `systemctl reload UNIT`; there is no reload-or-restart helper and no fallback to restart. A successful reload result means the reload command succeeded and the unit remained active. It does not prove application-specific configuration semantics were applied.

Unknown action IDs, unlisted targets, unsafe unit names, missing confirmation, unavailable audit storage, unavailable reload-capability evidence, and Docker deployment mode fail closed.

The Action Web/API surface is loopback-only even if the read-only HostSleuth UI is deliberately bound to a LAN address. State-changing requests require JSON plus an explicit `X-HostSleuth-Action: confirm` header. Use the local browser workflow, an SSH tunnel, or the local CLI.

M11/M18 do not add arbitrary command execution, generic systemd control, package/firewall/file administration, writable Docker-socket control, Docker/container mutation, or automatic remediation.

M18 security review: `docs/design/M18-SAFE-ACTIONS-II-SECURITY-REVIEW.md`.

## Workbench boundary

M9 Workbench adds bounded read-only helpers for file identity/checksums, file comparison, DNS inspection, HTTP HEAD/redirect inspection, and public certificate inspection/comparison.

Those Web/API operations are intentionally accepted only from loopback clients. This prevents an unauthenticated LAN-visible HostSleuth instance from becoming a remote file-hash oracle or a server-side HTTP/DNS probe. Use the local browser workflow, an SSH tunnel, or the local `hostsleuth workbench ...` CLI.

Workbench does not return file contents, edit files, accept arbitrary commands, accept HTTP credentials/cookies/custom headers, or expose a private-key viewer. Certificate inspection stops after the first public `CERTIFICATE` PEM block and refuses a private-key block encountered before a certificate.

The generic file identity tool does read the bytes of a user-selected readable regular file in order to calculate SHA-256/SHA-512. It returns metadata and digests only; it does not retain or return file contents. Treat file paths and fingerprints as potentially sensitive evidence.

## M13 DNS Detective boundary

M13 adds bounded read-only resolver comparison. The normal system resolver is always inspected; additional resolver views are queried only when the operator explicitly supplies resolver IPs for that request.

The first Web/API implementation accepts at most four custom resolver IPs and restricts them to DNS port 53. Hostnames are not silently sent to a hard-coded public DNS provider, and HostSleuth does not persist resolver query history server-side.

DNS Detective returns normalized answer metadata, lookup duration, bounded lookup errors, runtime-visible resolver/search-domain metadata, and deterministic agreement/divergence status. It does not perform zone transfers, DNS updates, resolver reconfiguration, arbitrary UDP probing, polling, alerting, or remediation.

Because a deliberately LAN-bound HostSleuth instance exposes read-only Diagnose and DNS Detective endpoints to clients that can already reach that UI, direct LAN exposure should still be treated as trusted-network access. The safer default remains loopback plus SSH tunneling.

## M14 Reverse Proxy / Upstream Story boundary

M14 adds bounded read-only HTTP request-path inspection for one explicitly supplied public URL and one explicitly supplied expected upstream URL.

The probe surface is intentionally narrow:

- HTTP method is HEAD only;
- no request body is sent and no response body is returned;
- URLs containing embedded credentials are refused;
- HostSleuth sends no cookies, Authorization header, or arbitrary operator-supplied request headers;
- environment HTTP proxy settings are not used;
- public same-host redirects may be followed only up to a small fixed limit;
- a redirect to a different hostname is recorded but not followed;
- upstream redirects are recorded but never followed;
- URL and Location query strings are omitted/redacted in returned evidence;
- selected response metadata is bounded to status, Location, Server, Content-Type, and duration.

When public and upstream hostnames differ, M14 may make a second request to the same explicitly supplied upstream dial target using the public hostname as HTTP Host and TLS SNI. That identity is derived only from the explicit public URL; M14 does not expose a generic custom-header field.

M14 does not parse proxy configuration, discover upstreams, scan networks, perform authentication flows, use body-bearing HTTP methods, poll/monitor endpoints, edit DNS/proxy/container/service configuration, reload/restart anything, or remediate automatically.

The M14 Web/API surface is part of the normal read-only troubleshooting UI rather than the loopback-only Workbench. Therefore, a deliberately LAN-bound HostSleuth deployment should be treated as trusted-network access: clients that can reach the UI can ask the HostSleuth host to perform these bounded probes against explicit URLs. The safer default remains loopback-only access or a trusted local tunnel.

## Redacted Evidence Bundle boundary

The Redacted Evidence Bundle is a local, operator-triggered support export. It is not a backup, forensic image, remote collection mechanism, or automatic sharing feature.

The first supported implementation is intentionally typed and bounded. It exports only the current snapshot, at most 200 recent events, and at most 100 Safe Action audit records. It does not provide a generic serializer or arbitrary include paths.

Before any archive is written, `hostsleuth evidence preview` builds the same redacted payload in memory and reports included/omitted evidence classes, record counts, redaction counts, files, bounded size, unavailable evidence, and warnings. Preview does not create an archive. Export is a separate explicit command.

The `redacted-v1` policy pseudonymizes non-loopback host/domain/IP identifiers, MAC addresses, arbitrary paths, service/container/network/interface identities, opaque machine IDs, image repository identity, e-mail addresses, and configuration fingerprints while preserving useful equality relationships inside one bundle. Ports, state/status fields, package names/versions, OS/kernel/architecture data, UTC timestamps, and loopback/unspecified-address semantics remain available where useful.

Exportable free-form evidence is scrubbed for password/token/API-key/cookie/session/authorization forms, URL credentials and query strings, PEM private-key material including truncated private-key blocks, plain domains, IPv4/IPv6 identifiers, service names, MAC/UUID values, and paths. Redaction is performed on copies only and never modifies retained HostSleuth state.

The first bundle explicitly excludes raw journal text, raw Safe Action command output, ActionPreview command/confirmation material, arbitrary file contents, arbitrary configuration contents, and arbitrary Workbench file/URL/DNS inputs. There is no cloud upload or automatic sending.

Bundles are limited to 1 MiB per JSON payload and 2 MiB total uncompressed payload. Temporary and final archive files use owner-only permissions (`0600`) where supported. The exporter refuses to overwrite an existing archive, writes through a temporary file, removes temporary output on failure where practical, emits SHA-256 checksums for bundle payloads, and prints the final archive SHA-256 after success.

Redaction reduces disclosure risk but cannot guarantee anonymity. Timing, package versions, topology shape, or other contextual evidence may still identify an environment to a recipient who already knows it. Operators should review the preview and bundle before sharing.

Full design/threat model: `docs/design/REDACTED-EVIDENCE-BUNDLE.md`.

## Docker socket

The supported Docker deployment mounts `/var/run/docker.sock` so HostSleuth can inventory Docker containers. Access to the Docker daemon socket is a powerful host capability: mounting the socket path read-only does not make Docker API access inherently read-only.

HostSleuth currently uses the socket only through read-only inventory commands such as `docker ps`. M11 does not use it for actions and does not request a writable socket. A compromised process with access to the socket could still potentially control the Docker daemon, so only use the Docker deployment when you trust the HostSleuth image/source and the host's Docker daemon is appropriate to expose to this container.

For the smallest HostSleuth security surface and fullest systemd/filesystem visibility, use the native systemd deployment.

## Reporting a vulnerability

Please open a GitHub security advisory for the repository when possible. Avoid posting secrets, private host data, credentials, or exploit details in a public issue.

## Supported versions

Until release policy is revised, the newest published stable release and current development branch are the supported references.
