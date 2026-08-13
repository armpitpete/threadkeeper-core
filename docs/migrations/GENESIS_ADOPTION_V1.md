# Legacy Genesis Adoption v1

The existing Threadkeeper development ledger predates the genesis contract. It must not rewrite its first commit to pretend otherwise.

For an existing ledger, genesis adoption is a governed migration record that identifies:

- the immutable pre-adoption ledger head;
- the project and ledger identities being asserted;
- the authority policy already in force at that head;
- the actor/mechanism authorised to make the adoption decision;
- the new genesis-contract version used for all later operations.

The historical pre-adoption chain remains valid under its recorded legacy contract. New production ledgers created after genesis v1 becomes normative must install the genesis root at creation.
