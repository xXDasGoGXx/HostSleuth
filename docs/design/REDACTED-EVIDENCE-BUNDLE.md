# Redacted Evidence Bundle — Threat Model and Redaction Contract

Status: DESIGN DRAFT — exporter implementation must not begin until this contract is accepted.

## Goal

Allow a HostSleuth operator to create a bounded, portable diagnostic bundle that can be shared with another person without silently exporting secrets or high-risk raw host data.

The bundle is for troubleshooting evidence, not backup, forensic imaging, remote administration, or configuration export.

## Security posture

The exporter must fail closed. The safe default is omission or pseudonymization, not disclosure.

The bundle must never silently contain:

- passwords or passphrases;
- API keys, bearer tokens, session tokens, cookies, OAuth material, or authorization headers;
- SSH private keys, TLS private keys, recovery keys, or seed material;
- PEM private-key blocks;
- arbitrary configuration-file contents;
- arbitrary local file contents;
- raw environment-variable dumps;
- raw shell history;
- raw command output that has not been explicitly classified as export-safe;
- raw journal excerpts by default;
- credentials embedded in URLs;
- Web/HTTP request headers supplied by users;
- writable control credentials or action confirmation material.

A redaction failure in one field must not cause the exporter to fall back to the original value.

## Threat model

### Intended recipient

Assume the recipient is a technically capable third party who needs diagnostic evidence but should not receive reusable credentials or unnecessary identifying data.

Examples:

- a system administrator;
- an application/vendor support engineer;
- a developer reviewing a bug report;
- a trusted forum/community helper.

### Adversary model

Assume a shared bundle may later be copied, uploaded, indexed, forwarded, or posted publicly by mistake.

The design therefore protects against accidental disclosure of:

- reusable secrets;
- private infrastructure addresses and hostnames;
- usernames and home-directory names;
- private registry/repository names;
- unique machine identifiers;
- configuration and certificate fingerprints that can act as correlation identifiers;
- user-specific absolute paths;
- internal URLs/domains;
- free-form evidence strings containing any of the above.

The design does not promise anonymity against a recipient who already knows the environment and can infer identity from timing, software versions, topology shape, or other contextual clues.

## Product boundaries

The first implementation must remain:

- local-first;
- single-host;
- explicit and operator-triggered;
- read-only with respect to host state;
- bounded in record count and output size;
- deterministic for the same source evidence and redaction policy;
- previewable before an export file is written;
- independent from Optional Safe Actions execution.

It must not add:

- cloud upload;
- automatic sending/sharing;
- arbitrary file inclusion;
- arbitrary directory traversal;
- arbitrary command execution;
- remote collection from other hosts;
- support-agent access;
- automatic remediation.

## Proposed first bundle format

The first implementation should produce one local ZIP archive containing JSON/text files plus a manifest. No archive is written during preview.

Proposed layout:

```text
hostsleuth-evidence-YYYYMMDDTHHMMSSZ.zip
  manifest.json
  summary.txt
  snapshot.json
  events.json
  action-audit.json          # only when records exist
  checksums.txt
```

Later evidence views such as Diagnosis, Service Story, Incident Lens, Workbench, and Reboot Story may be added only through explicit typed inclusion paths using the same redaction contract. The first implementation should not become a generic "serialize everything" facility.

## Manifest requirements

`manifest.json` must state at minimum:

- bundle schema version;
- HostSleuth version;
- generated-at time in UTC;
- redaction policy name/version;
- exact included evidence classes;
- exact omitted evidence classes;
- count of redactions by category;
- record limits applied;
- whether any source evidence was unavailable;
- SHA-256 for every payload file;
- a warning that redaction reduces risk but cannot guarantee anonymity.

The manifest must not contain the original values that were redacted.

## Default inclusion policy

### Included after redaction

The first bundle may include:

- current snapshot structural/runtime evidence;
- bounded recent change events;
- bounded Safe Action audit records, if present;
- HostSleuth version/schema/mode;
- timestamps in UTC;
- status/severity/state fields;
- counts and sizes that are not themselves secrets;
- ports and protocol names;
- package versions;
- OS/kernel/architecture information.

