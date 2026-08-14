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

## Authoritative actor-policy candidate under review

Issue #41 / PR #42 removes the remaining ambient authentication trust source:

- canonical digest-bound initial actor policy at `config/authority/actor-policy/root.json`;
- Genesis `initial_authorities` must exactly match actors granted by the root policy;
- only canonical raw-base64 Ed25519 public keys are trusted; duplicate/unsorted/malformed policy material fails closed;
- direct later root-policy mutation/removal/rename is rejected during replay;
- current policy is derived from the exact authoritative ledger snapshot, not runtime caller configuration;
- governed target `authority:actor-policy` / record kind `core.actor-auth-policy-v1` supports policy creation/rotation through existing exclusive reducer semantics;
- actor-policy values are semantically validated before candidate acceptance/CAS;
- governed policy revocation fails closed and never falls back to Genesis;
- supported Fresh Genesis bootstrap requires the root actor policy plus the event-schema/reducer-binding machinery needed to rotate/revoke it;
- exported service admission no longer accepts a caller-supplied `actorauth.Policy`;
- service admission requires request `ledger_id` and `expected_state` to match the exact policy-loading snapshot before Ed25519 authentication/authorization;
- `AUTHORITY_WRITES_DISABLED` remains the first service check.

This lane does not select real production keys/grants and does not enable public writes.

## Installed assurance/read capabilities

- genesis trust-root validation plus legacy-adoption validation contract;
- owner-selected Fresh Genesis deployment path;
- explicit threat model and single-authority-effect rule;
- content-addressed source escrow and exact-version source adapter;
- source registry, provenance graph, relationships/conflicts and evidence catalog;
- bitemporal time, coverage/completeness and confidentiality/redaction models;
- proposal/review bundles and deterministic policy simulation;
- recovery-fork classification and explicit operator-resolution workflow;
- destructive bare-ledger backup/restore proof comparing Genesis identity, head, replay and projection;
- verified derived replay checkpoints;
- Ed25519 external witness primitive;
- federated references with local authority disposition;
- deterministic portable export/import;
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

After PR #42 is accepted:

1. instantiate the actual production Fresh Genesis ledger with the approved real actor public keys/grants and prove ledger/quarantine filesystem ownership/non-user-writability on that deployment;
2. declare and prove the final load/resource envelope, including bounded memory/file-descriptor behavior and explicit overload/backpressure;
3. perform destructive restore from an independently operated secondary backup location and prove exact Genesis/head/replay/projection equivalence;
4. run full production-shaped end-to-end acceptance: fresh install → Genesis → authoritative policy → authenticate → write → restart → retry → conflict → independent restore → replay;
5. only then consider removing `AUTHORITY_WRITES_DISABLED` through a separately reviewed release decision.

No public authority-write transport is currently enabled.

Optional/non-v1 integrations remain external witness deployment, federation transport, checkpoint-accelerated replay, Recall/search/vector storage and GUI unless separately selected.

## Write status

`AUTHORITY_WRITES_DISABLED`
