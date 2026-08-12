# Threadkeeper Core

Threadkeeper Core is independent project-knowledge and governance infrastructure.

Its purpose is to preserve **what a project knows, what it has accepted, where that knowledge came from, what conflicts with it, how understanding changed through time, and enough durable evidence to reconstruct and challenge those conclusions without making an AI model the owner of project truth**.

## Invariants

- removing every AI component must not destroy or change authority;
- deleting Recall must not delete accepted truth, decisions or required provenance;
- retrieval relevance never promotes evidential authority;
- current state is a deterministic projection of preserved history;
- protected writes fail closed on stale or ambiguous state;
- preservation, integrity witnesses, checkpoints, simulation and federation do not become authority implicitly.

## Current implementation

Threadkeeper Core is implemented in Go with strict JSON validation, RFC 8785 canonicalisation, SHA-256 record identity, local-only JSON Schema, a hardened Git governance ledger, deterministic replay, reducer/policy bindings, and an internal exact-head candidate/CAS writer.

The internal writer remains behind the hard public gate:

`AUTHORITY_WRITES_DISABLED`

The CAS repair lane is independently reviewed before merge/enablement. The assurance layer adds explicit genesis, threat boundaries, source escrow policy, bitemporal time, coverage/completeness, confidentiality/retention, dissent/reopening context, fork recovery, candidate quarantine primitives, policy simulation, replay checkpoints, external witness verification, federation, build provenance and operability models.

## Planes

```text
Authoritative sources / sinks
        │
        ▼
Threadkeeper Core
├── exact identities + provenance
├── authority policy + decisions
├── conflict / temporal / coverage state
├── deterministic projections
├── recovery + integrity evidence
└── federated references
        │
        ├── Recall (disposable)
        └── protocol-neutral clients / AI
```

AI systems are clients. They may retrieve evidence, derive material and submit proposals. They receive no implicit privilege and cannot promote their own output to authority.

See `ARCHITECTURE.md`, `THREADKEEPER_STANDARD.md`, `docs/assurance/CORE_ASSURANCE_PROGRAMME_V0.1.md`, `docs/conformance/`, and `docs/decisions/` for the normative architecture and gates.
