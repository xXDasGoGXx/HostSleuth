# Security Policy

HostSleuth is early-stage software that inventories local system state. Treat its output as potentially sensitive because it can contain hostnames, IP addresses, mount paths, service names, listener addresses, and container metadata.

## Current security posture

- HostSleuth is read-only with respect to monitored services.
- The web interface binds to `127.0.0.1:8787` by default.
- Non-loopback exposure is not recommended until authentication is implemented.
- Snapshot and event files are created with owner-only permissions where supported.
- Automatic repair/remediation is intentionally out of scope for the current milestone.
- The supported Docker Compose deployment does not use unrestricted privileged mode, drops all Linux capabilities, enables `no-new-privileges`, and uses a read-only container filesystem.
- Docker mode deliberately reports systemd and host-filesystem evidence as unavailable instead of mounting broad host control surfaces to recreate native visibility.

## Docker socket

The supported Docker deployment mounts `/var/run/docker.sock` so HostSleuth can inventory Docker containers. Access to the Docker daemon socket is a powerful host capability: mounting the socket path read-only does not make Docker API access inherently read-only.

HostSleuth currently uses the socket only through read-only inventory commands such as `docker ps`, but a compromised process with access to the socket could potentially control the Docker daemon. Only use the Docker deployment when you trust the HostSleuth image/source and the host's Docker daemon is appropriate to expose to this container.

For the smallest HostSleuth security surface and fullest systemd/filesystem visibility, use the native systemd deployment.

## Reporting a vulnerability

Please open a GitHub security advisory for the repository when possible. Avoid posting secrets, private host data, credentials, or exploit details in a public issue.

## Supported versions

Until the first stable release, only the newest development/release build is supported.
