# Read-Only Durable Ledger Acceptance Gates

## Status

Candidate until merged to the default branch. These gates define the first read-only durable-ledger implementation boundary.

## Scope

This lane may inspect, validate and derive replay projections from the durable Git ledger. It may not create commits, update refs, accept events, or expose any authority-changing operation.

## RO-01 — Exact authoritative head

Resolve the configured authoritative ref to an exact commit object ID and expose that ID in the replay result.

## RO-02 — Git integrity before replay

Run `git fsck --full --strict` before treating ledger content as replayable. Corruption fails closed.

## RO-03 — Controlled Git environment

Read operations ignore system/global Git configuration, replacement refs, pagers, editors, prompts, alternate object directories and authority-relevant ambient Git path variables.

## RO-04 — No shell execution

Git is executed directly with an argument vector. No read operation uses `sh -c`, PowerShell, `cmd /c`, or equivalent shell interpolation.

## RO-05 — Linear authoritative history

The v1 authoritative ledger history must be a single parent chain. Merge commits or discontinuous ancestry fail with `INTEGRITY_FAILURE`.

## RO-06 — Immutable accepted event paths

An event file under `events/` may be added once. Modification, deletion, rename or copy history for an accepted event path fails closed.

## RO-07 — Historical schema snapshot

Each event is validated using the schema registry present in the exact Git commit that accepted that event, not a later schema snapshot.

## RO-08 — Offline schema validation

Schema compilation and event validation use only locally registered Draft 2020-12 resources. Replay performs no network schema retrieval.

## RO-09 — Canonical stored event

Every accepted event blob must already be RFC 8785 canonical JSON. Recanonicalization must be byte-identical.

## RO-10 — Durable digest verification

Every replayed event must pass the accepted `content_sha256` omission/JCS/SHA-256 verification procedure.

## RO-11 — Deterministic event sequence

Replay order is Git commit acceptance order. Multiple event additions in one commit are ordered lexicographically by durable path solely to make the audit projection deterministic; this ordering does not create domain precedence semantics.

## RO-12 — Deterministic replay manifest

Two replays of the same exact ledger head produce the same ordered event manifest and replay SHA-256.

## RO-13 — Read-only CLI

`threadkeeper-core ledger-inspect <ledger.git> [ref]` may emit inspection/replay data but cannot alter the ledger.

## RO-14 — Authority writes remain disabled

The existing `threadkeeper-core authority-write` command must continue to fail with `AUTHORITY_WRITES_DISABLED` after this lane is merged.

## RO-15 — No invented current-state semantics

This lane does not infer how arbitrary event types modify governed current state. It projects validated acceptance history only. Event-type-specific state reducers require separately accepted schemas/semantics.

## Completion gate

The lane is complete only when exact-head CI proves healthy replay, corruption rejection, immutable-event rejection, digest failure detection, CGO-free build, clean module metadata, and the continuing hard authority-write disable gate.
