# Threadkeeper Core

Threadkeeper Core is independent project-memory and governance infrastructure.

Its purpose is to preserve **where project truth comes from, how it relates, what state it is in, and what evidence supports it** without making any AI model the owner of that truth.

## Architectural invariant

> Removing or replacing every AI component must not destroy authoritative project information, change authority, or prevent Threadkeeper Core from operating and rebuilding its recall data.

A second invariant follows:

> Deleting the Threadkeeper Recall store must not delete accepted project truth, human decisions, or the evidence needed to reconstruct them.

## Current phase

**Contract definition only. No implementation technology is selected by this repository state.**

The first architecture package defines:

- `ARCHITECTURE.md` — system boundary and conceptual planes;
- `AUTHORITY_MODEL.md` — what may count as truth and how authority changes;
- `STORAGE_MODEL.md` — durable, rebuildable and transient storage classes;
- `INTERFACE_CONTRACT.md` — protocol-neutral client contract;
- `THREADKEEPER_STANDARD.md` — numbered conformance requirements;
- `docs/decisions/ADR-001-independent-from-ai.md` — the foundational architectural decision.

## System boundary

```text
Authoritative sources / authority sinks
        │
        │ immutable identity + provenance
        ▼
Threadkeeper Core
├── authority metadata
├── provenance
├── relationships
├── state projections
└── rebuildable recall
        │
        ▼
protocol-neutral interface boundary
        │
        ├── humans
        ├── CLI / tools
        ├── ChatGPT / Codex
        ├── local models
        └── future agents
```

AI systems are clients. They may retrieve evidence, reason over it and submit proposals. They do not become authoritative merely because they are confident, useful, automated or repeatedly consulted.

## Implementation rule

Implementation choices may begin only after the contract package is accepted. Storage engines, search engines, vector systems, API protocols, model providers and deployment topology are deliberately not selected here.
