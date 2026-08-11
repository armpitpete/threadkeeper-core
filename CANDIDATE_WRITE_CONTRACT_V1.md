# Candidate Write / Idempotency / CAS Contract v1

## Status

Candidate until merged to the default branch. When present on the default branch, this document defines the first Threadkeeper Core internal authority-write machinery contract. It does **not** enable a public or service authority-write interface.

## 1. Purpose

This contract defines the machinery required to turn one already-formed, semantically valid durable event into an accepted Git-ledger transition without weakening Threadkeeper's authority boundary.

The write path is split into two different acts:

1. **prepare** — validate the request and create unreachable Git objects;
2. **accept** — atomically advance the configured authoritative ref from the exact expected head to the prepared commit.

Only the second act changes authority.

## 2. Scope

v1 candidate writing is deliberately limited to event families whose write semantics are already accepted and executable. Initially this means `core.record.*` events governed by `exclusive-governed-record-v1` and an accepted reducer binding.

The machinery must not infer write semantics for another event family merely because its JSON validates.

This lane does not add:

- HTTP, gRPC or MCP write endpoints;
- actor authentication or credential verification;
- human-acceptance UI;
- automatic decision authority;
- schema/reducer-binding mutation writes;
- multi-event transactions;
- public `authority-write` enablement.

## 3. Fully formed event input

The internal candidate builder receives:

- exact `expected_head`;
- one safe, add-only `events/.../*.json` path;
- the exact completed canonical event bytes.

Before Git candidate construction, Core must require:

- strict JSON validity;
- RFC 8785 canonical storage bytes;
- valid `content_sha256`;
- non-empty `event_id` and `idempotency_key`;
- event `expected_ledger_commit` equal to the exact current head;
- historical schema validation at that head;
- an accepted reducer binding for the event's `record_kind`;
- matching event schema and authority-policy version;
- reducer validation against the current deterministic projection.

Invalid requests must fail before the authoritative ref can change.

## 4. Idempotency is checked before stale-state rejection

A retry may arrive after the original request was accepted but before its response reached the caller. Therefore Core must first search the validated durable ledger for the supplied `idempotency_key`.

If the key already exists:

### Exact retry

If the accepted event has the same logical `event_id` and `content_sha256`, Core returns `already_accepted` with the original accepted event identity. It does not construct or accept a second event.

The response must expose at least:

- status = `already_accepted`;
- event ID;
- idempotency key;
- content SHA-256;
- original event path;
- original accepting Git commit;
- current authoritative ledger commit.

This response must be reconstructable from the ledger after process restart or loss of all caches.

### Conflicting reuse

If the same idempotency key is already attached to different event content or a different event ID, Core fails closed with `IDEMPOTENCY_CONFLICT`.

It must not silently reinterpret the second request as a rebase or new decision.

## 5. Stale state

If no accepted event owns the idempotency key, the request's expected head must exactly equal the current authoritative Git commit.

Mismatch is `STALE_STATE`.

Threadkeeper does not automatically rewrite `expected_ledger_commit`, regenerate state assertions, or rebase a governance decision onto newer authority.

## 6. Candidate Git construction

Candidate construction starts from the exact tree at H0 and adds exactly one new event blob.

Implementation rules:

- use direct Git plumbing commands, never a shell;
- use a private temporary index rather than a user working tree;
- do not modify any ref while preparing;
- the event path must not already exist at H0;
- the candidate commit must have exactly one parent, H0;
- the candidate tree must differ from H0 only by the intended event addition;
- repository Git hooks must not execute;
- commit signing/configuration must not be inherited implicitly;
- ambient Git namespaces, replacement refs, alternate index paths and other previously forbidden environment state remain isolated.

Git objects created during prepare are not authority. If the process dies before ref advancement, they may remain unreachable and later be garbage-collected.

## 7. Candidate identity

A prepared candidate exposes:

- expected H0;
- candidate commit H1;
- event path;
- event ID;
- idempotency key;
- content SHA-256.

Git commit identity is not the logical event identity. The event ID and SHA-256 remain stable across future Git hash migrations.

A candidate handle is **untrusted input at acceptance time**. Immediately before CAS, Core must independently re-read the candidate commit and verify its exact parent, sole event-path addition, canonical event bytes, event ID, idempotency key, content digest, historical schema/binding and reducer transition against H0. A caller cannot gain authority by substituting a different commit or path into a previously prepared handle.

## 8. Exact-head compare-and-swap

Acceptance requires all of the following immediately before ref advancement:

1. authoritative ref still resolves to H0;
2. candidate H1 exists and is a commit;
3. H1 has exactly one parent and that parent is H0;
4. the request has not already been accepted under its idempotency key.

Core then performs the equivalent of:

```text
git update-ref <authoritative-ref> H1 H0
```

using direct process invocation with repository hooks disabled.

If the ref is no longer H0, acceptance fails with `STALE_STATE`.

No success response may be emitted from candidate preparation alone.

## 9. Post-CAS verification

After a successful compare-and-swap, Core must:

1. resolve the authoritative ref and require H1;
2. run the existing full read-only replay/integrity path;
3. recover the accepted event by idempotency key;
4. require its event ID, content digest and accepting commit to equal the prepared candidate.

Failure after a successful CAS is a `POST_ACCEPTANCE_VERIFICATION_FAILED` recovery condition, not permission to silently move the ref backwards.

## 10. Crash semantics

### Crash before CAS

Authority remains H0. Candidate objects may exist but are not accepted.

### Crash during failed CAS

Authority is whichever exact commit the ref actually names. Core re-reads the ref and reports stale/failure; it does not guess.

### Crash immediately after successful CAS

Authority is H1 even if no response was returned. A retry with the identical idempotency key/event digest reconstructs and returns the original acceptance as `already_accepted`.

This is why durable idempotency state comes from accepted events rather than a process-local response cache.

## 11. Competing candidates

Two candidates may be prepared against the same H0. At most one may be accepted.

After one candidate advances the authoritative ref to H1, every other candidate prepared against H0 must fail exact-head CAS with `STALE_STATE` unless its idempotency key resolves to an already accepted identical request.

There is no automatic winner selection, merge or rebase.

## 12. Hook isolation

Git repository hooks are not Threadkeeper authority policy.

Candidate construction and authoritative ref CAS must override `core.hooksPath` to a known-empty location so `reference-transaction`, `pre-commit`, or other repository hooks cannot execute or veto/augment an authority transition.

If future governance intentionally wants an external co-signing mechanism, it must be represented explicitly in the authority contract rather than smuggled in as a Git hook.

## 13. Public write gate remains closed

The existence of internal CAS-capable code does not enable users, AI clients, or service callers to perform authority writes.

`service.AuthorityWritesEnabled()` remains false and the executable `authority-write` command must continue to return `AUTHORITY_WRITES_DISABLED`.

Enabling an external write path is a later protected decision requiring actor/authentication policy, destructive/crash testing, deployment permissions, and explicit owner acceptance.

## 14. Next gate after this contract

After this machinery is accepted, Threadkeeper should run an independent adversarial review of candidate construction and crash/stale/idempotency behavior before defining actor authentication and the first intentionally enabled write interface.
