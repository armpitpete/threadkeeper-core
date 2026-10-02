# Core v1 Repository Reconciliation — 2026-10-02

## Status

This record reconciles the Threadkeeper Core source repository, the accepted IntoVPS production component, and the downstream `armpitpete/threadkeeper` product dependency.

It is a **read-only/repository-governance reconciliation**. It does not replace the production binary, mutate the authority ledger, activate a service, or enable authority writes.

## Source-of-truth precedence

When records disagree, use this order:

1. live immutable/cryptographic evidence from the accepted production component;
2. exact accepted commits, artifacts, hashes and preserved operational evidence;
3. normative documents on protected `main`;
4. issue/PR state and narrative.

Historical evidence is preserved even when later evidence supersedes its current-state conclusion.

## Three reconciled identities

### Accepted production Core

- authority domain: IntoVPS `threadkeeper-core`;
- host address used by the accepted product profile: `89.32.146.251`;
- binary: `/usr/local/bin/threadkeeper-core`;
- binary SHA-256: `7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37`;
- embedded source commit: `46f476fd4e0a346e45034310c423f6c1cd592f65`;
- Go: `go1.26.5`, `linux/amd64`;
- authority writes: `false`;
- ledger: `/var/lib/threadkeeper-core/authority/ledger.git`;
- authoritative ref: `refs/heads/main`;
- Genesis/head: `73fa0e66df2ae80b4b2a04247112470f6bb8e451`;
- actor-policy root SHA-256: `803e61858fe1dfae96b357845bed1b10644a5028801d307533fcf312d8b4a40a`;
- replay SHA-256: `6316bde6bf6f2caa0bc33f9cd495c3bf222c35956c8403e55c890818f74fea12`;
- accepted canonical RecoveryProof SHA-256 recorded by the product profile: `3806b7c8a94d9521a991927781ead2ca78ab4ec12b66cbceb641086e9e66cad4`;
- `/var/lib/threadkeeper-core`: `0755 root:root`;
- `/var/lib/threadkeeper-core/authority`: `0700 threadkeeper-core:threadkeeper-core`;
- no `threadkeeper-core.service`, daemon or Core listener is accepted.

A fresh read-only host check on 2026-10-02 reconfirmed the binary SHA-256, embedded source commit, platform/write gate and filesystem ownership/modes above.

### Core source repository

Reconciliation started from protected `threadkeeper-core/main`:

`96652a53700c44234367f5955f3a24c363aa32af`

This is seven commits after the deployed production source `46f476fd...`.

### Threadkeeper product dependency

The main `armpitpete/threadkeeper` repository deliberately treats the production Core identity above as a frozen accepted profile.

Relevant product surfaces include:

- `docs/CORE_OBSERVATION_RECEIPT_V0_1.md`;
- `docs/CORE_OBSERVATION_REAL_HOST_PACKAGE_V0_1.md`;
- `threadkeeper/core_observation.py`;
- `deploy/core-observation/run-one-shot-core-observation.sh`;
- associated schema, tests, requirements and impact manifests.

A future Core binary replacement therefore requires a reviewed downstream product-profile change. Updating Core without reconciling those bindings would intentionally fail the product observation gate.

## Production operational acceptance

Issue #51 is the authoritative operational acceptance record for the deployed Core profile.

Reconciled status:

- **Gate A — production load/resource proof: PASS** on 2026-08-20.
  - accepted envelope: `threadkeeper-core-production-initial-v1`;
  - 4 concurrent workers × 25 iterations;
  - 100 completed operations;
  - 991 resource samples;
  - zero unavailable open-handle samples;
  - peak growth: heap 2,914,064 bytes; goroutines 19; open handles 29;
  - settled growth: heap 45,024 bytes; goroutines 0; open handles 0;
  - `passed:true`, exit 0, empty stderr.
- **Gate B — independent-secondary destructive restore: PASS** on 2026-08-20.
  - secondary custody on OCI `server.vaelinya.uk`;
  - restore source fetched from the secondary after the primary authority path was made unavailable;
  - restored repository passed strict fsck;
  - Core equivalence passed;
  - original/restored canonical RecoveryProof SHA-256 matched exactly.
- **Gate C — direct Core service activation: N/A by accepted architecture / PASS by review.**
  - Core v1 is intentionally CLI/library-only;
  - no direct `threadkeeper-core.service`;
  - persistent transports/adapters remain outside Core.

Issue #65 was stale after Gate A PASS and has been closed during this reconciliation.

`AUTHORITY_WRITES_DISABLED` remains mandatory. Operational acceptance did not enable a public authority-write transport.

## Seven post-deployment commits

The commits between deployed source `46f476fd...` and reconciled source-main `96652a53...` are classified for the **next binary replacement**, not retroactively for the already accepted production profile.

