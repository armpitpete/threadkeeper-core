# Candidate Quarantine v1

Non-authoritative candidate material may contain sensitive data and must not be treated as harmless merely because it has not been accepted.

Before any public authority-write interface is enabled, candidate bytes must have a bounded quarantine lifecycle: private storage, content digest, explicit identity, no path traversal, explicit expiry/cleanup policy and promotion to authority only through the reviewed acceptance path.

## Integrated writer contract

The candidate writer treats quarantine as part of the authority boundary:

1. the durable event is fully validated against the exact current ledger snapshot before candidate materialisation;
2. the exact validated canonical event bytes are first staged in the ledger-bound private quarantine under a per-Prepare random staging identity, so identical concurrent prepares do not share temporary cleanup ownership;
3. Git candidate objects are created only from bytes read back from that staging entry;
4. after H1 is created and revalidated, preparation finalises the same raw bytes under a quarantine capability ID derived from the complete prepared identity: expected H0, prepared H1, event path, event ID, idempotency key, semantic content digest, raw-byte SHA-256 and raw byte size;
5. final capability publication is never performed by writing directly to the final filename. Each publisher writes to a private file inside the pinned quarantine root, completes write + file `fsync` + close, atomically hard-links that completed file to the final capability name without overwrite, and syncs the pinned quarantine directory before reporting durable publication success. An identical caller that encounters an already-visible matching final name establishes directory durability itself and re-reads/revalidates before converging;
6. a failed publisher owns and removes only its private publication file. It cannot remove another publisher's final capability. Private publication residue left by abrupt process death is retention cleanup rather than authority state and remains subject to bounded pruning;
7. the temporary staging entry is removed before a prepared handle is returned; if H0 moved before candidate construction or during final bound materialisation and the exact request is recovered as already accepted, the abandoned staging entry is removed on that recovery path as well;
8. immediately before returning a prepared candidate, Prepare rechecks current authority; if H0 moved, it reconciles against a fresh bounded replay instead of returning an already-stale candidate;
9. acceptance recomputes the expected capability ID from the untrusted handle and its claimed raw digest/size, then requires that exact bound file to exist and verifies its real raw digest/size before reading the event;
10. the bound bytes must also match the candidate Git event bytes and all durable event/idempotency/content identities before compare-and-swap can run;
11. a missing, altered, expired, commit-substituted, path-substituted or otherwise mismatched quarantine capability cannot itself move authority.

The capability hash is not a secret and is not treated as authentication. Its security property is that preparation materialises the bound file only for the exact prepared H0/H1/path identity. A caller may calculate the ID for a forged candidate, but cannot make the required file exist through the candidate handle. The separate production deployment gate must prove that the ledger and sibling quarantine storage are service-owned and not writable by untrusted callers/processes; this filesystem assumption is not claimed to be solved by hashing.

The quarantine root is opened through Go's traversal-resistant `os.Root` API and its filesystem identity is checked while opening. Candidate filenames are restricted to safe IDs and candidate entries must remain regular files. Final publication uses root-relative hard-link creation so both the private source and published capability stay inside that pinned root. Filesystems that cannot provide the required file-sync, directory-sync and no-overwrite hard-link semantics fail closed and are not valid production targets without a separately reviewed storage design.

## Lifecycle

Candidate quarantine retention is 24 hours. Prepare and acceptance prune entries older than that window, and `Reader.PruneCandidateQuarantine` is the explicit maintenance hook for the eventual service loop. Private publication files left by process death before final publication are also covered by bounded pruning.

Successful acceptance removes the exact bound quarantine entry after post-CAS replay verification. If a process crashes after CAS but before normal cleanup, an idempotent retry reconstructs the accepted result, derives the bound cleanup ID from the durable accepted event rather than a caller-supplied handle, and completes cleanup. Cleanup failure after authority has already moved is reported explicitly as `POST_ACCEPTANCE_QUARANTINE_CLEANUP_FAILED`; it is never represented as a rollback.

## Concurrent snapshot, materialisation, cleanup and idempotency

Prepare and Accept both begin from an exact replay snapshot, but that snapshot can become stale while work is still in flight. An identical invocation may accept H1 and clean the shared bound quarantine entry after another request captured H0.

Acceptance therefore does not blindly return a candidate/quarantine failure observed after its initial H0/idempotency snapshot. It first performs a fresh bounded authoritative replay and idempotency lookup.

Prepare applies the same rule across the final bound-materialisation window. A failure finalising or reading the exact bound capability after the original H0 snapshot is reconciled against fresh durable authority before it is returned. This covers the schedule in which bound `Ensure` succeeds, another identical request accepts H1 and deletes Q, and the losing Prepare then reaches bound `Read`. Missing material alone is never evidence of acceptance.

