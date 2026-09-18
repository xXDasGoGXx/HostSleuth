# M13 — DNS Detective + Admin Console v1

Status: implemented and locally accepted; PR/CI/release remain separate gates.

## Admin question

> Which DNS view is this host actually seeing, which resolver disagrees, and is the difference consistent with split-view DNS?

M9 Workbench can inspect one system-resolver view. M13 turns DNS into a deterministic comparison story.

## Privacy model

HostSleuth never silently sends a hostname to a third-party public resolver.

Every DNS Detective run uses:

1. the host process's normal system resolver; and
2. only the custom resolver IPs explicitly supplied by the operator for that request.

Custom resolvers are bounded to DNS port 53 in the Web/API surface.

Recent resolver/target values may be remembered by the browser UI only; they are not persisted in HostSleuth server state.

## Bounded input

- DNS name or IP address;
- zero to four custom resolver specifications;
- custom syntax: label=IP or bare IP;
- IPv4 and IPv6 accepted;
- DNS port 53 only for the first Web/API version.

Examples:

    192.168.2.5
    pihole-a=192.168.2.5
    pihole-b=192.168.3.5
    public=1.1.1.1

## Evidence collected per resolver

For a hostname:

- A/AAAA answers;
- canonical CNAME target when it differs from the input;
- lookup duration;
- success/error evidence.

For an IP input:

- PTR answers;
- lookup duration;
- success/error evidence.

Every address is normalized and classified as:

- loopback;
- private;
- link-local;
- global;
- unspecified/multicast/other where applicable.

The first version does not claim TTL evidence because Go's standard resolver API does not expose TTLs. TTL-aware wire queries may be a later bounded enhancement.

## Runtime resolver context

Read and parse a bounded /etc/resolv.conf view:

- nameserver addresses;
- search/domain suffixes;
- selected resolver options such as ndots, timeout, and attempts when present.

This is metadata only; HostSleuth does not rewrite resolver configuration.

Docker mode must label this honestly as the resolver configuration visible inside the running HostSleuth container, not necessarily the host's native resolver file.

## Deterministic comparison

Statuses:

- agree — two or more successful resolvers returned the same normalized answer;
- diverge — successful resolvers returned different normalized answers;
- partial — one or more resolver lookups failed and no proven divergence exists;
- single — only one resolver result is available.

A divergence where one answer set is private/link-local and another is global may be described as:

> Resolver views differ between private/local and global addresses; this is consistent with split-view DNS or resolver-specific overrides.

HostSleuth must not claim that split-horizon configuration exists unless direct configuration evidence proves it.

## Interfaces

CLI target:

    hostsleuth dns [--resolver LABEL=IP] NAME_OR_IP

HTTP API:

    GET /api/dns-detective?name=...&resolver=...

Web UI:

- dedicated DNS Detective view;
- system resolver always shown first;
- add/remove custom resolver fields;
- side-by-side resolver cards;
- prominent verdict;
- exact differing answer sets highlighted;
- one-click handoff to Diagnose or Expected Endpoint Contract when the input can be used as an endpoint host.

## Admin Console v1

M13 also upgrades the existing shell without rewriting feature logic:

- desktop navigation becomes a sticky left rail;
- main evidence workspace grows wider;
- global target command bar appears above views;
- target actions: Diagnose, Expectations, DNS;
- recent targets are browser-local only;
- density toggle: Comfortable / Compact;
- / keyboard shortcut focuses the command bar;
- mobile returns to horizontal navigation;
- existing dynamic feature tabs continue to work.

The visual direction is dense enough for admins but not cramped: clear status colors, strong typography, restrained motion, and evidence-first hierarchy.

## Security boundary

M13 is read-only.

No:

- DNS configuration changes;
- zone transfers;
- dynamic DNS updates;
- arbitrary UDP service probes;
- automatic public DNS queries;
- persistent server-side resolver history;
- DNS polling/monitoring;
- alerts;
- remediation.

The custom resolver field is an IP address only and is constrained to port 53.

## Acceptance

Required before merge:

- gofmt;
- go vet ./...;
- go test ./...;
- native build;
- Web JS syntax checks;
- deterministic unit tests for normalization/comparison/failure behavior;
- disposable DNS fixture proving agree/diverge/partial cases;
- disposable Web/API acceptance;
- existing Docker/native Safe Actions smoke unchanged;
- responsive Admin Console sanity check via served HTML/CSS/JS.
