# Candidate Quarantine v1

Non-authoritative candidate material may contain sensitive data and must not be treated as harmless merely because it has not been accepted.

Before any public authority-write interface is enabled, candidate bytes must have a bounded quarantine lifecycle: private storage, content digest, explicit identity, no path traversal, explicit expiry/cleanup policy and promotion to authority only through the reviewed acceptance path.

## Integrated writer contract

The candidate writer treats quarantine as part of the authority boundary:

1. the durable event is fully validated against the exact current ledger snapshot before candidate materialisation;
2. the exact validated canonical event bytes are first staged in the ledger-bound private quarantine;
3. Git candidate objects are created only from bytes read back from that staging entry;
4. after H1 is created and revalidated, preparation finalises the same raw bytes under a quarantine capability ID derived from the complete prepared identity: expected H0, prepared H1, event path, event ID, idempotency key, semantic content digest, raw-byte SHA-256 and raw byte size;
5. the temporary staging entry is removed before a prepared handle is returned;
6. immediately before returning a prepared candidate, Prepare rechecks current authority; if H0 moved, it reconciles against a fresh bounded replay instead of returning an already-stale candidate;
7. acceptance recomputes the expected capability ID from the untrusted handle and its claimed raw digest/size, then requires that exact bound file to exist and verifies its real raw digest/size before reading the event;
8. the bound bytes must also match the candidate Git event bytes and all durable event/idempotency/content identities before compare-and-swap can run;
9. a missing, altered, expired, commit-substituted, path-substituted or otherwise mismatched quarantine capability cannot itself move authority.

The capability hash is not a secret and is not treated as authentication. Its security property is that preparation materialises the bound file only for the exact prepared H0/H1/path identity. A caller may calculate the ID for a forged candidate, but cannot make the required file exist through the candidate handle. The separate production deployment gate must prove that the ledger and sibling quarantine storage are service-owned and not writable by untrusted callers/processes; this filesystem assumption is not claimed to be solved by hashing.

The quarantine root is opened through Go's traversal-resistant `os.Root` API and its filesystem identity is checked while opening. Candidate filenames are restricted to safe IDs and candidate entries must remain regular files.

## Lifecycle

Candidate quarantine retention is 24 hours. Prepare and acceptance prune entries older than that window, and `Reader.PruneCandidateQuarantine` is the explicit maintenance hook for the eventual service loop.

Successful acceptance removes the exact bound quarantine entry after post-CAS replay verification. If a process crashes after CAS but before normal cleanup, an idempotent retry reconstructs the accepted result, derives the bound cleanup ID from the durable accepted event rather than a caller-supplied handle, and completes cleanup. Cleanup failure after authority has already moved is reported explicitly as `POST_ACCEPTANCE_QUARANTINE_CLEANUP_FAILED`; it is never represented as a rollback.

## Concurrent snapshot, cleanup and idempotency

Prepare and Accept both begin from an exact replay snapshot, but that snapshot can become stale while work is still in flight. An identical invocation may accept H1 and clean the shared bound quarantine entry after another request captured H0.

Acceptance therefore does not blindly return a candidate/quarantine failure observed after its initial H0/idempotency snapshot. It first performs a fresh bounded authoritative replay and idempotency lookup. Prepare likewise reconciles a stale `PrepareEventCommit` and performs a final current-authority check before returning a candidate. If its H0 has moved, it must not return a stale candidate or leave a newly materialised accepted quarantine entry behind.

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

Independent hostile re-review Issue #25 then **FAILED** that exact repaired commit on an exactly-once race: a winning identical request could accept H1 and remove Q after a losing request's H0 snapshot but before its quarantine read, causing the loser to return `CANDIDATE_INVALID` instead of `already_accepted`. The current repair adds fresh snapshot reconciliation to both acceptance and equivalent Prepare/Accept races, with deterministic tests for winner cleanup before a losing acceptance read and winner acceptance before a racing Prepare's final authority check.

This repair is another authority-boundary candidate and must receive fresh independent hostile review at its final exact merged commit. Internal tests, CI and self-review cannot satisfy that gate.

`AUTHORITY_WRITES_DISABLED` remains mandatory until the fresh independent review and all other release gates are separately satisfied.
