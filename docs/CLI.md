# Command line

Back to the [README](../README.md).


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

Summarize containers (troubled first), or explain one:

```bash
hostsleuth containers
hostsleuth containers qbittorrent
```

Show recent events (add `--all` to include routine listener churn; the Web API equivalent is `/api/events?all=1`):

```bash
hostsleuth events
hostsleuth events --all
```

Show the installed version:

```bash
hostsleuth version
```
