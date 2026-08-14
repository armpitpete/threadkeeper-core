# Fresh Genesis Decision v1

## Owner decision

On 2026-08-14 the project owner selected **fresh Genesis** for Threadkeeper Core.

Threadkeeper Core will **not** invent or infer a legacy dedicated governance ledger, legacy governance-ledger head, or legacy Genesis-adoption record from the source repository, `.threadkeeper/state.json`, or any other substitute.

## Consequence

The first production Threadkeeper Core governance ledger will be created as a new dedicated ledger. Its Genesis root will be the first authoritative record in that ledger.

No pre-adoption governance-ledger head exists for this fresh-ledger path, and the legacy `genesis.Adoption` migration record is therefore not used for that ledger.

## Remaining deployment evidence

This decision selects the migration path; it does **not** fabricate a production ledger or claim Genesis has already been physically instantiated.

Actual Genesis instantiation remains a deployment operation and must bind the real production project identity, ledger identity, authority policy, storage location, and resulting Genesis root/head. Those values must come from the created production ledger and deployment evidence.

Until that deployment operation is performed and verified, the accurate state is:

> **Fresh Genesis selected; production Genesis not yet instantiated.**

`AUTHORITY_WRITES_DISABLED` remains unchanged.
