# Candidate Write / CAS v1 Acceptance Gates

A candidate-write implementation conforms only if all applicable gates pass on the exact reviewed head.

1. **CW-01 — Public gate closed.** `authority-write` still fails with `AUTHORITY_WRITES_DISABLED`.
2. **CW-02 — Exact head required.** A new request with a non-current expected head fails `STALE_STATE`.
3. **CW-03 — Idempotency first.** An exact accepted retry is returned before stale-state handling.
4. **CW-04 — Durable retry.** Exact retry works after constructing a new Reader/process-equivalent state with no response cache.
5. **CW-05 — Conflict.** Same idempotency key with different event ID or digest fails `IDEMPOTENCY_CONFLICT`.
6. **CW-06 — Event-ID uniqueness.** A logical `event_id` already present in accepted history cannot be accepted again under a new idempotency key; both prepare and acceptance-side revalidation must fail before CAS.
7. **CW-07 — Canonical input.** Non-canonical event bytes are rejected before candidate creation.
8. **CW-08 — Digest.** Invalid `content_sha256` is rejected before candidate creation.
9. **CW-09 — Schema.** Event must validate against the schema accepted at H0.
10. **CW-10 — Policy.** Record kind, event schema and authority-policy version must match the accepted binding at H0.
11. **CW-11 — Reducer.** Prior/resulting state assertions must pass the accepted reducer against H0 projection.
12. **CW-12 — Event family.** Event families without accepted write semantics fail closed.
13. **CW-13 — Add only.** Candidate cannot overwrite an existing durable event path.
14. **CW-14 — Safe path.** Candidate event path rejects traversal, shell/pathspec metacharacters and non-JSON destinations.
15. **CW-15 — No ref during prepare.** Preparing a valid candidate leaves the authoritative ref exactly at H0.
16. **CW-16 — One-tree delta.** Candidate differs from H0 only by the intended event addition.
17. **CW-17 — Exact parent.** Candidate has exactly one parent and it is H0.
18. **CW-18 — Competing candidate.** Two valid candidates against H0 can be prepared; after one wins CAS, the other fails stale.
19. **CW-19 — Crash before CAS.** Reopening/replaying the ledger after prepare-only still returns H0 and no candidate event.
20. **CW-20 — Crash after CAS.** Direct CAS followed by retry reconstructs `already_accepted` with original accepting commit.
21. **CW-21 — Hook isolation.** A repository `reference-transaction` hook that would reject/write a sentinel is not executed by CAS.
22. **CW-22 — Acceptance revalidation.** Treat the candidate handle as untrusted: immediately before CAS, re-read the candidate commit and require the exact H0 parent, sole event-path delta, event ID, idempotency key, digest, schema/policy binding and reducer semantics. A forged/substituted candidate handle must fail without moving the ref.
23. **CW-23 — CAS child guard.** A commit not having H0 as its sole parent cannot be passed to authoritative CAS.
24. **CW-24 — Post-CAS replay.** Success is returned only after full replay verifies the new head/event identity.
25. **CW-25 — Current-head response.** Idempotent retry returns original accepting commit and separately reports current ledger head.
26. **CW-26 — Direct Git.** Candidate construction and CAS use direct Git process invocation, not shell command strings.
27. **CW-27 — Existing read safety.** Read-only fsck/history/config/environment protections continue to pass.
28. **CW-28 — Build.** CGO-free Linux build succeeds.
29. **CW-29 — Module lock.** `go mod tidy` does not change committed module metadata.
30. **CW-30 — Full suite.** `go test ./...` passes, including all earlier reducer/replay conformance tests.
31. **CW-31 — Repository object isolation.** `objects/info/alternates` and `objects/info/http-alternates` are forbidden in an authoritative ledger. Replay, candidate verification and CAS must reject them; a repository-local alternate cannot supply H1/tree/blob objects used for authoritative ref movement.
32. **CW-32 — Snapshot-consistent idempotency.** An idempotency lookup is bound to the exact replay snapshot commit supplied by its caller and never re-resolves the authoritative ref internally. A retry response cannot combine an acceptance found at H1 with `ledger_commit=H0`.
33. **CW-33 — Common-dir isolation.** Git `commondir` metadata is forbidden and must be rejected before Git is invoked, so a façade Git directory cannot redirect refs/config/objects to another common directory.
34. **CW-34 — Canonical tree mode.** Durable event JSON and versioned schema/reducer-binding JSON must be exact `100644 blob` tree entries. Forged `100755`, `120000` or other modes fail replay/acceptance before ref movement.
35. **CW-35 — No promisor/lazy objects.** Promisor/partial-clone configuration is forbidden and controlled Git invocations disable lazy object fetching; authoritative replay and candidate validation cannot materialize missing objects from a remote.
36. **CW-36 — No filesystem authority indirection.** Critical authoritative Git paths, including `HEAD`, `config`, packed refs, and the `objects` and `refs` trees, must not be symlinks. Static redirection outside the service-owned ledger boundary is an integrity failure.
37. **CW-37 — One ref/config backend.** Threadkeeper v1 accepts only Git's classic `files` ref backend and the repository's primary `config` file. Reftable/other ref-storage extensions, `$GIT_DIR/reftable`, `extensions.worktreeConfig`, and `config.worktree` fail closed before replay, prepare, candidate verification or CAS.
38. **CW-38 — Canonical repository root.** The supplied Git-directory root and every ancestor component used to reach it must be non-symlinked. `New` must reject root/ancestor aliases, retain only the canonical self-resolving path, and every Core Git invocation must recheck that root boundary so later symlink replacement cannot silently redirect reads or authoritative `update-ref` writes.
39. **CW-39 — Repository filesystem identity pinning.** `New` must pin the underlying filesystem identity of the canonical Git-directory root. Every later Core Git invocation must require the path to resolve to that same filesystem object; replacing the repository with another ordinary directory at the same pathname is an integrity failure.
40. **CW-40 — Direct authority ref / no-deref CAS.** The configured authoritative ref must be a direct ref, never a Git symbolic ref. Static symbolic refs must fail before replay/prepare/CAS, and the actual CAS must use no-deref ref-update semantics so a symbolic ref introduced by a check-to-exec race cannot redirect mutation to another ref target.
