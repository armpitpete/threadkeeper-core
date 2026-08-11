# Reducer v1 Acceptance Gates

## Status

Candidate until merged to the default branch.

These gates define conformance for the first Threadkeeper Core current-state reducer family.

## R-01 — Opt-in policy binding

No record kind may use the exclusive reducer without an accepted binding to `exclusive-governed-record-v1`.

## R-02 — Exact event family

Only `core.record.created`, `core.record.replaced` and `core.record.revoked` are defined. Unknown `core.record.*` events fail closed.

## R-03 — Unrelated events do not mutate projection

Events outside the `core.record.` namespace are not applicable to this reducer and cannot alter its state.

## R-04 — Single target

Every reducer event contains exactly one target.

## R-05 — Stable target identity

The sole target identifier is the state key and may not change across replacements or revocation.

## R-06 — Target IDs are never recycled

Create fails if the target has ever been created, including when its current lifecycle state is revoked.

## R-07 — Create from explicit absence

Create requires an exact absent-state assertion and computes revision 1 active state.

## R-08 — Replace active only

Replace requires an existing active target. Missing or revoked targets fail closed.

## R-09 — Revoke active only

Revoke requires an existing active target. Missing or revoked targets fail closed.

## R-10 — Record kind immutability

`record_kind` is established at create and cannot change for the target.

## R-11 — Sequential revision

Create sets revision 1. Every valid replace or revoke increments revision by exactly one.

## R-12 — Exact lineage

Every transition after create sets `previous_event_id` to the immediately prior current event ID and `current_event_id` to the applying event ID.

## R-13 — Prior state is an assertion

The event's `prior_state` must canonically equal the actual reducer state immediately before the event. Create uses the explicit absent-state assertion.

## R-14 — Resulting state is derived

The reducer computes the permitted result from current state and transition payload. The event's `resulting_state` must canonically equal that computed result.

## R-15 — Revocation produces a tombstone

A revoked state contains lifecycle identity, revision and lineage but no `value` member. Historical value remains recoverable from accepted history.

## R-16 — Revocation is terminal in v1

No create, replace or revoke transition may advance a revoked target.

## R-17 — Explicit null remains a value

An active record may have JSON `null` as its value. This is distinct from the absent `value` member of a revoked tombstone.

## R-18 — Same-value replacement is permitted

A replace event may establish the same semantic value as before because event-level authority/provenance may have changed. It must still advance revision and lineage normally.

## R-19 — Deterministic pure reduction

The same ordered accepted event sequence under the same accepted reducer/policy version must produce the same canonical projection and digest without time, randomness, network or AI dependence.

## R-20 — Failure is non-mutating

If an event fails reducer validation, no partial state mutation is observable.

## R-21 — Evidence is out of scope

The reducer must not be automatically applied to observations, competing evidence, conflicting claims or other plural record sets.

## R-22 — No authority implication

Successful reduction proves semantic validity only. It does not make an event accepted or authoritative; future writer acceptance still requires the protected durable-ledger procedure and exact-head compare-and-swap.

## Required executable tests

The reference implementation must prove at least:

1. create -> replace -> revoke produces revisions 1 -> 2 -> 3 and correct lineage;
2. duplicate create fails;
3. missing-target replace fails;
4. missing-target revoke fails;
5. replace after revoke fails;
6. recreate after revoke fails;
7. repeated revoke fails;
8. record-kind change fails;
9. prior-state mismatch fails;
10. resulting-state mismatch fails;
11. zero or multiple targets fail;
12. unbound record kinds fail;
13. unknown `core.record.*` event type fails;
14. unrelated event type reports not-applicable without mutation;
15. explicit-null active value is preserved;
16. same-value replacement is allowed and advances lineage;
17. failure leaves input projection unchanged;
18. equivalent reductions produce canonical-equivalent output.

## Completion gate

This lane is complete when:

- the normative reducer contract and ADR are reviewable;
- a pure reference reducer encodes the semantics;
- the required tests pass on exact-head CI;
- module metadata remains clean;
- the CGO-free Core build remains green;
- `authority-write` continues to fail with `AUTHORITY_WRITES_DISABLED`.
