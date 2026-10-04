# Threadkeeper Core v1 Read-Only Release Boundary

**Status:** normative protected-main release boundary. Core v1 RC1 has completed protected promotion through Gate I with authority writes still hard-disabled; final Gate J dispositions remain to be recorded.

## Purpose

The Threadkeeper Core v1 release boundary is:

> **Threadkeeper Core v1 read-only production authority kernel: deterministic validation, replay, recovery, restore verification and production load assurance, with authority writes hard-disabled.**

This document governed the protected Core v1 RC1 replacement now accepted through Gate I. It does not authorize any later binary replacement, service activation, public transport, ledger mutation or authority-write enablement; each remains a separate protected transition.

## Current accepted production profile

The accepted production Core is now:

- release: `threadkeeper-core-v1-rc1`;
- source: `84c3e983768f67b518c6c84f2eb62f0bf4babae7`;
- binary SHA-256: `0bcfc7afd0632fcbdb62bc421c7ca80ee51a67e2e91dd78562ae0659d0c83bac`;
- active path: `/usr/local/bin/threadkeeper-core`;
- ledger/ref: `/var/lib/threadkeeper-core/authority/ledger.git` / `refs/heads/main`;
- Genesis/head: `73fa0e66df2ae80b4b2a04247112470f6bb8e451`;
- actor-policy root: `803e61858fe1dfae96b357845bed1b10644a5028801d307533fcf312d8b4a40a`;
- replay SHA-256: `6316bde6bf6f2caa0bc33f9cd495c3bf222c35956c8403e55c890818f74fea12`;
- canonical RecoveryProof SHA-256: `3806b7c8a94d9521a991927781ead2ca78ab4ec12b66cbceb641086e9e66cad4`;
- authority writes: disabled;
- direct Core service: not applicable by accepted CLI/library-only architecture.

Gate I promotion and post-promotion evidence are preserved in Issue #94. The prior accepted binary SHA-256 `7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37` remains retained as rollback material; its earlier operational acceptance evidence remains preserved in Issue #51.

## In scope

A future Core v1 read-only release MUST preserve/prove:

- strict JSON parsing, including duplicate-member rejection;
- RFC 8785 canonicalisation and SHA-256 content identity;
- local-only schema validation;
- Fresh Genesis validation/bootstrap semantics;
- ledger-derived actor-policy identity;
- hardened local bare-Git authority boundary;
- exact-head/linear history and semantic-path integrity;
- deterministic ledger replay and governed-record projection;
- source-file provenance and exact-version validation;
- evidence/proposal/health/portable read-validation surfaces;
- RecoveryProof generation and strict proof input semantics;
- independent-secondary restore verification;
- load/resource proof machinery;
- quarantine/CAS correctness and hostile regression preservation even while writes remain disabled;
- build provenance and exact source/binary identity;
- hard `AUTHORITY_WRITES_DISABLED` behaviour before authority-changing admission work.

## Out of scope

These do not block Core v1 read-only release unless a later reviewed dependency proves otherwise:

- Recall/search/vector storage;
- GUI;
- Manager replacement;
- long-running Core daemon/listener;
- HTTP/MCP/gRPC public Core transport;
- authority-write enablement;
- optional external witness deployment;
- optional federation transport;
- checkpoint-accelerated replay while disabled;
- Policy Pack product rollout;
- Specialist Pack rollout;
- broader Threadkeeper product-v1 work.

MCP interoperability and Policy Pack contracts may remain in the source repository as non-blocking architecture/contracts. Their presence does not widen the Core production release boundary.

## Resolved next-release correctness prerequisites

Before any new Core binary may be called a read-only release candidate, the #76 and #77 correctness contracts must be present in its exact protected lineage.

Current protected-main lineage satisfies both:

1. **Issue #76 — strict standalone RecoveryProof comparison: PASS / MERGED**
   - PR #79 merged as `f1cb0a8309cd397227f767d786d1c80c2cd073d7`;
   - incomplete/malformed proofs fail before equivalence comparison;
   - valid identical proofs still pass;
   - tampered valid proofs still fail.

2. **Issue #77 — authoritative-ledger namespace policy: PASS / MERGED**
   - PR #80 selected and enforced the closed-world Core v1 ledger namespace;
   - merged as `deb2cd42a69f61f30c2588b9cea18c75088aed97`;
   - unknown committed paths fail deterministically, including paths later deleted and mixed valid/unknown commits.

3. **Repository status must be internally consistent**
   - `IMPLEMENTATION_STATUS.md`, this boundary, the reconciliation record, release checklist and protected-gate status must not describe #76/#77 as unresolved current blockers;
   - historical findings may remain only when clearly marked as superseded/resolved.

Fresh post-merge conformance run `37065581798` on `deb2cd42a69f61f30c2588b9cea18c75088aed97` passed both required jobs. No production binary, ledger, product observer profile, service or authority-write state changed.

