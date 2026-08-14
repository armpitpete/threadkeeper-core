# Recovery Proof Acceptance v1

A Threadkeeper recovery claim is accepted only when a restored durable ledger passes the same read-only integrity and replay path as the original and produces an equivalent machine-readable recovery proof.

Minimum proof fields:

1. exact authoritative ledger commit;
2. authoritative ref identity;
3. Git object format;
4. exact Genesis commit;
5. Genesis `project_id`;
6. Genesis `ledger_id`;
7. Genesis content SHA-256;
8. history commit count;
9. accepted event count;
10. reducer-binding count;
11. governed-record count;
12. canonical governed-record projection SHA-256;
13. complete replay SHA-256.

Recovery equivalence therefore includes trust-root equivalence. A copy with identical later event data but a different Genesis commit/project/ledger/content identity is not the same recovered authority.

## Destructive test

The automated recovery test must:

1. construct a non-empty Genesis-rooted authoritative ledger with an accepted governed event;
2. produce its original recovery proof;
3. create a separate bare backup with no hardlinks to the original;
4. destroy the original ledger directory;
5. restore a fresh bare ledger from the backup;
6. run the complete repository-safety/fsck/Genesis/replay path on the restored copy;
7. require exact recovery-proof equivalence.

A different Genesis identity, head or derived-state digest is `RECOVERY_PROOF_MISMATCH`, never a best-effort success.

This proves file-level destructive restore equivalence in CI. Production readiness additionally requires a restore drill from an independently operated secondary backup location and fork handling when multiple restored copies diverge.
