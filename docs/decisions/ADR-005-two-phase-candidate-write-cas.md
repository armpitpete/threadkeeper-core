# ADR-005 — Two-phase candidate construction with exact-head Git CAS

## Status

Candidate until merged to the default branch. Accepted when present on the default branch.

## Context

Threadkeeper now has a Git authority root, strict canonical event format, historical schemas/bindings, and deterministic reducer replay. The next implementation step needs write machinery without making an internal implementation primitive equivalent to public decision authority.

The design must survive crashes, retries and competing clients without duplicate decisions or silent rebases.

## Decision

Use a two-phase internal write model:

1. validate a fully formed event against exact H0 and create an unreachable single-event candidate commit H1;
2. accept H1 only through atomic `update-ref <ref> H1 H0` compare-and-swap.

Idempotency is reconstructed from accepted ledger events and checked before stale-state rejection. Exact retries return the original accepting commit; conflicting reuse fails closed.

Repository Git hooks are explicitly bypassed for write plumbing. The public/service authority-write gate remains disabled.

## Why

This maps Threadkeeper semantics directly onto Git's strongest useful primitive: candidate objects may be created speculatively, while authority is represented by one exact ref transition guarded by its previous object ID.

It also gives clear crash boundaries:

- before CAS: no authority change;
- after CAS: accepted whether or not a response was delivered;
- retry: reconstruct result from ledger.

## Rejected alternatives

### Commit directly to a working tree branch

Rejected because working-tree state, index state and implicit Git behavior would become part of authority semantics.

### Lock in process and then update the ref unconditionally

Rejected because a process lock does not protect against another process, delayed client or administrative ref movement. Exact-head CAS is still required.

### Store idempotency only in SQLite/cache

Rejected because deletion of rebuildable storage would then destroy knowledge of whether a governance decision had already been accepted.

### Automatically rebase stale decisions

Rejected because rebuilding `prior_state`, `resulting_state`, policy identity or evidence against a newer head is a new decision candidate, not a transport retry.

### Let Git hooks enforce governance

Rejected because hooks are opaque repository-local executable behavior and would create a second, poorly inspectable authority mechanism.

## Consequences

The codebase now contains internal ref-mutating primitives, so testing and access boundaries become more important. Their presence does not by itself authorise callers to use them. External enablement requires a later reviewed service/authentication contract and explicit protected acceptance.
