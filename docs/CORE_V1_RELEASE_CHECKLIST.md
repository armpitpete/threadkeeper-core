# Threadkeeper Core v1 Read-Only Release Checklist

Use this checklist for any **future Core binary replacement**. The currently accepted production profile is already production-read-only accepted under Issue #51 and is not required to re-run this checklist merely because source `main` advanced.

Normative scope: `docs/CORE_V1_RELEASE_BOUNDARY.md`.

Every row requires one of `PASS`, `FAIL`, or `N/A`, plus exact evidence. A blank row is not acceptance.

## A. Repository reconciliation

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| A1 | Protected source base identified by exact SHA | **PASS** | repository/reconciliation closure base protected-main `6fd56b1143c3acb9addba860db0fcc4879608244`; PR #82 merged at that exact protected-main state |
| A2 | Current production source/binary/profile recorded | **PASS** | deployed source `46f476fd4e0a346e45034310c423f6c1cd592f65`; binary SHA-256 `7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37`; ledger `/var/lib/threadkeeper-core/authority/ledger.git` / `refs/heads/main`; no Core service; `AUTHORITY_WRITES_DISABLED` |
| A3 | Downstream `armpitpete/threadkeeper` Core-profile impact enumerated | **PASS** | `docs/CORE_V1_RECONCILIATION_2026-10-02.md`; product observer remains deliberately bound to the accepted deployed Core profile; future RC requires a separate protected product-profile change |
| A4 | No unresolved contradictory current-status document | **PASS** | PR #82 merge `6fd56b1143c3acb9addba860db0fcc4879608244`; repository-wide closure scan repairs the remaining release-boundary/Fresh-Genesis status residue and records exact-head scan evidence in the closure PR |
| A5 | Optional/deferred work cannot silently become a release blocker | **PASS** | Issues #70/#74 remain explicitly deferred/non-blocking; `docs/CORE_V1_RELEASE_BOUNDARY.md` keeps Recall/MCP/Policy Pack/Specialist Pack and other optional product lanes outside the Core v1 read-only release boundary |

## B. Release-blocking correctness

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| B1 | Issue #76 strict RecoveryProof comparison contract closed | PASS | PR #79; protected-main `f1cb0a8309cd397227f767d786d1c80c2cd073d7` |
| B2 | Incomplete/malformed RecoveryProof regression fails closed | PASS | PR #79 exact-head regression; merged in `f1cb0a8309cd397227f767d786d1c80c2cd073d7` |
| B3 | Valid identical RecoveryProof comparison passes | PASS | PR #79 exact-head conformance; merged in `f1cb0a8309cd397227f767d786d1c80c2cd073d7` |
| B4 | Tampered valid RecoveryProof comparison fails | PASS | PR #79 exact-head conformance; merged in `f1cb0a8309cd397227f767d786d1c80c2cd073d7` |
| B5 | Issue #77 ledger namespace policy explicitly accepted | PASS | closed-world Core v1 namespace; PR #80; protected-main `deb2cd42a69f61f30c2588b9cea18c75088aed97` |
| B6 | #77 implementation/regressions pass for the accepted policy | PASS | post-merge conformance run `37065581798`: `test` + `windows-git-environment-isolation` GREEN |

## C. Exact-head conformance

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| C1 | PR head SHA frozen |  |  |
| C2 | Base remains current protected `main` |  |  |
| C3 | required `test` check GREEN |  |  |
| C4 | required `windows-git-environment-isolation` check GREEN |  |  |
| C5 | whole-tree Go tests/race coverage required by workflow GREEN |  |  |
| C6 | strict JSON/canonical/digest/schema regressions GREEN |  |  |
| C7 | Git repository/environment hostile regressions GREEN |  |  |
| C8 | Genesis/actor-policy/replay regressions GREEN |  |  |
| C9 | quarantine/CAS/idempotency/post-CAS regressions GREEN |  |  |
| C10 | recovery/restore-verification regressions GREEN |  |  |
| C11 | load/resource reference proof GREEN |  |  |
| C12 | hard `AUTHORITY_WRITES_DISABLED` proof GREEN |  |  |
| C13 | unresolved review threads = 0 |  |  |

