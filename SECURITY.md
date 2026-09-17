# Security Policy

HostSleuth is early-stage software that inventories local system state. Treat its output as potentially sensitive because it can contain hostnames, IP addresses, mount paths, service names, listener addresses, container metadata, file metadata/fingerprints, DNS answers, HTTP endpoint metadata, certificate metadata/fingerprints, and optional-action audit evidence.

## Current security posture

- HostSleuth remains read-only by default. Optional Safe Actions require explicit enablement and an explicit target allowlist.
- The web interface binds to `127.0.0.1:8787` by default.
- Non-loopback exposure is not recommended until authentication is implemented.
- Snapshot, event, and action-audit files are created with owner-only permissions where supported.
- Automatic repair/remediation remains out of scope. M11 actions are explicit operator requests, never automatic responses to a diagnosis.
- The supported Docker Compose deployment does not use unrestricted privileged mode, drops all Linux capabilities, enables `no-new-privileges`, and uses a read-only container filesystem.
- Docker mode deliberately reports systemd and host-filesystem evidence as unavailable instead of mounting broad host control surfaces to recreate native visibility.

## M11 Optional Safe Actions boundary

M11 introduces one deliberately narrow state-changing action in development source: `service.restart`.

The action framework is disabled unless HostSleuth is started or invoked with `--enable-actions`. A systemd service must also be explicitly named with `--allow-restart-service UNIT`; no service is permitted by default. Allowlisted service names are normalized to `.service` and must pass a conservative unit-name validator.

A restart requires a preview, an exact confirmation value from that preview, a durable pre-execution audit record, direct argv execution of `systemctl restart UNIT` without a shell, bounded before/after systemd evidence, and an observed `ActiveState=active` postcondition before HostSleuth reports success. Unknown action IDs, unlisted targets, unsafe unit names, missing confirmation, unavailable audit storage, and Docker deployment mode fail closed.

The Action Web/API surface is loopback-only even if the read-only HostSleuth UI is deliberately bound to a LAN address. State-changing requests require JSON plus an explicit `X-HostSleuth-Action: confirm` header. Use the local browser workflow, an SSH tunnel, or the local CLI.

M11 does not add arbitrary command execution, generic systemd control, package/firewall/file administration, writable Docker-socket control, or automatic remediation.

## Workbench boundary

M9 Workbench adds bounded read-only helpers for file identity/checksums, file comparison, DNS inspection, HTTP HEAD/redirect inspection, and public certificate inspection/comparison.

Those Web/API operations are intentionally accepted only from loopback clients. This prevents an unauthenticated LAN-visible HostSleuth instance from becoming a remote file-hash oracle or a server-side HTTP/DNS probe. Use the local browser workflow, an SSH tunnel, or the local `hostsleuth workbench ...` CLI.

Workbench does not return file contents, edit files, accept arbitrary commands, accept HTTP credentials/cookies/custom headers, or expose a private-key viewer. Certificate inspection stops after the first public `CERTIFICATE` PEM block and refuses a private-key block encountered before a certificate.

The generic file identity tool does read the bytes of a user-selected readable regular file in order to calculate SHA-256/SHA-512. It returns metadata and digests only; it does not retain or return file contents. Treat file paths and fingerprints as potentially sensitive evidence.

## Docker socket

The supported Docker deployment mounts `/var/run/docker.sock` so HostSleuth can inventory Docker containers. Access to the Docker daemon socket is a powerful host capability: mounting the socket path read-only does not make Docker API access inherently read-only.

HostSleuth currently uses the socket only through read-only inventory commands such as `docker ps`. M11 does not use it for actions and does not request a writable socket. A compromised process with access to the socket could still potentially control the Docker daemon, so only use the Docker deployment when you trust the HostSleuth image/source and the host's Docker daemon is appropriate to expose to this container.

For the smallest HostSleuth security surface and fullest systemd/filesystem visibility, use the native systemd deployment.

## Reporting a vulnerability

Please open a GitHub security advisory for the repository when possible. Avoid posting secrets, private host data, credentials, or exploit details in a public issue.

## Supported versions

Until release policy is revised, the newest published stable release and current development branch are the supported references.
