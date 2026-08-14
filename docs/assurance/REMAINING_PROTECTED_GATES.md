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
- Fresh Genesis authority/bootstrap merged as `69b0c3b5f51c9891a78a623621bb64159b9672de` after exact-head conformance #165 and hostile self-review;
- authoritative actor-policy sourcing merged as `f4ea4d7a7ab286446ca560a67619c181605fc189` after exact-head conformance #204 and hostile self-review.

Historical governance remains explicit: PR #11 was owner-authorised and merged without a genuinely independent full Issue #9 PASS. That exception is not rewritten as a review that occurred.

None of these facts enables public authority writes.

## 1. Load/resource proof machinery

Issue #43 / PR #44 is the current independent code-side gate while real production deployment coordinates are unresolved.

Acceptance requires:

- strict machine-readable workload/resource envelopes;
- sampled peak and settled Go heap/goroutine/process descriptor-or-handle evidence;
- Linux and Windows process-handle metrics exercised in conformance;
- concurrent full replay/recovery equality under a bounded reference envelope;
- restored-copy replay-under-load equality;
- explicit bounded overload proof;
- hard write-kill-switch proof under concurrency;
- `ledger-load-proof` production handoff command;
- an evidence matrix mapping all load-safety clauses;
- checkpoint acceleration treated conditionally: full replay remains authoritative while acceleration is disabled;
- complete exact-head conformance and hostile self-review.

Passing this gate closes **proof machinery/reference conformance only**. It does not establish production capacity.

## 2. Production Fresh Genesis + filesystem ownership

Perform one real deployment gate that both:

1. creates the actual dedicated production governance ledger with Fresh Genesis and its real approved actor public-key/grant policy; and
2. proves ledger and sibling quarantine storage are service-owned and non-writable by untrusted users/processes.

Evidence must bind the actual host, paths, service identity, project/ledger IDs, authority-policy version, actor IDs/public keys/grants, initial schema/binding seed, authoritative ref, Genesis/policy digests, Genesis commit/head, replay/recovery proof and platform-native permissions/ACLs.

Private signing keys are never ledger or deployment-evidence material.

This gate is currently blocked on real deployment coordinates/material, not on missing Core code. See `docs/operations/FRESH_GENESIS_DEPLOYMENT_V1.md`.

## 3. Actual production load/resource envelope

After the production-shaped target exists and Issue #43 is accepted, declare the selected production envelope and run:

```text
threadkeeper-core ledger-load-proof <ledger.git> <envelope.json> [authoritative-ref]
```

Require exact RecoveryProof stability plus passing explicit peak/settled resource ceilings. Repository reference results are not production capacity evidence.

See `docs/operations/LOAD_RESOURCE_PROOF_V1.md`.

## 4. Independent secondary restore

Perform a destructive restore from an independently operated secondary backup location and prove exact Genesis identity, actor-policy identity, authoritative head, replay and projection equivalence. Another directory on the same authority boundary is not sufficient evidence.

Local destructive/restored-copy tests prove machinery only and do not close this operational gate.

## 5. End-to-end release acceptance

Run the complete production-shaped sequence:

fresh install → create Fresh Genesis ledger → load authoritative actor policy → authenticate → authorised write → restart → idempotent retry → concurrent conflict → independent restore → replay.

Require identical final Genesis/actor-policy identity, authoritative state and deterministic projection.

## 6. Release decision

Only after every preceding gate passes may a separate reviewed release decision consider removing `AUTHORITY_WRITES_DISABLED` and exposing any public authority-write transport.

## Public write status

`AUTHORITY_WRITES_DISABLED`

No merge, deployment step, test result or authentication success silently authorises public authority writes.

## Optional operational integrations

Not Core v1 prerequisites unless selected: external witness service/key deployment, federation transport, checkpoint-accelerated replay, Recall/search/vector storage and human GUI.
