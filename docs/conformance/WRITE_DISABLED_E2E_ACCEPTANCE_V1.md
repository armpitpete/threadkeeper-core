# Write-Disabled Core v1 End-to-End Acceptance

This conformance lane proves that the major Core v1 authority components compose coherently **while the exported authority-write service remains hard disabled**.

It is a code-side acceptance proof, not a production deployment, not an independent-secondary backup proof, and not permission to expose or enable authority writes.

## Deterministic sequence

`internal/service/release_e2e_test.go::TestWriteDisabledCoreV1EndToEndReleaseAcceptance` performs one coherent authority lifecycle:

1. create a brand-new bare ledger with the supported Fresh Genesis bootstrap;
2. bind Genesis to an exact test ledger ID, authority-policy version, actor, Ed25519 public key and reducer bindings;
3. reopen the real ledger through the hardened Reader and require H0 to equal the Genesis commit;
4. load actor policy from that exact authoritative replay snapshot;
5. sign a real Ed25519 proof with a deterministic test-only private key and authenticate/authorise it against exact ledger ID, action, target, H0 and idempotency identity;
6. call exported `service.AdmitAuthorityWrite` with the valid proof and a nil Reader and require `AUTHORITY_WRITES_DISABLED`, proving the release gate fires before ledger/policy/authentication work;
7. use the already independently reviewed internal candidate/quarantine/CAS path to accept one governed H0→H1 event;
8. capture the authoritative RecoveryProof and deterministic replay/projection at H1;
9. close and reopen the on-disk ledger and require the same Genesis, actor-policy identity, H1 and RecoveryProof;
10. retry the exact candidate and require `already_accepted` for H1;
11. submit different content under the same idempotency key and require deterministic `IDEMPOTENCY_CONFLICT` with H1 unchanged;
12. create a local bare backup, close and destroy the original temporary authority store, restore from the backup into a new path, and require exact RecoveryProof and replay/projection equality;
13. finish by requiring `AuthorityWritesEnabled() == false`.

The complete test also runs under whole-tree `go test -race ./...` conformance.

## Authentication/write composition boundary

The test deliberately does **not** add an enabled service write route.

The cryptographic proof is checked using the existing package-private pure `authenticateAuthoritativePolicySnapshot` helper over a policy snapshot derived from authoritative replay. The subsequent authority transition uses `ledger.PrepareWriteCandidate` / `ledger.AcceptWriteCandidate` directly inside the test.

Therefore this lane proves that:

- Fresh Genesis, ledger-derived actor policy, exact request authentication, candidate validation, quarantine, CAS, replay and recovery agree on one exact authority history; and
- the exported service path is still unavailable.

It does **not** claim that a production authenticated transport exists or that authentication and CAS have been exposed as one public transaction. Any future enabled transport remains subject to the separate release decision and must preserve the same exact snapshot/request bindings.

## Key material

The Ed25519 private key is deterministic **test fixture material only**. It has no production authority and must never be reused as a production signing key. Production private keys remain outside the ledger and repository.

## Restore boundary

The test performs a genuinely destructive restore of its temporary authority directory, but the backup is another local test directory under the same test/runtime authority boundary.

That proves restore orchestration and exact semantic continuity only. It does **not** satisfy the Core v1 requirement for restore from an independently operated secondary backup location.

## Gates still open after this passes

A passing code-side E2E lane does not close:

1. actual production Fresh Genesis creation plus service-owned/non-user-writable ledger and quarantine filesystem proof;
2. actual production load/resource envelope measurement;
3. destructive restore from an independently operated secondary location;
4. the production-shaped E2E run using the actual deployment and real approved signing key;
5. the separately reviewed decision on whether to remove `AUTHORITY_WRITES_DISABLED` and expose a public write transport.

## Write status

`AUTHORITY_WRITES_DISABLED`

No result in this conformance lane authorises changing that status.
