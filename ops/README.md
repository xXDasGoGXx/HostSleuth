# Operations Registry Convention

The `ops/` directory is the machine-readable part of HostSleuth's recovery documentation.

The goal is simple: a future administrator should not need a chat transcript or personal memory to understand how the project is assembled.

## Files

### `project.yaml`

Canonical machine-readable inventory for:

- project identity and lifecycle;
- current stable release;
- source-of-truth files;
- distribution endpoints;
- automation/workflows;
- credential names and storage locations;
- deployment mode;
- recovery repository;
- security and documentation rules.

It stores **references to credentials, never credential values**.

## Human documentation

Machine-readable inventory is not a substitute for explanation.

The paired human runbook is:

- `docs/OPERATIONS.md`

The exact current state is:

- `CURRENT-HANDOFF.md`

Historical implementation and release records are:

- `docs/history/`

## Reuse this convention

For other repositories, use the same pattern:

```text
project/
├── CURRENT-HANDOFF.md
├── docs/
│   └── OPERATIONS.md
└── ops/
    ├── README.md
    └── project.yaml
```

Each project should be understandable on its own.

A future central private registry can then index the `ops/project.yaml` files from multiple repositories without becoming the only copy of critical knowledge.

## Required rule

An integration is not operationally complete until the repository records:

1. what it does;
2. why it exists;
3. where configuration lives;
4. what triggers it;
5. which named credentials it uses;
6. what it depends on;
7. how to validate it;
8. how to roll it back or recreate it;
9. what sensitive information must never be committed.

This is documentation-as-infrastructure, not optional cleanup.
