# Legacy Genesis Adoption v1

The existing Threadkeeper development ledger predates the genesis contract. It must not rewrite its first commit to pretend otherwise.

For an existing ledger, genesis adoption is a governed migration record that identifies:

- the immutable pre-adoption ledger head;
- the project and ledger identities being asserted;
- the authority policy already in force at that head;
- the actor/mechanism authorised to make the adoption decision;
- the new genesis-contract version used for all later operations.

The historical pre-adoption chain remains valid under its recorded legacy contract. New production ledgers created after genesis v1 becomes normative must install the genesis root at creation.

## Validation

`internal/genesis.Adoption` is the canonical validation surface for the legacy-adoption record. It requires strict canonical JSON, the Threadkeeper content digest, a full SHA-1 or SHA-256 pre-adoption Git object ID, project and ledger identities, the existing policy identity, the authorised adoption mechanism, the Genesis contract version, and an RFC 3339 adoption timestamp.

## Evidence boundary

The pre-adoption head must come from the actual dedicated governance ledger. A Threadkeeper source-code repository commit or a `.threadkeeper/state.json` lane-state file is not a substitute. If the real governance ledger and exact head cannot be inspected, Core is Genesis-adoption ready but Genesis is not yet adopted; the missing identity must not be fabricated and history must not be rewritten.
