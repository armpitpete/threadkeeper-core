# Exclusive Governed Record Reducer v1

## Status

Candidate until merged to the default branch. When present on the default branch, this document defines the first Threadkeeper Core current-state reducer semantics.

## 1. Purpose

This reducer provides one deliberately narrow current-state model for governed objects that are explicitly declared to have **one exclusive current state**.

It is not a generic "latest value wins" mechanism.

It must not be used to collapse:

- observations;
- competing evidence;
- conflicting authoritative claims;
- testimony;
- hypotheses;
- source versions;
- any domain where multiple simultaneously valid records must be preserved.

Those remain separate records until a separately authorised governance event establishes a supersession or other effect.

## 2. Reducer family

The v1 event family contains exactly three event types:

- `core.record.created`
- `core.record.replaced`
- `core.record.revoked`

Any other event beginning with `core.record.` is an unknown reducer event and must fail closed.

Events outside the `core.record.` namespace are not inputs to this reducer and have no effect on this projection.

## 3. Policy binding

A `record_kind` may use this reducer only when an accepted authority/policy binding explicitly declares that kind to use the **exclusive-governed-record-v1** state model.

The reducer must never infer applicability from:

- the shape of a JSON object;
- a filename;
- a client request;
- an AI interpretation;
- the presence of `core.record.*` text alone.

Until such a binding exists for a record kind, transitions for that kind fail closed.

This contract defines reducer semantics; it does not itself authorise any concrete `record_kind`.

## 4. Target cardinality

Every v1 reducer event affects exactly one logical target.

The durable event envelope therefore must contain exactly one value in `targets` for these event types.

That target identifier is the stable projection key.

A target identifier is never recycled. A revoked target remains historically occupied.

## 5. Projected state

For an active target, the canonical logical state is:

```json
{
  "target_id": "...",
  "record_kind": "...",
  "status": "active",
  "revision": 1,
  "current_event_id": "...",
  "previous_event_id": null,
  "value": {}
}
```

For a revoked target, the canonical logical state is a tombstone:

```json
{
  "target_id": "...",
  "record_kind": "...",
  "status": "revoked",
  "revision": 3,
  "current_event_id": "...",
  "previous_event_id": "..."
}
```

The revoked state contains no `value` member. The historical value remains recoverable from prior accepted events and projections.

`value: null` is a valid active value and is therefore distinct from the absence of the `value` member in a revoked tombstone.

## 6. Absent-state assertion

Creation begins from explicit absence:

```json
{
  "exists": false,
  "target_id": "..."
}
```

This object is an assertion about the reducer projection. It is not a durable target state after creation.

## 7. Event-specific transition payload

The reducer consumes a validated event-specific transition payload in addition to the common durable event envelope.

### `core.record.created`

Transition payload:

```json
{
  "record_kind": "...",
  "value": {}
}
```

### `core.record.replaced`

Transition payload:

```json
{
  "record_kind": "...",
  "value": {}
}
```

`record_kind` must equal the current state's `record_kind`.

### `core.record.revoked`

Transition payload:

```json
{
  "record_kind": "..."
}
```

A revoke transition carries no replacement `value`.

The exact durable JSON Schema for the common envelope and these transition payloads is a later implementation artifact. The semantics in this document are normative for the reducer family.

## 8. `prior_state` and `resulting_state`

The durable event envelope already requires `prior_state` and `resulting_state`.

For this reducer they are **checked assertions**, not instructions.

The reducer must:

1. derive the actual current state from previously accepted events;
2. require the event's `prior_state` to equal that state exactly, or equal the explicit absent-state assertion for create;
3. independently compute the only permitted resulting state from the transition semantics;
4. require the event's `resulting_state` to equal that computed state;
5. fail closed on any mismatch.

A client therefore cannot smuggle an arbitrary state transition into the projection by supplying a desired `resulting_state`.

Comparison is over canonical JSON semantics using the accepted Threadkeeper strict JSON/JCS rules.

## 9. Create semantics

`core.record.created` is valid only when:

- exactly one target is named;
- the target has never had a projected state;
- the supplied `record_kind` has an accepted exclusive-reducer binding;
- `prior_state` is the exact absent-state assertion for that target;
- a transition `value` is present, including explicit JSON `null` if that is the intended value.

