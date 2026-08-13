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
- adversarial repository-isolation regressions for alternates, `commondir`, partial clones/lazy fetch, symlinked authority stores, reftable/worktree config, hostile Git environments and non-regular JSON tree modes;
- explicit known-safe Git subprocess environment, including Windows case-insensitive environment isolation;
- concurrent replay/write/idempotency/cancellation safety tests and explicit overload signalling;
- protocol-neutral Ed25519 actor authentication with proofs bound to ledger, action, exact target, expected prior state and idempotency identity;
- exact actor + ledger + action + target authorisation grants with revoked-key and expiry handling;
- service-level authority-write admission that evaluates the hard release kill-switch before actor authentication/authorisation;
- explicit hard public/service gate: `AUTHORITY_WRITES_DISABLED`.

## Installed assurance/read capabilities

- genesis trust-root validation plus legacy adoption contract;
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
- deterministic recovery-fork classification;
- destructive non-empty bare-ledger backup/restore proof comparing exact head, replay and projection digests;
- private candidate quarantine storage primitive;
- verified derived replay checkpoint digests;
- Ed25519 external witness signing/verification;
- federated exact references with mandatory local authority disposition;
- deterministic portable Core export/import with canonical round-trip validation;
- Core build-provenance model; CI artifact/SBOM packaging remains pending because the current repository automation boundary did not permit workflow mutation in this integration lane;
- five-domain health model, incident lifecycle and key lifecycle;
- read-only reference CLI for genesis, evidence, review bundles, health and recovery proof/compare.

## Integration and protected work

- PR #11 was owner-authorised and merged to `main` as `38ea7c28f2b0f5c5ff0ca38b8da94eff17bfec5b`; an independent full Issue #9 PASS was **not** recorded before that merge. The exception is explicit and does not enable public authority writes;
- PR #12 assurance expansion is integrated on `main`;
- the Genesis legacy-adoption validator is integrated, but actual Genesis adoption remains blocked until the real legacy governance ledger and exact pre-adoption head are identified;
- before any public authority writer is enabled, the merged CAS boundary still requires fresh independent hostile review;
- no public authority-write transport is enabled; actor authentication/authorisation is installed only as a fail-closed admission prerequisite;
- quarantine is not yet wired into Git candidate materialisation because that would change the CAS boundary and requires a separate review;
- broad existing durable event/config schemas have not yet been migrated to require temporal/coverage/confidentiality fields universally;
- checkpoint-accelerated replay has not replaced full replay; checkpoint build/verification is installed as an optimisation boundary only;
- external witness deployment/key service is optional and not configured;
- recovery-fork operator resolution and restore from an independently operated secondary backup location remain deployment gates;
- federation transport is not configured;
- Recall/search/vector storage remains deliberately separate and unimplemented;
- repository `main` protection and required conformance checks remain a release-governance gate until proven enabled.

See `docs/assurance/REMAINING_PROTECTED_GATES.md` for the exact sequence.

## Write status

`AUTHORITY_WRITES_DISABLED`

This remains true until the Genesis evidence boundary, deployment ownership, quarantine integration, fresh CAS review, load safety, independent recovery proof and release-governance gates are separately accepted.
