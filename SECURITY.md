# Security Policy

HostSleuth is early-stage software that inventories local system state. Treat its output as potentially sensitive because it can contain hostnames, IP addresses, mount paths, service names, listener addresses, container metadata, file metadata/fingerprints, DNS answers, HTTP endpoint metadata, and certificate metadata/fingerprints.

## Current security posture

- HostSleuth is read-only with respect to monitored services.
- The web interface binds to `127.0.0.1:8787` by default.
- Non-loopback exposure is not recommended until authentication is implemented.
- Snapshot and event files are created with owner-only permissions where supported.
- Automatic repair/remediation is intentionally out of scope for the current milestone.
- The supported Docker Compose deployment does not use unrestricted privileged mode, drops all Linux capabilities, enables `no-new-privileges`, and uses a read-only container filesystem.
- Docker mode deliberately reports systemd and host-filesystem evidence as unavailable instead of mounting broad host control surfaces to recreate native visibility.

## Workbench boundary

M9 Workbench adds bounded read-only helpers for file identity/checksums, file comparison, DNS inspection, HTTP HEAD/redirect inspection, and public certificate inspection/comparison.

Those Web/API operations are intentionally accepted only from loopback clients. This prevents an unauthenticated LAN-visible HostSleuth instance from becoming a remote file-hash oracle or a server-side HTTP/DNS probe. Use the local browser workflow, an SSH tunnel, or the local `hostsleuth workbench ...` CLI.

Workbench does not return file contents, edit files, accept arbitrary commands, accept HTTP credentials/cookies/custom headers, or expose a private-key viewer. Certificate inspection stops after the first public `CERTIFICATE` PEM block and refuses a private-key block encountered before a certificate.

The generic file identity tool does read the bytes of a user-selected readable regular file in order to calculate SHA-256/SHA-512. It returns metadata and digests only; it does not retain or return file contents. Treat file paths and fingerprints as potentially sensitive evidence.

## Docker socket

The supported Docker deployment mounts `/var/run/docker.sock` so HostSleuth can inventory Docker containers. Access to the Docker daemon socket is a powerful host capability: mounting the socket path read-only does not make Docker API access inherently read-only.

HostSleuth currently uses the socket only through read-only inventory commands such as `docker ps`, but a compromised process with access to the socket could potentially control the Docker daemon. Only use the Docker deployment when you trust the HostSleuth image/source and the host's Docker daemon is appropriate to expose to this container.

For the smallest HostSleuth security surface and fullest systemd/filesystem visibility, use the native systemd deployment.

## Reporting a vulnerability

Please open a GitHub security advisory for the repository when possible. Avoid posting secrets, private host data, credentials, or exploit details in a public issue.

## Supported versions

Until release policy is revised, the newest published stable release and current development branch are the supported references.
