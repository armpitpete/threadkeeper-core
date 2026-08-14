# Fresh Genesis Production Deployment v1

This runbook begins only after the Fresh Genesis bootstrap implementation has been accepted and merged. It does not authorise production deployment by itself.

## Stop conditions before execution

Do not run production bootstrap until all of the following are explicitly resolved for the actual target:

- production host/environment;
- dedicated ledger Git directory;
- service account/process identity;
- ledger and quarantine parent directories and ownership/permission model;
- unique `project_id` and `ledger_id`;
- initial authority-policy identity;
- initial authorised actor/mechanism identities;
- exact initial schema/reducer-binding seed directory;
- authoritative ref, normally `refs/heads/main`;
- runtime actor-auth trusted-key/grant policy and its durable-authority binding plan.

The target ledger path must not already exist. Fresh bootstrap is create-only; an existing path is not adoption evidence and must not be overwritten.

## Preflight evidence

Before creating the ledger, record:

1. exact Threadkeeper Core release/source commit being deployed;
2. executable build metadata;
3. host and filesystem identity;
4. intended service account UID/SID and groups where applicable;
5. target ledger path and sibling quarantine path;
6. parent-directory ownership and permissions;
7. Genesis input SHA-256 and validated `genesis-check` output;
8. recursive inventory/digests of the bootstrap seed root;
9. confirmation that `AUTHORITY_WRITES_DISABLED` is still reported by the deployed build.

Any unexpected existing target, symlinked/canonicality-changing parent, writable-by-untrusted-user storage, invalid Genesis or seed mismatch is a deployment FAIL.

## Creation

Use the deployed binary's create-only command:

```text
threadkeeper-core fresh-genesis-init <ledger.git> <genesis.json> <seed-root> [authoritative-ref]
```

Do not substitute a source-repository commit, a worktree, `.threadkeeper/state.json`, a copied test ledger or a manually advanced ref.

A successful command emits machine-readable evidence containing the real storage path, project ID, ledger ID, authoritative ref, Genesis content digest, Genesis/root commit, authoritative head, Git object format and initial schema/binding counts.

## Immediate independent verification

Using a fresh process after creation:

1. run `ledger-inspect` on the exact created ledger/ref;
2. run `ledger-recovery-proof`;
3. require `history_commit_count == 1` at the initial state;
4. require `genesis_commit == ledger_commit`;
5. require replayed project/ledger/content identity to equal the approved Genesis input;
6. require the root schema set and reducer-binding policy versions to match Genesis;
7. confirm no durable event exists in the root commit;
8. confirm the source repository commit was not used as the ledger Genesis identity;
9. confirm `AUTHORITY_WRITES_DISABLED` remains hard false.

## Filesystem ownership proof

Creation is not sufficient. On the actual production filesystem, prove that:

- the service identity owns or exclusively controls the ledger and quarantine storage;
- ordinary/untrusted users cannot create, rename, replace, symlink, hard-link into, modify or delete the ledger/quarantine trees;
- parent directories cannot be used by an untrusted process to replace the protected roots;
- the filesystem supports the file sync, directory sync and no-overwrite hard-link semantics required by quarantine publication;
- restart preserves the exact ledger/quarantine identities and permissions.

Record platform-native permission/ACL commands and results as deployment evidence. Do not infer this property from application checks alone.

## Acceptance result

The production Fresh Genesis + filesystem gate passes only when the creation evidence, fresh replay/recovery proof and platform ownership/permission evidence all agree on the same actual deployment.

Passing this gate does **not** enable authority writes. Remaining release gates include durable actor-policy sourcing, declared load/resource proof, independently operated secondary restore and full end-to-end release acceptance.

`AUTHORITY_WRITES_DISABLED` remains mandatory.
