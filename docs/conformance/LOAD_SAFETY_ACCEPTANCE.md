# Load Safety Acceptance

## Current production status

The production load gate described below was subsequently executed on the dedicated IntoVPS production host and **PASSED** under Issue #51 on 2026-08-20 using the accepted `threadkeeper-core-production-initial-v1` envelope. The accepted production profile remains source `46f476fd4e0a346e45034310c423f6c1cd592f65` with `AUTHORITY_WRITES_DISABLED`.

For any future Core binary replacement, this contract becomes active again: repository/reference evidence does not substitute for a fresh production-shaped measurement of the replacement candidate.

Performance is acceptable only if load cannot change authority semantics.

Before a public write interface is enabled, evidence must demonstrate:

1. concurrent read replay returns internally consistent exact heads;
2. competing writes preserve exact-head CAS and never auto-rebase;
3. same-key retry/conflict semantics remain deterministic under concurrency;
4. no request can observe or gain authority from another request's temporary candidate/quarantine material or credentials;
5. **if checkpoint-accelerated replay is enabled**, its result must equal full authoritative replay and an invalid/missing checkpoint must fail closed or fall back to full replay. When acceleration is not enabled, full replay is authoritative and this clause is not applicable rather than a release blocker;
6. cancellation/timeouts cannot create a successful-looking partial authority transition;
7. bounded overload returns explicit backpressure/failure rather than dropping governance events or admitting work beyond the declared service limit;
8. restore/replay under load preserves exact recovery/projection identity;
9. Go heap, goroutine and process descriptor/handle growth are bounded for the **declared measured envelope**. A repository/CI reference envelope demonstrates implementation behavior only and is not a production capacity claim;
10. `AUTHORITY_WRITES_DISABLED` remains effective regardless of concurrency until the separate enablement gate is accepted.

## Two evidence layers

### Code/reference conformance

Core must provide repeatable tests and machine-readable proof machinery that exercise these semantics under an explicit bounded reference envelope. Passing the repository reference envelope closes the implementation/proof-machinery gate only.

### Production deployment envelope

For the accepted production profile, the production load gate is closed PASS by Issue #51 evidence. A future replacement build must again declare/review its production-shaped concurrency/workload/resource ceilings and run the same proof machinery against the actual or equivalently production-shaped ledger/storage environment. The recorded evidence must identify the exact Core build, ledger/recovery identity and supplied envelope.

No reference CI result may be extrapolated into an unmeasured production capacity claim.

## Authority status

Passing either load test layer does not enable authority writes. Public write activation remains a separate final release decision and `AUTHORITY_WRITES_DISABLED` remains mandatory until then.