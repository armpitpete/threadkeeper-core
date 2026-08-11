# ADR-001: Threadkeeper Core Is Independent From AI

- **Status:** Proposed
- **Date:** 2026-08-11
- **Decision scope:** Foundational architecture

## Context

Threadkeeper is intended to provide durable project memory, authority tracking, provenance, relationships and retrieval across long-lived work.

AI systems are useful for reasoning, summarisation, semantic retrieval and proposing actions, but they are unsuitable as the sole owner of project truth because:

- model/runtime providers can change;
- model state and context are not durable authority records;
- generated material may be wrong while sounding confident;
- an AI session can disappear without project history disappearing;
- semantic indexes may need to be rebuilt when models change;
- human acceptance and governed decisions must remain independently auditable.

A design in which Threadkeeper's meaning depends on one model would turn project governance into a side effect of an AI product.

## Decision

Threadkeeper Core will be architecturally independent from AI.

AI systems are **clients and optional derived-data producers**, not owners of authority, accepted project state or indispensable durable memory.

The architecture therefore separates:

1. **Authoritative sources/sinks** — preserve governed project truth and decision records.
2. **Core durable configuration** — preserves how sources, authority and relationships are interpreted.
3. **Recall** — disposable, rebuildable retrieval/index data.
4. **AI/model environments** — replaceable clients and optional transformation providers.

Deleting or replacing every AI component must not change authority or prevent Core recovery.

Deleting the complete Recall store must not destroy accepted truth, decisions or provenance required to reconstruct governed state.

## Consequences

### Positive

- Model/provider replacement does not rewrite project truth.
- AI experimentation can be aggressive without making experimentation authoritative.
- Recall indexes can be rebuilt or changed independently.
- Human decisions remain inspectable outside chat history.
- Non-AI tooling can consume the same project evidence.
- Recovery has a clear test: rebuild without AI.

### Costs

- Authority and provenance require explicit modelling.
- Governed writes need stronger state/version checks than ordinary note-taking systems.
- Some data must be duplicated as derived indexes while retaining source references.
- Technology selection must prioritise recoverability and auditability over convenience.

These costs are accepted because they are necessary to prevent AI convenience from becoming hidden authority.

## Rejected alternatives

### Alternative A — AI conversation/model memory is the project memory

Rejected because it couples truth to a replaceable model/session and cannot provide a sufficiently inspectable authority boundary.

### Alternative B — One combined database for AI state and Threadkeeper state

Rejected as the default architecture because lifecycle operations on model/runtime storage could affect governance data and because the boundary becomes difficult to prove.

Physical co-location may later be permitted only if logical separation, independent backup/restore and destructive-isolation tests prove the same invariants.

### Alternative C — Treat the search/vector store as authoritative

Rejected because retrieval structures are derived and should be replaceable. Search rank is not evidential authority.

### Alternative D — Allow AI to auto-accept when confidence is high

Rejected because confidence is not authority and because explicit human/governance boundaries would become probabilistic.

## Reopening conditions

This decision may be reconsidered only if a proposed architecture can prove all of the following without weakening the contract:

- accepted truth survives model/runtime removal;
- human-required decisions remain independently attributable;
- exact provenance remains inspectable;
- Recall remains rebuildable;
- storage/model replacement cannot silently change authority;
- stale or ambiguous protected writes still fail closed.

Convenience, model quality or lower implementation effort alone are not sufficient reopening conditions.

## Validation

ADR-001 is implemented only when the conformance suite can demonstrate:

1. all AI/model services disabled;
2. exact source and authority inspection still works;
3. Recall completely deleted;
4. Recall rebuilt from authoritative sources plus durable Core configuration;
5. accepted state, decision history and provenance recovered unchanged.
