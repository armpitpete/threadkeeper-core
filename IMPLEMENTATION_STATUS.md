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

Installed code-side machinery:

- immutable root Genesis at `config/genesis/root.json` validated during every replay;
- exact root schema-set and reducer-policy binding to Genesis;
- Genesis-rooted replay and recovery identity;
- create-only isolated Fresh Genesis bare-ledger bootstrap;
- parent-path, Git-template, direct-ref and restart verification;
- `fresh-genesis-init` machine-readable evidence;
- production deployment runbook.

Production Fresh Genesis is now instantiated and Issue #45 is closed PASS:

- production ledger: `/var/lib/threadkeeper-core/authority/ledger.git`;
- Genesis commit: `73fa0e66df2ae80b4b2a04247112470f6bb8e451`;
- replay SHA-256: `6316bde6bf6f2caa0bc33f9cd495c3bf222c35956c8403e55c890818f74fea12`;
- independent reopen/replay + FSCK passed;
- `/var/lib/threadkeeper-core` is `root:root` `0755`;
- `/var/lib/threadkeeper-core/authority` is `threadkeeper-core:threadkeeper-core` `0700`;
- no service activation or post-Genesis authority mutation is authorised.

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

The production Genesis roots the accepted production actor/key identity and minimum grant selected in Issue #45. Private signing-key material remains outside GitHub, repository, CI and the production server.

## Load/resource proof machinery installed

Issue #43 / PR #44 merged as protected `main` `46f476fd4e0a346e45034310c423f6c1cd592f65` after exact-head conformance and hostile self-review.

Installed:

- strict explicit load/resource envelopes with no silent defaults;
- sampled peak and settled Go heap/goroutine/process descriptor-or-handle evidence;
- Linux `/proc/self/fd` and Windows `GetProcessHandleCount` process metrics;
- a repository/CI reference envelope explicitly separated from production sizing;
- concurrent full `ProveRecovery` replay requiring exact baseline RecoveryProof equivalence;
- restored-copy replay-under-load equivalence proof;
- synchronized limiter burst proof with explicit overload rather than unbounded queueing;
- 128-way concurrent proof that `AUTHORITY_WRITES_DISABLED` dominates before ledger access;
- `ledger-load-proof` CLI for production-envelope measurement;
- load-safety evidence matrix and production runbook;
- checkpoint acceleration clarified as conditional/optional: full replay remains authoritative while acceleration is disabled.

The initial production envelope `threadkeeper-core-production-initial-v1` is accepted in Issue #51: 4 concurrent workers × 25 iterations, 128 MiB/64 MiB peak/settled heap-growth ceilings, 32/8 goroutine-growth ceilings, 64/16 open-handle-growth ceilings, with the open-handle metric mandatory. The read-only production measurement itself remains OPEN; acceptance of the envelope is not a load PASS.

## Restore-verification machinery installed

Issue #46 / PR #49 squash-merged as protected `main` `a51a6ccfdecc64797bdd263fa9bd9fc5f2d15b71` from exact reviewed head `c164628feeba5e9ef937a2fbeaf84102066efa40` after exact-head conformance #239 and hostile self-review.

Installed:

- strict canonical/digest-bound secondary provenance declarations;
- exact original RecoveryProof parsing and semantic hashing;
- restored-ledger verification through the hardened Reader;
- exact Genesis, actor-policy, head, replay and governed-projection RecoveryProof equality;
- machine-readable restore reports separating Core equivalence from fixed `requires_external_review` operational-independence status;
- read-only `recovery-restore-verify` CLI;
- hostile regressions for malformed/contradictory provenance and altered authority state;
- operational runbook requiring real external custody/provider/operator evidence.

This machinery does not itself prove that a backup is genuinely independently operated. Issue #51 keeps the destructive production restore behind a new explicit owner authorization and external evidence review.

## Core v1 E2E acceptance candidate under review

Issue #50 / PR #52 is the final code-side/reference E2E lane. The disposable test sequence combines Fresh Genesis, ledger-derived actor policy, exact Ed25519 authentication, hard service-gate rejection, real quarantine/CAS acceptance, a stale competing H0 candidate with no rebase, restart/retry/idempotency conflict, and the merged restore-verification path.

The candidate emits machine-readable `CORE_V1_E2E_ACCEPTANCE` evidence and ends with `authority_writes_enabled: false`. Its local temporary backup/restore is implementation evidence only, not independent-secondary operational evidence.

## Installed assurance/read capabilities

- genesis trust-root validation plus legacy-adoption validation contract;
- owner-selected Fresh Genesis deployment path;
- explicit threat model and single-authority-effect rule;
- content-addressed source escrow and exact-version source adapter;
- source registry, provenance graph, relationships/conflicts and evidence catalog;
- bitemporal time, coverage/completeness and confidentiality/redaction models;
- proposal/review bundles and deterministic policy simulation;
- recovery-fork classification and explicit operator-resolution workflow;
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

The final consolidated quarantine/CAS repair merged as `fde19f4c03a1915f7d26da493593566a6017bc49`. Independent hostile re-review Issue #36 **PASSED** that exact commit, including independently constructed real-Git CAS/ref-lock attacks. That correctness gate is closed subject to the separately proved production filesystem-ownership boundary.

## Remaining protected release work

1. complete review/integration of Issue #50 / PR #52 E2E machinery;
2. run the accepted `threadkeeper-core-production-initial-v1` envelope through the read-only production `ledger-load-proof` and preserve its exact evidence under Issue #51;
3. select and evidence a genuinely independent secondary custody boundary, then obtain new explicit authorization before any destructive production restore/replacement and prove exact recovery equivalence under Issue #51;
4. only after the production operational gates pass, prepare and separately review service activation while `AUTHORITY_WRITES_DISABLED` remains closed;
5. only after all release gates pass consider removal of `AUTHORITY_WRITES_DISABLED` through a separate explicit decision and review.

No public authority-write transport or long-running production service is currently enabled.

Optional/non-v1 integrations remain external witness deployment, federation transport, checkpoint-accelerated replay, Recall/search/vector storage and GUI unless separately selected.

## Write status

`AUTHORITY_WRITES_DISABLED`
