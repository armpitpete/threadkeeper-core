# Threadkeeper Core v1 Read-Only Release Checklist

Use this checklist for any **future Core binary replacement**. The exact `threadkeeper-core-v1-rc1` profile completed Gates C–I and is terminally dispositioned by Gate J / Issue #96; source `main` advancing later does not itself change production.

Normative scope: `docs/CORE_V1_RELEASE_BOUNDARY.md`.

Every row requires one of `PASS`, `FAIL`, or `N/A`, plus exact evidence. A blank row is not acceptance.

## A. Repository reconciliation

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| A1 | Protected source base identified by exact SHA | **PASS** | Gate C Issue #85 freezes exact RC1 source `84c3e983768f67b518c6c84f2eb62f0bf4babae7`, tree `9032ec44810ec2b48b2baf10c116add17bdcce1f`, derived from protected PR #84 lineage |
| A2 | Current production source/binary/profile recorded | **PASS** | deployed `threadkeeper-core-v1-rc1`; source `84c3e983768f67b518c6c84f2eb62f0bf4babae7`; binary SHA-256 `0bcfc7afd0632fcbdb62bc421c7ca80ee51a67e2e91dd78562ae0659d0c83bac`; ledger `/var/lib/threadkeeper-core/authority/ledger.git` / `refs/heads/main`; no Core service/listener; `AUTHORITY_WRITES_DISABLED` |
| A3 | Downstream `armpitpete/threadkeeper` Core-profile impact enumerated | **PASS** | Gate G / PR #581 updated the frozen observer profile through protected review; product protected main `7a8b43481b70de9646ed882e5a950246781479e5` accepts exactly RC1 source `84c3e983...` and binary `0bcfc7af...` |
| A4 | No unresolved contradictory current-status document | **PASS** | PR #95 reconciled the promoted RC1 profile to protected main `0f6f4f6e1fd61b933425707b5e5b89ba1e67020c`; Gate J / Issue #96 backfills the remaining C–F/J checklist dispositions and repairs present-tense pre-RC1 residue without rewriting dated historical evidence |
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
| C1 | PR head SHA frozen | **PASS** | Gate C Issue #85: PR #84 frozen head `ae2b748cf6805a014e68af70bca1c50e21e400af` merged to exact release source `84c3e983768f67b518c6c84f2eb62f0bf4babae7`, tree `9032ec44810ec2b48b2baf10c116add17bdcce1f` |
| C2 | Base remains current protected `main` | **PASS** | Gate C entry/exit protected `main` remained exactly `84c3e983768f67b518c6c84f2eb62f0bf4babae7`; no source-changing bookkeeping commit was created |
| C3 | required `test` check GREEN | **PASS** | exact-source post-merge conformance run `37116081295`, job `test`: SUCCESS |
| C4 | required `windows-git-environment-isolation` check GREEN | **PASS** | exact-source run `37116081295`, job `windows-git-environment-isolation`: SUCCESS |
| C5 | whole-tree Go tests/race coverage required by workflow GREEN | **PASS** | exact-source run `37116081295`: `go test ./...` and `go test -race ./...` GREEN |
| C6 | strict JSON/canonical/digest/schema regressions GREEN | **PASS** | Gate C Issue #85 terminal evidence: exact-source Linux normal/race suites GREEN for strict JSON, canonicaljson, digest and schema; independent clean-tree package pass also GREEN |
| C7 | Git repository/environment hostile regressions GREEN | **PASS** | exact-source `internal/gitledger` GREEN in normal/race tree; Windows Git-environment-isolation GREEN; independent exact-source `internal/gitledger` pass GREEN |
| C8 | Genesis/actor-policy/replay regressions GREEN | **PASS** | exact-source normal/race tree GREEN for `internal/genesis`, `internal/actorauth`, `internal/ledger`; Issue #85 terminal evidence |
| C9 | quarantine/CAS/idempotency/post-CAS regressions GREEN | **PASS** | exact-source Linux normal/race `internal/ledger` + `internal/quarantine` GREEN, including stale-head/post-CAS/stage-isolation regressions; Issue #85 |
| C10 | recovery/restore-verification regressions GREEN | **PASS** | exact-source normal/race `internal/recovery`, `internal/recoveryproof`, `internal/restoreproof` GREEN; independent exact-source packages GREEN |
| C11 | load/resource reference proof GREEN | **PASS** | exact-source `internal/loadproof` GREEN in Linux normal/race tree and Windows load-resource step GREEN; Issue #85 |
| C12 | hard `AUTHORITY_WRITES_DISABLED` proof GREEN | **PASS** | run `37116081295` hard kill-switch step GREEN; `authority-write` returned `AUTHORITY_WRITES_DISABLED` |
| C13 | unresolved review threads = 0 | **PASS** | PR #84 unresolved review threads = 0; no open Core PR at Gate C exit |

