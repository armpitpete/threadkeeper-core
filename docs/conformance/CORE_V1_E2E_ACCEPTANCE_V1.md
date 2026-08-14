# Core v1 End-to-End Acceptance v1

Issue #50 closes the final **code-side/reference** Core v1 E2E acceptance gap. It does not activate production service operation and it does not enable authority writes.

## Deterministic reference sequence

`internal/service/release_e2e_consolidated_test.go` performs one disposable sequence using the real Core primitives:

1. create Fresh Genesis with canonical actor policy and reducer bindings;
2. reopen through the hardened Git Reader and prove Genesis/root-policy identity;
3. sign a real Ed25519 request proof for the exact ledger/head/action/target/idempotency context;
4. prove exported `AdmitAuthorityWrite` stops at `AUTHORITY_WRITES_DISABLED` before ledger/auth work;
5. authenticate the same proof through the already-reviewed pure ledger-derived policy primitive;
6. prepare two valid H0 candidates through the real quarantine/Git-candidate path;
7. accept one H0→H1 candidate through the exact-head CAS boundary;
8. require the competing H0 candidate to fail stale with no rebase and no authority movement;
9. close/reopen and require exact RecoveryProof equality;
10. retry the accepted candidate and require `already_accepted` for the same durable H1 identity;
11. attempt conflicting reuse of the accepted idempotency key and require `IDEMPOTENCY_CONFLICT` with H1 unchanged;
12. preserve the original RecoveryProof, make a disposable local bare backup, remove only the temporary test authority store and restore a fresh bare copy;
13. run the merged `restoreproof.Verify` path against strict provenance and require exact RecoveryProof equivalence;
14. require final Genesis, root actor policy, head, governed projection and replay identity to match;
15. emit one machine-readable `CORE_V1_E2E_ACCEPTANCE` JSON record and require `authority_writes_enabled: false`.

## Evidence record

The test-visible JSON binds at least:

- Genesis commit and content digest;
- root/current actor-policy content identity;
- authenticated actor/key and exact signed request context;
- accepted event, idempotency and content identity;
- accepted H1 commit;
- restart/retry disposition;
- stale competing-write disposition;
- same-key conflict disposition;
- pre-restore and restored RecoveryProof SHA-256 values;
- restored Core-equivalence result;
- fixed restore operational-independence status (`requires_external_review`);
- explicit hard write-gate state.

## Safety boundary

The test private key is deterministic test material only. The authority store, backup and destructive restore are under `t.TempDir()` and are never production paths.

The local backup is deliberately **not** claimed as independent-secondary operational evidence. Operational independence remains externally reviewed and the real production destructive restore remains protected by Issue #51.

No production `enabled` boolean, environment override, test mode or alternate exported write path is introduced. `AUTHORITY_WRITES_DISABLED` remains mandatory.

Passing this lane closes E2E machinery/reference conformance only. Production load measurement, genuine independent-secondary recovery, service activation and any future write-enablement decision remain separate protected gates.