### Excluded by default

The first bundle must exclude:

- raw previous/current boot journal text;
- raw systemd journal excerpts;
- raw action command output;
- arbitrary Workbench file metadata for user-selected paths;
- arbitrary Workbench HTTP URLs or DNS query inputs unless a later typed evidence exporter explicitly redacts them;
- arbitrary file contents;
- arbitrary configuration contents;
- ActionPreview command arrays and confirmation strings;
- any field the exporter cannot classify safely.

## Deterministic pseudonymization

Redaction must preserve equality relationships inside one bundle without preserving the original values.

Use bundle-local deterministic aliases assigned from a sorted set of observed values. The alias table exists only in memory and is not written to the bundle.

Examples:

```text
openmediavault        -> host-01.invalid
192.168.2.181         -> ipv4-01
10.0.8.2              -> ipv4-02
2001:db8::1234        -> ipv6-01
AA:BB:CC:DD:EE:FF     -> mac-01
/home/matthew/project -> path-01
abc123...docker-id    -> id-01
```

The same source value must map to the same alias everywhere in the bundle. Different source values must not share an alias.

Aliases must not encode or hash the source value in a reversible/dictionary-testable form.

## Identifier rules

### Hostnames and domain names

- Preserve `localhost`.
- Pseudonymize all other hostnames/FQDNs.
- Preserve ports separately.
- Do not preserve original domain suffixes.

### IP addresses and CIDRs

- Preserve loopback/unspecified semantics (`127.0.0.1`, `::1`, `0.0.0.0`, `::`) because these values are not host-identifying and are diagnostically important.
- Pseudonymize all other IPv4/IPv6 addresses.
- For CIDRs, preserve prefix length while replacing the address/network value with an alias.
- Preserve port numbers.

### MAC addresses

Pseudonymize all MAC addresses.

### Paths

The following HostSleuth-defined configuration paths may remain literal because they are fixed product evidence targets, not operator-selected paths:

- `/etc/hosts`
- `/etc/fstab`
- `/etc/ssh/sshd_config`
- `/etc/docker/daemon.json`
- `/etc/nftables.conf`

All other absolute paths are pseudonymized as `path-NN` by default.

No original path-to-alias lookup table is exported.

### Usernames and e-mail addresses

Pseudonymize usernames when they appear in structured or free-form evidence. Pseudonymize all e-mail addresses.

### Docker/container identifiers

- Container IDs: pseudonymize.
- Container names: pseudonymize.
- Docker network names: pseudonymize unless they are the fixed default names `bridge`, `host`, or `none`.
- Image references: preserve only the tag/digest/version portion required for diagnosis; pseudonymize the registry/repository identity by default.

### systemd service names

Service names are diagnostically important but may contain organization-specific names. Default bundle policy pseudonymizes service unit names while preserving the `.service` suffix and equality relationships.

### Package names

Package names and versions are included as-is. They are software inventory rather than reusable secrets and are important for supportability.

### Boot IDs, UUIDs, long opaque IDs

Pseudonymize machine-unique opaque identifiers, including kernel boot IDs and Docker-style IDs.

### Configuration fingerprints

Do not export raw configuration SHA-256 fingerprints by default. Replace each distinct fingerprint with a bundle-local alias such as `config-fingerprint-01` so equality/change relationships remain visible without exporting the stable correlation identifier.

### Certificate fingerprints and serials

Pseudonymize certificate fingerprints and serial numbers by default while preserving equality relationships. Certificate validity dates, issuer class after hostname/domain redaction, and status may remain.

## URL rules

When a URL is included by a future typed evidence exporter:

- remove username/password userinfo unconditionally;
- remove query string unconditionally by default;
- remove fragment unconditionally;
- pseudonymize host/domain/IP;
- preserve scheme and port;
- preserve only a bounded path shape if explicitly needed by that evidence type; otherwise replace path with `/[redacted]`.

