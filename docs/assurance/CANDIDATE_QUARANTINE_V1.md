# Candidate Quarantine v1

Non-authoritative candidate material may contain sensitive data and must not be treated as harmless merely because it has not been accepted.

Before any public authority-write interface is enabled, candidate bytes must have a bounded quarantine lifecycle: private storage, content digest, explicit identity, no path traversal, explicit expiry/cleanup policy and promotion to authority only through the reviewed acceptance path.

## Integrated writer contract

The candidate writer now treats quarantine as part of the authority boundary:

1. the durable event is fully validated against the exact current ledger snapshot before candidate materialisation;
2. the exact validated canonical event bytes are stored in the ledger-bound private quarantine;
3. the quarantine entry records an explicit candidate identity, raw-byte SHA-256 and byte size;
4. Git candidate objects are created from the bytes read back from that quarantine entry, not directly from caller memory;
5. acceptance reopens the quarantine derived from the same pinned ledger identity and revalidates the stored entry;
6. the quarantined bytes must match the candidate Git event bytes and all durable event/idempotency/content identities before compare-and-swap can run;
7. a missing, altered, expired or substituted quarantine entry fails before authority can move.

The quarantine root is opened through Go's traversal-resistant `os.Root` API and its filesystem identity is checked while opening. Candidate filenames are restricted to safe IDs and candidate entries must remain regular files.

## Lifecycle

Candidate quarantine retention is 24 hours. Prepare and acceptance prune entries older than that window, and `Reader.PruneCandidateQuarantine` is the explicit maintenance hook for the eventual service loop.

Successful acceptance removes the quarantine entry after post-CAS replay verification. If a process crashes after CAS but before normal cleanup, an idempotent retry reconstructs the accepted result and completes cleanup. Cleanup failure after authority has already moved is reported explicitly as `POST_ACCEPTANCE_QUARANTINE_CLEANUP_FAILED`; it is never represented as a rollback.

## Authority boundary

This integration changes the candidate/CAS boundary and therefore requires fresh hostile exact-head review before public authority writes may be enabled. It does not itself enable a transport, actor authority, or public write operation.

`AUTHORITY_WRITES_DISABLED` remains mandatory until all remaining release gates are separately satisfied.
