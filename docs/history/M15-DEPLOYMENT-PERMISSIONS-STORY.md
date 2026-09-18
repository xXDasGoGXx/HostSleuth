# M15 — Deployment / Permissions Story Closeout

Status: source implementation complete and merged on 2026-09-17. Stable/public/live/recovery remain aligned on v0.8.0; publication and rollout are separate gates.

## Product question

> The process is running; why can it not use this path/socket/port?

M15 delivers the filesystem and Unix-domain-socket permission part of that question for one native systemd service and one explicit absolute path. Existing Diagnose and Service Story remain the source for TCP/listener/port reachability rather than duplicating that logic here.

## Delivered evidence

For the selected running native service, HostSleuth records bounded process identity evidence:

- MainPID;
- effective UID;
- effective GID;
- supplementary GIDs;
- configured systemd User/Group when present;
- current working directory;
- executable identity;
- effective Linux capability mask.

For the explicit path, HostSleuth inspects only the exact component chain from / to the target and records:

- node type;
- owner UID/GID;
- permission mode;
- owner/group/other relation for the observed service identity.

No sibling enumeration or recursive filesystem walk occurs.

## Deterministic permission reasoning

Every parent directory receives an explicit traverse/search decision.

Target decisions are:

- directory: read, write, traverse;
- regular/other file: read, write, execute;
- Unix-domain socket: connect using the socket write bit.

Owner/group/other selection uses effective UID/GID plus supplementary groups.

Effective UID 0 is not treated as unconditional proof of access. HostSleuth consults the observed effective capability mask for CAP_DAC_OVERRIDE and CAP_DAC_READ_SEARCH. Missing capability evidence becomes unknown rather than a guessed pass.

The first proven failed decision is returned as first_problem.

## Reasoning boundary

M15 is a deterministic POSIX ownership/mode story, not a claim that every Linux access-control layer has been modeled.

The result explicitly notes that real access can still be affected by:

- POSIX ACLs;
- AppArmor/SELinux or other Linux Security Modules;
- namespaces;
- mount flags;
- application-level policy.

M15 does not read ACL databases, security-policy files, process environments, or file contents.

## Docker bind-mount context

When native Docker inventory and the Docker CLI are already accessible, HostSleuth may return only bind mounts relevant to the explicit requested/resolved path.

Returned fields are bounded to:

- container name;
- host source;
- container destination;
- read-write/read-only flag;
- which side matched the path.

The implementation considers at most 64 containers from the existing snapshot and returns at most 16 matching bind mounts.

Docker deployment mode does not attempt to cross the container boundary to manufacture native systemd process identity. It reports native permission evidence as unavailable.

## Interfaces

CLI:

    hostsleuth permissions --service UNIT /absolute/path

HTTP API:

    GET /api/permissions-story?service=UNIT&path=/absolute/path

Web UI:

- dedicated Permissions navigation/view;
- service + explicit path form;
- process identity cards;
- visual permission chain;
- searchable evidence table;
- relevant Docker bind-mount context;
- visible reasoning-boundary disclosure.

## Security / privacy boundary

M15 remains read-only and adds no privilege.

It does not add:

- file contents;
- recursive directory crawling;
- arbitrary command execution;
- chmod/chown/setfacl;
- service restart/reload;
- container mutation;
- writable Docker-socket access;
- credential/environment-variable reads;
- automatic remediation.

Safe Actions remain unchanged.

## Local engineering acceptance

Local validation used Go 1.24.13 on an isolated /tmp clone.

Passed:

- gofmt;
- git diff --check;
- go test ./...;
- go test -race ./internal/core;
- go vet ./...;
- native build;
- JavaScript syntax checks.

Focused tests cover:

- owner/group/other mode-class selection;
- supplementary-group access;
- parent traversal failure;
- Unix-domain-socket connect semantics;
- effective-UID-0 capability handling;
- component-aware bind-mount path matching;
- bounded path metadata collection.

## Native acceptance

Bounded acceptance on the OMV host changed no services and no filesystem permissions.

Passing case:

- service: ssh.service;
- path: /etc/ssh;
- result: pass;
- observed UID/GID: 0/0;
- parent and target decisions permitted by the observed mode/capability evidence.

Failing case:

- service: dbus.service;
- path: /root;
- observed configured identity: messagebus/messagebus;
- observed UID/GID: 990/990;
- target mode: 0700 owned by UID/GID 0/0;
- relation: other;
- result: fail;
- first_problem: read:/root.

This exercised both a real running-service success path and a real running-service DAC denial without changing the host.

## Disposable HTTP / UI acceptance

A temporary HostSleuth server was started only on 127.0.0.1:18787 from the locally built M15 binary.

Acceptance confirmed:

- GET /api/permissions-story reproduced the dbus.service -> /root deterministic failure;
- first_problem remained read:/root;
- the served JavaScript contained the Deployment / Permissions Story view;
- the served CSS contained the permission-chain layout.

The disposable server was signaled to stop and was not the production HostSleuth deployment.

## GitHub CI acceptance

PR #55 exact head:

    78bd31d8cdb34f9f5e592e57f14f2e8537e79b41

CI run:

    35302370815

All jobs passed:

- test / format / vet / Web JavaScript syntax / native build;
- native Safe Actions smoke;
- Docker smoke, including the M15 Docker-mode boundary assertion;
- linux/amd64 image build;
- linux/arm64 image build.

The Docker smoke requires M15 to report native service identity as unavailable in Docker deployment mode instead of weakening the deployment security boundary.

## Merge acceptance

PR #55 squash-merged to main at:

    d64bb387c8f9efcf5dcb9f814f54b519ee231205

## Release boundary

M15 source is now complete on main.

Stable/public/live/recovery are still:

    v0.8.0 / mjmalleo/hostsleuth:0.8.0

No image publication, Arcane production redeploy, recovery change, Safe Action expansion, or other live-system change was performed for M15.

The next approved source milestone is M16 — STARTTLS / Mail Service Story.