## Code-side candidate gates

The exact candidate SHA MUST pass, without waiver:

- protected-main PR process;
- `test`;
- `windows-git-environment-isolation`;
- complete Go test tree and race coverage where required by the repository conformance workflow;
- strict JSON/canonical/digest/schema tests;
- Git repository/environment isolation hostile tests;
- Genesis/actor-policy/replay tests;
- quarantine/CAS/idempotency/stale-head/post-CAS regressions;
- recovery and restore-verification tests;
- load/resource proof reference tests;
- the #76 regression;
- the #77 regression/contract tests selected by the accepted policy;
- hard write-kill-switch proof.

A passing CI/reference envelope is not a production-capacity claim.

## Build identity

A release candidate MUST be identified by the tuple:

- exact protected source SHA;
- binary SHA-256;
- version/release name;
- Go/toolchain identity;
- target OS/architecture;
- dependency versions;
- build flags/provenance;
- exact required-check evidence.

Future release names SHOULD use an immutable convention such as:

`threadkeeper-core-v1-rcN`

and MUST NOT rely on the ambiguous label `production-candidate` as the sole release identity.

## Branch strategy

Use an exact protected-main SHA as the release source unless a separately documented reason requires a release branch.

No direct production update may be triggered merely by merging source changes.

## Production-candidate gates

After code-side acceptance, before replacing the current production binary:

1. build the exact frozen candidate;
2. preserve the current accepted binary/profile as rollback material;
3. test both old and candidate binaries against the same **copy** of the production ledger;
4. require equal authority-relevant identity:
   - Genesis;
   - actor-policy root;
   - authoritative head;
   - governed projection;
   - replay identity;
   - valid RecoveryProof;
5. run the accepted production-shaped load envelope on the candidate in a safe production-shaped context;
6. prove independent-secondary restore/equivalence for the candidate where the candidate changes recovery interpretation;
7. update the downstream `armpitpete/threadkeeper` Core observation profile through protected review;
8. install the candidate side-by-side on IntoVPS;
9. run live read-only equivalence checks;
10. require `authority-write` to remain disabled;
11. only then seek the exact protected promotion authorization.

## Downstream compatibility

The current Threadkeeper product intentionally hard-binds the accepted Core source/binary/ledger/ref/proof profile.

A new Core production identity MUST NOT be promoted while the product observer still accepts only the old profile.

The product-side change must remain fail-closed and must not widen the observer command allowlist merely to accommodate a new Core version.

## Rollback

Until post-promotion acceptance completes:

- retain the accepted `46f476fd...` binary or exact recoverable artifact;
- retain its SHA/build evidence;
- retain the pre-change RecoveryProof and production authority evidence;
- retain a verified independent authority-ledger backup.

Rollback does not authorize ledger-history rewriting.

## Service boundary

A Core read-only release does not imply service activation.

Accepted Core v1 architecture is CLI/library-only. Creating a long-running Core service or transport requires a separate architecture/release decision.

## Write boundary

A Core read-only release does not imply authority writes.

`AUTHORITY_WRITES_DISABLED` remains a hard release invariant.

Removing or weakening it requires a separate exact release decision with its own authentication/authorization, transport, recovery, operational and owner-authority gates.

## Evidence form

Every release checklist item MUST preserve:

- exact SHA/artifact identity;
- exact check/command;
- expected result;
- actual result;
- evidence/log/artifact reference;
- PASS/FAIL/N/A disposition;
- reason for any N/A.

“Tested”, “looks good”, or an issue state without exact evidence is insufficient.

## Cross-references

- current reconciliation: `docs/CORE_V1_RECONCILIATION_2026-10-02.md`;
- implementation status: `IMPLEMENTATION_STATUS.md`;
- durable storage: `DURABLE_STORAGE_ARCHITECTURE.md`;
- Fresh Genesis: `docs/operations/FRESH_GENESIS_DEPLOYMENT_V1.md`;
- actor authentication/authority: `docs/assurance/ACTOR_AUTH_V1.md`;
- quarantine/CAS: `docs/assurance/CANDIDATE_QUARANTINE_V1.md`;
- load proof: `docs/operations/LOAD_RESOURCE_PROOF_V1.md`;
- E2E conformance: `docs/conformance/CORE_V1_E2E_ACCEPTANCE_V1.md`;
- independent restore: `docs/operations/INDEPENDENT_SECONDARY_RESTORE_V1.md`;
- accepted product roadmap: `armpitpete/threadkeeper#214`;
- product Core observation contract: `armpitpete/threadkeeper/docs/CORE_OBSERVATION_RECEIPT_V0_1.md`.

## Frozen stop boundaries

This release boundary authorizes none of the following by itself:

- production binary replacement;
- production-ledger mutation;
- product observer profile change;
- service activation;
- public transport;
- authority-write enablement.

Each remains a later exact protected transition.
