# Threadkeeper Core v1 Read-Only Release Boundary

**Status:** normative protected-main release boundary. Correctness gates #76/#77 satisfied on 2026-10-02; release-candidate deployment has not begun.

## Purpose

The next Threadkeeper Core release boundary is:

> **Threadkeeper Core v1 read-only production authority kernel: deterministic validation, replay, recovery, restore verification and production load assurance, with authority writes hard-disabled.**

This document governs a **future Core binary replacement**. It does not invalidate the currently accepted production profile and does not itself authorize deployment.

## Current accepted production profile

The accepted production Core remains:

- source: `46f476fd4e0a346e45034310c423f6c1cd592f65`;
- binary SHA-256: `7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37`;
- ledger/ref: `/var/lib/threadkeeper-core/authority/ledger.git` / `refs/heads/main`;
- Genesis/head: `73fa0e66df2ae80b4b2a04247112470f6bb8e451`;
- replay SHA-256: `6316bde6bf6f2caa0bc33f9cd495c3bf222c35956c8403e55c890818f74fea12`;
- authority writes: disabled;
- direct Core service: not applicable by accepted CLI/library-only architecture.

Operational acceptance evidence is preserved in Issue #51.

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

## Pre-candidate correctness gates

The repository correctness gates that had to close before release-candidate construction are now satisfied:

1. **Issue #76 / PR #79 — strict standalone RecoveryProof comparison: PASS**
   - incomplete/malformed proofs cannot compare as equivalent;
   - the complete proof shape is validated before comparison through one shared strict decoder;
   - valid identical proofs pass;
   - valid unequal proofs fail with `RECOVERY_PROOF_MISMATCH`.

2. **Issue #77 / PR #80 — authoritative-ledger namespace policy: PASS**
   - Core v1 uses a closed-world committed-tree namespace;
   - unknown paths fail closed across the complete authoritative history, including paths later deleted;
   - the accepted v1 grammar is defined by `docs/assurance/LEDGER_NAMESPACE_CONTRACT_V1.md`;
   - future namespace extension requires a separately reviewed format/migration contract.

3. **Repository current-state consistency**
   - current status/protected-gate documents record #76/#77 as closed;
   - historical reconciliation evidence remains historical rather than being treated as current blocker state.

Passing these correctness gates authorizes only progression to the separately protected release-candidate sequence. It does not authorize building/deploying a production candidate, changing the downstream product profile, replacing the live binary, activating a Core service or enabling authority writes.

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