## D. Build identity

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| D1 | immutable release/RC name assigned |  |  |
| D2 | exact source SHA embedded |  |  |
| D3 | binary SHA-256 recorded |  |  |
| D4 | Go/toolchain identity recorded |  |  |
| D5 | target OS/architecture recorded |  |  |
| D6 | dependency versions recorded |  |  |
| D7 | build flags/provenance recorded |  |  |
| D8 | `authority_writes_enabled:false` in built candidate |  |  |

## E. Production-ledger copy compatibility

This stage MUST use a copy, not the live authority path.

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| E1 | old accepted binary tested against exact ledger copy |  |  |
| E2 | candidate binary tested against identical ledger copy |  |  |
| E3 | Genesis identity equal |  |  |
| E4 | actor-policy root equal |  |  |
| E5 | authoritative head equal |  |  |
| E6 | governed projection equal |  |  |
| E7 | replay identity equal |  |  |
| E8 | valid RecoveryProof authority identity equal |  |  |
| E9 | no authority-store mutation occurred |  |  |

## F. Candidate operational proof

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| F1 | accepted production-shaped envelope reviewed for candidate |  |  |
| F2 | declared operation count completed |  |  |
| F3 | required resource metrics available at every sample |  |  |
| F4 | resource ceilings PASS without relaxation-after-failure |  |  |
| F5 | replay/RecoveryProof remains stable throughout load run |  |  |
| F6 | independent-secondary restore/equivalence re-proved if candidate changes recovery interpretation |  |  |
| F7 | failure evidence, if any, preserved rather than overwritten |  |  |

## G. Downstream product compatibility

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| G1 | exact new Core profile proposed in `armpitpete/threadkeeper` |  |  |
| G2 | observer binary/source/path/ref/proof bindings updated through protected review |  |  |
| G3 | observer command allowlist remains bounded/read-only |  |  |
| G4 | observer hostile/freshness/profile tests GREEN |  |  |
| G5 | product change does not enable Core writes/service/second Manager |  |  |
| G6 | exact product protected-main SHA accepting the new profile recorded |  |  |

## H. Side-by-side IntoVPS validation

The existing production binary MUST remain available as rollback material during this stage.

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| H1 | candidate installed at a separate non-authoritative path |  |  |
| H2 | candidate binary SHA matches frozen build |  |  |
| H3 | candidate read-only inspection against live ledger succeeds |  |  |
| H4 | live Genesis/actor-policy/head/replay identity matches accepted state |  |  |
| H5 | candidate authority-write path remains hard-disabled |  |  |
| H6 | no unexpected Core listener/service introduced |  |  |
| H7 | no unexpected Core external network dependency introduced |  |  |
| H8 | live read-only comparison old vs candidate accepted |  |  |

## I. Protected promotion

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| I1 | exact candidate source + binary hashes named in promotion decision |  |  |
| I2 | exact current production binary retained/recoverable |  |  |
| I3 | exact pre-promotion RecoveryProof preserved outside authority store |  |  |
| I4 | verified independent authority-ledger backup available |  |  |
| I5 | explicit protected promotion authorization obtained |  |  |
| I6 | active binary replaced only with exact tested artifact |  |  |
| I7 | post-promotion SHA/version/write-gate reverified |  |  |
| I8 | post-promotion RecoveryProof/replay identity reverified |  |  |
| I9 | downstream product observation succeeds against exact promoted profile |  |  |
| I10 | rollback remains available until acceptance recorded |  |  |

## J. Final dispositions

| State | Required conclusion | Status | Exact evidence |
|---|---|---|---|
| J1 | `CORE IMPLEMENTATION COMPLETE` for exact source SHA |  |  |
| J2 | `READ-ONLY RELEASE CANDIDATE` for exact binary SHA |  |  |
| J3 | `PRODUCTION READ-ONLY ACCEPTED` for exact deployed profile |  |  |
| J4 | `SERVICE ACTIVATED` | N/A for Core v1 unless architecture changes | accepted CLI/library-only decision |
| J5 | `AUTHORITY WRITES ENABLED` | **NO / outside this release** | `AUTHORITY_WRITES_DISABLED` |

## Stop rules

Stop rather than infer acceptance when:

- a required check is not exact-head GREEN;
- an authority/replay identity differs without an accepted migration;
- the candidate cannot be tied to exact source/build evidence;
- the downstream product still binds only the old Core profile;
- rollback/independent recovery evidence is missing;
- live read-only comparison diverges;
- any step would require enabling authority writes or inventing a Core service.

No row in this checklist grants authority to remove or weaken `AUTHORITY_WRITES_DISABLED`.
