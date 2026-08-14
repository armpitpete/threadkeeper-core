# Genesis Contract v1

Threadkeeper must not have an implicit first authority. Every durable ledger has a single inspectable Genesis root that identifies the authority domain from which later policy derives.

A Genesis record MUST identify `project_id`, `ledger_id`, creation time, initial authority policy, initial schema identities, initial authorised actors/mechanisms, and its own `content_sha256`.

Genesis arrays are sorted and duplicate-free so two implementations cannot assign different meaning to ordering noise. The record is RFC 8785 canonical JSON and uses Threadkeeper's ordinary digest rule.

## Authoritative ledger representation

For a fresh v1 ledger, Genesis is not a detached validation document. It is authoritative Git history:

- the first/root commit is the only permissible Genesis commit;
- the exact record path is `config/genesis/root.json`;
- that path must be an ordinary `100644` blob containing the exact validated canonical Genesis bytes;
- the root commit may also contain initial immutable schemas and reducer bindings, but it may not contain durable events;
- the exact set of schema `$id` values present in the root commit must equal `initial_schemas`;
- every reducer binding present in the root commit must use the Genesis `initial_authority_policy` version;
- later commits may not modify, remove, rename or add material under `config/genesis/`.

`ledger.Replay` validates these properties on every authoritative replay and exposes both the Genesis root and its exact Git commit. The replay digest and recovery proof include Genesis identity, so a restored ledger with a different Genesis is not equivalent authority.

Genesis establishes the starting trust root only. It does not make later evidence authoritative, and it cannot be rewritten by later policy. Replacing a Genesis root creates a different ledger identity rather than a migration of the same ledger.

## Fresh bootstrap boundary

Fresh bootstrap is create-only. It refuses an existing target and checks the existing target parent against the same no-symlink/canonical filesystem boundary used by the hardened reader before creating anything. It creates a new bare SHA-1 v1 ledger, a deterministic parentless Genesis commit and the configured direct authoritative ref, then reopens that repository through the normal hardened Reader and complete replay/fsck path before success is reported.

Bootstrap seed material is restricted to initial immutable schema and reducer-binding namespaces. It cannot seed events or arbitrary authority state. A failed bootstrap may leave incomplete create-only residue, but that residue is not accepted as a production ledger unless normal replay validates it.

The bootstrap result is machine-readable evidence binding the actual storage path, project ID, ledger ID, authoritative ref, Genesis content digest, Genesis/root commit, ledger head, Git object format and initial schema/binding counts.

## Deployment assumptions still outside Genesis

Genesis `initial_authorities` and `initial_authority_policy` identify the intended starting trust domain. The current actor-auth primitive still receives trusted keys/grants as a runtime `actorauth.Policy`; binding that runtime policy durably to authoritative ledger state is a separate release gate and must be resolved before public authority writes are enabled.

The real production host, path, service identity, filesystem ownership/permissions, project/ledger IDs, authority policy, authorised actors and initial seed set are deployment decisions. The source repository commit is never substituted for the governance-ledger Genesis identity.

`AUTHORITY_WRITES_DISABLED` remains mandatory until all release gates are separately satisfied.
