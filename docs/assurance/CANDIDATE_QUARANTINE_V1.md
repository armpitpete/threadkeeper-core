# Candidate Quarantine v1

Non-authoritative candidate material may contain sensitive data and must not be treated as harmless merely because it has not been accepted.

Before any public authority-write interface is enabled, candidate bytes must have a bounded quarantine lifecycle: private storage, content digest, explicit identity, no path traversal, explicit expiry/cleanup policy and promotion to authority only through the reviewed acceptance path.

## Integrated writer contract

The candidate writer treats quarantine as part of the authority boundary:

1. the durable event is fully validated against the exact current ledger snapshot before candidate materialisation;
2. the exact validated canonical event bytes are first staged in the ledger-bound private quarantine under a per-Prepare random staging identity, so identical concurrent prepares do not share temporary cleanup ownership;
3. Git candidate objects are created only from bytes read back from that staging entry;
4. after H1 is created and revalidated, preparation finalises the same raw bytes under a quarantine capability ID derived from the complete prepared identity: expected H0, prepared H1, event path, event ID, idempotency key, semantic content digest, raw-byte SHA-256 and raw byte size;
5. final capability publication is never performed by writing directly to the final filename. Each publisher writes to a private file inside the pinned quarantine root, completes write + `fsync` + close, then atomically hard-links that completed file to the final capability name without overwrite. Identical concurrent publishers may converge on an already-published completed file; a failed publisher can remove only its private file and cannot invalidate another publisher's final capability;
6. the temporary staging entry is removed before a prepared handle is returned; if H0 moved before candidate construction and the exact request is recovered as already accepted, the abandoned staging entry is removed on that recovery path as well;
7. immediately before returning a prepared candidate, Prepare rechecks current authority; if H0 moved, it reconciles against a fresh bounded replay instead of returning an already-stale candidate;
8. acceptance recomputes the expected capability ID from the untrusted handle and its claimed raw digest/size, then requires that exact bound file to exist and verifies its real raw digest/size before reading the event;
9. the bound bytes must also match the candidate Git event bytes and all durable event/idempotency/content identities before compare-and-swap can run;
10. a missing, altered, expired, commit-substituted, path-substituted or otherwise mismatched quarantine capability cannot itself move authority.

The capability hash is not a secret and is not treated as authentication. Its security property is that preparation materialises the bound file only for the exact prepared H0/H1/path identity. A caller may calculate the ID for a forged candidate, but cannot make the required file exist through the candidate handle. The separate production deployment gate must prove that the ledger and sibling quarantine storage are service-owned and not writable by untrusted callers/processes; this filesystem assumption is not claimed to be solved by hashing.

The quarantine root is opened through Go's traversal-resistant `os.Root` API and its filesystem identity is checked while opening. Candidate filenames are restricted to safe IDs and candidate entries must remain regular files. Final publication uses root-relative hard-link creation so both the private source and published capability stay inside that pinned root.

## Lifecycle

Candidate quarantine retention is 24 hours. Prepare and acceptance prune entries older than that window, and `Reader.PruneCandidateQuarantine` is the explicit maintenance hook for the eventual service loop. Private publication files left by process death before final publication are also covered by bounded pruning.

Successful acceptance removes the exact bound quarantine entry after post-CAS replay verification. If a process crashes after CAS but before normal cleanup, an idempotent retry reconstructs the accepted result, derives the bound cleanup ID from the durable accepted event rather than a caller-supplied handle, and completes cleanup. Cleanup failure after authority has already moved is reported explicitly as `POST_ACCEPTANCE_QUARANTINE_CLEANUP_FAILED`; it is never represented as a rollback.

## Concurrent snapshot, cleanup and idempotency

Prepare and Accept both begin from an exact replay snapshot, but that snapshot can become stale while work is still in flight. An identical invocation may accept H1 and clean the shared bound quarantine entry after another request captured H0.

Acceptance therefore does not blindly return a candidate/quarantine failure observed after its initial H0/idempotency snapshot. It first performs a fresh bounded authoritative replay and idempotency lookup. Prepare likewise reconciles both a stale `PrepareEventCommit` and a final moved-authority check before returning a candidate. If its H0 has moved, it must not return a stale candidate, resurrect a cleaned accepted quarantine entry, or retain staging material abandoned because the exact request became durably accepted. Per-invocation staging IDs additionally ensure that one concurrent Prepare cannot remove another Prepare's temporary staging file.

For both paths:

- if the key is now durably bound to the same event ID and content digest, the operation resolves as `already_accepted` with the durable accepted commit and current replay snapshot;
- if the key is durably bound to different identity, the result is `IDEMPOTENCY_CONFLICT`;
- if no accepted event owns the key, the original candidate/stale/quarantine failure remains the result;
- if authoritative recovery itself cannot complete, Core returns an explicit `acceptance_unknown` response (`CONCURRENT_ACCEPTANCE_RECOVERY_REQUIRED` or `PREPARE_SNAPSHOT_RECOVERY_REQUIRED`) rather than pretending the old H0 observation proves that no concurrent acceptance occurred.

Missing, tampered or substituted quarantine material is still not evidence of acceptance. Only the durable authoritative ledger can convert a stale or pre-CAS failure into `already_accepted`.

## Post-CAS caller cancellation

Once `git update-ref` has been invoked, the request context is no longer treated as evidence that authority did or did not move. Core rechecks repository safety and resolves the authoritative ref under a fresh bounded recovery context. If Git moved H0 to H1 before caller cancellation became observable, the operation continues as an accepted transition. If successful CAS cannot be fully verified, it is reported as `POST_CAS_VERIFICATION_FAILED`; if the result of an attempted CAS cannot be determined, it is reported as `POST_CAS_RECOVERY_REQUIRED` rather than as an ordinary failed write. Ledger replay and idempotency recovery after a known acceptance also run under a fresh bounded context.

## Review status

Independent hostile review Issue #21 **FAILED** merged commit `747f30b4e2af0109f592220aa03b43e1ca1f0543` because quarantine was not bound to the exact prepared commit/path and caller cancellation could make a successful CAS look like an ordinary failure. Those defects were repaired in merged commit `bfe7686856ddec54c2be3e71aa8bc020d2b7a38e`.

Independent hostile re-review Issue #25 then **FAILED** that exact repaired commit on an exactly-once race: a winning identical request could accept H1 and remove Q after a losing request's H0 snapshot but before its quarantine read, causing the loser to return `CANDIDATE_INVALID` instead of `already_accepted`. That race and related Prepare/staging concurrency surfaces were repaired in merged commit `6710cb1b5f9d591f7e1653a5adc409581d34a858`.

Hostile re-review Issue #28 **FAILED** `6710cb1b5f9d591f7e1653a5adc409581d34a858` because the then-current `Ensure` could observe all bytes at the final filename and report success before the first creator had successfully completed `fsync`/close; a later creator failure could therefore remove material another Prepare had already treated as settled. The current repair removes in-progress writes from the final namespace entirely: publication occurs only from a completed private file through atomic no-overwrite hard-link creation, with deterministic failure injection proving a failed creator cannot invalidate another identical publisher's completed final capability.

This repair again changes the authority boundary and must receive fresh hostile review at its final exact merged commit. Internal tests, CI and self-review cannot convert the prior FAIL into PASS.

`AUTHORITY_WRITES_DISABLED` remains mandatory until the fresh review and all other release gates are separately satisfied.
