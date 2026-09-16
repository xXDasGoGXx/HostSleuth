# HostSleuth — Consumer Opportunity Landscape

Status: research input only. This document does not change the approved roadmap or authorize implementation.

Last researched: 2026-09-16

## Research question

What do administrators already get from monitoring products, Linux admin consoles, reverse proxies, DNS tools, container dashboards, ACME clients, TLS scanners, scripts, and one-off websites — and where could HostSleuth provide something materially more useful instead of becoming another flavor of an existing tool?

The owner explicitly clarified that certificate/Certbot ideas are only one example among many. HostSleuth's opportunity research must therefore stay broad across Linux service, network, DNS, HTTP/proxy, container, file/deployment, mail/protocol, boot/recovery, and certificate workflows.

HostSleuth's differentiator remains deterministic correlation and verification: connect evidence that users otherwise gather from several commands, logs, dashboards, and websites, then show where the observed chain stops matching the expected one.

## What already exists and should not simply be copied

### Linux administration consoles

Cockpit already provides a browser-accessible Linux administration interface, journal views, systemd service management, networking, firewall integration, storage management, and a web terminal. HostSleuth should not become a second Cockpit or a generic system-control panel.

Research:
- https://docs.cockpit-project.org/cockpit-guide/latest/guide/features.html
- https://docs.cockpit-project.org/cockpit-guide/latest/guide/api-system.html
- https://docs.cockpit-project.org/cockpit-guide/main/guide/feature-systemd.html

### Uptime/status monitoring

Uptime Kuma already monitors HTTP(S), TCP, DNS, Docker containers, ping, certificates, and other availability signals, with notifications and status pages. HostSleuth should not become another polling/alerting dashboard or time-series uptime product.

Research:
- https://github.com/louislam/uptime-kuma

### Reverse proxies and routing dashboards

Traefik already exposes active routers/services through its dashboard and manages routing/load-balancing configuration. Caddy already provides reverse proxying and automatic HTTPS. Nginx Proxy Manager provides a management UI around reverse proxy and certificate workflows.

Research:
- https://doc.traefik.io/traefik/master/routing/services/
- https://doc.traefik.io/traefik/v2.10/operations/dashboard/
- https://caddyserver.com/docs/quick-starts/reverse-proxy
- https://caddyserver.com/docs/automatic-https
- https://nginxproxymanager.com/

Conclusion: HostSleuth should not become a reverse proxy. Its opportunity is to explain the path through a proxy when the user sees the wrong endpoint, redirect, certificate, status code, or backend failure.

### DNS visualization and validation

DNSViz already specializes in DNS/DNSSEC resolution-path visualization and configuration-error analysis. Command-line tools such as `dig` and `resolvectl` already expose raw DNS answers.

Research:
- https://dnsviz.net/

Conclusion: HostSleuth should not try to replace a full DNSSEC analyzer. It can add value by connecting DNS answers to the actual host route, endpoint, listener, proxy, certificate, and recent host changes.

### ACME clients and certificate automation

Certbot, acme.sh, and lego already issue/renew certificates and support hooks or deployment workflows. Caddy automatically manages TLS for endpoints it owns.

Research:
- https://eff-certbot.readthedocs.io/en/latest/using.html
- https://github.com/acmesh-official/acme.sh
- https://go-acme.github.io/lego/references/ref-flags/
- https://caddyserver.com/docs/automatic-https

Conclusion: HostSleuth should not win by inventing another generic renewal client or arbitrary hook mechanism.

### TLS scanners

testssl.sh already performs deep TLS/security assessment and supports STARTTLS protocols such as SMTP, POP3, and IMAP.

Research:
- https://github.com/testssl/testssl.sh

Conclusion: HostSleuth should not become a cipher/vulnerability scanner. Protocol-aware TLS evidence is useful only when it helps connect an endpoint to the local service/deployment story.

## Broad recurring troubleshooting gaps

The same pattern appears across many self-hosted/Linux problems: each individual component has a tool, but the administrator still has to mentally join the evidence.

Examples:

- DNS resolves, but to which address did this host actually connect and why?
- A reverse proxy is reachable, but is the failure at DNS, the proxy listener, the route rule, the container publication, the upstream listener, or upstream TLS?
- A systemd service says active, but is it actually listening where expected and does the endpoint respond?
- A container is running, but is its host port published, bound to the intended address, and reachable through the route/proxy path?
- A file was copied or regenerated, but does the destination match the source, do permissions/ownership make sense, and did the consumer actually adopt it?
- A host rebooted or an automated update ran, but which service/listener/container failed to return?
- Mail looks healthy locally, but do MX/PTR, SMTP banner, STARTTLS, certificate, and the local Postfix/Dovecot state agree?
- A remote service disappeared, but were several ingress/auth/tunnel containers removed or restarted together?

