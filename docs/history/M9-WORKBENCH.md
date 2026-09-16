# M9 — HostSleuth Workbench

Date: 2026-09-16

## Goal

Reduce several common troubleshooting workflows that normally require separate shell commands or one-off websites while keeping HostSleuth small, deterministic, read-only, and explicitly not a browser shell or generic utilities portal.

The owner also clarified during this milestone that certificate/Certbot ideas are only one example among many possible future enhancements. M9 therefore deliberately spans multiple troubleshooting domains rather than making certificate management the product direction.

## Delivered workflows

### File identity and integrity

`InspectFile` reads a selected regular file only to calculate evidence and returns:

- absolute path;
- size;
- mode/permissions;
- mtime;
- UID/GID and owner/group when resolvable;
- SHA-256;
- SHA-512;
- optional expected-checksum match/mismatch.

No file contents are returned or persisted by the Workbench result.

`CompareFiles` independently fingerprints two readable regular files and reports whether their SHA-256 fingerprints match. Filename, path, timestamp, and requested permissions are not treated as proof of identical contents.

### DNS inspection

`InspectDNS` uses the host system resolver with bounded time and result size. It can return:

- A;
- AAAA;
- CNAME when distinct from the requested canonical name;
- MX;
- NS;
- TXT;
- PTR for IP input.

Records are deduplicated and deterministically sorted. Individual unavailable lookups remain explicit rather than erasing successful record types.

### HTTP redirect/status inspection

`InspectHTTP` performs a direct HEAD request for complete `http://` or `https://` URLs and returns a bounded redirect chain with selected response metadata.

Boundaries:

- HEAD only;
- no request body;
- no proxy environment use;
- no credentials embedded in the URL;
- no custom headers/cookies/authentication input;
- bounded dial/TLS/header/total timeout;
- bounded redirect count.

The Workbench is meant to answer where a URL goes and what status/selected headers are observed, not become an HTTP client or proxy administration tool.

### Public certificate identity

`InspectCertificateFile` parses the first public `CERTIFICATE` PEM block and returns certificate metadata/fingerprint plus public-key/signature algorithm evidence.

`CompareCertificateFileToServed` uses the existing direct-TLS evidence path and compares the public certificate file's exact SHA-256 fingerprint with the certificate actually presented by the endpoint.

If a private-key PEM block appears before a certificate, inspection refuses the file. HostSleuth does not expose a private-key viewer.

Certificate identity is one M9 workflow alongside files, DNS, and HTTP; it is not the sole future enhancement direction.

## Interfaces

CLI:

- `hostsleuth workbench file [-expect CHECKSUM] PATH`
- `hostsleuth workbench compare LEFT RIGHT`
- `hostsleuth workbench dns NAME_OR_IP`
- `hostsleuth workbench http URL`
- `hostsleuth workbench cert [-target HOST:PORT] CERT_PATH`

JSON API:

- `/api/workbench/file`
- `/api/workbench/compare`
- `/api/workbench/dns`
- `/api/workbench/http`
- `/api/workbench/cert`

Web UI:

- a dedicated **Workbench** tab;
- separate cards for file identity, file comparison, DNS, HTTP, and certificate identity;
- no arbitrary command input or generic launcher.

## Web/API security boundary

M9's browser/API capabilities are more sensitive than passive dashboard viewing. File hashing can reveal whether a readable file matches a known value, and server-side DNS/HTTP operations can turn a host into a probe if exposed without authentication.

For that reason all `/api/workbench/*` handlers require the HTTP peer to be loopback.

This preserves:

- local browser use;
- the recommended SSH tunnel, which arrives via loopback;
- local CLI use.

It rejects unauthenticated LAN clients even if the general HostSleuth UI was deliberately bound to a LAN address.

M9 adds no file writes, service controls, package/firewall actions, arbitrary commands, credentials, or private-key output.

## Validation defects found before closeout

### Host umask test assumption

The first real-host test created a temporary file requesting mode `0640`, but the OMV acceptance user's umask produced `0600`. HostSleuth correctly reported the actual observed mode; the test was wrong to assume a particular host umask.

The test was changed to verify truthful permission evidence without requiring one environment-specific mode.

### API multiple-return wiring

The first CLI/API integration build exposed a Go compile error because `(value, error)` returns from the core helpers were passed directly into an API helper expecting separate arguments.

Handlers now explicitly unpack `value, err` before writing the response.

### CI UI blind spot

Earlier CI syntax-checked only the base `app.js`, while Service Story, Incident Lens, and Workbench JavaScript are concatenated into the served script at runtime.

M9 strengthens CI to syntax-check:

- base `app.js`;
- `service_story.js`;
- `incident_lens.js`;
- `workbench.js`;
- the exact concatenated script in runtime order.

Docker smoke also exercises a Workbench API path over loopback.

## Isolated real-host acceptance

Acceptance used the actual OMV Debian environment with an isolated clone and temporary state under `/tmp/hostsleuth-m9-accept`. It did not modify the live HostSleuth deployment, Compose definition, persistent production state, or recovery repository.

Validated:

- branch `go test ./...`;
- `go vet ./...`;
- native build;
- `gofmt` cleanliness;
- individual and concatenated JavaScript syntax;
- selected-file SHA-256 match against independent `sha256sum` plus SHA-512 output;
- exact file-to-file fingerprint comparison;
- real DNS A/AAAA evidence;
- real HTTPS HEAD/status evidence;
- public certificate inspection;
- exact local-public-cert vs served-certificate fingerprint matching against an isolated temporary self-signed TLS listener;
- isolated loopback API/UI smoke.

A temporary private key was generated only to start the isolated TLS acceptance listener. HostSleuth itself was pointed only at the public certificate file.

## Product/research boundary

M9 deliberately validates a broader principle: HostSleuth should not merely put existing commands behind buttons. A useful Workbench tool must answer one concrete troubleshooting question with bounded evidence.

Ongoing consumer research remains broader than certificates and should continue evaluating areas such as:

- endpoint expectation contracts;
- DNS resolver/delegation/split-view mismatches;
- HTTP redirect/reverse-proxy/upstream mismatches;
- file permissions/ownership/deployment paths;
- STARTTLS and other protocol-specific endpoint behavior;
- certificate deployment/rollout verification;
- later narrow safe actions with observable postcondition verification.

These research ideas do not automatically become implementation scope.

## Closeout boundary

M9 remains read-only. Stable `v0.3.0` is still the current published release and does not contain M7/M8/M9 source capabilities. Live OMV and `OMV-Docker-Rebuild` remain intentionally pinned to the known-good `0.1.0` deployment/recovery state.

The next approved milestone is **M10 — Reboot Story**.
