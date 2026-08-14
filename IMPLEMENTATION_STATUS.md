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
- crash-durable quarantine publication candidate: private write + file sync + close, no-overwrite root-relative hard-link final publication, pinned-directory sync before success, independent durable convergence for identical visible final material, and fail-closed directory-sync errors;
- fresh bounded authoritative reconciliation for stale Prepare results and pre-CAS Accept candidate/quarantine failures that may race a concurrent identical acceptance, including the final bound `Ensure`→`Read` window, preserving durable `already_accepted` recovery instead of ordinary candidate/quarantine failure;
- per-Prepare private random staging identities prevent identical concurrent Prepare operations from sharing temporary cleanup ownership;
- caller-independent bounded recovery once `update-ref` may have changed authority, including explicit post-CAS verification/unknown-outcome classifications and bounded ref-lock contention settling when an update error is followed by an initial H0 observation;
- whole-tree `go test -race ./...` is part of the conformance workflow;
- adversarial repository-isolation regressions for alternates, `commondir`, partial clones/lazy fetch, symlinked authority stores, reftable/worktree config, hostile Git environments and non-regular JSON tree modes;
- explicit known-safe Git subprocess environment, including Windows case-insensitive environment isolation;
- concurrent replay/write/idempotency/cancellation safety tests and explicit overload signalling;
- protocol-neutral Ed25519 actor authentication with proofs bound to ledger, action, exact target, expected prior state and idempotency identity;
- exact actor + ledger + action + target authorisation grants with revoked-key and expiry handling;
- service-level authority-write admission that evaluates the hard release kill-switch before actor authentication/authorisation;
- explicit hard public/service gate: `AUTHORITY_WRITES_DISABLED`.

## Installed assurance/read capabilities

- genesis trust-root validation plus legacy-adoption validation contract;
- owner-selected fresh-Genesis deployment path: no legacy governance ledger/head will be fabricated;
- explicit threat model and single-authority-effect rule;
- content-addressed source escrow store plus preservation modes;
- exact-version filesystem source adapter with digest and symlink/traversal protection;
- stable source registry with immutable version identities;
- acyclic provenance graph with exact source/version lineage;
- typed relationship graph and durable conflict-set representation;
- evidence catalog/envelope with authority separate from retrieval score;
- bitemporal effective/knowledge time;
- coverage/completeness and bounded absence claims;
- confidentiality clearance and governed redaction tombstones;
- proposal/review bundle with alternatives, dissent and reopening conditions and zero authority effect;
- deterministic policy-impact simulation comparison;
- deterministic recovery-fork classification plus explicit operator-resolution candidate workflow that preserves rejected history;
- destructive non-empty bare-ledger backup/restore proof comparing exact head, replay and projection digests;
- verified derived replay checkpoint digests;
- Ed25519 external witness signing/verification;
- federated exact references with mandatory local authority disposition;
- deterministic portable Core export/import with canonical round-trip validation;
- Core build-provenance model; CI artifact/SBOM packaging remains pending because the current repository automation boundary did not permit workflow mutation in that lane;
- five-domain health model, incident lifecycle and key lifecycle;
- read-only reference CLI for genesis, evidence, review bundles, health and recovery proof/compare.

## Release governance already established

- repository `main` is protected by an active repository ruleset;
- deletion and non-fast-forward/force-push updates are blocked;
- pull requests are required;
- review conversations must be resolved;
- `test` and `windows-git-environment-isolation` are required status checks;
- required checks are strict against current `main`.

## Integration and protected work

- PR #11 was owner-authorised and merged to `main` as `38ea7c28f2b0f5c5ff0ca38b8da94eff17bfec5b`; an independent full Issue #9 PASS was **not** recorded before that merge. The exception remains explicit and does not enable public authority writes;
- the fresh-Genesis path is selected and recorded, but the production governance ledger and its first authoritative Genesis record do not exist yet; that is a deployment-evidence gate;
- Issue #21 **FAILED** `747f30b4e2af0109f592220aa03b43e1ca1f0543` on exact quarantine commit/path binding and post-CAS cancellation reporting; repaired in `bfe7686856ddec54c2be3e71aa8bc020d2b7a38e`;
- Issue #25 **FAILED** `bfe7686856ddec54c2be3e71aa8bc020d2b7a38e` on the identical-accept winner-cleanup race; repaired in `6710cb1b5f9d591f7e1653a5adc409581d34a858`;
- Issue #28 **FAILED** `6710cb1b5f9d591f7e1653a5adc409581d34a858` because matching final bytes could be observed while still creator-owned before file sync/close; private completed-file hard-link publication was merged as `a7214cbbbfb8d28732c5aff48eeb78bbe4103d52`;
- Issue #31 then **FAILED** that publication design because final-name visibility was not crash durability without a quarantine-directory sync, and required the whole-tree race gate to be robust to Git versions that reject hostile `refstorage` metadata earlier while still failing closed;
- Issue #32 **FAILED** merged `a7214cbbbfb8d28732c5aff48eeb78bbe4103d52` on two remaining concurrency windows: an identical Prepare could lose Q after bound `Ensure` but before bound `Read` and return ordinary `CANDIDATE_INVALID`, and ref-lock contention could make a losing identical CAS recover H0 too early and return an ordinary Git error before the winner published H1;
- open PR #34 consolidates the Issue #31 durability/race-gate repair with both Issue #32 concurrency repairs. Exact-head CI and internal hostile review are required before merge, and any merged commit still requires one fresh genuinely independent hostile review before this authority boundary can PASS;
- no public authority-write transport is enabled; actor authentication/authorisation remains a fail-closed admission prerequisite only;
- service-owned/non-writable-by-untrusted-users production ledger and quarantine filesystem ownership still require deployment proof;
- the final declared load/resource envelope still requires bounded memory/file-descriptor and overload/backpressure evidence;
- restore from an independently operated secondary backup location remains a deployment/recovery gate;
- broad existing durable event/config schemas have not yet been migrated to require temporal/coverage/confidentiality fields universally;
- checkpoint-accelerated replay has not replaced full replay; checkpoint build/verification is installed as an optimisation boundary only;
- external witness deployment/key service is optional and not configured;
- federation transport is not configured;
- Recall/search/vector storage remains deliberately separate and unimplemented;
- full fresh-install → Genesis → authenticate → write → restart → retry → conflict → restore → replay release acceptance remains outstanding.

See `docs/assurance/REMAINING_PROTECTED_GATES.md` for the release-critical sequence.

## Write status

`AUTHORITY_WRITES_DISABLED`

This remains true until production Genesis instantiation, deployment ownership, a fresh independent PASS on the final repaired quarantine/CAS boundary, final load/resource proof, independent secondary restore proof and end-to-end release acceptance are separately satisfied.