A representative self-hosted incident described multiple containers disappearing during an automated update, taking Portainer, remote-access, authentication, and tunnel components with them while persistent definitions remained. The useful HostSleuth question is not "how do I display Docker containers?"; it is "what disappeared together, which access path depended on it, and what evidence shows when that happened?"

Research signal:
- https://www.reddit.com/r/selfhosted/comments/1salkni/laugh_at_my_pain_and_learn_from_my_mistakes/

Community reports are workflow signals, not authoritative specifications.

## High-value differentiated opportunity lanes

### 1. Endpoint Path Story

Question answered:

> I typed this hostname/URL/host:port. Show me the actual path and the first place reality stops matching expectations.

Possible bounded evidence chain:

- resolver answer(s);
- selected destination address;
- kernel route/source interface;
- TCP result;
- TLS/STARTTLS result when applicable;
- local listener ownership when the endpoint is local;
- container port publication/network evidence;
- reverse-proxy/front-door evidence when explicitly configured or safely discoverable;
- current HTTP status/redirect destination when relevant;
- nearby retained service/listener/container/configuration changes.

This extends HostSleuth's existing Diagnose model rather than becoming network monitoring.

### 2. Reverse Proxy / Upstream Story

Question answered:

> The public URL gives me 404/502/redirect-loop/wrong service. Where is the break between front door and backend?

Potential evidence:

- DNS answer and destination;
- front-door listener/process/container;
- current HTTP redirect/status chain;
- expected Host/SNI identity;
- declared upstream host:port when available from an explicitly supported proxy integration;
- backend listener/container publication;
- backend reachability/TLS result;
- recent proxy/config/container/listener changes.

Caddy and Traefik already route traffic. HostSleuth should explain mismatches across the route, not configure the proxy.

### 3. DNS Reality Story

Question answered:

> Is this a DNS problem, a split-view problem, or is DNS correct and the failure is later in the path?

Potential evidence:

- system-resolver answer;
- configured resolver identity when safely obtainable;
- authoritative NS/delegation evidence for selected common records;
- A/AAAA/CNAME/MX/PTR/TXT evidence;
- optional comparison between system-resolver and authoritative answers;
- connection attempt tied to the address actually selected by the host;
- explicit wording when differences may be intentional split DNS rather than "wrong DNS."

Do not duplicate DNSViz's deep DNSSEC visualization; connect DNS evidence to the host/service path.

### 4. Port / Bind / Ownership Story

Question answered:

> Something should be on this port. What owns it, what address is it bound to, and what changed?

Potential evidence:

- expected address/port;
- current listener(s);
- PID/process/cgroup/service ownership where available;
- container publication and bind address;
- competing owner only when positively proven;
- route/interface relationship;
- retained listener/service/container changes;
- current endpoint reachability.

M7 already provides part of this for systemd services. A future generalized story should reuse that evidence rather than inventing a second implementation.

### 5. File / Deployment Story

Question answered:

> I copied/generated/updated a file that an application depends on. Is the right artifact actually in the right place with the right identity and metadata?

Potential evidence:

- source and destination hashes;
- owner/group/mode/mtime/size;
- symlink target where relevant and safe;
- associated service/container when explicitly declared;
- retained configuration fingerprint/change evidence;
- optional postcondition probe proving the consumer adopted the artifact.

This is broader than certificates: configuration fragments, keys/public material, web assets, database dumps, model files, binaries, and generated artifacts can all have "source vs destination vs consumer" failure modes. Private/secret content must not be returned merely because a file can be hashed.

### 6. Container Disappearance / Dependency Story

Question answered:

> Several services vanished or became unreachable together. What container/listener changes happened and which ingress path disappeared?

Potential evidence:

- containers removed/stopped/recreated in the incident window;
- image/name/network/port-publication changes already available from retained evidence;
- listener disappearance;
- associated reverse-proxy/tunnel/auth container changes;
- endpoint checks for a small explicit set of user-selected services;
- persistent definitions/configuration existence only where safely and explicitly supported.

Do not become Portainer. The purpose is incident correlation and recovery evidence, not container lifecycle management.

### 7. Mail Endpoint Story

Question answered:

