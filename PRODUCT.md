# HostSleuth — Product Definition

## One sentence

**HostSleuth is a small, local-first Linux flight recorder that remembers meaningful host changes and gives deterministic, evidence-backed explanations for why a host/service/port is or is not reachable.**

## The two questions

HostSleuth exists to answer:

1. **What changed on this Linux host?**
2. **Why can I not reach this host/service/port?**

Everything in the product should materially improve one of those two answers.

## Core workflow

### 1. Observe

Periodically capture a compact snapshot of useful Linux state: host identity, networking, listeners, systemd services, Docker containers, filesystems, and related evidence.

### 2. Remember

Compare snapshots and record meaningful changes as a concise local event timeline. Avoid noisy changes that do not help explain incidents.

### 3. Diagnose

Given `host:port`, follow a deterministic evidence path. Current evidence can include:

- DNS resolution;
- kernel route lookup;
- TCP connectivity;
- target-aware local listener state;
- Docker port publication, bind address, and network context;
- bounded nftables candidate evidence;
- bounded failed-systemd candidate evidence.

Strong evidence outranks weak candidate evidence. Missing optional evidence is reported as unavailable/unknown rather than converted into a false diagnosis.

## Product boundaries

HostSleuth is intentionally:

- local-first;
- single-host first;
- read-only;
- deterministic before explanatory;
- useful without cloud services, accounts, API keys, or AI;
- small enough to inspect and understand.

HostSleuth is **not** intended to become:

- a Prometheus/Grafana replacement;
- a general monitoring platform;
- HomeCommander;
- a Docker or firewall manager;
- an automatic repair agent;
- a generic root shell;
- an AI-dependent troubleshooting system;
- a distributed enterprise observability platform.

## Read-only rule

V0.1 observes, records, compares, and explains. It does not automatically restart services, modify firewall rules, reconfigure applications, repair containers, or mutate the monitored host.

Any future repair capability must be a separate, explicitly designed workflow and must not weaken deterministic diagnosis.

## User-facing success test

A useful HostSleuth build should pass this simple test:

1. Install one binary on a Linux host.
2. Open the local dashboard or use the CLI.
3. See meaningful recent changes without noise.
4. Enter a host and port.
5. Receive a concise explanation of what HostSleuth can prove, what it cannot prove, and the strongest evidence supporting that conclusion.

If a proposed feature does not materially improve that experience, it is not a priority yet.

## Current product stage

M0 repository foundation, M1 deployable single-host MVP, and M2 deeper deterministic diagnosis are complete.

The next phase is **product hardening through real use**, not another large subsystem. Use the current build on real incidents, improve clarity where actual users get confused, and add new collectors only when real troubleshooting exposes a concrete gap.
