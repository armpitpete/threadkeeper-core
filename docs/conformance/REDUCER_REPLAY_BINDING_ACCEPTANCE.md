# Reducer Replay + Binding Acceptance Gates

## Status

Candidate until merged to the default branch.

## Scope

This lane integrates the accepted `exclusive-governed-record-v1` reducer into read-only ledger replay and defines its first machine-readable policy-binding/event-schema contract. It does not enable authority writes.

## RB-01 — Ledger-owned interpretation

Reducer replay must obtain schemas and bindings from the accepted ledger snapshot. Runtime embedded/reference schemas may not be used as a fallback for missing ledger data.

## RB-02 — Binding schema identity

Reducer bindings must validate against `urn:threadkeeper:schema:reducer-binding:v1`.

## RB-03 — Event schema identity

Exclusive governed-record events must validate against the exact event schema named by their accepted binding; v1 uses `urn:threadkeeper:schema:exclusive-governed-record-event:v1`.

## RB-04 — Canonical binding bytes

Stored reducer bindings must be strict RFC 8785 canonical JSON.

## RB-05 — Binding digest

Every reducer binding must pass `content_sha256` verification.

## RB-06 — Append-only schemas

Accepted files beneath `config/schemas/` may only be added. Modification, deletion or in-place replacement fails closed.

## RB-07 — Append-only bindings

Accepted files beneath `config/authority/reducer-bindings/` may only be added. Modification, deletion or in-place replacement fails closed.

## RB-08 — Unique schema IDs

Two accepted schema resources may not define the same `$id` in one historical snapshot.

## RB-09 — Unique binding IDs

A binding ID may occur only once in the accepted binding snapshot.

## RB-10 — One binding per record kind

A `record_kind` may have at most one reducer binding in v1. Rebinding requires a future migration contract.

## RB-11 — Explicit reducer model

The binding must name `exclusive-governed-record-v1`; unknown reducer models fail closed.

## RB-12 — Exact schema binding

A `core.record.*` event's `schema_version` must equal the selected binding's `event_schema`.

## RB-13 — Exact authority policy binding

A reducer event's `authority_policy_version` must equal the selected binding's `authority_policy_version`.

## RB-14 — Exact expected ledger parent

For an event accepted in commit H1 with parent H0, `expected_ledger_commit` must equal H0 exactly. Reducer events in a root commit fail.

## RB-15 — Historical snapshot

Schema and binding selection must occur at the exact accepting Git commit, never by looking only at the current head snapshot.

## RB-16 — Reducer application

Every `core.record.*` event must be applied through the accepted reducer. Unknown/malformed transitions may not be ignored.

## RB-17 — Checked assertions

Reducer `prior_state` and `resulting_state` must be independently checked as defined by `CURRENT_STATE_REDUCER_V1.md`.

## RB-18 — Idempotency uniqueness

No two accepted events may share the same non-empty `idempotency_key`.

## RB-19 — Deterministic projection

Two replays of the same exact ledger head must produce byte-equivalent canonical governed-record projections and the same projection SHA-256.

## RB-20 — Projection identity exposed

The replay manifest must expose reducer-binding count, governed-record count, governed-record projection and governed-record projection SHA-256.

## RB-21 — Replay identity covers reducer projection

The replay SHA-256 must incorporate the governed-record projection identity so a state-changing reducer result changes replay identity.

## RB-22 — Non-reducer events preserved

Events outside `core.record.*` remain in the audit replay sequence but do not silently mutate the governed-record projection.

## RB-23 — Failure is terminal for projection

A binding/schema/reducer invariant failure stops replay. Threadkeeper may not skip the invalid event and continue from guessed state.

## RB-24 — No authority side effects

This lane may not create ledger commits, update Git refs, authenticate acceptance requests or expose any authority-changing operation.

## RB-25 — Write gate remains disabled

`threadkeeper-core authority-write` must continue to fail with `AUTHORITY_WRITES_DISABLED` after this lane is merged.

## Completion gate

Exact-head CI must prove the full existing test suite plus bound lifecycle replay, missing-binding rejection, expected-head rejection, authority-policy mismatch rejection, schema/binding mutation rejection, duplicate-idempotency rejection, clean module metadata, CGO-free build, and the continuing hard authority-write disable gate.
