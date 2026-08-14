# Remaining Protected Gates

These are the release-critical boundaries still open for Threadkeeper Core v1. Completed historical gates are recorded here only where they materially constrain the remaining work.

## Completed release prerequisites

Established:

- assurance expansion PR #12 is integrated;
- actor-auth cryptographic/exact-grant primitives are implemented behind the hard service gate;
- the owner selected Fresh Genesis; no legacy governance ledger/head will be fabricated;
- recovery-fork classification and explicit operator-resolution workflow are implemented;
- repository `main` is protected by an active ruleset requiring pull requests, resolved conversations, strict `test` and `windows-git-environment-isolation` checks, while blocking deletion and non-fast-forward/force-push updates;
- the consolidated quarantine/CAS boundary merged as `fde19f4c03a1915f7d26da493593566a6017bc49` and received an independent hostile **PASS** in Issue #36, including independently constructed real-Git ref/CAS attacks.

Historical governance remains explicit: PR #11 was owner-authorised and merged as `38ea7c28f2b0f5c5ff0ca38b8da94eff17bfec5b` without a genuinely independent full Issue #9 PASS. That exception is preserved as history; it is not rewritten as a review that occurred.

None of these facts enables public authority writes by itself.

## 1. Fresh Genesis bootstrap integration

Issue #37 / PR #40 is the current code-side gate. The candidate makes Genesis part of actual authoritative history and adds a create-only Fresh Genesis initializer.

Acceptance requires exact-head conformance and hostile self-review of:

- fixed root Genesis identity/path/mode and immutability;
- no durable events in the Genesis commit;
- exact initial schema-set binding and initial reducer-policy binding;
- create-only/no-overwrite behavior;
- controlled Git environment and direct ref creation;
- pre-creation parent-path safety and post-creation hardened Reader/replay/fsck verification;
- recovery-proof binding to Genesis identity;
- preservation of the hard write kill-switch.

This gate creates deployment machinery, not the production ledger.

## 2. Production Genesis + filesystem ownership

After the bootstrap implementation is accepted, perform one real deployment gate that both:

1. creates the actual dedicated production governance ledger with Fresh Genesis; and
2. proves the ledger and sibling quarantine storage are service-owned and non-writable by untrusted users/processes.

The evidence must bind the real host, path, service identity, project/ledger IDs, authority-policy identity, initial authorities, initial schema/binding seed, authoritative ref, Genesis commit/head, replay/recovery proof and platform-native permissions/ACLs.

Source-repository commits, copied test ledgers and `.threadkeeper/state.json` are not substitutes for production ledger identity.

See `docs/operations/FRESH_GENESIS_DEPLOYMENT_V1.md`.

## 3. Durable actor-policy sourcing

The Ed25519 proof and exact-grant primitive is implemented, but service admission currently receives trusted keys/grants as an in-memory `actorauth.Policy` supplied by its caller.

Before public writes can be enabled, the deployed service must load and validate its trusted actor keys/grants from an authoritative, versioned source bound to the ledger trust domain. Runtime configuration must not be able to silently substitute a different authority policy than the one the ledger/Genesis identifies.

This is a code/review gate, not merely an operations checkbox.

## 4. Final load/resource envelope

The existing concurrency, CAS, idempotency, cancellation and explicit-overload tests are necessary but are not the full production envelope. Declare and prove bounded resource behavior for the selected deployment, including memory and file-descriptor growth and explicit overload/backpressure behavior.

## 5. Independent secondary restore

Perform a destructive restore test from an independently operated secondary backup location and prove exact Genesis identity, authoritative head, replay and projection equivalence. A backup copied to another path on the same authority boundary is not sufficient evidence for this gate.

## 6. End-to-end release acceptance

After the preceding gates are satisfied, run the complete production-shaped acceptance sequence:

fresh install → create Fresh Genesis ledger → load authoritative actor policy → authenticate → authorised write → restart → idempotent retry → concurrent conflict → independent restore → replay.

The final Genesis identity, authoritative state and deterministic projection must agree across the sequence.

## Public write status

`AUTHORITY_WRITES_DISABLED`

It remains mandatory until every release-critical gate above is accepted. No merge, deployment step or test result silently authorises a public authority-write transport.

## Optional operational integrations

These are not prerequisites for Core v1 unless the deployment selects them: external witness service/key deployment, federation transport, checkpoint-accelerated replay, Recall/search/vector storage, and human GUI. Their contracts/primitives must not be confused with enabled services.
