# HostSleuth — Consumer Opportunity Landscape

Status: research input only. This document does not change the approved roadmap or authorize implementation.

Last researched: 2026-09-16

## Research question

What do administrators already get from monitoring products, Linux admin consoles, ACME clients, TLS scanners, reverse proxies, scripts, and one-off websites — and where could HostSleuth provide something materially more useful instead of becoming another flavor of an existing tool?

HostSleuth's differentiator remains deterministic correlation: connect local host evidence, change history, service ownership, listener state, certificate state, and what an endpoint is actually presenting into one understandable story.

## What already exists and should not simply be copied

### Linux administration consoles

Cockpit already provides a browser interface for systemd services and journal logs, including service-filtered journal views and privileged service control through systemd/PolicyKit. HostSleuth should not become a second Cockpit or generic service control panel.

Research:
- https://docs.cockpit-project.org/cockpit-guide/latest/guide/api-system.html
- https://docs.cockpit-project.org/cockpit-guide/main/guide/feature-systemd.html

### ACME clients and renewal hooks

Certbot already supports automated renewal plus pre/post/deploy hooks. A successful deploy hook can run after certificate issuance/renewal. Executable renewal hook directories are also supported.

Research:
- https://eff-certbot.readthedocs.io/en/latest/using.html
- https://eff-certbot.readthedocs.io/en/stable/man/certbot.html

acme.sh already supports issuing, renewing, copying a certificate to application-specific paths, and an arbitrary reload command. Its documentation explicitly warns that renewal alone is not enough if the service is not reloaded and therefore continues serving the previous certificate.

Research:
- https://github.com/acmesh-official/acme.sh

lego likewise supports pre, deploy, and post hooks around issuance/renewal.

Research:
- https://go-acme.github.io/lego/references/ref-flags/
- https://go-acme.github.io/lego/index.print.html

Conclusion: HostSleuth should not win by inventing yet another generic hook mechanism. Existing clients already do that well.

### Automatic-TLS servers and reverse proxies

Caddy automatically obtains and renews certificates for the endpoints it owns. Nginx Proxy Manager wraps Certbot and DNS provider plugins in a management UI. These products are strongest when they directly terminate TLS.

Research:
- https://caddyserver.com/docs/automatic-https
- https://nginxproxymanager.com/certbot/

Conclusion: HostSleuth should not require users to move TLS termination into HostSleuth and should not become a reverse proxy.

### TLS scanners

testssl.sh can test TLS/SSL on arbitrary ports and supports STARTTLS protocols such as SMTP, POP3, and IMAP. It is an excellent deep protocol/security scanner.

Research:
- https://github.com/testssl/testssl.sh

Conclusion: HostSleuth should not try to replace a full cipher/vulnerability scanner. Its value is correlating protocol/certificate observations with the local service, listener, deployment, and retained host-change evidence.

## Repeated workflow pain visible in real use

Representative self-hosted discussions repeatedly describe the same operational gap:

1. a certificate is renewed on one machine;
2. one or more applications need a copied certificate rather than the ACME client's live path;
3. one or more services or containers must reload;
4. administrators use rsync/scp/scripts/hooks/systemd path units to distribute the certificate;
5. there is no single deterministic answer proving every consumer received the new certificate and is actually serving it.

Representative discussions:
- https://www.reddit.com/r/selfhosted/comments/gp575d/restart_services_after_letsencrypt_certificate/
- https://www.reddit.com/r/selfhosted/comments/1kf89kg/automating_tls_certificate_updates_across_multiple/
- https://www.reddit.com/r/selfhosted/comments/187h4rj/does_acmesh_remember_how_i_deployed_certificates/

These discussions are workflow signals, not authoritative technical specifications.

## High-value differentiated opportunities

### 1. Certificate Delivery Story

Question answered:

> The CA renewed my certificate. Where did that certificate need to go, what consumed it, and what is each endpoint actually serving now?

Potential evidence chain:

- ACME lineage/source certificate fingerprint and validity;
- declared or discovered destination certificate file fingerprint;
- destination metadata such as existence, owner/group/mode and mtime without exposing private-key contents;
- associated service/container identity;
- reload/restart expectation;
- endpoint protocol and host:port;
- certificate fingerprint actually served by that endpoint;
- deterministic result: source matches destination / destination differs / served matches destination / served is stale / evidence unavailable.

This is stronger than a hook because it verifies the postcondition instead of assuming that a copy or reload command worked.

### 2. Protocol-aware endpoint certificate inspection

Immediate TLS is not enough for common mail services. A future bounded protocol layer could explicitly support STARTTLS negotiation for common protocols such as:

- SMTP 25/587;
- IMAP 143;
- POP3 110;
- LDAP 389 if later justified.