The reducer computes:

- `target_id` = sole event target;
- `record_kind` = transition `record_kind`;
- `status` = `active`;
- `revision` = `1`;
- `current_event_id` = this event's logical event ID;
- `previous_event_id` = `null`;
- `value` = transition value.

Creation of a target that already exists, including a revoked target, fails closed.

## 10. Replace semantics

`core.record.replaced` is valid only when:

- exactly one target is named;
- the target currently exists and is `active`;
- the transition `record_kind` equals the existing `record_kind`;
- that record kind remains bound to this reducer;
- `prior_state` exactly equals the current projected state;
- a transition `value` is present.

The reducer computes a new active state with:

- unchanged `target_id`;
- unchanged `record_kind`;
- `status` = `active`;
- `revision` = previous revision + 1;
- `current_event_id` = this event ID;
- `previous_event_id` = previous `current_event_id`;
- `value` = transition value.

The replacement value may be semantically identical to the previous value. That can still represent a meaningful governance transition because event-level provenance, authority policy, actor, source versions or reason may differ.

Replacement supersedes the prior state's **active effect**. It does not erase the prior event or prove the prior value false.

## 11. Revoke semantics

`core.record.revoked` is valid only when:

- exactly one target is named;
- the target currently exists and is `active`;
- the transition `record_kind` equals the existing `record_kind`;
- that record kind remains bound to this reducer;
- `prior_state` exactly equals the current projected state;
- no transition value is supplied.

The reducer computes a tombstone with:

- unchanged `target_id`;
- unchanged `record_kind`;
- `status` = `revoked`;
- `revision` = previous revision + 1;
- `current_event_id` = this event ID;
- `previous_event_id` = previous `current_event_id`;
- no `value` member.

Revocation ends current active effect. It does not delete history and does not necessarily assert that earlier content was factually false.

## 12. Terminal revocation

Revocation is terminal in v1.

After a target is revoked:

- it cannot be replaced;
- it cannot be revoked again;
- it cannot be recreated under the same target ID.

A future reinstatement/reopening transition, if needed, requires a separately reviewed event type and semantics. It must not be inferred by treating create as an "unrevoke" operation.

## 13. Determinism

For a fixed accepted event sequence and fixed accepted reducer/policy version, replay must produce byte-equivalent canonical projection output and the same projection digest.

The reducer may not depend on:

- wall-clock time during replay;
- random values;
- network calls;
- AI/model output;
- mutable external state;
- iteration order of unordered language maps;
- local timezone or locale.

## 14. Failure semantics

A reducer violation makes the governed-record projection invalid at that event.

Threadkeeper must stop that projection and expose a typed failure rather than skip the event and continue with a guessed state.

Minimum failure classes are:

- `REDUCER_NOT_APPLICABLE`
- `UNKNOWN_REDUCER_EVENT`
- `REDUCER_POLICY_UNBOUND`
- `REDUCER_TARGET_CARDINALITY`
- `REDUCER_ALREADY_EXISTS`
- `REDUCER_NOT_FOUND`
- `REDUCER_TERMINAL_STATE`
- `REDUCER_RECORD_KIND_MISMATCH`
- `REDUCER_PRIOR_STATE_MISMATCH`
- `REDUCER_RESULTING_STATE_MISMATCH`
- `REDUCER_TRANSITION_INVALID`

## 15. Relationship to future authority writes

These semantics are intentionally defined before authority-write implementation.

A future writer may prepare a candidate event only if it can prove that the exact same reducer applied at the expected ledger head yields the event's asserted `prior_state` and `resulting_state`.

The eventual Git compare-and-swap remains a separate acceptance boundary. Reducer validity alone never makes an event authoritative.

## 16. Relationship to evidence and conflict preservation

This reducer is appropriate for things such as a single active project setting, policy selection, or other object explicitly governed as exclusive state.

It is not appropriate for statements such as:

- "Source A says X";
- "Source B says not-X";
- "two measurements disagree";
- "two reviewers reached different conclusions".

Those are plural records. Replacing one with another would destroy evidence.

If governance later decides one record supersedes another for operational effect, that decision may itself be represented through an explicitly bound governed object while the underlying evidence remains preserved.
