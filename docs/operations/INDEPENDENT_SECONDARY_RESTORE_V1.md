# Independent Secondary Restore Runbook v1

This runbook is for the protected Core v1 operational restore gate. It must be performed against a real independently operated secondary backup. The repository's local clone/restore tests do not satisfy this gate.

## Before backup capture

Record and preserve:

- exact Threadkeeper Core build/source commit;
- exact authoritative ledger path/ref;
- `threadkeeper-core version` output proving `authority_writes_enabled: false`;
- pre-backup `ledger-recovery-proof` output;
- identity of the primary authority boundary;
- secondary provider/account/location/operator arrangement and the evidence that makes it operationally separate from the primary.

The pre-backup RecoveryProof is evidence and must be retained outside the authority store that will later be destroyed/restored.

## Secondary evidence

The external reviewer should be able to inspect evidence such as:

- secondary provider/account or storage tenancy identity;
- access-control/ACL/IAM evidence showing the secondary custody boundary;
- who operates the backup and restore capability;
- backup capture/upload receipt or immutable object/version identity;
- backup artifact SHA-256 calculated on the transferred artifact;
- timestamps and logs for capture and later restore;
- evidence that the source used for restore came from the declared secondary rather than a surviving primary/local copy.

The exact evidence varies by deployment. Core records references to that material but cannot prove its external truth.

## Provenance document

Create the canonical digest-bound `threadkeeper.secondary-restore-provenance.v1` document containing the declared authority domains, location/operator, backup identifiers/hashes, original RecoveryProof SHA, capture/restore times and sorted external evidence references.

Do not add a Boolean such as `operational_independence_verified`. The schema rejects self-certification fields.

## Destructive boundary

A real acceptance run must demonstrate recovery after loss/unavailability of the primary authority store. Before destructive action:

1. preview the exact delete/remove/replace operation;
2. verify the preserved original RecoveryProof and secondary artifact are reachable independently of the primary store;
3. confirm the target restore path and permission/ownership plan;
4. record the primary state immediately before destruction;
5. keep `AUTHORITY_WRITES_DISABLED` hard false.

Do not perform destructive production work merely to exercise this runbook unless that operation is explicitly authorised for the actual target.

## Restore verification

After restoring the secondary artifact into the intended restored ledger path, run:

```text
threadkeeper-core recovery-restore-verify \
  <original-recovery-proof.json> \
  <restored-ledger.git> \
  <secondary-provenance.json> \
  [authoritative-ref]
```

The command is read-only. It:

1. strictly decodes the preserved original RecoveryProof;
2. strictly decodes and digest-verifies the provenance declaration;
3. opens the restored ledger through the hardened Reader;
4. recomputes RecoveryProof from the restored authority store;
5. requires exact authority-state equality;
6. emits a machine-readable report;
7. leaves `operational_independence_status` as `requires_external_review` regardless of the result.

On an authority-state mismatch, the command emits the comparison report when possible and exits non-zero.

## Acceptance

The operational gate requires **both**:

- `core_equivalence_passed: true`; and
- independent review of the external evidence establishing that the backup source/custody/operator boundary was genuinely secondary to the primary authority boundary.

Distinct strings in the provenance document are necessary consistency inputs but are not proof of independence.

## Still not authorised

Passing this gate does not by itself enable authority writes or expose a public write transport. Full production-shaped E2E release acceptance and the separate reviewed write-enablement decision remain required.

`AUTHORITY_WRITES_DISABLED` remains mandatory.
