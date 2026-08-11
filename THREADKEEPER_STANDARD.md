# Threadkeeper Core Contract Standard v0.1

**Status:** Normative only when present on the repository default branch; unmerged copies are candidates  
**Scope:** Authority, storage and interface semantics  
**Technology selection:** Not yet authorised by this standard

Normative terms **MUST**, **MUST NOT**, **SHOULD** and **MAY** are used deliberately.

## A. Authority

### TK-AUTH-001 — Explicit authority
Every governed record MUST expose an authority class determined by declared policy, not by model confidence or retrieval rank.

**Test:** retrieve representative records and verify authority class plus policy basis are inspectable.

### TK-AUTH-002 — Minimum classes
The implementation MUST distinguish `AUTHORITATIVE`, `DERIVED`, `ADVISORY` and `EPHEMERAL` semantics or provide an exactly equivalent mapping.

**Test:** construct one record of each class and prove clients cannot confuse them through the interface.

### TK-AUTH-003 — No automatic AI promotion
AI-generated or AI-transformed material MUST NOT become authoritative solely because an AI produced, repeated, ranked or endorsed it.

**Test:** submit generated material through every AI-facing write path and verify it remains non-authoritative absent a permitted decision transition.

### TK-AUTH-004 — Explicit decision event
Any governed transition to accepted/authoritative state that requires approval MUST have an attributable authorised decision event.

**Test:** attempt the transition without the required actor/mechanism and verify failure.

### TK-AUTH-005 — Exact source version
Authority claims MUST be traceable to an immutable source version and/or content digest.

**Test:** move a mutable reference after ingestion and prove the earlier evidence still resolves to its original exact version.

### TK-AUTH-006 — History preservation
Supersession, revocation or replacement MUST preserve enough history to explain prior authoritative state.

**Test:** supersede a record and retrieve both prior and current states with the transition between them.

### TK-AUTH-007 — Conflict preservation
Materially conflicting authoritative records MUST remain inspectable until policy or an authorised decision resolves them.

**Test:** ingest conflicting authoritative records and verify neither silently overwrites the other.

### TK-AUTH-008 — Uncertainty states
The system MUST be able to represent unknown, stale, incomplete, conflicting, unverified and proposed-but-unaccepted states.

**Test:** create each condition and verify the interface does not collapse it into accepted truth.

### TK-AUTH-009 — Current is a projection
A `current` or equivalent view MUST be traceable to the exact events/versions from which it is computed.

**Test:** request current state and inspect supporting history.

### TK-AUTH-010 — Authority policy is governed
Changes to authority policy MUST themselves be versioned and attributable.

**Test:** change policy and prove both previous and replacement policy versions remain identifiable.

## B. Provenance

### TK-PROV-001 — Derived lineage
Every derived record MUST identify the source versions used to create it.

**Test:** inspect an arbitrary derived record and follow its lineage to source versions.

### TK-PROV-002 — Transformation identity
Derived records MUST identify the transformation/tool version or equivalent reproducibility identity.

**Test:** show which transformation produced a selected derived record.

### TK-PROV-003 — Producer identity
Generated or transformed records MUST identify their producing actor/tool class.

**Test:** distinguish human, deterministic tool and AI-produced outputs in provenance.

### TK-PROV-004 — Observation/inference separation
The system MUST NOT represent an inference or summary as though it were a direct source observation.

**Test:** retrieve both source text and a derived summary and verify record type/provenance differs.

### TK-PROV-005 — Evidence survives presentation
Convenience UIs MAY simplify display, but exact provenance MUST remain retrievable through the underlying interface.

**Test:** move from a displayed result to its complete evidence envelope.

## C. Storage

### TK-STOR-001 — AI storage independence
Threadkeeper Core data MUST be independently manageable from AI model/runtime storage.

**Test:** remove the AI/model environment and prove Core durable data remains intact.

### TK-STOR-002 — Recall is disposable
Complete deletion of Recall MUST NOT delete accepted project truth, decision history or required provenance.

**Test:** destructive Recall rebuild test.

### TK-STOR-003 — Rebuild without AI
Core MUST recover authoritative state and non-model-dependent provenance without requiring an AI model.

**Test:** perform rebuild with all AI/model services unavailable.

### TK-STOR-004 — Durable authority sink
A governance-significant decision MUST NOT exist solely in Recall, transient storage or an AI conversation.

**Test:** accept a governed proposal, delete Recall and sessions, then recover the decision from its authoritative sink.

### TK-STOR-005 — Versioned durable configuration
Source registry, authority policy and other configuration required for reconstruction MUST be durable and versioned outside disposable Recall.

**Test:** rebuild using the preserved configuration set.

### TK-STOR-006 — Stable logical identity
Changing physical storage or rebuilding indexes MUST NOT change the logical identity or authority of source-backed records.

**Test:** migrate/rebuild storage and compare stable record/source identities.

### TK-STOR-007 — Incomplete write detection
The implementation MUST detect incomplete governed writes and MUST NOT expose a partially committed authority transition as valid state.

**Test:** fault-inject a write between stages and verify fail-closed recovery.

### TK-STOR-008 — Secrets separation
Operational secrets MUST NOT be stored in Recall, generated prompts, logs intended as project records, or versioned repository content.

**Test:** scan representative persisted stores/exports for configured test credentials.

