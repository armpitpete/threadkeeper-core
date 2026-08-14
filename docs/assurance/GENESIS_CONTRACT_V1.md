# Genesis Contract v1

Threadkeeper must not have an implicit first authority. Every durable ledger has a single inspectable Genesis root that identifies the authority domain from which later policy derives.

A Genesis record MUST identify `project_id`, `ledger_id`, creation time, initial authority policy, initial schema identities, initial authorised actors/mechanisms, and its own `content_sha256`.

Genesis arrays are sorted and duplicate-free so two implementations cannot assign different meaning to ordering noise. The record is RFC 8785 canonical JSON and uses Threadkeeper's ordinary digest rule.

## Authoritative ledger representation

For a fresh v1 ledger, Genesis is not a detached validation document. It is authoritative Git history:

- the first/root commit is the only permissible Genesis commit;
- the exact record path is `config/genesis/root.json`;
- that path must be an ordinary `100644` blob containing the exact validated canonical Genesis bytes;
- the root commit may also contain initial immutable schemas, reducer bindings and the initial actor-auth policy, but it may not contain durable events;
- the exact set of schema `$id` values present in the root commit must equal `initial_schemas`;
- every reducer binding present in the root commit must use the Genesis `initial_authority_policy` version;
- the root actor-auth policy must exist at `config/authority/actor-policy/root.json`, must declare the exact Genesis `ledger_id` and `initial_authority_policy`, and its granted actor set must exactly equal Genesis `initial_authorities`;
- later commits may not modify, remove, rename or add material under `config/genesis/` or directly mutate the root actor-policy path.

`ledger.Replay` validates these properties on every authoritative replay and exposes both the Genesis root and its exact Git commit. It also exposes the root actor-policy version/content digest and includes them in the deterministic replay digest. Recovery proof carries the same explicit identity, so a restored ledger with a different Genesis or root actor policy is not equivalent authority.

Genesis establishes the starting trust root only. It does not make later evidence authoritative, and it cannot be rewritten by later policy. Replacing a Genesis root creates a different ledger identity rather than a migration of the same ledger.

## Initial actor-auth trust

`initial_authority_policy` identifies the authority-policy version used by root reducer bindings. `initial_authorities` is not decorative metadata: it must equal the set of actors granted authority by the canonical root actor policy.

The actor-policy value contains its exact `ledger_id`, exact `authority_policy_version`, bounded proof lifetime, trusted Ed25519 keys and exact action/target grants. Those identity fields are part of the canonical digest-bound bytes and must match Genesis. Supplying different runtime context cannot transplant the same policy bytes into another trust domain.

The root actor policy is the starting policy. Later key/grant rotation is performed only through governed events at target `authority:actor-policy` with record kind `core.actor-auth-policy-v1`; direct file mutation remains forbidden. Governed create/replacement values must declare the same ledger and accepted authority-policy version and are rejected during replay/prepare before CAS if they do not. A governed policy revocation fails closed and does not fall back to the root policy.

## Fresh bootstrap boundary

Fresh bootstrap is create-only. It refuses an existing target and checks the existing target parent against the same no-symlink/canonical filesystem boundary used by the hardened reader before creating anything. It creates a new bare SHA-1 v1 ledger, a deterministic parentless Genesis commit and the configured direct authoritative ref, then reopens that repository through the normal hardened Reader and complete replay/fsck path before success is reported.

Bootstrap seed material is restricted to initial immutable schema/reducer-binding namespaces plus the exact root actor-policy path. It cannot seed events or arbitrary authority state.

The supported production bootstrap additionally requires:

- the canonical root actor policy bound to the exact Genesis ledger/policy identity;
- the exclusive governed-record event schema;
- the reducer-binding schema;
- a reducer binding for `core.actor-auth-policy-v1` using the Genesis authority-policy version.

This prevents creating a production trust policy that cannot later be rotated or revoked through the governed write path.

All semantic seed validation occurs before target creation. A failed semantic bootstrap validation therefore leaves no newly created ledger path. Failures after repository creation still do not qualify as production success unless normal replay verifies the exact expected identity.

The bootstrap result is machine-readable evidence binding the actual storage path, project ID, ledger ID, authoritative ref, Genesis content digest, actor-policy content digest, Genesis/root commit, ledger head, Git object format and initial schema/binding counts.

## Deployment assumptions still outside Genesis

The real production host, path, service identity, filesystem ownership/permissions, project/ledger IDs, authority-policy identity, authorised actor IDs, public keys/grants and exact initial seed set are deployment decisions. Private signing keys are never stored in the ledger; only trusted public keys are authoritative policy material. The source repository commit is never substituted for the governance-ledger Genesis identity.

`AUTHORITY_WRITES_DISABLED` remains mandatory until all release gates are separately satisfied.
