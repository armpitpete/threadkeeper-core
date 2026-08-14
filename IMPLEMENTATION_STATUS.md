# Implementation Status

## Installed authority kernel

Implemented and exercised by conformance tests:

- Go CLI/service skeleton and machine-readable build metadata;
- strict raw JSON validation, duplicate-member/UTF-8/negative-zero rejection;
- RFC 8785 canonicalisation and SHA-256 content identity;
- local-only JSON Schema Draft 2020-12 registry;
- hardened bare Git ledger opening, repository safety checks and strict fsck;
- linear history inspection and immutable semantic JSON tree rules;
- Git ledger replay with event/schema/policy validation;
- exclusive governed-record reducer and accepted reducer bindings;
- internal candidate-write construction, exact-head Git compare-and-swap, durable idempotency and post-CAS replay verification;
- candidate quarantine integrated into preparation and acceptance: pinned private store, exact-byte storage, prepared H0/H1/path capability binding, raw digest/size verification, 24-hour retention, acceptance cleanup and crash/retry cleanup;
- crash-durable quarantine publication with file + directory durability and fail-closed recovery;
- fresh bounded reconciliation for stale Prepare/Accept observations, including bound `Ensure`→`Read` cleanup races;
- caller-independent post-`update-ref` recovery and bounded ref-lock contention settling;
- whole-tree `go test -race ./...` conformance;
- hardened Git repository/environment isolation including Windows case-insensitive variants;
- explicit overload signalling and concurrency safety tests;
- protocol-neutral Ed25519 proofs bound to actor/key, ledger, action, exact target, expected state and idempotency identity;
- exact actor + ledger + action + target grants, revoked-key handling and bounded proof lifetime;
- service-level authority-write admission with the hard release kill-switch evaluated before policy/authentication machinery;
- explicit hard public/service gate: `AUTHORITY_WRITES_DISABLED`.

## Fresh Genesis authority installed

Merged PR #40 / Issue #37 as `69b0c3b5f51c9891a78a623621bb64159b9672de` after exact-head conformance #165 and hostile self-review.

Installed:

- immutable root Genesis at `config/genesis/root.json` validated during every replay;
- exact root schema-set and reducer-policy binding to Genesis;
- Genesis-rooted replay and recovery identity;
- create-only isolated Fresh Genesis bare-ledger bootstrap;
- parent-path, Git-template, direct-ref and restart verification;
- `fresh-genesis-init` machine-readable evidence;
- production deployment runbook.

The real production Genesis ledger has **not** yet been instantiated.

## Authoritative actor policy installed

Issue #41 / PR #42 merged as `f4ea4d7a7ab286446ca560a67619c181605fc189` after exact-head conformance #204 and hostile self-review.

Installed:

- canonical digest-bound actor policy at `config/authority/actor-policy/root.json` whose bytes declare exact `ledger_id` and `authority_policy_version`;
- exact Genesis `initial_authorities` reconciliation;
- canonical raw-base64 Ed25519 public keys and strict duplicate/ordering/active-key validation;
- immutable root-policy path and replay/recovery binding to its version/content digest;
- current policy derived from the exact authoritative replay snapshot rather than runtime caller configuration;
- governed target `authority:actor-policy` / record kind `core.actor-auth-policy-v1` for policy creation/replacement/revocation;
- wrong-ledger/wrong-policy-version governed policy values rejected before CAS;
- terminal policy revocation with no Genesis fallback;
- exported service admission with no caller-supplied `actorauth.Policy` path;
- exact request ledger/head binding before Ed25519 authentication/authorization;
- `AUTHORITY_WRITES_DISABLED` still evaluated before ledger/policy/authentication work.

Real production actor IDs/public keys/grants remain deployment inputs and have not been selected by the code lane.

## Load/resource proof candidate under review

Issue #43 / PR #44 adds the remaining code-side load/resource proof machinery without claiming production capacity:

- strict explicit load/resource envelopes with no silent defaults;
- sampled peak and settled Go heap/goroutine/process descriptor-or-handle evidence;
- Linux `/proc/self/fd` and Windows `GetProcessHandleCount` process metrics;
- a repository/CI reference envelope explicitly separated from production sizing;
- concurrent full `ProveRecovery` replay requiring exact baseline RecoveryProof equivalence;
- restored-copy replay-under-load equivalence proof;
- synchronized limiter burst proof with explicit overload rather than unbounded queueing;
- 128-way concurrent proof that `AUTHORITY_WRITES_DISABLED` dominates before ledger access;
- `ledger-load-proof` CLI for later production-envelope measurement;
- load-safety evidence matrix and production runbook;
- checkpoint acceleration clarified as conditional/optional: full replay remains authoritative while acceleration is disabled.

Passing this lane will close **load-proof machinery only**. The actual production load/resource envelope remains open until the real production-shaped deployment supplies and passes its reviewed envelope.

## Installed assurance/read capabilities

- genesis trust-root validation plus legacy-adoption validation contract;
- owner-selected Fresh Genesis deployment path;
- explicit threat model and single-authority-effect rule;
- content-addressed source escrow and exact-version source adapter;
- source registry, provenance graph, relationships/conflicts and evidence catalog;
- bitemporal time, coverage/completeness and confidentiality/redaction models;
- proposal/review bundles and deterministic policy simulation;
- recovery fork classification and explicit operator-resolution workflow;
- destructive non-empty bare-ledger backup/restore proof comparing Genesis identity, head, replay and projection;
- verified derived replay checkpoints;
- Ed25519 external witness primitive;
- federated references with local authority disposition;
- deterministic portable Core export/import;
- build provenance model, health, incident and key-lifecycle models;
- read-only reference CLI for assurance/recovery inspection.

## Release governance already established

- protected `main` ruleset blocks deletion and force/non-fast-forward updates;
- pull requests are required;
- review conversations must be resolved;
- strict required checks are `test` and `windows-git-environment-isolation`.

Historical governance remains explicit: PR #11 was owner-authorised without a genuinely independent full Issue #9 PASS.

## Independent quarantine/CAS review status

The final consolidated quarantine/CAS repair merged as `fde19f4c03a1915f7d26da493593566a6017bc49`. Independent hostile re-review Issue #36 **PASSED** that exact commit, including independently constructed real-Git CAS/ref-lock attacks. That correctness gate is closed subject to the separate production filesystem-ownership assumption.

## Remaining protected release work

1. instantiate the actual production Fresh Genesis ledger with approved real actor public keys/grants and prove ledger/quarantine filesystem ownership/non-user-writability on that deployment;
2. after Issue #43 is accepted, declare and measure the **actual production** load/resource envelope using `ledger-load-proof` on that production-shaped target;
3. perform destructive restore from an independently operated secondary backup location and prove exact Genesis/actor-policy/head/replay/projection equivalence;
4. run full production-shaped end-to-end acceptance: fresh install → Genesis → authoritative policy → authenticate → write → restart → retry → conflict → independent restore → replay;
5. only then consider removing `AUTHORITY_WRITES_DISABLED` through a separately reviewed release decision.

No public authority-write transport is currently enabled.

Optional/non-v1 integrations remain external witness deployment, federation transport, checkpoint-accelerated replay, Recall/search/vector storage and GUI unless separately selected.

## Write status

`AUTHORITY_WRITES_DISABLED`
