# Fresh Genesis Production Deployment v1

This runbook begins only after the Fresh Genesis bootstrap and authoritative actor-policy implementation have been accepted and merged. It does not authorise production deployment by itself.

## Stop conditions before execution

Do not run production bootstrap until all of the following are explicitly resolved for the actual target:

- production host/environment;
- dedicated ledger Git directory;
- service account/process identity;
- ledger and quarantine parent directories and ownership/permission model;
- unique `project_id` and `ledger_id`;
- initial authority-policy version;
- initial authorised actor identities;
- exact trusted Ed25519 **public** keys, key IDs, grant actions/targets and proof lifetime;
- secure private-key custody outside the ledger;
- exact initial schema/reducer-binding seed directory, including the actor-policy rotation binding;
- authoritative ref, normally `refs/heads/main`.

The target ledger path must not already exist. Fresh bootstrap is create-only; an existing path is not adoption evidence and must not be overwritten.

## Required initial seed

The seed root must contain, at minimum:

- `config/authority/actor-policy/root.json` — canonical digest-bound actor policy;
- the exclusive governed-record event schema;
- the reducer-binding schema;
- a reducer binding for record kind `core.actor-auth-policy-v1`, state model `exclusive-governed-record-v1`, event schema equal to the exclusive record schema and authority-policy version equal to Genesis `initial_authority_policy`.

Genesis `initial_authorities` must exactly equal the actors granted by the root actor policy. Private signing keys are never seed material.

## Preflight evidence

Before creating the ledger, record:

1. exact Threadkeeper Core release/source commit being deployed;
2. executable build metadata;
3. host and filesystem identity;
4. intended service account UID/SID and groups where applicable;
5. target ledger path and sibling quarantine path;
6. parent-directory ownership and permissions;
7. Genesis input SHA-256 and validated `genesis-check` output;
8. actor-policy content SHA-256 plus public actor/key/grant inventory;
9. recursive inventory/digests of the bootstrap seed root;
10. confirmation that `AUTHORITY_WRITES_DISABLED` is still reported by the deployed build.

Any unexpected existing target, symlinked/canonicality-changing parent, writable-by-untrusted-user storage, invalid Genesis/policy, authority mismatch or seed mismatch is a deployment FAIL.

## Creation

Use the deployed binary's create-only command:

```text
threadkeeper-core fresh-genesis-init <ledger.git> <genesis.json> <seed-root> [authoritative-ref]
```

Do not substitute a source-repository commit, a worktree, `.threadkeeper/state.json`, a copied test ledger or a manually advanced ref.

A successful command emits machine-readable evidence containing the real storage path, project ID, ledger ID, authoritative ref, Genesis content digest, actor-policy content digest, Genesis/root commit, authoritative head, Git object format and initial schema/binding counts.

## Immediate independent verification

Using a fresh process after creation:

1. run `ledger-inspect` on the exact created ledger/ref;
2. run `ledger-recovery-proof`;
3. require `history_commit_count == 1` at the initial state;
4. require `genesis_commit == ledger_commit`;
5. require replayed project/ledger/content identity to equal the approved Genesis input;
6. require the root schema set and reducer-binding policy versions to match Genesis;
7. require root actor-policy digest and granted actor set to match the approved policy/Genesis;
8. confirm the actor-policy reducer binding exists so keys/grants can be rotated/revoked through governed events;
9. confirm no durable event exists in the root commit;
10. confirm the source repository commit was not used as the ledger Genesis identity;
11. confirm `AUTHORITY_WRITES_DISABLED` remains hard false.

## Actor-key custody

The ledger contains public verification keys only. Production private Ed25519 signing keys must be separately protected and must not be copied into the ledger, repository, seed directory, CI artifacts or deployment evidence.

Key/grant changes after Genesis must use the governed `authority:actor-policy` record. Direct edits to the root policy are invalid. Revoking the governed policy is an emergency fail-closed state and prevents admission until a separately governed recovery path is defined/accepted; it does not silently restore the old root grants.

## Filesystem ownership proof

Creation is not sufficient. On the actual production filesystem, prove that:

- the service identity owns or exclusively controls the ledger and quarantine storage;
- ordinary/untrusted users cannot create, rename, replace, symlink, hard-link into, modify or delete the ledger/quarantine trees;
- parent directories cannot be used by an untrusted process to replace the protected roots;
- the filesystem supports the file sync, directory sync and no-overwrite hard-link semantics required by quarantine publication;
- restart preserves the exact ledger/quarantine identities and permissions.

Record platform-native permission/ACL commands and results as deployment evidence. Do not infer this property from application checks alone.

## Acceptance result

The production Fresh Genesis + filesystem gate passes only when the creation evidence, actor-policy identity, fresh replay/recovery proof and platform ownership/permission evidence all agree on the same actual deployment.

Passing this gate does **not** enable authority writes. Remaining release gates are the declared load/resource envelope, independently operated secondary restore and full end-to-end release acceptance.

`AUTHORITY_WRITES_DISABLED` remains mandatory.
