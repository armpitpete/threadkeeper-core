# Threadkeeper Product v1 Roadmap

## Product thesis

Threadkeeper is a project continuity system.

Its first job is not generic task management. Its first job is to let a person return to a long-running project after an interruption and immediately know:

1. what is currently true;
2. what has already been decided;
3. what remains uncertain;
4. what is blocked;
5. what can be done now;
6. what comes next;
7. what the AI may continue doing without asking;
8. where human authority is required.

Threadkeeper Core remains the authority/integrity kernel. The product layer should live in a separate `armpitpete/threadkeeper` repository.

## v1 scope

### In scope

- single-user projects;
- project goals and definitions of done;
- authoritative current state;
- work states: Now, Next, Blocked, Later, Done and Candidate;
- decisions;
- open questions;
- evidence references;
- gates and stop conditions;
- authority presentation;
- immutable project history derived from Core;
- minimal CLI;
- dogfood on three real projects;
- minimal web UI;
- GitHub integration;
- AI continuation loop;
- portfolio view;
- export/recovery sufficient for a stable single-user v1.

### Explicitly out of scope until the core user journey is proven

- teams and enterprise permissions;
- billing;
- mobile applications;
- integrations marketplace;
- social features;
- elaborate workflow customization;
- vector search as an authority source;
- broad SaaS scaling work;
- speculative Core expansion.

## Delivery phases

### Phase 0 — Core boundary and release housekeeping

- freeze unnecessary Core expansion;
- reconcile implementation status after the merged Core v1 E2E harness;
- keep remaining production operational gates separate;
- maintain `AUTHORITY_WRITES_DISABLED` until a separately reviewed decision.

Exit: product work can proceed without treating unfinished production write-enable work as a product dependency.

### Phase 1 — Project model

Define first-class application records for:

- Project;
- Goal;
- Work;
- Decision;
- Question;
- EvidenceReference;
- Gate;
- AuthorityBoundary;
- Transition.

Define legal transitions and invariants before implementation.

Exit: synthetic project history can express goal, state, work, decisions, blockers and gates without ambiguity.

### Phase 2 — Project reducer

Implement deterministic derivation of user-facing project state from accepted history.

Required output includes:

- project status;
- current goal/definition of done;
- Now;
- Next;
- Blocked;
- Later;
- unresolved questions;
- current gates;
- current authority boundary;
- recent meaningful transitions.

Exit: identical accepted history always yields identical project state.

### Phase 3 — Minimal CLI

Provide the smallest useful interface, including equivalents of:

- `tk project show <project>`;
- `tk now <project>`;
- add/accept decision;
- add/complete/block work;
- record evidence reference;
- show authority/gates;
- show history.

Exit: a user can manage a project without reading raw Core records.

### Phase 4 — Three-project dogfood

Use three real projects with different failure modes:

- Vaelinya — long-running creative/canon state;
- TWIS — evidence/provenance/uncertainty/publication gates;
- System 55 Guide — physical-world dependencies, suppliers and owner-only physical acceptance.

Primary test:

> After a meaningful interruption, can the user open each project and correctly determine where it stands and what to do next without reconstructing old chats?

Exit criteria:

- current state correct;
- decisions not needlessly re-litigated;
- blockers visible;
- `Now` is actionable;
- `Next` is justified;
- human-only gates are correctly exposed;
- no hidden dependence on conversation memory.

This is the first kill/continue gate for the product.

### Phase 5 — Minimal web interface

Initial sections:

- Overview;
- State;
- Work;
- Decisions;
- Evidence;
- Authority;
- History.

Add a simple portfolio page only after single-project views work.

Exit: ordinary use no longer requires the CLI.

### Phase 6 — GitHub integration

Observe repository state such as commits, PRs, issues, checks and merges as evidence.

GitHub observations MUST NOT silently become project authority merely because they exist externally.

Exit: software projects can update their observable state without manual transcription while preserving explicit acceptance semantics.

### Phase 7 — AI continuation engine

Implement the controlled loop:

1. load accepted project state;
2. identify executable Now action;
3. verify preconditions;
4. verify authority;
5. execute or prepare the action;
6. capture evidence;
7. propose/accept the resulting transition according to authority;
8. recalculate state;
9. continue until a protected stop condition.

Mandatory stop classes include:

- owner/human judgement;
- insufficient evidence;
- destructive operation;
- spending money;
- publication;
- physical acceptance;
- unexpected state change;
- missing or ambiguous authority.

Exit: `continue <project>` performs useful bounded work and stops correctly.

### Phase 8 — Portfolio

Show projects primarily by actionability:

- Ready to move;
- Waiting/Blocked;
- Owner required;
- Recently changed.

Do not optimize the portfolio around overdue-task pressure.

Exit: the user can see where effort can produce progress across projects.

### Phase 9 — External beta

Start with approximately five people who have real multi-week projects.

Measure:

- time from return to meaningful work;
- correctness of reconstructed state;
- repeated-decision rate;
- manual state-repair rate;
- trust in decisions/history;
- correctness of AI stop behavior;
- whether `Now` remains genuinely useful.

Exit: evidence that the system helps users other than its creator.

### Phase 10 — Product hardening

Only after beta evidence:

- authentication/account boundaries;
- deployment/monitoring;
- encrypted secrets;
- backup automation;
- import/export;
- onboarding;
- accessibility;
- performance;
- error recovery;
- security review.

Exit: stable single-user public v1.

## Success metric

Primary product metric:

> Time from returning to an interrupted project to beginning the correct meaningful next action.

Secondary metrics:

- percentage of returns where `Now` is correct without repair;
- number of previously settled decisions reopened unnecessarily;
- number of hidden blockers discovered only after attempting work;
- number of AI actions that reach a human-only gate without crossing it;
- manual project-state repair frequency.

## Product kill rule

Do not assume the product deserves completion merely because Core is strong.

At the end of Phase 4, continue only if dogfood evidence shows that Threadkeeper materially reduces project-reconstruction cost and reliably identifies a useful next action.

If it does not, preserve Core and stop or redesign the product layer rather than expanding scope.

## Recommended repository split

### `threadkeeper-core`

Stable protected kernel. Changes slowly.

### `threadkeeper`

Product/application repository. Changes quickly and owns project semantics, CLI, UI, integrations and AI orchestration.

Do not create additional `threadkeeper-web`, `threadkeeper-api`, `threadkeeper-cli` or similar repositories until a concrete technical boundary requires it.