## D. Build identity

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| D1 | immutable release/RC name assigned | **PASS** | annotated tag `threadkeeper-core-v1-rc1`, tag object `95b1ae34fecd7af1212c4c7b3db2a89a14c80671`, bound to exact source `84c3e983...`; Issue #86 |
| D2 | exact source SHA embedded | **PASS** | candidate `version` and Go build info report exact `84c3e983768f67b518c6c84f2eb62f0bf4babae7`, `vcs.modified=false` |
| D3 | binary SHA-256 recorded | **PASS** | `0bcfc7afd0632fcbdb62bc421c7ca80ee51a67e2e91dd78562ae0659d0c83bac`; Issue #86 |
| D4 | Go/toolchain identity recorded | **PASS** | `go1.26.5`, compiler `gc`; Issue #86 |
| D5 | target OS/architecture recorded | **PASS** | `linux/amd64`; `GOOS=linux`, `GOARCH=amd64`, `GOAMD64=v1` |
| D6 | dependency versions recorded | **PASS** | direct compiled modules recorded: `github.com/gowebpki/jcs v1.0.1`, `github.com/santhosh-tekuri/jsonschema/v6 v6.0.2`, `golang.org/x/text v0.14.0`; full `MODULES.txt` preserved |
| D7 | build flags/provenance recorded | **PASS** | Issue #86 records exact `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOTOOLCHAIN=local go build -trimpath -buildvcs=true` provenance/ldflags; independent repeat build produced identical SHA-256 |
| D8 | `authority_writes_enabled:false` in built candidate | **PASS** | exact candidate reports `authority_writes_enabled:false`; `authority-write` exits 1 with `AUTHORITY_WRITES_DISABLED` |

## E. Production-ledger copy compatibility

This stage MUST use a copy, not the live authority path.

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| E1 | old accepted binary tested against exact ledger copy | **PASS** | Gate E Issue #87: exact old binary SHA `7d823828...` ran `ledger-inspect` + `ledger-recovery-proof` against an exact accepted production-ledger copy |
| E2 | candidate binary tested against identical ledger copy | **PASS** | exact RC1 SHA `0bcfc7af...` ran identical operations against a separate identical copy |
| E3 | Genesis identity equal | **PASS** | both binaries: Genesis `73fa0e66df2ae80b4b2a04247112470f6bb8e451`; Genesis content SHA-256 `0d484af0586f747a7e6444ee1b63b747f8bafb4bc4a0505650f1f2157eb72ceb` |
| E4 | actor-policy root equal | **PASS** | both: `803e61858fe1dfae96b357845bed1b10644a5028801d307533fcf312d8b4a40a` |
| E5 | authoritative head equal | **PASS** | both: `73fa0e66df2ae80b4b2a04247112470f6bb8e451` at `refs/heads/main` |
| E6 | governed projection equal | **PASS** | both governed-record count 0; projection SHA-256 `44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a`; inspect outputs byte-identical SHA-256 `4c5c148c...` |
| E7 | replay identity equal | **PASS** | both: `6316bde6bf6f2caa0bc33f9cd495c3bf222c35956c8403e55c890818f74fea12` |
| E8 | valid RecoveryProof authority identity equal | **PASS** | proof files byte-identical SHA-256 `978801ca94138353572a37d948db888e8b1585042c98bee85c3f4d7f43e2db88`; RC1 `recovery-compare` returned `{"equivalent":true}` |
| E9 | no authority-store mutation occurred | **PASS** | live path not used; baseline/old/RC1 copies retain identical 22-file manifest SHA-256 `a96157e2b9039342f746b726c05e8f9585e23e868603f1681735fa6ce456a0c9`, unchanged head, strict fsck PASS |

## F. Candidate operational proof