| Commit | Purpose | Classification |
|---|---|---|
| `a51a6ccfdecc64797bdd263fa9bd9fc5f2d15b71` | strict independent-secondary restore-verification machinery and `recovery-restore-verify` | **REQUIRED** |
| `d39feadbe7c01258b2d36ff5d0675d994d991c68` | Core v1 code-side/reference E2E acceptance harness | **REQUIRED** |
| `d1a8fa8b2d3925c7863866d9e9f1d693ad1346df` | bounded MCP interoperability profile | **ALLOWED / non-blocking** |
| `ec62a32d472d3bc86c016b0400a71b5cd9ce6e85` | generic Policy Pack v0.1 contract/schema/example | **ALLOWED / non-blocking** |
| `5b8a764535fd114844b80b3b24f97822471263a5` | Policy Pack repository-content security invariant | **ALLOWED / non-blocking** |
| `74cc8889cab2b9c9a0a2056b5f3a1f10ce929165` | close Policy Pack checkout/materialisation security gap | **ALLOWED / non-blocking** |
| `96652a53700c44234367f5955f3a24c363aa32af` | merge the Policy Pack repository-content security lane | **ALLOWED / non-blocking** |

Decision: use current protected-main lineage as the base for future release-blocking repairs. Do not cherry-pick an older release branch merely to exclude already-merged optional documentation/contracts.

Being present on `main` does not make MCP or Policy Pack product evolution a Core v1 production prerequisite.

## Issue and PR reconciliation

### Closed/superseded historical authority-boundary work

The following old gates/repair issues were left open after their substantive work was overtaken by the consolidated independent hostile PASS in Issue #36 at exact merged commit `fde19f4c03a1915f7d26da493593566a6017bc49`:

- #9;
- #21;
- #25;
- #26;
- #28;
- #31;
- #32;
- #35.

They are closed as superseded/completed while retaining their historical FAIL/repair evidence.

PR #10 was a review-only frozen snapshot whose own contract prohibited merge; it is closed unmerged as superseded by later consolidated review.

PR #13's Git-alternates repair is also superseded. Current `main` rejects `objects/info/alternates` and `objects/info/http-alternates`, and Issue #36 explicitly re-tested repository-local alternates. PR #13 is closed unmerged.

### Release-status correction at reconciliation time

At reconciliation time, Issue #54 remained the Core-side status-correction gate pending the reconciliation/status documentation merge. PR #78 later merged and Issue #54 closed.

### Newly demonstrated next-release decisions at reconciliation time

- **#76 — Harden recovery-compare against incomplete RecoveryProof inputs.**
  - live hostile case proved that incomplete documents can compare as equivalent in the standalone command;
  - the newer restore-verification decoder is already strict;
  - this does not invalidate the frozen production observer, which does not call `recovery-compare`;
  - **RELEASE BLOCKER for any future Core binary replacement**.
- **#77 — Decide and enforce Core v1 authoritative-ledger namespace policy.**
  - unrelated committed content changes ledger/replay identity but is currently tolerated by replay;
  - protected semantic files remain integrity-enforced;
  - the contract must explicitly choose closed-world namespace enforcement or documented inert-content tolerance;
  - **RELEASE BLOCKER (policy decision, then implementation if required) for any future Core binary replacement**.

### Deferred programme work

- #70 Specialist Pack integration boundary: **DEFERRED / non-Core-v1-blocking**.
- #74 MiMo-Code persistent-memory research: **DEFERRED / research only**.

The product-v1 roadmap remains `armpitpete/threadkeeper#214`; Core does not maintain a competing product roadmap.

## Relevant branch disposition

Only branches with open PR/security/release relevance were considered release-significant.

- review/CAS historical branches: superseded by Issue #36;
- `agent/reject-repository-alternates`: superseded by merged later protection and Issue #36;
- later Policy Pack/MCP branches: already represented by protected-main commits;
- inert historical agent branches are not release blockers merely because the refs still exist.

No release decision depends on deleting historical branches.

## Repository protection

The active `threadkeeper-core` ruleset protects the default branch by:

- blocking deletion;
- blocking non-fast-forward updates;
- requiring pull requests;
- requiring review-thread resolution;
- requiring strict checks:
  - `test`;
  - `windows-git-environment-isolation`;
- no bypass actor is configured.

The branch endpoint's legacy protection summary is not used to override the active repository ruleset.

## Command-surface reconciliation

The deployed `46f476fd...` binary exposes:

- `version`;
- `check-json`;
- `canonicalize`;
- `digest`;
- `validate`;
- `genesis-check`;
- `fresh-genesis-init`;
- `evidence-check`;
- `review-bundle`;
- `health-check`;
- `source-file-check`;
- `portable-check`;
- `ledger-inspect`;
- `ledger-recovery-proof`;
- `ledger-load-proof`;
- `recovery-compare`;
- hard-gated `authority-write`.

Current source-main additionally contains read-only `recovery-restore-verify`.

No accepted Core v1 daemon/listener command exists.

## Completion-state vocabulary

These states MUST remain distinct:

