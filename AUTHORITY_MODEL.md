# Authority Model

## 1. Principle

Threadkeeper Core does not decide truth by confidence, frequency, model agreement or retrieval rank.

Authority is a property of **declared sources, exact source versions and explicit authorised decisions**.

## 2. Record types

Threadkeeper must be able to distinguish at least these logical record types:

- **Source** — an external or Threadkeeper-governed origin of information.
- **Source Version** — an immutable identity for one state of a source.
- **Observation** — a fact observed from a source version.
- **Derived Record** — material computed, extracted, summarised or inferred from one or more source versions.
- **Proposal** — a suggested change, interpretation or decision that has not been accepted.
- **Decision Event** — an attributable authorised act that changes governed project state.
- **Projection** — a computed view such as current status, latest accepted version or unresolved conflict.

These distinctions are semantic contracts and do not prescribe database tables or classes.

## 3. Authority classes

Every retrievable record must expose an authority class. The minimum classes are:

### AUTHORITATIVE

May establish governed project truth because its source/version or decision event is permitted by declared authority policy.

### DERIVED

Computed from other records. It may be useful and reproducible, but it cannot silently inherit authority from its inputs.

### ADVISORY

Human or machine material intended to guide reasoning but not establish project truth.

### EPHEMERAL

Temporary working material, caches or session state that must never be treated as durable project truth.

Additional classes may be introduced later, but these meanings must not be weakened.

## 4. Authority policy

Authority must be explicit and inspectable. A policy must be able to answer:

- which source is permitted to establish which kind of truth;
- which exact version was used;
- who or what may record an acceptance or revocation;
- where the resulting authoritative record is durably preserved;
- what conditions make an authority claim stale, superseded or disputed.

Changing authority policy is itself a governed change.

## 5. No silent promotion

The following must never cause automatic promotion to AUTHORITATIVE:

- an AI generated the record;
- multiple models agree;
- a record has high confidence or retrieval score;
- a statement appears repeatedly;
- a derived record is based only on authoritative inputs;
- a client labels something “final” or “approved” without an authorised decision event;
- a branch, tag, filename or URL appears current without an exact immutable identity.

## 6. Explicit transitions

An authority-changing transition must record:

- the affected object or proposition;
- prior authority/state where applicable;
- resulting authority/state;
- authorised actor or mechanism;
- exact time;
- reason or decision reference;
- exact source versions involved;
- destination authoritative sink;
- idempotency identity or equivalent replay protection.

If governance requires human acceptance, an AI client may prepare the proposal but cannot impersonate that acceptance.

## 7. Provenance

Every DERIVED record must preserve enough lineage to reproduce or audit it. At minimum:

- source identifiers;
- immutable source versions or content digests;
- transformation identity/version;
- creation time;
- producing actor/tool identity;
- relationships to intermediate derived records where relevant.

A summary must remain distinguishable from the text it summarises. An inference must remain distinguishable from an observation.

## 8. Conflict model

Threadkeeper must preserve materially conflicting authoritative evidence.

It must not resolve conflict by overwriting an older record or by selecting the most recent timestamp unless authority policy explicitly defines that rule.

A conflict view must expose:

- the competing records;
- their authority classes;
- exact versions;
- known temporal relationship;
- supersession information if any;
- whether the conflict is resolved, unresolved or intentionally preserved.

## 9. Current state

“Current” is a projection over history, not a replacement for history.

A current-state result must be traceable to the exact authoritative events and versions from which it was projected.

Mutable references such as branch names may be convenient locators, but they are insufficient as sole evidence of exact state.

## 10. Decisions and durable truth

A Decision Event that matters to project governance must not exist only inside disposable Recall data or an AI conversation.

Before such a decision is treated as durable project truth, it must be persisted in a configured authoritative sink from which Threadkeeper can recover it.

## 11. Uncertainty

Threadkeeper must be able to represent:

- unknown;
- conflicting;
- stale;
- incomplete;
- unverified;
- proposed but unaccepted.

It must not collapse these states into a confident single answer merely to make retrieval simpler.