For mail servers this would let HostSleuth answer whether Postfix/Dovecot-facing endpoints present the certificate expected from the local source, without pretending every TLS-capable service speaks HTTPS or immediate TLS.

This should remain certificate/endpoint evidence, not a full mail diagnostic suite.

### 3. Expected Endpoint Contract

Allow an administrator to state a small deterministic expectation such as:

- `postfix.service` should own/listen on TCP/25 and TCP/587;
- `dovecot.service` should own/listen on TCP/993;
- `mail.example.com:587` should negotiate STARTTLS;
- the served certificate should match a named local lineage/destination;
- the service should be active and the endpoint reachable.

HostSleuth could then explain which expectation drifted and when related listener/service/configuration evidence changed. This is an evidence contract, not uptime monitoring or alerting by default.

### 4. Renewal/Deployment Readiness Story

Before any future write action exists, HostSleuth can eventually make the existing automation understandable:

- ACME client present and renewal mechanism visible;
- renewal timer/cron state when safely discoverable;
- deploy hook present or absent;
- destination mapping known or unknown;
- destination certificate matches source or not;
- associated service known or unknown;
- served fingerprint matches destination or not.

For third-party ACME clients, discovery must avoid leaking DNS API credentials, tokens, account private material, or arbitrary environment values.

### 5. Bounded Certificate Deployment Recipe — M11-or-later candidate

This is the natural safe-action evolution of the Certificate Delivery Story, but it crosses the read-only boundary and therefore belongs only behind the M11 design/security gate or a later explicitly approved milestone.

A HostSleuth recipe should be declarative and narrow rather than an arbitrary shell hook:

- choose one known certificate source;
- choose one explicit destination set;
- preview exact files/services/endpoints affected;
- preserve safe ownership/mode semantics rather than inventing permissions;
- perform atomic copy/update where feasible;
- reload only a specifically associated known service when explicitly allowed;
- record an audit event;
- re-probe the real endpoint after the operation;
- report success only when the served fingerprint/postcondition is verified.

No free-form command box, no hidden shell execution, and no generic file manager.

### 6. Certificate Rollout Consistency

For one certificate used by several local services/endpoints, show a compact consistency story:

- source fingerprint;
- each destination fingerprint;
- each served fingerprint;
- which endpoints are current;
- which endpoint is stale or unavailable.

This is especially useful for mail stacks where SMTP submission, IMAPS, POP3S, webmail/reverse proxy, and management interfaces may consume copies of the same certificate differently.

### 7. Explain an ACME failure with host evidence

ACME clients can report errors such as an address already being in use, but the administrator still has to determine who owns the port or why the challenge path cannot work. HostSleuth can correlate the ACME-facing symptom with the same listener/process/container/firewall/service evidence it already collects.

This should explain the obstruction; it should not automatically stop the conflicting service during read-only milestones.

## Product-shaping rule from this research

A future feature should pass all three tests:

1. **Correlation:** does it connect evidence users currently gather from multiple commands/tools?
2. **Verification:** does it prove the final observable state instead of merely running an action?
3. **Boundedness:** can it be implemented without becoming a generic shell/admin/monitoring platform?

If the answer is only "another UI for a command that already exists," it is probably not a HostSleuth feature.

## Roadmap fit without reordering

- **M7 Service Story:** continue exactly as approved; service/journal/listener/container/change correlation only.
- **M8 Incident Lens:** makes the timing around service/certificate/deployment changes easier to understand; no actions.
- **M9 Workbench:** a natural home for read-only STARTTLS/PEM/file-vs-served inspection if explicitly approved when M9 is designed.
- **M10 Reboot Story:** unchanged.
- **M11 Optional Safe Actions:** design/security review can evaluate a bounded certificate deployment recipe, Certbot dry-run/renewal, and tightly scoped reload/postcondition verification.
- **Later:** expected endpoint contracts, rollout consistency, or broader certificate-delivery stories can be scheduled only after roadmap review and demand evidence.

## Important Let's Encrypt account note

Let's Encrypt is primarily consumed through the ACME protocol and local ACME account credentials maintained by clients such as Certbot; it is not normally an OAuth-style web account that HostSleuth would simply "link" to like a SaaS dashboard. HostSleuth can potentially discover safe ACME account/CA metadata or integrate with an ACME client later, but storing or handling ACME account keys, DNS API credentials, or private certificate keys requires an explicit security design and should not be added casually.

## Current decision

Do not alter M7 implementation scope because of this research. Keep the approved roadmap order. Preserve these opportunities as design input for M8/M9/M11 and later roadmap review, where each can be evaluated against real use and HostSleuth's security boundary.
