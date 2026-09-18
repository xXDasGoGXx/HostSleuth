# M15 — Deployment / Permissions Story

Status: implementation in progress; local validation passed, PR/CI acceptance pending.

## Product question

> The process is running; why can it not use this path/socket/port?

M15 answers the filesystem-permission part of that question for one native systemd service and one explicit absolute path. Existing Diagnose and Service Story remain responsible for TCP/listener reachability; M15 does not duplicate them.

## Bounded inputs

- one native systemd service unit;
- one explicit absolute filesystem path or Unix-domain socket path.

No directory selection, globbing, recursive crawl, arbitrary process selector, or bulk scan is accepted.

## Runtime identity evidence

HostSleuth reads bounded service metadata from systemctl show, then uses the running service MainPID to inspect process metadata:

- effective UID;
- effective GID;
- supplementary GIDs;
- configured systemd User/Group when present;
- current working directory from /proc/PID/cwd;
- executable identity from /proc/PID/exe;
- effective capability mask from /proc/PID/status.

No process memory, environment, command output beyond the fixed systemd property set, or file contents are exposed.

## Path evidence

The requested path must be absolute.

HostSleuth resolves symlinks when possible and then inspects only the exact component chain from / to the target. For each component it records:

- path;
- node type;
- owner UID;
- owner GID;
- permission mode;
- whether the process is evaluated through owner, group, or other class.

The implementation never enumerates sibling entries and never recursively walks a directory.

## Deterministic access reasoning

For every parent directory, HostSleuth evaluates search/traverse (x) permission for the observed service identity.

For the target:

- directory: read, write, and traverse;
- regular/other file: read, write, and execute;
- Unix-domain socket: connect using the socket write bit.

Owner/group/other selection uses effective UID/GID plus supplementary groups.

For effective UID 0, HostSleuth does not assume unconditional root bypass. It consults the effective capability mask:

- CAP_DAC_OVERRIDE may bypass ordinary DAC read/write/search denials;
- CAP_DAC_READ_SEARCH may bypass read/search denials;
- executing a non-directory with no execute bit set is not promoted to a pass merely because CAP_DAC_OVERRIDE is present;
- missing capability evidence yields unknown rather than a guessed pass.

The first proven failed chain decision becomes the first problem.

## Reasoning boundary

The result is intentionally a POSIX ownership/mode story, not a claim that every kernel access-control layer has been modeled.

The UI/API explicitly notes that these can still alter real access:

- POSIX ACLs;
- Linux Security Modules such as AppArmor/SELinux;
- namespaces;
- mount flags;
- application-level policy.

M15 does not read ACL databases, security-policy files, or mount contents. Those are not silently inferred.

## Docker bind-mount metadata

When native Docker inventory and the Docker CLI are already available, HostSleuth inspects only the bounded container inventory from the current snapshot and returns bind mounts relevant to the explicit requested/resolved path.

Returned fields are limited to:

- container name;
- bind source;
- bind destination;
- read-write/read-only flag;
- whether the match was on host source or container destination.

Limits:

- at most 64 current snapshot containers are considered;
- at most 16 matching bind mounts are returned;
- no container filesystem walk;
- no writable Docker socket requirement is introduced;
- Docker deployment mode does not attempt to escape the container to obtain native systemd identity.

## CLI

    hostsleuth permissions --service UNIT /absolute/path

Output is typed JSON.

## API

    GET /api/permissions-story?service=UNIT&path=/absolute/path

The endpoint is read-only.

## Admin Console

M15 adds a dedicated Permissions view with:

- native service + absolute path form;
- process identity cards;
- visual parent-directory/target permission chain;
- searchable evidence table;
- relevant Docker bind-mount cards;
- visible reasoning-boundary disclosure.

The evidence search is browser-side filtering of the already returned bounded result; it does not issue broader filesystem queries.

## Privacy and security boundary

M15 does not add:

- file contents;
- recursive listing/crawling;
- arbitrary command execution;
- chmod/chown/setfacl;
- service restart/reload;
- container mutation;
- writable Docker socket access;
- privilege escalation;
- credential/environment-variable reads;
- automatic remediation.

Safe Actions remain unchanged.

## Validation

Local validation on Go 1.24.13:

- gofmt;
- go test ./...;
- go test -race ./internal/core;
- go vet ./...;
- native build;
- JavaScript syntax checks;
- git diff --check.

Permanent CI also checks the M15 JavaScript and verifies Docker mode reports native permission identity as unavailable instead of crossing the deployment boundary.

## Acceptance still required

Before M15 can be called complete:

1. push the focused implementation branch;
2. pass the full GitHub CI matrix on the exact PR head;
3. update README/handoff/TODO/roadmap and create the M15 closeout record only after CI acceptance;
4. merge the focused PR.

## Native acceptance observed

Bounded native acceptance on the OMV host passed without changing services or filesystem permissions:

- `ssh.service` against `/etc/ssh` returned `pass` with the observed UID/GID/capability identity and a fully permitted traversal/target chain;
- `dbus.service` against `/root` returned `fail` with `read:/root` as the first problem because the observed `messagebus` identity was evaluated through the `other` class of mode `0700`.

This acceptance exercised both a real running-service success path and a real running-service denial path.

Publication, production rollout, and recovery alignment remain separate owner-controlled gates.