1. **CORE IMPLEMENTATION COMPLETE** — required source-side primitives/contracts/tests exist on an exact protected-main SHA.
2. **READ-ONLY RELEASE CANDIDATE** — an exact build from an accepted SHA has complete build provenance and all release-blocking code gates PASS.
3. **PRODUCTION READ-ONLY ACCEPTED** — the exact deployed build/profile has passed live identity, load/recovery and operational gates with writes disabled.
4. **SERVICE ACTIVATED** — a separately reviewed long-running runtime exists and is activated. For Core v1 direct activation is currently N/A by architecture.
5. **AUTHORITY WRITES ENABLED** — a separate protected decision has removed/altered the hard write gate and admitted an exact write transport. This has **not** occurred.

The deployed `46f476fd...` profile is **PRODUCTION READ-ONLY ACCEPTED**.

## Reconciled findings

### Standalone RecoveryProof comparison

At reconciliation time, standalone `recovery-compare` validated raw JSON syntax but did not require a complete RecoveryProof before struct comparison. Issue #76 was opened for the required next-release repair.

The strict `restoreproof.DecodeRecoveryProof` path already requires the full closed proof shape and is the accepted restore-verification path.

### Ledger namespace

At reconciliation time, Core strongly enforced known semantic paths but did not fail merely because an unrelated path existed in history. Issue #77 was opened for the contract decision.

No production ledger mutation was performed to establish this finding; testing used a disposable ledger.

## Product-repository contradiction scan

The main `armpitpete/threadkeeper` repository consistently binds its Core observation mechanism to the accepted `46f476fd...` production profile and explicitly requires a reviewed repository change for a different binary/path/ref/profile.

No current contradiction requiring a product-repository patch was found in this reconciliation.

The product binding is therefore deliberately left unchanged.

## Next-release baseline decision

There is **no immediate production-update requirement merely because Core source-main is seven commits ahead**.

The current deployed profile remains operationally accepted and is intentionally bound by the product observer.

At reconciliation time, the planned sequence for a future Core binary replacement was:

1. begin from current protected-main lineage;
2. close #76;
3. close #77 with an explicit policy and any required implementation;
4. require exact-head `test` and `windows-git-environment-isolation` PASS plus the release checklist in `CORE_V1_RELEASE_BOUNDARY.md`;
5. build/freeze an exact candidate;
6. prove compatibility against a copy of the real authority ledger;
7. update the downstream `threadkeeper` frozen Core profile through protected review;
8. only then perform side-by-side/live read-only comparison and a separately authorised promotion.

## Reconciliation terminal state

```text
THREADKEEPER CORE V1

Core source repository state: RECONCILED
Threadkeeper product dependency state: RECONCILED
Live IntoVPS baseline: VERIFIED

Current production Core:
  source: 46f476fd4e0a346e45034310c423f6c1cd592f65
  binary: 7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37
  authority writes: DISABLED
  production read-only acceptance: PASS

Next Core release base: current protected-main lineage
Next-release blockers at reconciliation time: #76, #77, exact release-candidate gates
Deferred: MCP/Policy Pack/Specialist Pack/research unless a proven authority-kernel dependency appears

Production binary changed: NO
Production ledger changed: NO
Product observer baseline changed: NO
Service activated: NO
Authority writes enabled: NO
```

## Post-reconciliation correctness closure — 2026-10-02

The findings above are preserved as the state discovered during reconciliation. They are no longer current blockers.

Subsequent protected repairs closed both findings:

- **#76 / PR #79 — PASS / MERGED**
  - strict shared RecoveryProof decoder/validator;
  - incomplete, null, unknown, duplicate and trailing proof content fails closed;
  - exact discovered malformed-equivalence regression closed;
  - merge commit: `f1cb0a8309cd397227f767d786d1c80c2cd073d7`.

- **#77 / PR #80 — PASS / MERGED**
  - Core v1 authoritative history is closed-world;
  - unknown committed paths fail with `INTEGRITY_FAILURE` across the complete history;
  - future namespace extension requires a reviewed format/migration contract;
  - merge commit: `deb2cd42a69f61f30c2588b9cea18c75088aed97`.

Final exact-main read-only compatibility evidence at code-main `deb2cd42a69f61f30c2588b9cea18c75088aed97`:

- preserved production Gate B ledger bundle remained unchanged;
- Genesis/head reproduced exactly as `73fa0e66df2ae80b4b2a04247112470f6bb8e451`;
- actor-policy root SHA-256 reproduced exactly as `803e61858fe1dfae96b357845bed1b10644a5028801d307533fcf312d8b4a40a`;
- replay SHA-256 reproduced exactly as `6316bde6bf6f2caa0bc33f9cd495c3bf222c35956c8403e55c890818f74fea12`;
- identical complete RecoveryProofs compared equivalent;
- malformed/incomplete RecoveryProof comparison failed with `RECOVERY_PROOF_INVALID`.

Current disposition:

```text
CORE V1 NEXT-RELEASE CORRECTNESS GATES

#76 RecoveryProof strict comparison: PASS
#77 Closed-world authoritative ledger: PASS

Production binary: UNCHANGED
Production ledger: UNCHANGED
Product observer profile: UNCHANGED
Core service: NOT ACTIVATED
Authority writes: DISABLED

NEXT:
separately protected Core v1 read-only release-candidate sequence
```

This closure does not authorize release-candidate deployment or production promotion.