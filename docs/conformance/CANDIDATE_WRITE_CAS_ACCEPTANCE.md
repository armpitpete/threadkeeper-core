# Candidate Write / CAS v1 Acceptance Gates

A candidate-write implementation conforms only if all applicable gates pass on the exact reviewed head.

1. **CW-01 — Public gate closed.** `authority-write` still fails with `AUTHORITY_WRITES_DISABLED`.
2. **CW-02 — Exact head required.** A new request with a non-current expected head fails `STALE_STATE`.
3. **CW-03 — Idempotency first.** An exact accepted retry is returned before stale-state handling.
4. **CW-04 — Durable retry.** Exact retry works after constructing a new Reader/process-equivalent state with no response cache.
5. **CW-05 — Conflict.** Same idempotency key with different event ID or digest fails `IDEMPOTENCY_CONFLICT`.
6. **CW-06 — Canonical input.** Non-canonical event bytes are rejected before candidate creation.
7. **CW-07 — Digest.** Invalid `content_sha256` is rejected before candidate creation.
8. **CW-08 — Schema.** Event must validate against the schema accepted at H0.
9. **CW-09 — Policy.** Record kind, event schema and authority-policy version must match the accepted binding at H0.
10. **CW-10 — Reducer.** Prior/resulting state assertions must pass the accepted reducer against H0 projection.
11. **CW-11 — Event family.** Event families without accepted write semantics fail closed.
12. **CW-12 — Add only.** Candidate cannot overwrite an existing durable event path.
13. **CW-13 — Safe path.** Candidate event path rejects traversal, shell/pathspec metacharacters and non-JSON destinations.
14. **CW-14 — No ref during prepare.** Preparing a valid candidate leaves the authoritative ref exactly at H0.
15. **CW-15 — One-tree delta.** Candidate differs from H0 only by the intended event addition.
16. **CW-16 — Exact parent.** Candidate has exactly one parent and it is H0.
17. **CW-17 — Competing candidate.** Two valid candidates against H0 can be prepared; after one wins CAS, the other fails stale.
18. **CW-18 — Crash before CAS.** Reopening/replaying the ledger after prepare-only still returns H0 and no candidate event.
19. **CW-19 — Crash after CAS.** Direct CAS followed by retry reconstructs `already_accepted` with original accepting commit.
20. **CW-20 — Hook isolation.** A repository `reference-transaction` hook that would reject/write a sentinel is not executed by CAS.
21. **CW-21 — Acceptance revalidation.** Treat the candidate handle as untrusted: immediately before CAS, re-read the candidate commit and require the exact H0 parent, sole event-path delta, event ID, idempotency key, digest, schema/policy binding and reducer semantics. A forged/substituted candidate handle must fail without moving the ref.
22. **CW-22 — CAS child guard.** A commit not having H0 as its sole parent cannot be passed to authoritative CAS.
23. **CW-23 — Post-CAS replay.** Success is returned only after full replay verifies the new head/event identity.
24. **CW-24 — Current-head response.** Idempotent retry returns original accepting commit and separately reports current ledger head.
25. **CW-25 — Direct Git.** Candidate construction and CAS use direct Git process invocation, not shell command strings.
26. **CW-26 — Existing read safety.** Read-only fsck/history/config/environment protections continue to pass.
27. **CW-27 — Build.** CGO-free Linux build succeeds.
28. **CW-28 — Module lock.** `go mod tidy` does not change committed module metadata.
29. **CW-29 — Full suite.** `go test ./...` passes, including all earlier reducer/replay conformance tests.