Prepare also reconciles a stale `PrepareEventCommit` and performs a final moved-authority check before returning a candidate. If its H0 has moved, it must not return a candidate that it already knows is stale, resurrect a cleaned accepted quarantine entry, or retain staging material abandoned because the exact request became durably accepted. Per-invocation staging IDs additionally ensure that one concurrent Prepare cannot remove another Prepare's temporary staging file.

For these recovery paths:

- if the key is now durably bound to the same event ID and content digest, the operation resolves as `already_accepted` with the durable accepted commit and current replay snapshot;
- if the key is durably bound to different identity, the result is `IDEMPOTENCY_CONFLICT`;
- if no accepted event owns the key, the original candidate/stale/quarantine failure remains the result;
- if authoritative recovery itself cannot complete, Core returns an explicit `acceptance_unknown` response (`CONCURRENT_ACCEPTANCE_RECOVERY_REQUIRED` or `PREPARE_SNAPSHOT_RECOVERY_REQUIRED`) rather than pretending the old H0 observation proves that no concurrent acceptance occurred.

Missing, tampered or substituted quarantine material is still not evidence of acceptance. Only the durable authoritative ledger can convert a stale or pre-CAS failure into `already_accepted`.

## CAS contention and post-CAS recovery

Once `git update-ref` has been invoked, the caller context is no longer treated as evidence that authority did or did not move. Core resolves the authoritative ref under fresh bounded recovery contexts.

If `update-ref` reports an error and recovery finds H1 in the authoritative history, the result is explicit recovered acceptance and the ledger layer reconstructs the durable request as `already_accepted`.

If `update-ref` reports an error while the first detached recovery still sees H0, that H0 observation is not treated as proof of failure: another legitimate writer may still own the Git ref lock and complete H0→H1 immediately afterwards. Core therefore performs a bounded detached settle/recovery window:

- H1 appears in authoritative history: recovered acceptance;
- a different exact-head winner appears and H1 is absent: `STALE_STATE` and normal idempotency/conflict recovery;
- H0 remains unresolved through the bounded window, or recovery itself fails: `POST_CAS_RECOVERY_REQUIRED` and API status `acceptance_unknown`, never an ordinary nil-response write failure.

If `update-ref` itself succeeds, a current H1 or later linear H2 containing H1 proves the original acceptance. Failure to verify a successful CAS is `POST_CAS_VERIFICATION_FAILED`; it is never represented as rollback or ordinary failure. Ledger replay and idempotency recovery after known or recovered acceptance likewise use caller-independent bounded contexts.

## Review status

Independent hostile review Issue #21 **FAILED** merged commit `747f30b4e2af0109f592220aa03b43e1ca1f0543` because quarantine was not bound to the exact prepared commit/path and caller cancellation could make a successful CAS look like an ordinary failure. Those defects were repaired in merged commit `bfe7686856ddec54c2be3e71aa8bc020d2b7a38e`.

Independent hostile re-review Issue #25 then **FAILED** that exact repaired commit on an exactly-once race: a winning identical request could accept H1 and remove Q after a losing request's H0 snapshot but before its quarantine read, causing the loser to return `CANDIDATE_INVALID` instead of `already_accepted`. That race and related Prepare/staging concurrency surfaces were repaired in merged commit `6710cb1b5f9d591f7e1653a5adc409581d34a858`.

Hostile re-review Issue #28 **FAILED** `6710cb1b5f9d591f7e1653a5adc409581d34a858` because the then-current `Ensure` could observe all bytes at the final filename and report success before the first creator had successfully completed file sync/close; a later creator failure could therefore remove material another Prepare had already treated as settled. Private write/sync/close plus no-overwrite hard-link publication was merged as `a7214cbbbfb8d28732c5aff48eeb78bbe4103d52`.

Independent review Issue #31 then **FAILED** the corresponding publication code because final-name visibility was not yet crash durability: the quarantine directory was not synced after the hard-link namespace mutation. The same review also required the whole-tree race gate to tolerate Git versions that reject the hostile `refstorage` fixture earlier while still failing closed.

Hostile review Issue #32 **FAILED** merged commit `a7214cbbbfb8d28732c5aff48eeb78bbe4103d52` on two further concurrency windows: Prepare could return ordinary `CANDIDATE_INVALID` if a winner removed Q after its bound `Ensure` but before bound `Read`, and a losing identical CAS could return an ordinary Git error when ref-lock contention caused its immediate recovery to observe H0 before the winner published H1.

The current PR #34 candidate consolidates the Issue #31 crash-durability/race-gate repair and both Issue #32 concurrency repairs. Exact-head CI, race detection and internal hostile review remain evidence only; they cannot satisfy the required fresh independent hostile review of the final merged commit.

`AUTHORITY_WRITES_DISABLED` remains mandatory until that fresh independent review and all other release gates are separately satisfied.
