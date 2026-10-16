# Installing HostSleuth

Back to the [README](../README.md).

## Quick start — recommended native install

Native Linux is the recommended deployment because it gives HostSleuth the fullest view of the machine.

Requirements: Linux with systemd. Go is **not** required when installing a release. By default, the installer selects the latest published GitHub release, downloads its `SHA256SUMS`, and verifies the architecture-specific binary before installing it. Set `HOSTSLEUTH_VERSION` only when you intentionally want to pin a specific release.

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

The Docker Hub storefront copy is version-controlled under `docker/`. The public short description is aligned with the Git-tracked source. A manual-only metadata-sync workflow is retained as optional future automation; it is not required for image publication or normal operation.

`latest` represents the newest stable release. Stable releases also receive an explicit numeric tag such as `1.0.1`.

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
HOSTSLEUTH_IMAGE=mjmalleo/hostsleuth:1.0.1 docker compose up -d
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
- M11/M18 native systemd Safe Actions are unavailable in Docker mode;
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

## Install a specific release

```bash
sudo HOSTSLEUTH_VERSION=v1.0.1 ./scripts/install.sh
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
