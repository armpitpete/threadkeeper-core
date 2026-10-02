# RecoveryProof Input Contract v1

## Purpose

A RecoveryProof is a complete machine-readable fingerprint of one fully validated authoritative replay. Missing JSON members are not equivalent to legitimate zero values.

This contract applies to every command that accepts a serialized RecoveryProof, including:

- `recovery-compare`;
- `recovery-restore-verify`.

## Closed document shape

A serialized RecoveryProof MUST contain every field defined by `ledger.RecoveryProof`:

- `ledger_commit`;
- `authoritative_ref`;
- `git_object_format`;
- `genesis_commit`;
- `project_id`;
- `ledger_id`;
- `genesis_content_sha256`;
- `actor_policy_version`;
- `actor_policy_root_content_sha256`;
- `history_commit_count`;
- `event_count`;
- `reducer_binding_count`;
- `governed_record_count`;
- `governed_records_sha256`;
- `replay_sha256`.

Required fields MUST NOT be omitted or `null`. Unknown and duplicate JSON members are invalid. Trailing JSON values/data are invalid.

Required authority-identity strings MUST be non-empty. `history_commit_count` MUST be positive. Event, reducer-binding and governed-record counts MUST NOT be negative.

This contract intentionally does not add a new lexical restriction on Git object IDs or SHA-256 strings beyond existing Core contracts. Such changes require their own compatibility decision.

## Comparison semantics

`recovery-compare` MUST:

1. strictly decode and validate both RecoveryProof documents;
2. fail with `RECOVERY_PROOF_INVALID` if either document is invalid;
3. compare only two valid proofs;
4. report success only when all RecoveryProof fields are equal;
5. report `RECOVERY_PROOF_MISMATCH` for valid but unequal proofs.

Two identically malformed/incomplete documents are never evidence of recovery equivalence.

## Shared decoder rule

Core MUST use one strict RecoveryProof decode/validation implementation for standalone comparison and restore verification. A caller-specific decoder MUST NOT silently weaken the document contract.

## Authority boundary

RecoveryProof validation and comparison are read-only. They do not:

- advance the authority ref;
- mutate the ledger;
- create a Core service;
- enable a public transport;
- enable authority writes.

`AUTHORITY_WRITES_DISABLED` remains mandatory.
