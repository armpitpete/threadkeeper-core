# Core v1 Protected Gates — Reconciled Status

This document was originally the pre-production gate list. The 2026-10-02 reconciliation preserves those boundaries while recording which gates were subsequently satisfied.

Normative current release-boundary document: `docs/CORE_V1_RELEASE_BOUNDARY.md`.

## Completed authority-kernel prerequisites

Established:

- assurance/recovery foundation integrated;
- Ed25519 actor proof and exact-grant primitives implemented;
- consolidated quarantine/CAS boundary merged as `fde19f4c03a1915f7d26da493593566a6017bc49` and independently **PASSED** Issue #36;
- Fresh Genesis bootstrap merged and the real production Genesis instantiated under Issue #45;
- authoritative actor-policy sourcing installed and bound to production Genesis;
- production filesystem ownership/durability semantics proved on the dedicated IntoVPS host;
- code/reference load-resource machinery merged at `46f476fd4e0a346e45034310c423f6c1cd592f65`;
- independent-secondary restore-verification machinery merged at `a51a6ccfdecc64797bdd263fa9bd9fc5f2d15b71`;
- Core v1 code-side/reference E2E acceptance merged at `d39feadbe7c01258b2d36ff5d0675d994d991c68`.

Historical governance remains explicit: PR #11 was owner-authorised and merged without a genuinely independent full Issue #9 PASS at that historical point. That history is not rewritten. The later consolidated authority boundary was independently re-tested and passed under Issue #36.

None of these facts enables public authority writes.

## Production Gate A — load/resource proof

**PASS — 2026-08-20, Issue #51.**

Accepted production envelope:

- `threadkeeper-core-production-initial-v1`;
- 4 workers × 25 iterations;
- 100 completed operations;
- 991 resource samples;
- zero unavailable open-handle samples;
- measured peak/settled resource growth inside the accepted ceilings;
- exact production RecoveryProof remained stable;
- `passed:true`, exit 0, empty stderr.

Issue #65 was later closed as stale/duplicative of this evidence.

## Production Gate B — independent secondary restore

**PASS — 2026-08-20, Issue #51.**

The production authority was backed up to separately custodied OCI storage, the primary authority path was made unavailable, restore input was fetched from that declared secondary, the restored repository passed strict fsck, and Core reproduced exact Genesis/actor-policy/head/replay/projection RecoveryProof identity.

Core correctly reports operational independence as requiring external review; the external evidence review separately recorded Gate B PASS with its caveats preserved.

## Production Gate C — direct Core service activation

**N/A by accepted architecture / PASS by review — 2026-08-20, Issue #51.**

Core v1 remains intentionally CLI/library-only in production.

Do not create `threadkeeper-core.service`, a daemon or listener merely to satisfy an obsolete service-activation assumption. Persistent transports/adapters belong outside Core unless a later protected architecture decision proves a missing Core runtime primitive.

## Current accepted production state

- deployed source: `46f476fd4e0a346e45034310c423f6c1cd592f65`;
- binary SHA-256: `7d823828262e18d1ab6398687e451ddbb6ca536f4b460b8a767f55bc45348a37`;
- authority writes: disabled;
- direct Core service: none;
- production read-only operational acceptance: PASS.

The fact that protected source `main` later advanced does not itself supersede this accepted production profile.

## Future binary-replacement gates

A future Core binary replacement must follow `docs/CORE_V1_RELEASE_BOUNDARY.md`.

The 2026-10-02 correctness blockers are closed in protected-main lineage:

- Issue #76 — strict standalone RecoveryProof comparison — **PASS / MERGED** in PR #79 as `f1cb0a8309cd397227f767d786d1c80c2cd073d7`;
- Issue #77 — closed-world authoritative-ledger namespace — **PASS / MERGED** in PR #80 as `deb2cd42a69f61f30c2588b9cea18c75088aed97`;
- fresh post-merge conformance run `37065581798` on `deb2cd42a69f61f30c2588b9cea18c75088aed97` passed both required jobs.

Remaining future binary-replacement gates are:

1. exact protected-main release-candidate conformance and build provenance;
2. production-ledger-copy compatibility proof;
3. downstream `armpitpete/threadkeeper` frozen Core-profile update;
4. side-by-side live read-only equivalence before promotion;
5. separately authorised protected promotion.

No item above authorises starting the release-candidate deployment phase before repository-status reconciliation itself is merged and accepted.

## Final write-enable decision

`AUTHORITY_WRITES_DISABLED` remains the hard public/service gate.

Removing or weakening it is **not** a remaining step of the already accepted read-only Core v1 profile. It is a distinct future protected release decision requiring its own exact transport, authentication/authorization, recovery and operational evidence.

## Optional integrations

Not Core v1 read-only prerequisites unless separately selected:

- external witness deployment;
- federation transport;
- checkpoint-accelerated replay;
- Recall/search/vector storage;
- GUI;
- MCP product transport;
- Policy Pack/Specialist Pack product rollout.
