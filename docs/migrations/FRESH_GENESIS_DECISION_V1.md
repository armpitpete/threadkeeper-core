# Fresh Genesis Decision v1

## Owner decision

On 2026-08-14 the project owner selected **fresh Genesis** for Threadkeeper Core.

Threadkeeper Core will **not** invent or infer a legacy dedicated governance ledger, legacy governance-ledger head, or legacy Genesis-adoption record from the source repository, `.threadkeeper/state.json`, or any other substitute.

## Consequence

The first production Threadkeeper Core governance ledger will be created as a new dedicated ledger. Its Genesis root will be the first authoritative commit/record in that ledger.

No pre-adoption governance-ledger head exists for this fresh-ledger path, and the legacy `genesis.Adoption` migration record is therefore not used for that ledger.

## Code-side bootstrap

Issue #37 / PR #40 merged the Fresh Genesis authority/bootstrap machinery as `69b0c3b5f51c9891a78a623621bb64159b9672de`:

- Genesis is validated from the actual root commit on every replay;
- `config/genesis/root.json` is the fixed immutable Genesis path;
- replay/recovery output is bound to project, ledger and Genesis commit/content identity;
- a create-only Fresh Genesis bootstrap creates a new dedicated bare ledger and verifies it through the normal hardened Reader/replay/fsck path;
- `fresh-genesis-init` emits machine-readable creation evidence;
- normal event/CAS authority writes remain disabled.

Issue #41 / PR #42 adds the final trust-source prerequisite for production Genesis:

- the root commit must contain canonical authoritative actor policy at `config/authority/actor-policy/root.json`;
- Genesis `initial_authorities` must exactly match the policy's granted actors;
- the supported bootstrap requires the schemas/reducer binding needed to rotate or revoke that policy through governed events;
- service admission derives trusted keys/grants from exact authoritative replay state rather than caller/runtime policy injection.

Neither code lane itself instantiates production Genesis.

## Remaining deployment evidence

Actual Genesis instantiation remains a protected deployment operation and must bind the real:

- production host/environment;
- dedicated ledger storage path;
- service identity and filesystem ownership/permissions;
- `project_id` and unique `ledger_id`;
- initial authority-policy version;
- initial authorised actor identities;
- trusted Ed25519 public keys, exact grants and proof lifetime;
- private-key custody outside the ledger;
- exact initial schema/reducer-binding/actor-policy seed set;
- authoritative ref;
- resulting Genesis and actor-policy digests, Genesis commit/head and replay/recovery proof.

Fresh Genesis instantiation and proof that ledger/quarantine storage is service-owned and non-writable by untrusted users/processes should be performed together on the actual target. A source-repository commit is never a governance-ledger Genesis identity.

Until that deployment operation is performed and verified, the accurate state is:

> **Fresh Genesis selected; bootstrap machinery installed; authoritative actor-policy sourcing under review; production Genesis not yet instantiated.**

`AUTHORITY_WRITES_DISABLED` remains unchanged.
