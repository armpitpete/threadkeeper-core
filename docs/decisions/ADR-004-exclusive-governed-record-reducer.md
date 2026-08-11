# ADR-004 — Exclusive Governed Record Reducer v1

## Status

This ADR is accepted when present on the repository default branch. An unmerged copy is a candidate only.

## Context

Threadkeeper Core can now validate and deterministically replay accepted durable ledger events, but replay intentionally does not invent domain-specific current-state meaning.

Before an authority-write path can be designed safely, Core needs at least one explicit state-transition model that answers:

- what an event is allowed to change;
- how current state is derived;
- what counts as a stale or contradictory transition;
- which invariants must hold across replay;
- what must remain historical rather than being overwritten.

A generic latest-event-wins reducer would be unsafe because Threadkeeper must preserve conflicting evidence and distinguish authority from recency.

## Decision

Adopt **exclusive-governed-record-v1** as the first current-state reducer family.

It is opt-in by accepted authority/policy binding and applies only to record kinds declared to have one exclusive current state.

The family defines exactly:

- `core.record.created`
- `core.record.replaced`
- `core.record.revoked`

The state key is one stable logical target ID. Creation establishes revision 1. Replacement advances revision by one and preserves lineage to the prior current event. Revocation advances revision by one and produces a terminal tombstone. Revoked target IDs cannot be reused in v1.

`prior_state` and `resulting_state` are verified assertions. The reducer independently derives current state and the permitted result; it never applies arbitrary client-supplied state objects.

The reducer is deterministic and pure with respect to the accepted event sequence and accepted policy/reducer version.

## Scope restriction

This reducer must not be used for observations, competing evidence, conflicting claims or any record set whose cardinality is legitimately greater than one.

Replacement changes active governance effect; it does not erase or falsify historical evidence.

No concrete record kind becomes authorised merely because this ADR exists. Each kind requires a separate accepted binding to this state model.

## Consequences

### Positive

- Current-state semantics become inspectable and testable before writes exist.
- Future candidate writes can be validated against the same reducer used for replay.
- Exact prior-state assertions expose stale or contradictory transitions.
- History remains append-only; current state is a projection.
- Revocation cannot be silently treated as deletion or recreation.
- Evidence/conflict preservation is protected from accidental latest-value semantics.

### Costs

- Reinstatement after revocation requires a future explicitly designed event type.
- Multi-target atomic reducer events are deferred.
- Policy binding machinery must exist before real record kinds use the reducer.
- Other domain state models will require separate reducers rather than being forced into this one.

## Rejected alternatives

### Latest event wins

Rejected because timestamp/order alone is not authority and would collapse conflicts.

### Generic patch events

Rejected for v1 because arbitrary JSON Patch-style mutation makes invariants and audit meaning harder to reason about and can make resulting state client-controlled.

### Delete events

Rejected because deletion conflicts with Threadkeeper's requirement to preserve accepted history. Revocation creates a tombstone instead.

### Automatic recreation after revocation

Rejected because it creates ambiguous lifecycle semantics. Reopening must be explicit if later required.

### Universal reducer for all record types

Rejected because evidence, observations, relationships and plural authoritative records have different cardinality and conflict semantics.

## Non-decision

This ADR does not enable authority writes, select concrete governed record kinds, define client authentication, or replace the protected Git compare-and-swap acceptance boundary.
