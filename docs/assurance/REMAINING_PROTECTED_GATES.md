# Remaining Protected Gates

These are the release-critical boundaries still open for Threadkeeper Core v1. Completed historical gates are recorded only where they constrain what follows.

## Completed prerequisites

Established:

- assurance/recovery foundation integrated;
- Ed25519 actor proof and exact-grant primitives implemented;
- owner-selected Fresh Genesis path; no legacy governance ledger/head will be fabricated;
- recovery-fork workflow implemented;
- protected `main` ruleset active;
- consolidated quarantine/CAS boundary merged as `fde19f4c03a1915f7d26da493593566a6017bc49` and independently **PASSED** Issue #36;
- Fresh Genesis authority/bootstrap merged as `69b0c3b5f51c9891a78a623621bb64159b9672de` after exact-head conformance #165 and hostile self-review.

Historical governance remains explicit: PR #11 was owner-authorised and merged without a genuinely independent full Issue #9 PASS. That exception is not rewritten as a review that occurred.

None of these facts enables public authority writes.

## 1. Authoritative actor-policy sourcing

Issue #41 / PR #42 is the current code-side gate.

Acceptance requires:

- strict canonical/digest-bound root actor policy;
- exact Genesis `initial_authorities` reconciliation;
- immutable root policy path;
- authoritative current policy derived from exact replay state;
- governed rotation/replacement at fixed target `authority:actor-policy`;
- malformed rotations rejected before CAS;
- revocation failing closed rather than reverting to root policy;
- exported service admission unable to accept a caller-supplied substitute policy;
- exact ledger/head binding before Ed25519 authentication;
- hard `AUTHORITY_WRITES_DISABLED` check still first;
- complete exact-head conformance and hostile self-review.

The supported Fresh Genesis bootstrap must include the actor-policy root and reducer binding needed to rotate/revoke it.

## 2. Production Fresh Genesis + filesystem ownership

After actor-policy sourcing is accepted, perform one real deployment gate that both:

1. creates the actual dedicated production governance ledger with Fresh Genesis and its real approved actor public-key/grant policy; and
2. proves ledger and sibling quarantine storage are service-owned and non-writable by untrusted users/processes.

Evidence must bind the actual host, paths, service identity, project/ledger IDs, authority-policy version, actor IDs/public keys/grants, initial schema/binding seed, authoritative ref, Genesis/policy digests, Genesis commit/head, replay/recovery proof and platform-native permissions/ACLs.

Private signing keys are never ledger or deployment-evidence material.

See `docs/operations/FRESH_GENESIS_DEPLOYMENT_V1.md`.

## 3. Final load/resource envelope

Declare the selected production envelope and prove bounded resource behavior under it, including:

- concurrent replay and write/conflict behavior;
- explicit overload/backpressure;
- bounded memory growth;
- bounded file-descriptor/handle growth;
- restart/recovery behavior under load;
- hard write kill-switch effectiveness under concurrency.

The existing logical concurrency tests are necessary but not sufficient for this operational gate.

## 4. Independent secondary restore

Perform a destructive restore from an independently operated secondary backup location and prove exact Genesis identity, authoritative head, replay and projection equivalence. Another directory on the same authority boundary is not sufficient evidence.

## 5. End-to-end release acceptance

Run the complete production-shaped sequence:

fresh install → create Fresh Genesis ledger → load authoritative actor policy → authenticate → authorised write → restart → idempotent retry → concurrent conflict → independent restore → replay.

Require identical final Genesis identity, authoritative state and deterministic projection.

## 6. Release decision

Only after every preceding gate passes may a separate reviewed release decision consider removing `AUTHORITY_WRITES_DISABLED` and exposing any public authority-write transport.

## Public write status

`AUTHORITY_WRITES_DISABLED`

No merge, deployment step, test result or authentication success silently authorises public authority writes.

## Optional operational integrations

Not Core v1 prerequisites unless selected: external witness service/key deployment, federation transport, checkpoint-accelerated replay, Recall/search/vector storage and human GUI.
