# Interface Contract

## 1. Principle

Threadkeeper Core exposes a **protocol-neutral evidence and governance boundary**.

The contract defines semantics, required fields and failure behaviour. It does not choose HTTP, RPC, MCP, CLI, sockets, message queues or any other transport.

## 2. Client classes

Clients may include humans, user interfaces, scripts, services and AI agents.

Trust is granted by authenticated role and authority policy, not by client type. An AI client receives no implicit privilege. A human-facing application receives no implicit privilege merely because a human is present.

The minimum logical capabilities are:

- **read** — retrieve records and evidence;
- **propose** — submit non-authoritative candidate material or requested changes;
- **decide** — record an authorised decision where policy permits;
- **operate** — perform administrative/rebuild functions without changing project truth unless separately authorised.

Implementations may use finer-grained permissions.

## 3. Evidence envelope

A read operation that returns project knowledge must return enough metadata to prevent naked text from being mistaken for truth.

The logical response must be able to expose:

- record identity;
- record type;
- authority class;
- source identity;
- exact immutable source version and/or digest;
- provenance/derivation chain where applicable;
- relevant timestamps;
- supersession state;
- conflict state;
- staleness or verification state where known;
- relationships used to construct the result;
- projection identity/version where the result is computed;
- retrieval score separately from authority, if a score exists.

Retrieval relevance and evidential authority must never share one undifferentiated confidence field.

## 4. Required read behaviours

The interface must support the logical ability to:

- retrieve a known record by stable identity;
- inspect its exact sources and provenance;
- query/search without losing authority metadata;
- retrieve current-state projections with their supporting history;
- enumerate materially conflicting records;
- inspect superseded and historical states;
- ask why a result is considered authoritative, derived, advisory or ephemeral;
- retrieve exact source-version references suitable for independent inspection.

A convenience interface may omit fields from its display, but the underlying evidence must remain obtainable.

## 5. Proposal contract

A proposal is explicitly non-authoritative until an authorised transition occurs.

A proposal submission must be attributable and must identify, where applicable:

- proposing actor/client;
- proposed content/change;
- target object/state;
- source evidence used;
- exact expected target version/state;
- rationale;
- creation time;
- proposal identity;
- idempotency identity or equivalent replay protection.

AI-generated proposals must be labelled as generated/derived rather than silently represented as observations.

## 6. Decision contract

A decision-changing write must require an authority policy match.

The logical request must include or resolve:

- authenticated decision actor/mechanism;
- action being authorised;
- exact target;
- expected prior state/version;
- proposal or evidence reference where applicable;
- decision reason/reference;
- idempotency identity;
- required durable authoritative sink.

The operation must fail if the caller lacks authority, the expected state has changed, the target is materially ambiguous, or durable persistence cannot be completed.

## 7. Optimistic state protection

Protected writes must support an **expected-state precondition** or equivalent mechanism.

If a caller reasons over version A and the governed target has moved to version B, Threadkeeper must reject the stale write rather than silently applying the decision to B.

This applies to AI and non-AI clients equally.

## 8. Idempotency and replay

Authority-changing writes must be safely retryable without accidentally recording the same decision twice.

The interface must therefore support an idempotency key, stable decision identity or equivalent replay-protection mechanism.

## 9. Failure semantics

Protected operations must fail closed when any material requirement is unresolved, including:

- unknown authority;
- ambiguous source identity;
- stale expected state;
- incomplete provenance required by policy;
- unresolved required conflict;
- unavailable authoritative sink;
- failed authentication/authorisation;
- partial durable write.

Failures must be inspectable. The interface must not convert these conditions into a successful-looking natural-language answer.

## 10. AI-specific boundary

AI systems may:

- search and retrieve;
- request evidence;
- derive summaries and relationships;
- detect possible conflicts;
- submit proposals;
- prepare decision candidates;
- invoke authorised operations only where an independently configured policy grants that exact capability.

AI systems must not, merely by being AI systems:

- promote their own output to authority;
- forge or infer human acceptance;
- hide provenance;
- reinterpret a stale mutable reference as the exact version originally reviewed;
- overwrite conflicting authoritative evidence;
- make Recall data the sole durable record of a governed decision.

## 11. Explainability contract

For any governed answer or current-state projection, a client must be able to request a machine-readable explanation equivalent to:

1. **What is the result?**
2. **What authority class is it?**
3. **Which exact records support it?**
4. **How was it derived?**
5. **What conflicts or uncertainty remain?**
6. **What decision event, if any, made it accepted?**

Natural-language explanation is optional. The underlying provenance is not.

## 12. Compatibility and versioning

The interface contract itself must be versioned.

A client must be able to discover the contract/schema version it is speaking to. Breaking semantic changes require an explicit version change or migration path; they must not silently change the meaning of authority, provenance or decisions.

## 13. No implementation selection

Nothing in this contract selects a network protocol, serialization format, database, programming language, AI provider or deployment architecture. Those choices must be evaluated later against these requirements.