| Gate | Required result | Status | Exact evidence |
|---|---|---|---|
| F1 | accepted production-shaped envelope reviewed for candidate | **PASS** | Gate F Issue #88 reused unchanged `threadkeeper-core-production-initial-v1`; envelope SHA-256 `5f0c2c96a8bd4a47d9a8d5f9bde84fae8ed0ff2b675ad1bf7b4385546309c1b6` |
| F2 | declared operation count completed | **PASS** | 4 workers × 25 iterations = 100 required; `completed_operations:100` |
| F3 | required resource metrics available at every sample | **PASS** | 994 resource samples; 0 unavailable open-handle samples; required before/peak/settled handle metrics available |
| F4 | resource ceilings PASS without relaxation-after-failure | **PASS** | peak growth: heap 2,851,168 B, goroutines 20, handles 29; settled: heap 45,160 B, goroutines 0, handles 0; `passed:true`, exit 0, empty stderr; envelope unchanged |
| F5 | replay/RecoveryProof remains stable throughout load run | **PASS** | pre/post RecoveryProof byte-identical SHA-256 `978801ca...`; `recovery-compare` equivalent; copied-ledger pre/post manifest `02074670d448142de7e5ec5612f91a97ef23e73245b7c98d06110de0adff046c`; strict fsck PASS |
| F6 | independent-secondary restore/equivalence re-proved if candidate changes recovery interpretation | **PASS** | required for RC1; OCI bundle SHA-256 `3966077b...` freshly restored/fscked; RC1 `recovery-restore-verify` reports `core_equivalence_passed:true` with identical canonical RecoveryProof `3806b7c8...`; Issue #88 |
| F7 | failure evidence, if any, preserved rather than overwritten | **PASS** | load had no failed attempt; two F6 evidence-format preparation failures preserved with stderr SHA-256 `c404dfc1...` and `ea4518bb...`; tooling sharp edge tracked as Issue #89 |

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
| I1 | exact candidate source + binary hashes named in promotion decision | **PASS** | RC1 source `84c3e983768f67b518c6c84f2eb62f0bf4babae7`; binary SHA-256 `0bcfc7afd0632fcbdb62bc421c7ca80ee51a67e2e91dd78562ae0659d0c83bac`; exact promotion tuple recorded in Issue #94 |
| I2 | exact current production binary retained/recoverable | **PASS** | pre-promotion binary preserved at `/opt/threadkeeper-core/rollback/threadkeeper-core-pre-rc1-7d823828` and local rollback copy; SHA-256 `7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37` |
| I3 | exact pre-promotion RecoveryProof preserved outside authority store | **PASS** | local preserved proof `I:\ORDER\GitHub\_threadkeeper-core-gate-i\prepromotion-recovery-proof.json`; raw SHA-256 `978801ca94138353572a37d948db888e8b1585042c98bee85c3f4d7f43e2db88` |
| I4 | verified independent authority-ledger backup available | **PASS** | OCI `/srv/threadkeeper-core-secondary/threadkeeper-core-production.bundle`; SHA-256 `3966077b7539c8826265278dd7be22ae0465e9fc35428d559f3854723417dc06`; strict fsck PASS; restored head `73fa0e66df2ae80b4b2a04247112470f6bb8e451` |
| I5 | explicit protected promotion authorization obtained | **PASS** | exact owner authorization recorded in Issue #94 comment `5978584410`, bound to old SHA `7d8238...`, RC1 SHA `0bcfc7...`, source `84c3e983...` and fixed no-write/no-service/rollback boundaries |
| I6 | active binary replaced only with exact tested artifact | **PASS** | `/usr/local/bin/threadkeeper-core` now SHA-256 `0bcfc7afd0632fcbdb62bc421c7ca80ee51a67e2e91dd78562ae0659d0c83bac`, exactly the side-by-side accepted RC1 artifact |
| I7 | post-promotion SHA/version/write-gate reverified | **PASS** | version `threadkeeper-core-v1-rc1`; source `84c3e983768f67b518c6c84f2eb62f0bf4babae7`; `linux/amd64`; writes false; `authority-write` exit 1 with exact `AUTHORITY_WRITES_DISABLED` marker |
| I8 | post-promotion RecoveryProof/replay identity reverified | **PASS** | raw RecoveryProof SHA-256 `978801ca94138353572a37d948db888e8b1585042c98bee85c3f4d7f43e2db88`; canonical `3806b7c8a94d9521a991927781ead2ca78ab4ec12b66cbceb641086e9e66cad4`; head/Genesis `73fa0e66...`; actor-policy root `803e6185...`; replay `6316bde6...`; ledger manifest unchanged `e98fea09...` |
| I9 | downstream product observation succeeds against exact promoted profile | **PASS** | exact `armpitpete/threadkeeper@7a8b43481b70de9646ed882e5a950246781479e5` observer receipt SHA-256 `1e4ab01332101da18dd02a08e9ad7d306302a18ab3489bd044d5bf58443d3cb3`; `state_matches_expected:true` |
| I10 | rollback remains available until acceptance recorded | **PASS** | root rollback artifact remains SHA-256 `7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37`; independent OCI backup and pre-promotion RecoveryProof also retained through acceptance |

## J. Final dispositions

| State | Required conclusion | Status | Exact evidence |
|---|---|---|---|
| J1 | `CORE IMPLEMENTATION COMPLETE` for exact source SHA | **PASS** | exact RC1 source `84c3e983768f67b518c6c84f2eb62f0bf4babae7`; release-blocking source/conformance/build gates C–F passed and that exact source is embedded in the promoted binary |
| J2 | `READ-ONLY RELEASE CANDIDATE` for exact binary SHA | **PASS** | immutable `threadkeeper-core-v1-rc1`; binary SHA-256 `0bcfc7afd0632fcbdb62bc421c7ca80ee51a67e2e91dd78562ae0659d0c83bac`; Gate D build identity and Gates E–H compatibility/operational/live-read-only proofs passed |
| J3 | `PRODUCTION READ-ONLY ACCEPTED` for exact deployed profile | **PASS** | Gate I Issue #94 FINAL PASS / ACCEPTED / FROZEN; exact RC1 active at `/usr/local/bin/threadkeeper-core`; PR #95 reconciled production status to protected main `0f6f4f6e1fd61b933425707b5e5b89ba1e67020c` with post-merge conformance GREEN |
| J4 | `SERVICE ACTIVATED` | **N/A** | accepted Core v1 CLI/library-only architecture; no direct Core service, persistent Core process or listener is required or authorised |
| J5 | `AUTHORITY WRITES ENABLED` | **NO / outside this release** | promoted RC1 reports `authority_writes_enabled:false`; post-promotion `authority-write` remains hard rejected by `AUTHORITY_WRITES_DISABLED` |

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
