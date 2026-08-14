# Fresh Genesis Decision v1

## Owner decision

On 2026-08-14 the project owner selected **fresh Genesis** for Threadkeeper Core.

Threadkeeper Core will **not** invent or infer a legacy dedicated governance ledger, legacy governance-ledger head, or legacy Genesis-adoption record from the source repository, `.threadkeeper/state.json`, or any other substitute.

## Consequence

The first production Threadkeeper Core governance ledger will be created as a new dedicated ledger. Its Genesis root will be the first authoritative commit/record in that ledger.

No pre-adoption governance-ledger head exists for this fresh-ledger path, and the legacy `genesis.Adoption` migration record is therefore not used for that ledger.

## Code-side bootstrap candidate

Issue #37 / PR #40 adds the code-side machinery required before that deployment can be performed honestly:

- Genesis is validated from the actual root commit on every replay;
- `config/genesis/root.json` is the fixed immutable Genesis path;
- replay/recovery output is bound to project, ledger and Genesis commit/content identity;
- a create-only Fresh Genesis bootstrap creates a new dedicated bare ledger and verifies it through the normal hardened Reader/replay/fsck path;
- `fresh-genesis-init` emits machine-readable creation evidence;
- normal event/CAS authority writes remain disabled.

This implementation work does **not** itself instantiate production Genesis.

## Remaining deployment evidence

Actual Genesis instantiation remains a protected deployment operation and must bind the real:

- production host/environment;
- dedicated ledger storage path;
- service identity and filesystem ownership/permissions;
- `project_id` and unique `ledger_id`;
- initial authority-policy identity;
- initial authorised actors/mechanisms;
- exact initial schema and reducer-binding seed set;
- authoritative ref;
- resulting Genesis commit/head and replay/recovery proof.

Fresh Genesis instantiation and proof that ledger/quarantine storage is service-owned and non-writable by untrusted users/processes should be performed together on the actual target. A source-repository commit is never a governance-ledger Genesis identity.

Until that deployment operation is performed and verified, the accurate state is:

> **Fresh Genesis selected; bootstrap machinery under review; production Genesis not yet instantiated.**

The actor-auth primitive currently accepts runtime-injected trusted keys/grants. Durable binding of runtime authentication/authorisation policy to authoritative ledger state remains a separate release gate before public writes can be enabled.

`AUTHORITY_WRITES_DISABLED` remains unchanged.
