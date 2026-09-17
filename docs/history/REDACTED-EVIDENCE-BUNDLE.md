# Redacted Evidence Bundle — Implementation and Acceptance

Status: complete in source on PR #40; release/publication/deployment remain separate decisions.

## Goal

Provide one bounded, local, operator-triggered support package that makes selected HostSleuth evidence portable without silently exporting reusable secrets or unnecessary raw host identifiers.

The accepted threat model and redaction contract are documented in `docs/design/REDACTED-EVIDENCE-BUNDLE.md`.

## Delivered first version

CLI:

```text
hostsleuth evidence preview
hostsleuth evidence export [--output PATH]
```

`preview` builds the same redacted payload in memory and reports included/omitted evidence classes, record counts, redaction counts, files, estimated bounded payload size, unavailable evidence, and warnings. It does not write an archive.

`export` creates one local ZIP and refuses to overwrite an existing output file.

The first typed evidence surface contains only:

- current `Snapshot`;
- at most 200 recent `Event` records;
- at most 100 `ActionAudit` records when present.

There is no generic serializer and no arbitrary include path.

## Bundle format

The first bundle contains:

```text
manifest.json
summary.txt
snapshot.json
events.json
action-audit.json   # only when records exist
checksums.txt
```

`manifest.json` records bundle schema/policy, HostSleuth version, UTC generation time, included and omitted evidence classes, redaction counts, configured limits, unavailable evidence, payload file sizes, and payload SHA-256 values.

`checksums.txt` covers bundle payloads and the manifest. The CLI also prints the SHA-256 of the completed archive.

## Redaction model

The `redacted-v1` policy uses bundle-local deterministic aliases so repeated values remain correlatable inside one bundle without exporting the original lookup table.

The first implementation pseudonymizes or removes:

- non-loopback hostnames/domains and IPv4/IPv6 addresses;
- MAC addresses;
- arbitrary paths;
- service/container/network/interface identities;
- opaque machine/boot/container-style IDs;
- image registry/repository identity while retaining useful version/tag/digest suffixes;
- e-mail addresses;
- configuration fingerprints;
- credentials in URL userinfo;
- URL query strings/fragments;
- password/token/API-key/cookie/session/authorization forms;
- full and truncated PEM/private-key material.

It preserves diagnostically useful semantics where safe, including ports, prefix lengths, loopback/unspecified addresses, state/status fields, package names/versions, OS/kernel/architecture data, UTC ordering, and URL scheme/port/path-shape information.

## Explicit exclusions

The first version does not export:

- raw journal text;
- raw Safe Action command output;
- ActionPreview command arrays or confirmation strings;
- arbitrary local file contents;
- arbitrary configuration contents;
- arbitrary Workbench file/URL/DNS inputs;
- cloud uploads or automatic sharing.

## Bounds and failure behavior

Limits:

- 200 recent events;
- 100 action audits;
- 1 MiB maximum individual JSON payload;
- 2 MiB maximum total uncompressed payload.

Temporary and final archive output use owner-only permissions (`0600`) where supported. Export writes through a temporary file, refuses an existing destination, and removes temporary output on failure where practical. Unsafe/oversized payload construction fails before a successful archive is presented.

## Security bugs caught during development

Adversarial testing caught and fixed a bearer-token leak in the first secret-regex draft: `Authorization: Bearer <token>` could redact the word `Bearer` while leaving the token. The parser was tightened and a permanent regression test added.

A later security pass also added direct handling for:

- plain domains in free-form evidence;
- unseeded IPv6 identifiers;
- truncated private-key blocks without a closing PEM marker.

Structural acceptance additionally caught two non-leak evidence-quality defects and fixed them:

- IPv6 CIDR suffixes could be misclassified as paths;
- a path pass could mangle an already-redacted URL.

Regression tests now preserve readable redacted URL and IPv6-CIDR shape.

## Validation

An isolated OMV `/tmp` clone used Go 1.24.13 without installing or changing production software.

Local validation passed:

- `gofmt` cleanliness;
- `go vet ./...`;
- `go test ./...`;
- native build;
- `git diff --check`.

GitHub CI on the completed implementation passed:

- main test/format/vet/build job;
- linux/amd64 image build;
- linux/arm64 image build;
- supported Docker smoke;
- native-actions smoke.

Disposable end-to-end CLI acceptance used synthetic state only. It confirmed:

- preview created no ZIP;
- export created the expected bounded bundle;
- final archive mode was `0600`;
- bundle checksums verified;
- planted private host/domain/IP/IPv6/interface/service/container/network/registry/path/password/Bearer-token/URL-credential/query/private-key values did not survive the leak scan;
- redacted URL scheme/host/path shape remained useful;
- IPv6 CIDR prefix length remained useful without being mistaken for a path.

No live HostSleuth process, Arcane project, production state, stable image tag, or disaster-recovery pin was changed by this development/acceptance work.

## Release boundary

Completion of this source milestone does not publish a new HostSleuth release and does not authorize a production redeploy. Stable/public/live/recovery remain on `v0.4.0` until a separate release decision is made.

Any future raw/unredacted mode, journal inclusion, arbitrary file inclusion, broader Workbench evidence export, cloud upload, or automatic sharing requires a separate explicit design decision.
