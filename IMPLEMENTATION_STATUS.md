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
- fresh bounded authoritative recovery for pre-CAS candidate/quarantine failures that may race a concurrent identical acceptance, preserving exactly-one `accepted` / recovered `already_accepted` semantics;
- caller-independent bounded recovery once `update-ref` may have changed authority, with explicit post-CAS verification/unknown-outcome classifications;
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
- independent hostile review Issue #21 **FAILED** exact merged commit `747f30b4e2af0109f592220aa03b43e1ca1f0543` on two gate-blocking defects: quarantine did not bind the exact prepared commit/path, and caller cancellation after successful CAS could be reported as an ordinary failed write;
- those Issue #21 defects were repaired and merged as `bfe7686856ddec54c2be3e71aa8bc020d2b7a38e`;
- independent hostile re-review Issue #25 **FAILED** exact merged commit `bfe7686856ddec54c2be3e71aa8bc020d2b7a38e` on an exactly-once race: after two identical calls captured H0, the winner could accept H1 and clean Q before the loser read Q, causing the loser to return `CANDIDATE_INVALID` instead of `already_accepted`;
- the current repair fresh-replays authoritative state before returning pre-CAS candidate/quarantine failures and reconstructs exact concurrent acceptance as `already_accepted`. Because this again changes the reviewed authority boundary, its final exact merged commit requires a completely fresh independent hostile review; CI/self-review cannot convert the prior FAIL into PASS;
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
