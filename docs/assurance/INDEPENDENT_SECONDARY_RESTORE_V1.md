# Independent Secondary Restore v1

Threadkeeper Core separates two questions that must not be collapsed:

1. **Did the restored authority store reproduce the exact authoritative state?**
2. **Was the backup actually held and operated on an independent secondary authority boundary?**

Core can answer the first question mechanically. Core cannot establish the second merely because an operator writes different names into a JSON file.

## Core-verifiable equivalence

The restore verifier recomputes `ledger.RecoveryProof` from the restored ledger through the hardened Reader and requires exact equality with the preserved pre-restore proof.

That equality includes:

- authoritative ledger head/ref;
- Git object format;
- Genesis commit, project/ledger identity and Genesis content digest;
- root actor-policy version/content digest;
- history/event/binding/governed-record counts;
- governed projection digest;
- deterministic replay digest.

Any difference fails `core_equivalence_passed`.

## Secondary provenance declaration

The verifier also accepts a strict canonical/digest-bound provenance document containing:

- primary authority-domain ID;
- declared secondary authority-domain ID;
- declared secondary location ID;
- declared secondary operator ID;
- backup-set and backup-artifact identity;
- backup artifact SHA-256;
- SHA-256 of the preserved original RecoveryProof;
- capture and restore timestamps;
- one or more sorted, unique external evidence references.

The document rejects unknown, duplicate, missing and null members; unsafe identifiers; malformed hashes; duplicate/unsorted evidence references; reversed capture/restore chronology; and a secondary authority-domain ID identical to the primary authority-domain ID.

These checks establish that the declaration is internally coherent and bound to the exact original RecoveryProof. They do **not** establish that the declaration is true in the outside world.

## No self-certification

There is no `operational_independence_verified` field or equivalent input. Such an unknown field is rejected.

Every restore verification report sets:

```text
operational_independence_status = requires_external_review
```

Core has no code path that upgrades that status based on provenance strings, evidence-reference strings, successful replay or exact state equivalence.

A human/external reviewer must inspect evidence demonstrating the real authority separation—for example provider/account ownership, access-control boundaries, backup custody, operator identity/role, storage location/account, transfer/capture records and restore logs.

## Proof binding

The provenance document includes `original_recovery_proof_sha256`, defined as SHA-256 of the RFC 8785 canonical JSON representation of the original RecoveryProof.

The restore report records:

- the same original-proof digest;
- the recomputed restored-proof digest;
- provenance content digest;
- restored storage path and authoritative ref;
- all declared secondary provenance identifiers/evidence references;
- both complete RecoveryProof objects.

The report therefore provides one inspectable package linking Core-verifiable state equivalence to the external-evidence claims that still need review.

## Acceptance boundary

The Core restore harness passes when it can fail closed on malformed provenance and altered authority state, emit exact machine-readable equivalence evidence and preserve the external-review boundary.

The actual Core v1 **independent-secondary restore gate** passes only after a destructive restore is performed from a genuinely independently operated secondary and the external provenance evidence is reviewed alongside a Core equivalence PASS.

A same-host/same-account/local-copy test is useful implementation evidence but is never operational independence evidence.

`AUTHORITY_WRITES_DISABLED` remains mandatory throughout restore verification.