### TK-STOR-009 — Portable durable export
Durable configuration, authority metadata, provenance and relationship definitions MUST have a deterministic export path independent of a specific storage engine's internal files.

**Test:** export and validate logical equivalence after import into a clean environment.

### TK-STOR-010 — Backup does not create authority
Restored copies MUST retain original authority/provenance rather than gaining authority merely by being a backup.

**Test:** restore a copy and inspect preserved source/version identity.

## D. Interface

### TK-IF-001 — Protocol neutrality
Core semantics MUST NOT depend on a particular client protocol.

**Test:** contract tests operate against a logical adapter boundary rather than transport-specific meanings.

### TK-IF-002 — Evidence envelope
Knowledge reads MUST make authority, source version, provenance, conflict/supersession state and projection identity obtainable.

**Test:** retrieve a governed result and validate required evidence fields.

### TK-IF-003 — Relevance is not authority
Retrieval/search score MUST be represented separately from evidential authority.

**Test:** high-rank non-authoritative material remains visibly non-authoritative.

### TK-IF-004 — Stable record retrieval
Clients MUST be able to retrieve a known record by stable logical identity.

**Test:** retrieve the same record before and after Recall rebuild.

### TK-IF-005 — Historical inspection
Clients MUST be able to inspect superseded and historical governed states where retained by policy.

**Test:** retrieve prior accepted state after a later transition.

### TK-IF-006 — Conflict enumeration
Clients MUST be able to discover material conflicts affecting a governed result.

**Test:** query an object with conflicting authority and enumerate both sides.

### TK-IF-007 — Proposal is non-authoritative
Proposal submission MUST NOT itself change accepted project truth.

**Test:** submit proposal and verify current authoritative projection remains unchanged.

### TK-IF-008 — Expected-state protection
Protected writes MUST use an expected-state/version precondition or equivalent stale-write protection.

**Test:** change target after proposal preparation and verify the stale write is rejected.

### TK-IF-009 — Replay protection
Authority-changing writes MUST support idempotent retry or equivalent duplicate-decision protection.

**Test:** repeat identical decision request and verify one logical decision event.

### TK-IF-010 — Fail closed
Protected writes MUST fail on unresolved authority, target ambiguity, stale state, required-provenance failure, authorisation failure or unavailable authoritative persistence.

**Test:** exercise each failure condition and verify no successful state transition is exposed.

### TK-IF-011 — Inspectable failure
A protected-operation failure MUST be machine-inspectable and MUST NOT be hidden behind successful-looking natural-language output.

**Test:** trigger a stale write and verify explicit error state reaches the client.

### TK-IF-012 — Contract version discovery
Clients MUST be able to discover the active semantic contract/schema version.

**Test:** connect a client and retrieve contract version before governed operations.

## E. AI boundary

### TK-AI-001 — AI is a client
No model provider, model runtime or AI session is required to own Threadkeeper authority or durable memory.

**Test:** remove all AI components and execute source inspection, authority inspection and Recall rebuild.

### TK-AI-002 — No inferred human acceptance
AI clients MUST NOT create a human-required decision event by claiming, inferring or paraphrasing approval.

**Test:** submit text resembling approval through an AI path without authenticated decision authority and verify no transition.

### TK-AI-003 — Replaceable model-derived data
Model-derived indexes or metadata MUST identify the producing model/transformation sufficiently to invalidate or rebuild them when that producer changes.

**Test:** replace a model-derived transformation and selectively mark/rebuild affected derived records without changing source authority.

### TK-AI-004 — Degraded no-model operation
Loss of AI functionality MAY reduce semantic retrieval quality but MUST NOT prevent exact source retrieval, authority inspection, provenance inspection or governance-state recovery.

**Test:** disable AI and perform those operations.

## F. Operations and recovery

### TK-OPS-001 — Rebuild procedure
A documented, executable procedure MUST rebuild Recall from authoritative sources and durable Core configuration.

**Test:** run it in a clean environment.

### TK-OPS-002 — Rebuild verification
Rebuild completion MUST include integrity checks comparing source identities/digests and governed-state expectations.

**Test:** corrupt a rebuild input and verify integrity checks detect divergence.

### TK-OPS-003 — Schema/contract migration history
Breaking changes to the meaning of authority, provenance, decisions or relationships MUST use an explicit version/migration path.

**Test:** attempt to open incompatible durable data and verify the system refuses silent reinterpretation.

### TK-OPS-004 — Technology replacement test
Replacing a storage engine, search engine, model provider or transport MUST NOT require redefining authority semantics.

**Test:** architecture review/contract suite demonstrates semantic equivalence across adapters.

## G. Technology-selection gate

No implementation technology is selected by this standard.

A future technology proposal MUST be evaluated against these requirements. Any candidate that requires one of the following is disqualified unless the standard itself is explicitly revised:

- AI/model runtime as the sole memory or authority store;
- non-rebuildable Recall;
- silent authority promotion;
- inability to preserve exact provenance;
- destructive current-state replacement without history;
- inability to reject stale protected writes;
- coupling the meaning of authority to one storage engine or protocol.

## H. Core conformance gate

A release MUST NOT claim Threadkeeper Core conformance until the contract suite proves at minimum:

1. AI-off operation;
2. full Recall deletion and rebuild;
3. exact-version provenance;
4. conflict preservation;
5. proposal/decision separation;
6. stale-write rejection;
7. decision replay protection;
8. durable recovery of accepted state.
