# Security Policy

HostSleuth is early-stage software that inventories local system state. Treat its output as potentially sensitive because it can contain hostnames, IP addresses, mount paths, service names, listener addresses, and container metadata.

## Current security posture

- HostSleuth is read-only with respect to monitored services.
- The web interface binds to `127.0.0.1:8787` by default.
- Non-loopback exposure is not recommended until authentication is implemented.
- Snapshot and event files are created with owner-only permissions where supported.
- Automatic repair/remediation is intentionally out of scope for the current milestone.

## Reporting a vulnerability

Please open a GitHub security advisory for the repository when possible. Avoid posting secrets, private host data, credentials, or exploit details in a public issue.

## Supported versions

Until the first stable release, only the newest development/release build is supported.
