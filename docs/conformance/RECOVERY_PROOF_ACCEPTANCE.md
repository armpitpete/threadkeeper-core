# Recovery Proof Acceptance v1

A Threadkeeper recovery claim is accepted only when a restored durable ledger passes the same read-only integrity and replay path as the original and produces an equivalent machine-readable recovery proof.

Minimum proof fields:

1. exact authoritative ledger commit;
2. authoritative ref identity;
3. Git object format;
4. history commit count;
5. accepted event count;
6. reducer-binding count;
7. governed-record count;
8. canonical governed-record projection SHA-256;
9. complete replay SHA-256.

## Destructive test

The automated recovery test must:

1. construct a non-empty authoritative ledger with an accepted governed event;
2. produce its original recovery proof;
3. create a separate bare backup with no hardlinks to the original;
4. destroy the original ledger directory;
5. restore a fresh bare ledger from the backup;
6. run the complete safety/fsck/replay path on the restored copy;
7. require exact recovery-proof equivalence.

A different head or derived-state digest is `RECOVERY_PROOF_MISMATCH`, never a best-effort success.

This proves file-level destructive restore equivalence in CI. Production readiness additionally requires a restore drill from an independently operated secondary backup location and fork handling when multiple restored copies diverge.