> My mail service is "up," but which part of SMTP/IMAP/DNS/TLS identity is actually wrong?

Possible bounded read-only evidence:

- MX resolution;
- A/AAAA of the selected MX;
- PTR for the connecting/public address when explicitly supplied;
- SMTP banner/EHLO capability evidence;
- STARTTLS availability and served certificate on 25/587;
- implicit TLS certificate on 465/993/995 where applicable;
- local Postfix/Dovecot service/listener ownership on the host;
- SPF/DMARC TXT visibility as raw evidence without pretending to be a full deliverability scoring service;
- recent service/config/listener/certificate events.

This should diagnose protocol identity and host evidence, not become a mail server or reputation service.

### 8. Certificate Delivery Story

Question answered:

> The CA renewed my certificate. Where did it need to go, what consumed it, and what is each endpoint actually serving now?

Potential evidence chain:

- ACME/source certificate fingerprint and validity;
- declared/discovered destination certificate fingerprint;
- destination metadata without exposing private-key contents;
- associated service/container;
- endpoint protocol/host:port;
- served certificate fingerprint;
- deterministic source/destination/served match states.

This remains one opportunity lane, not the center of the product.

### 9. Boot / Recovery Story

Question answered:

> The machine came back. What did not?

This is the approved M10 direction and can reuse:

- boot/shutdown evidence;
- failed services;
- listeners that did not return;
- container changes;
- package/kernel/configuration changes around the boot;
- selected endpoint post-boot checks.

Do not claim why the reboot occurred without direct evidence.

### 10. Expected-State Contracts

Question answered:

> What should be true for this service, and which expectation is currently broken?

A small local contract might state:

- service should be active;
- expected listener/bind should exist;
- optional container should be running/publishing a port;
- hostname should resolve as expected;
- endpoint should be reachable;
- optional protocol/TLS/certificate expectation should hold;
- optional file fingerprint/destination should match.

The result is a deterministic checklist with change context, not continuous uptime monitoring by default.

## Safe-action research must also remain broad

M11 is the first approved milestone that may cross the read-only boundary, and it requires a separate design/security review. Certificate deployment is only one candidate family.

Other possible action families should be evaluated by the same standard, for example:

- reload one explicitly associated known service after a validated configuration/deployment change;
- re-run one bounded probe or validation after an operator changes something outside HostSleuth;
- apply a predefined local deployment recipe with exact source/destination/postcondition checks;
- restore one explicitly supported service/container state only if the operation can be narrowly modeled, previewed, confirmed, audited, and verified.

These are research candidates, not commitments. Generic start/stop/restart buttons, package management, firewall editing, file editing, Docker administration, or arbitrary commands remain out of scope unless the product direction is explicitly changed later.

## Product-shaping rule

A future feature should pass all three tests:

1. **Correlation:** does it connect evidence users currently gather from multiple commands/tools?
2. **Verification:** does it prove the final observable state instead of merely running an action?
3. **Boundedness:** can it be implemented without becoming a generic shell/admin/monitoring platform?

A fourth practical test is now useful:

4. **Consumer payoff:** does the result answer a common troubleshooting question faster or more clearly than opening another terminal tab or website?

If the feature is merely "an existing command with a button," it is probably not a HostSleuth feature.

## Roadmap fit without reordering

- **M7 Service Story — complete:** systemd/journal/listener/container/change correlation.
- **M8 Incident Lens — complete:** bounded non-causal historical context.
- **M9 Workbench — current branch:** bounded file, DNS, HTTP, and certificate identity tools; deliberately broader than certificates.
- **M10 Reboot Story — next:** boot/shutdown and what failed to return.
- **M11 Optional Safe Actions:** security/design review of a very small set of evidence-backed actions across multiple candidate workflow families.
- **Later roadmap review:** Endpoint Path Story, Reverse Proxy Story, DNS Reality Story, Mail Endpoint Story, Expected-State Contracts, Container Disappearance Story, and richer deployment verification can be scheduled only after explicit prioritization.

## Important security note

Many attractive troubleshooting features can accidentally create a remote probe, file oracle, credential leak, or administration console. HostSleuth should continue preferring loopback/local execution for sensitive Workbench-style operations, never expose secret file contents by default, avoid private-key/token/environment-variable collection, and never add arbitrary command execution as a shortcut.

## Current decision

Continue the owner-approved roadmap without deviation. Maintain broad consumer research in parallel, and deliberately assign a researched idea to a future milestone only when it strengthens HostSleuth's two core jobs and preserves the product/security boundary.
