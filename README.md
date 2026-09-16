# HostSleuth

**A small, local-first Linux flight recorder for two questions: what changed, and why can I not reach this host/service/port?**

HostSleuth records meaningful Linux state changes and performs deterministic, evidence-backed `host:port` diagnosis. It does not require a cloud account, external database, API key, or AI model.

## Quick start — recommended native install

Native Linux is the recommended deployment because it gives HostSleuth the fullest view of the machine.

Requirements: Linux with systemd. Go is **not** required when installing a release.

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

`latest` represents the newest stable release. Stable releases also receive an explicit numeric tag such as `0.2.0`.

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
HOSTSLEUTH_IMAGE=mjmalleo/hostsleuth:0.2.0 docker compose up -d
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
- host filesystem inventory is reported as unavailable;
- host `apt` / `dpkg` package-history evidence is unavailable in the default Docker deployment because host package logs are not mounted;
- host configuration-fingerprint evidence is unavailable in the default Docker deployment because those host configuration files are not mounted;
- firewall evidence may be unavailable without elevated network-administration privileges;
- native installation remains the recommended choice when full host visibility matters.

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
- filesystems and capacity in native mode;
- interfaces and routes;
- listening sockets;
- systemd services in native mode;
- Docker containers, published ports, state, and network names when Docker is readable;
- bounded Debian/Ubuntu package install, update, and removal history from local `dpkg` logs, with `apt` history as a fallback, in native mode;
- SHA-256 fingerprints for a small explicit set of high-value configuration files in native mode, storing path/state/fingerprint/size rather than file contents.

It compares snapshots and records meaningful service/container/listener/package/configuration changes in one local event timeline. Known noisy changes such as Docker uptime progression are suppressed. Package history and configuration fingerprints are baselined across their schema upgrades so existing evidence is not falsely replayed as new changes.

Configuration fingerprinting currently covers `/etc/hosts`, `/etc/fstab`, `/etc/ssh/sshd_config`, `/etc/docker/daemon.json`, and `/etc/nftables.conf`. Missing or unreadable files remain truthful and quiet, and HostSleuth does not recursively crawl `/etc`.

### Diagnose `host:port`

Enter a target such as `192.168.1.20:443` or `example.com:443`. HostSleuth can combine:

- DNS resolution;
- kernel route evidence;
- TCP connectivity;
- target-aware local TCP listener evidence;
- Docker port publication, bind address, and network context;
- bounded nftables candidate evidence when available;
- bounded failed-systemd candidate evidence in native mode.

The Web UI presents the answer first and keeps the underlying evidence available for inspection. Successful TCP is definitive, stronger local bind/listener evidence outranks weaker candidates, and unavailable optional evidence stays `unknown` rather than becoming a false failure.

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
sudo HOSTSLEUTH_VERSION=v0.2.0 ./scripts/install.sh
```

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
- read-only;
- deterministic before explanatory;
- loopback-only by default for the Web UI.

HostSleuth does **not** automatically restart services, modify firewall rules, repair containers, install/remove/update packages, edit configuration, or reconfigure the host.

## Current stage

M0 repository foundation, M1 deployable single-host MVP, M2 deeper deterministic diagnosis, M3 Product Experience, M3.4 Public Container Distribution, M4 package-change timeline, and M5 configuration fingerprinting are complete on `main`.

Stable `v0.2.0` remains the current published native/Docker release and predates M5. M5 is present in source on `main`; no newer public release is implied by this documentation.

The next active milestone is **M6 — Certificate Story / TLS Detective**. It adds bounded, read-only TLS/certificate diagnosis, native Certbot discovery, certificate fingerprints, and local-certificate-vs-served-certificate comparison. After M6 acceptance, the approved plan is to publish **v0.3.0** containing M5 + M6, then continue in order through Service Story, Incident Lens, HostSleuth Workbench, Reboot Story, Optional Safe Actions, and eventually a redacted evidence bundle.

See [`docs/ROADMAP.md`](docs/ROADMAP.md) for the exact approved order and guardrails.

## Security and privacy

HostSleuth can collect hostnames, IP addresses, mount paths, service names, listener addresses, container metadata, package names/versions, configuration paths, and configuration fingerprints. It does not store configuration file contents as part of M5 fingerprinting. Treat snapshots, event logs, and diagnostic output as potentially sensitive. See [`SECURITY.md`](SECURITY.md) for the current security posture and vulnerability-reporting guidance.

## Project files

- `README.md` — product definition, usage, and current scope.
- `CURRENT-HANDOFF.md` — current project state and next task.
- `TO-DO.md` — active checklist and product decisions.
- `docs/ROADMAP.md` — owner-approved ordered product roadmap.
- `docs/history/DEVELOPMENT-HISTORY.md` — milestone and validation history.

## License

HostSleuth is released under the MIT License. See `LICENSE`.
