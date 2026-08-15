# Threadkeeper Product Boundary v1

## Purpose

Threadkeeper Core remains the protected authority, integrity, provenance, replay and recovery kernel. The user-facing Threadkeeper product is a separate application layer.

The product exists to solve one primary problem:

> After an interruption, a user can immediately determine what is true about a project, why it is true, what is blocked, what should happen next, what an AI is authorised to continue, and where it must stop.

The product must not turn Threadkeeper Core into a general task manager or silently weaken Core's authority model.

## Repository boundary

### `threadkeeper-core`

Owns:

- immutable authority ledger and deterministic replay;
- canonical identity, schemas and reducer bindings;
- provenance and evidence primitives;
- actor/key/grant policy and cryptographic proofs;
- candidate preparation, quarantine, exact-head CAS and idempotency;
- recovery, integrity and assurance machinery;
- protocol-neutral interfaces required to use those capabilities safely.

Core changes slowly. Product work must not add Core primitives merely because they are convenient for one UI or workflow. A new Core primitive requires evidence that the application layer cannot safely express the required behaviour using existing contracts.

### `threadkeeper`

Owns:

- project model and product semantics;
- project/work reducer and derived actionable state;
- CLI and application service;
- `Now`, `Next`, `Blocked`, `Later` and completion workflows;
- goals and definition of done;
- project decisions, open questions, gates and evidence references;
- GitHub and later external integrations;
- AI continuation/execution orchestration;
- portfolio view;
- web UI;
- onboarding and eventual product operations.

Do not split CLI, API, web, AI or integrations into separate repositories until a concrete engineering constraint justifies it.

## Product v1 scope

### First-class records

The initial application model should support:

- `Project`
- `Goal`
- `Work`
- `Decision`
- `Question`
- `EvidenceReference`
- `Gate`
- `Authority`
- `Transition`

The application must define legal transitions and their preconditions rather than treating these as free-form task labels.

### Minimum actionable state

For one project, Threadkeeper must be able to derive:

- current project status;
- current goal / definition of done;
- `Now` — the currently executable action;
- `Next` — accepted work unlocked by completion of `Now`;
- `Blocked` — accepted work whose prerequisites are unsatisfied;
- `Later` — accepted work deliberately deferred;
- open questions requiring judgement or missing evidence;
- recent accepted decisions;
- authority available to an AI/client;
- stop conditions;
- meaningful transition history.

## Build order

### Phase 0 — close the Core v1 operational boundary

1. Reconcile `IMPLEMENTATION_STATUS.md` after PR #52 merge.
2. Run the accepted read-only production load proof and preserve exact evidence.
3. Establish a genuinely independent secondary custody boundary.
4. Obtain explicit owner authorisation before destructive production restore/replacement and prove exact recovery equivalence.
5. Separately review service activation while `AUTHORITY_WRITES_DISABLED` remains closed.
6. Only after all release gates pass consider any removal of `AUTHORITY_WRITES_DISABLED` as a separate explicit decision.
7. Freeze non-evidence-driven Core expansion during application dogfood.

### Phase 1 — project model

Define schemas and transition contracts for the initial application records.

Gate: a synthetic project can be reconstructed entirely from accepted history and invalid transitions fail closed.

### Phase 2 — deterministic project reducer

Build the application reducer that derives project status, Now, Next, Blocked, Later, questions, gates and authority from preserved records.

Gate: identical accepted history always produces identical project state.

### Phase 3 — minimal CLI

Initial commands should centre on:

```text
tk project show <project>
tk now <project>
```

Add only the mutation commands required to create/update project records through governed Core interfaces.

Gate: a user can operate a project without inspecting raw Core records.

### Phase 4 — dogfood on three dissimilar projects

Use:

1. Vaelinya — canon, candidates, language decisions and protected acceptance boundaries.
2. TWIS — evidence, uncertainty, provenance and publication gates.
3. System 55 Guide — physical-world dependencies, procurement and owner-only physical acceptance.

Primary validation gate: after a material interruption, Threadkeeper must correctly answer, without reconstructing old chat history:

- Where are we?
- Why?
- What is blocked?
- What happens next?
- What may AI continue without asking?
- Where must it stop?

Do not build the GUI until this gate is satisfied across all three projects.

### Phase 5 — minimal web interface

Only seven project areas are required initially:

- Overview
- State
- Work
- Decisions
- Evidence
- Authority
- History

Add one portfolio view showing actionable project state, not a global backlog of overdue tasks.

### Phase 6 — GitHub integration

Observe repositories, commits, branches, pull requests, issues, CI state and changed heads.

GitHub observations are evidence/source state. They never become authority merely because they were retrieved.

### Phase 7 — AI continuation engine

Execution loop:

```text
load authoritative state
→ derive Now
→ check preconditions
→ check authority
→ execute
→ preserve result/evidence
→ propose/accept transition when authorised
→ derive new state
→ continue
```

The loop stops on human judgement, insufficient evidence, destructive action, expenditure, publication, physical acceptance, changed/ambiguous state, or any other explicit authority boundary.

### Phase 8 — portfolio

Show which projects are ready to move, waiting on dependencies, or require owner action.

The portfolio optimises for actionability rather than task volume.

### Phase 9 — external beta

Use approximately five people with real multi-week projects.

Measure:

- time from return to meaningful work;
- accuracy of reconstructed project state;
- repeated-decision reduction;
- frequency of manual state repair;
- correctness of AI stop/continue behaviour;
- whether `Now` remains useful over time.

### Phase 10 — hardening and public v1

Only after beta evidence:

- authentication/account model;
- deployment and monitoring;
- secrets handling;
- automated backup/recovery;
- import/export;
- onboarding;
- accessibility;
- performance;
- security review.

## Explicit non-goals before dogfood passes

Do not prioritise:

- teams or enterprise permissions;
- billing;
- mobile applications;
- integrations marketplace;
- arbitrary third-party integrations;
- vector/semantic search as a dependency;
- social features;
- elaborate UI customisation;
- speculative Core assurance expansion unrelated to a demonstrated product need.

## Stop / continue decision

The product earns further investment only if Phase 4 demonstrates that it materially improves resumption and continuation of real projects.

If it cannot reliably preserve project state and produce a useful, authorised next action across Vaelinya, TWIS and System 55 Guide, stop product expansion and retain Core for other uses.