## Secret-pattern redaction

Before identifier pseudonymization, every exportable free-form string must pass a secret scrubber.

At minimum it must redact values associated with case-insensitive labels such as:

- `password`, `passwd`, `passphrase`;
- `token`, `access_token`, `refresh_token`, `id_token`;
- `api_key`, `apikey`, `secret`, `client_secret`;
- `authorization`, `bearer`, `basic`;
- `cookie`, `set-cookie`, `session`, `sessionid`;
- `private_key`, `ssh_key`.

It must also redact:

- PEM private-key blocks;
- common private-key headers even when malformed/truncated;
- URL userinfo;
- credential-looking `key=value` / `key: value` forms;
- long opaque credential-like strings when they appear next to a secret label.

The scrubber must operate on copies only and never modify retained HostSleuth state.

## Free-form evidence policy

Free-form strings are the highest-risk evidence class.

Rules:

1. secret scrubber first;
2. structured identifier redaction second;
3. bounded length after redaction;
4. no fallback to original content on parse/redaction error;
5. journal-derived text is omitted by default rather than merely scrubbed;
6. command output is omitted by default rather than merely scrubbed.

Event summaries and ordinary deterministic `Check.Evidence` strings may be included only after this full pipeline.

## Timestamp policy

Keep timestamps in UTC because event ordering and incident correlation depend on them.

Do not export local timezone names, locale, or geolocation metadata.

## Preview-before-export contract

`hostsleuth evidence preview` must be the first user workflow.

Preview must show:

- which evidence classes would be included;
- record counts;
- redaction counts by category;
- files that would be created;
- estimated/bounded output size;
- explicit warnings for omitted evidence;
- no original secret values.

Export must require a separate explicit command. Preview alone writes no archive.

## Fail-closed conditions

Export must stop with an error rather than write a partially unsafe bundle when:

- redaction initialization fails;
- a selected evidence class has no registered redactor;
- manifest/checksum generation fails;
- archive creation fails mid-write;
- output would exceed the configured bundle size limit;
- a payload attempts to include an explicitly forbidden field/class;
- redaction encounters a value it classifies as a private key or credential that cannot be safely transformed.

Temporary export files must be owner-only and removed on failure where practical.

## Output permissions and integrity

On platforms that support Unix permissions:

- temporary files/directories: owner-only;
- final archive: mode `0600`.

`checksums.txt` must use SHA-256 and cover every payload except itself.

The final archive itself may also be accompanied by a printed SHA-256 value after successful export.

## Bounded first implementation

Initial limits proposed for acceptance:

- latest snapshot only;
- at most 200 recent change events;
- at most 100 action-audit records;
- maximum uncompressed bundle payload: 2 MiB;
- maximum individual JSON payload: 1 MiB;
- no recursively discovered files;
- no user-specified arbitrary include paths.

These limits can be revisited only with explicit justification.

## Acceptance tests required before merge

The implementation must prove at least:

- password/token/cookie/authorization examples never survive export;
- PEM private keys never survive export;
- URL credentials/query strings never survive export;
- repeated hostnames/IPs/paths map consistently within one bundle;
- different identifiers map to different aliases;
- loopback/unspecified address semantics remain visible;
- configuration/certificate fingerprints are pseudonymized;
- service/container/custom network identifiers are pseudonymized;
- package names/versions remain useful;
- raw journal text and raw command output are absent by default;
- preview writes no archive;
- export creates owner-only output;
- manifest lists included/omitted evidence and redaction counts;
- manifest/checksums contain no original redacted values;
- size/record limits are enforced;
- failed redaction/archive generation leaves no apparently successful partial bundle;
- Docker and native source evidence both use the same redaction policy;
- existing HostSleuth snapshot/event/action behavior remains unchanged.

## Decision needed before implementation

Owner acceptance of this contract authorizes implementation of the bounded first version only. Any future option to expose raw identifiers, include journal text, include arbitrary Workbench evidence, or weaken a default redaction requires a separate explicit design decision.
