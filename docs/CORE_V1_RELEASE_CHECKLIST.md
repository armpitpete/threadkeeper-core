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
| G1 | exact new Core profile proposed in `armpitpete/threadkeeper` | **PASS** | PR #581 reconciled candidate head `0c96de13a86620e6e848be4a3741f05218987d9a` binds Core RC1 binary SHA-256 `0bcfc7afd0632fcbdb62bc421c7ca80ee51a67e2e91dd78562ae0659d0c83bac` and source `84c3e983768f67b518c6c84f2eb62f0bf4babae7` |
| G2 | observer binary/source/path/ref/proof bindings updated through protected review | **PASS** | PR #581 exact base `c5e72d8fce05dd94bd0c597e2ed8cac2acd4e731`, head `0c96de13a86620e6e848be4a3741f05218987d9a`; bindings accepted through governed merge to protected main `7a8b43481b70de9646ed882e5a950246781479e5` |
| G3 | observer command allowlist remains bounded/read-only | **PASS** | accepted profile keeps allowlist exactly `version` + `ledger-recovery-proof`; mutation commands remain forbidden by tests and implementation |
| G4 | observer hostile/freshness/profile tests GREEN | **PASS** | exact-head runs: Ledger `37144377584`, Release Integrity `37144377495`, Core observation package `37144377662`, production package `37144377758` all SUCCESS; Proofkeeper run `37145358765` PASS with proof key `pk-b4a980c49f7076f386c52acf8a7aa73107c9259c5961dec189169e8bcd13260f`; post-merge exact-main runs `37145823072`, `37145823054`, `37145823238`, `37145823249`, `37145823084` all SUCCESS |
| G5 | product change does not enable Core writes/service/second Manager | **PASS** | PR #581 changed only observer profile/docs/test/impact surfaces; no Core write path, listener/service/database, second Manager, deployment or authority expansion; ruleset `24422007` remains active with ordinary owner bypass `never` |
| G6 | exact product protected-main SHA accepting the new profile recorded | **PASS** | Release Authority decision `ra-a08cb81e1033f1e8abf89da3e4d8d2810e5cae32977451c7cbecd8fd265d64b8`; App-governed merge actor `merrin-threadkeeper-manager[bot]`; resulting protected main `7a8b43481b70de9646ed882e5a950246781479e5`; merge receipt SHA-256 `58fcba36ae7718ff17deae8b44c2b410f58cc3315e05a6a7f93c189aac17d088` |

## H. Side-by-side IntoVPS validation

The existing production binary MUST remain available as rollback material during this stage.

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| H1 | candidate installed at a separate non-authoritative path | **PASS** | root-owned `0555` candidate at `/opt/threadkeeper-core/candidates/threadkeeper-core-v1-rc1`; production path `/usr/local/bin/threadkeeper-core` retained unchanged |
| H2 | candidate binary SHA matches frozen build | **PASS** | live IntoVPS candidate SHA-256 `0bcfc7afd0632fcbdb62bc421c7ca80ee51a67e2e91dd78562ae0659d0c83bac`, exact frozen Gate D build |
| H3 | candidate read-only inspection against live ledger succeeds | **PASS** | candidate `ledger-inspect` and `ledger-recovery-proof` succeeded as `threadkeeper-core` against live `/var/lib/threadkeeper-core/authority/ledger.git` / `refs/heads/main`; inspection SHA-256 `4c5c148cbe66f37affadb39aa93286308a0efd642fbc0cd53582f9d7f8949cc6`; proof file SHA-256 `978801ca94138353572a37d948db888e8b1585042c98bee85c3f4d7f43e2db88` |
| H4 | live Genesis/actor-policy/head/replay identity matches accepted state | **PASS** | live head/Genesis `73fa0e66df2ae80b4b2a04247112470f6bb8e451`; actor-policy root `803e61858fe1dfae96b357845bed1b10644a5028801d307533fcf312d8b4a40a`; replay SHA-256 `6316bde6bf6f2caa0bc33f9cd495c3bf222c35956c8403e55c890818f74fea12` |
| H5 | candidate authority-write path remains hard-disabled | **PASS** | candidate `authority-write` exit `1`; exact `AUTHORITY_WRITES_DISABLED` rejection preserved |
| H6 | no unexpected Core listener/service introduced | **PASS** | no Core service file/unit/process; production binary remains SHA-256 `7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37`; listener-set hash unchanged before/after `886ee514cf47e46067a7aaadb58df2f30efe2a6c376b5184cc29ad4c1fc775eb` |
| H7 | no unexpected Core external network dependency introduced | **PASS** | static candidate; live candidate RecoveryProof under `strace -f -qq -e trace=network -e signal=none` produced 0 network-trace bytes; non-privileged version/write-gate paths also produced zero network syscalls |
| H8 | live read-only comparison old vs candidate accepted | **PASS** | old and RC1 live proof files byte-identical SHA-256 `978801ca94138353572a37d948db888e8b1585042c98bee85c3f4d7f43e2db88`; old and RC1 live inspection files byte-identical SHA-256 `4c5c148cbe66f37affadb39aa93286308a0efd642fbc0cd53582f9d7f8949cc6`; live ledger manifest unchanged before/after `e98fea09157b1b4f514a68447cdf078df8f3b58725dfc99f5c7fad0f45d5a99a` |

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
