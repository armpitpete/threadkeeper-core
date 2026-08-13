# Threadkeeper Core Assurance Standard v0.2

**Status:** candidate until merged to the repository default branch.  
**Relationship:** extends `THREADKEEPER_STANDARD.md`; it does not weaken any v0.1 requirement.

## TK-ASSURE-001 — Rooted genesis
Every new authority ledger MUST have one inspectable genesis identity and initial authority root. A legacy ledger may adopt this only through an explicit migration; history MUST NOT be rewritten to fabricate an earlier genesis record.

## TK-ASSURE-002 — Declared threat boundary
Each releasable security boundary MUST state its protected assets, in-scope adversaries, deployment assumptions and explicit non-claims. Findings are evaluated against that declared boundary; expanding the boundary requires a reviewed revision.

## TK-ASSURE-003 — Evidence preservation policy
Every source class MUST declare a preservation mode. A source reference whose evidence may disappear MUST expose that preservation limitation. Escrow MUST NOT promote authority.

## TK-ASSURE-004 — Bitemporal knowledge
Where effective time differs materially from observation/acceptance time, Core MUST preserve both dimensions and MUST NOT rewrite knowledge history when later evidence changes what is believed about an earlier date.

## TK-ASSURE-005 — Coverage-aware absence
Core MUST NOT turn a retrieval miss into an absence claim unless the relevant source domain is explicitly complete enough for that bounded conclusion.

## TK-ASSURE-006 — Confidentiality and retention
Content classification is independent from authority. Exports, backups, escrow and federation MUST preserve access/retention constraints. Required content destruction MUST leave only the minimum governed tombstone permitted by policy.

## TK-ASSURE-007 — Decision context
Where policy requires it, decisions MUST preserve serious alternatives, attributable dissent, unresolved uncertainty and reopening conditions. Acceptance MUST NOT erase rejected reasoning.

## TK-ASSURE-008 — Fork-aware recovery
Divergent valid restored histories MUST enter `RECOVERY_FORK`; timestamp or convenience MUST NOT choose a winner. Resolution is a new governed act and the rejected fork remains evidence.

## TK-ASSURE-009 — Candidate quarantine
Before public authority writes are enabled, non-authoritative candidate material MUST have a bounded private lifecycle and MUST NOT become durable authority except through the reviewed acceptance path.

## TK-ASSURE-010 — Policy simulation
Governance policy changes SHOULD expose deterministic impact simulation before acceptance. Simulation is derived and MUST NOT itself change authority.

## TK-ASSURE-011 — Verified checkpoints
Replay checkpoints MAY accelerate reconstruction only when cryptographically bound to exact ledger/projection/schema/binding state. Invalid checkpoints fall back to full replay and MUST NOT alter authority.

## TK-ASSURE-012 — External witness
An optional witness MAY attest to exact historical state. Witness evidence is integrity attestation, not project authority.

## TK-ASSURE-013 — Federation
Cross-project references MUST preserve exact source identity and MUST require an explicit local authority disposition. Authority from another project is never imported implicitly.

## TK-ASSURE-014 — Core supply-chain identity
Releasable Core binaries MUST be attributable to exact source/build identity and binary digest, with dependency/SBOM evidence where supported.

## TK-ASSURE-015 — Single authority effect
Every mechanism MUST declare one authority effect. Composition MUST NOT create authority implicitly. Preservation, retrieval, simulation, checkpoints, federation and witnesses are not authority transitions.

## TK-ASSURE-016 — Independent health domains
System, knowledge, authority, source and recovery health MUST remain separately inspectable. Process health MUST NOT mask stale evidence or failed recovery proof.

## TK-ASSURE-017 — Incident and key lifecycle
Security/governance incidents and cryptographic key lifecycle MUST have explicit state transitions. Compromise MUST NOT be repaired by silently reclassifying earlier evidence as trustworthy.

## TK-ASSURE-018 — Human review bundle
A decision review surface MUST be able to present the proposal, exact evidence envelopes, material conflicts, serious alternatives/dissent/reopening conditions and simulated consequences without making review itself an authority transition.

## TK-ASSURE-019 — Load preserves semantics
Concurrency and overload MUST NOT change exact-head, idempotency, evidence isolation, checkpoint equivalence or fail-closed semantics.

## TK-ASSURE-020 — Self-description
A release MUST NOT claim an implementation state contradicted by its executable/tested capabilities. Capability-status documentation is part of conformance evidence.

## Extended conformance gate
A production authority writer MUST NOT be enabled until, in addition to the v0.1 gate: destructive restore equivalence passes; the CAS boundary has an independent PASS; the deployment threat assumptions are proven; candidate quarantine is integrated/reviewed; actor authentication/authorisation is accepted; load-safety tests pass; and repository/release governance protects the exact accepted source and artifact identity.
