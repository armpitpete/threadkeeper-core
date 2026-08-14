# Load Safety Evidence v1

This matrix maps the Core v1 load-safety acceptance clauses to inspectable implementation evidence. It distinguishes semantic/reference proof from the later real production-envelope measurement.

## Evidence matrix

| # | Acceptance property | Code/reference evidence | Production status |
|---|---|---|---|
| 1 | Concurrent read replay returns internally consistent exact heads. | `internal/ledger/load_safety_test.go::TestConcurrentReplayReturnsExactSnapshot`; `internal/ledger/load_resource_test.go::TestReferenceReadLoadEnvelopePreservesRecoveryProof` requires every concurrent full recovery/replay to equal one exact baseline proof. | Requires rerun with the approved production envelope. |
| 2 | Competing writes preserve exact-head CAS and never auto-rebase. | `TestConcurrentCompetingWritersPreserveExactlyOnceCAS`; candidate/CAS regression suite including `TestCompetingCandidatesRejectStaleSecond` and exact-child checks. | Semantic gate covered; production-shaped E2E still required before writes can be enabled. |
| 3 | Same-key retry/conflict remains deterministic. | `TestConcurrentSameIdempotencyRetryIsDeterministic`; `TestConcurrentSameIdempotencyConflictIsDeterministic`; Issue #25 concurrent cleanup/recovery regressions. | Semantic gate covered; production-shaped E2E still required. |
| 4 | Requests cannot gain authority from another request's temporary material/credentials. | `internal/ledger/issue25_stage_isolation_test.go::TestIssue25IdenticalConcurrentPreparesUsePrivateStages`; quarantine binding/cleanup-isolation regressions; Issue #36 independently passed the final quarantine/CAS boundary. Core currently exposes no public write transport and stores no shared request credential scratch state. | Recheck transport/session isolation in the final production-shaped E2E if/when a public transport is introduced. |
| 5 | Checkpoint acceleration equals full replay if enabled. | Current Core does not enable checkpoint-accelerated replay. Full `ledger.Replay` is authoritative. `docs/assurance/CHECKPOINT_CONTRACT_V1.md` requires any future acceleration to verify the checkpoint and preserve replay semantics. | **N/A for Core v1 while acceleration is disabled.** Must become an active proof if acceleration is enabled later. |
| 6 | Cancellation/timeouts cannot create a successful-looking partial transition. | `TestCancelledPrepareCannotMutateAuthority`; Issue #25 cleanup races; Issue #32 prepare-bound/read and ref-lock uncertainty regressions; final Issue #36 hostile PASS of post-CAS reporting. | Semantic gate covered; final production E2E includes restart/retry recovery. |
| 7 | Bounded overload is explicit. | `internal/service/limiter_test.go`; `internal/service/load_concurrency_test.go::TestLimiterNeverAdmitsBeyondCapacityUnderBurst` proves exactly capacity admissions and explicit `ErrOverloaded` for the synchronized excess. | Actual service limit must be selected and recorded for production. |
| 8 | Restore/replay under load preserves exact projection/recovery identity. | `internal/ledger/recovery_proof_test.go::TestRecoveryProofSurvivesDestructiveBareRestore`; `internal/ledger/load_restore_test.go::TestRestoredCopiesPreserveRecoveryProofUnderLoad` requires multiple restored copies to reproduce the exact original RecoveryProof concurrently. | **Independent-secondary operational restore remains open.** Local copies prove machinery/semantics only. |
| 9 | Resource growth is bounded for the declared envelope. | `internal/loadproof` measures sampled peak and settled Go heap allocation, goroutines and process descriptors/handles; `TestReferenceReadLoadEnvelopePreservesRecoveryProof` runs the repository reference envelope; `ledger-load-proof` runs an operator-supplied envelope against a real ledger. | **Open until measured on the actual production-shaped deployment.** CI reference ceilings are not capacity claims. |
| 10 | Hard authority-write disable remains effective under concurrency. | `internal/service/load_concurrency_test.go::TestAuthorityWriteKillSwitchDominatesUnderConcurrency` runs 128 concurrent admissions with a nil Reader; every call must return `AUTHORITY_WRITES_DISABLED`, proving the gate fires before ledger/policy/auth work. | Must remain hard false throughout deployment and E2E until the final separately reviewed enablement decision. |

## Reference resource envelope

`loadproof.ReferenceEnvelope()` is intentionally small and deterministic enough for repository conformance:

- 8 workers;
- 4 full-replay operations per worker;
- sampled peak Go heap growth <= 128 MiB;
- settled Go heap growth <= 64 MiB;
- sampled peak goroutine growth <= 64;
- settled goroutine growth <= 8;
- sampled peak process descriptor/handle growth <= 128;
- settled descriptor/handle growth <= 16;
- descriptor/handle metric required.

The resource monitor samples every 5 ms and reports that interval in its machine-readable evidence. `heap_alloc_bytes` is Go live heap allocation from `runtime.MemStats`, **not RSS or total machine memory**. Descriptor/handle counts are process-global and may conservatively fail if unrelated work in the same process increases them during the measurement.

Reference conformance establishes that the implementation has bounded observable growth under that exact workload. It does not establish maximum throughput, latency, host sizing or production capacity.

## Production acceptance handoff

The production operator must provide a strict JSON envelope with the real selected concurrency/workload and explicit ceilings, then run:

```text
threadkeeper-core ledger-load-proof <ledger.git> <envelope.json> [authoritative-ref]
```

The resulting JSON binds the measured resource evidence to the exact RecoveryProof of the ledger used for the run. `passed: true` is necessary for the production load gate but does not by itself enable authority writes.

The independent-secondary restore gate and full production-shaped E2E gate remain separate.
