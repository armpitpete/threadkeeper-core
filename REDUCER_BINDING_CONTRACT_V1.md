# Reducer Binding and Event Contract v1

## Status

Candidate until merged to the default branch. When present on the default branch, this document defines the machine-readable binding between accepted ledger policy, the `exclusive-governed-record-v1` reducer, and its v1 durable event schema.

## 1. Purpose

Reducer semantics alone do not say which governed objects may use them. Threadkeeper therefore requires an accepted machine-readable binding before a `core.record.*` event can affect current-state projection.

The binding is part of durable governance truth. It is not Recall data, a client preference, an AI interpretation, or a filename convention.

## 2. Contract identities

The v1 reducer-binding schema ID is:

```text
urn:threadkeeper:schema:reducer-binding:v1
```

The v1 exclusive governed-record event schema ID is:

```text
urn:threadkeeper:schema:exclusive-governed-record-event:v1
```

The state-model identity remains:

```text
exclusive-governed-record-v1
```

Reference JSON Schemas are stored in this implementation repository under `schemas/`. Runtime replay does **not** use those files as a fallback. A durable ledger must contain its accepted schema snapshots beneath `config/schemas/`.

## 3. Reducer binding record

A reducer binding is an immutable canonical JSON record beneath:

```text
config/authority/reducer-bindings/
```

Its logical fields are:

```json
{
  "schema_version": "urn:threadkeeper:schema:reducer-binding:v1",
  "binding_id": "...",
  "record_kind": "...",
  "state_model": "exclusive-governed-record-v1",
  "event_schema": "urn:threadkeeper:schema:exclusive-governed-record-event:v1",
  "authority_policy_version": "...",
  "content_sha256": "..."
}
```

The stored record must be strict JSON, RFC 8785 canonical bytes, valid against its accepted historical schema, and pass the Threadkeeper `content_sha256` omission/JCS/SHA-256 procedure.

## 4. One binding per record kind in v1

A `record_kind` may have at most one reducer binding in v1.

Likewise, `binding_id` is globally unique within the reducer-binding snapshot.

Threadkeeper v1 does not implement rebinding, reducer migration, event-schema replacement, or authority-policy rebinding for an existing record kind. Such a change requires separately reviewed migration/governance semantics.

Adding a second binding for an already-bound record kind fails closed.

## 5. Immutable versioned configuration

Files beneath these v1 paths are append-only:

```text
config/schemas/
config/authority/reducer-bindings/
```

An accepted file may not later be modified, deleted, renamed, or replaced in place.

New schema identities may be added as new files. Existing schema `$id` values may not be silently redefined; loading two schema resources with the same `$id` fails closed.

This restriction prevents historical interpretation from changing because a later commit edited a schema or policy binding.

## 6. Event envelope

A v1 exclusive governed-record event contains the common authority-relevant envelope plus reducer transition data.

Required fields are:

- `schema_version`;
- `event_id`;
- `event_type`;
- `occurred_at`;
- `actor`;
- `expected_ledger_commit`;
- `authority_policy_version`;
- `targets`;
- `source_versions`;
- `record_kind`;
- `prior_state`;
- `resulting_state`;
- `reason`;
- `idempotency_key`;
- `content_sha256`;
- `value` for create/replace only.

The exact machine constraints are defined by `urn:threadkeeper:schema:exclusive-governed-record-event:v1`.

No additional top-level members are permitted in v1.

## 7. Event types and value rule

The schema permits exactly:

- `core.record.created`;
- `core.record.replaced`;
- `core.record.revoked`.

Create and replace require a `value`. Explicit JSON `null` is a valid value.

Revoke forbids the `value` member entirely.

Reducer transition semantics remain defined by `CURRENT_STATE_REDUCER_V1.md`.

## 8. Exact expected-head assertion

For a reducer event accepted in Git commit `H1` with parent `H0`:

```text
event.expected_ledger_commit MUST equal H0
```

A reducer event in a root commit is invalid because there is no prior authoritative ledger state to assert.

During read-only recovery this rule is rechecked from Git history. The event is therefore auditable as having been constructed against the exact parent state, independently of future writer implementation.

## 9. Authority-policy binding

For an event with `record_kind = K`, replay selects the unique accepted binding for `K` at the exact accepting commit.

The event is valid for reducer application only if:

```text
event.schema_version == binding.event_schema
```

and:

```text
event.authority_policy_version == binding.authority_policy_version
```

The reducer name must also be exactly `exclusive-governed-record-v1`.

No client, AI, path name, or event payload may override the accepted binding.

## 10. Historical interpretation

Each event is interpreted using:

1. the exact Git commit that accepted it;
2. schemas present in that commit;
3. reducer bindings present in that commit;
4. the accepted deterministic reducer implementation/semantics.

Later additions cannot retroactively change how an earlier event was interpreted.

## 11. Idempotency reconstruction

The v1 reducer event schema requires a non-empty `idempotency_key`.

Read-only replay reconstructs the accepted idempotency set from durable history. More than one accepted event with the same non-empty key is an integrity failure.

This does not yet implement writer-side retry handling. It establishes the durable invariant that the future writer must preserve.

## 12. Deterministic current-state projection

Replay applies valid bound reducer events in accepted Git history order.

The replay manifest exposes:

- reducer-binding count;
- governed-record count;
- governed-record projection;
- SHA-256 of the canonical governed-record projection;
- the existing audit event sequence and replay SHA-256.

For the same exact ledger head and accepted software semantics, these values must be deterministic.

## 13. Failure model

Replay fails closed on at least:

- missing reducer binding;
- duplicate `binding_id`;
- multiple bindings for one `record_kind`;
- unknown reducer model;
- event-schema mismatch;
- authority-policy-version mismatch;
- expected-ledger-commit mismatch;
- duplicate idempotency key;
- reducer prior/resulting-state mismatch;
- mutation/deletion of accepted schema or binding files;
- schema, canonicalization, or digest failure.

Threadkeeper must not skip a failed reducer event and continue with a guessed current state.

## 14. Authority boundary

This contract is read-only with respect to acceptance.

Successful schema validation, binding selection and reducer application prove only that already-accepted history is internally interpretable under the v1 contract.

They do not:

- construct an accepted Git commit;
- advance the authoritative ref;
- authenticate a client decision;
- authorise an actor;
- perform compare-and-swap;
- enable `authority-write`.

Those remain future protected lanes.

## 15. Migration boundary

Future support for any of the following requires a new reviewed contract:

- rebinding an existing `record_kind`;
- replacing its event schema;
- changing its state model;
- changing the bound authority-policy version;
- reopening a revoked target;
- migrating an existing projection to a new reducer family.

The v1 implementation must fail rather than infer those semantics.
