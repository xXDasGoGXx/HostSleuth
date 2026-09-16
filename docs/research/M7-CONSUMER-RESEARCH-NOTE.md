# Consumer research note — M7 boundary

This note exists only to prevent current market/user research from silently expanding M7.

The owner asked HostSleuth to keep the approved roadmap order while continuously studying what people currently assemble from monitoring products, admin consoles, certificate clients, TLS scanners, scripts, and single-purpose websites, then use that research to identify genuinely differentiated later workflows.

M7 remains complete as the read-only Service Story milestone. No research idea in this note is active implementation scope.

The research backlog currently points toward later investigation of:

- protocol-aware TLS/STARTTLS inspection for mail and other non-HTTPS services;
- source certificate -> destination certificate -> actually served certificate verification;
- explicit certificate deployment recipes with preview, bounded known-service reload, audit evidence, and postcondition re-probe rather than arbitrary shell hooks;
- endpoint/certificate rollout consistency checks;
- expected-endpoint contracts tying service/listener/protocol/certificate expectations to retained change evidence.

Any filesystem write, certificate deployment, service reload, ACME-account interaction, or private-key handling remains outside the read-only milestones and requires the M11 design/security boundary or a later explicitly approved milestone.

A broader sourced market/product research document should remain separate from implementation PRs so research can influence prioritization without becoming accidental scope.
