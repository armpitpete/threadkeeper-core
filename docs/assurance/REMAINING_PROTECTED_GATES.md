# Remaining Protected Gates

The assurance programme intentionally leaves a small number of boundaries unconsumed. They are not hidden implementation gaps.

## 1. Independent CAS acceptance

PR #11 at its exact current repair head must receive a genuinely independent full Issue #9 PASS. The repair author cannot self-clear this gate. Only after PASS may PR #11 merge.

## 2. Rebase/retarget assurance expansion

PR #12 is stacked on PR #11. After #11 merges, #12 must be retargeted/rebased onto the accepted `main`, rerun full conformance on the resulting exact head, and receive review appropriate to its security/authority-adjacent scope before merge.

## 3. Genesis adoption

The development ledger predates Genesis v1. Adoption requires an explicit migration decision; history must not be rewritten to fabricate genesis.

## 4. Public write enablement

Before any public authority writer can exist, all of the following remain required:

- actor authentication and authorisation contract + implementation;
- service-owned/non-writable ledger deployment proof from the threat model;
- candidate quarantine integrated into writer materialisation and independently re-reviewed because that changes the CAS boundary;
- final load/performance envelope including bounded resource growth and overload/backpressure evidence;
- destructive restore from an independently operated secondary backup location;
- recovery-fork operator resolution workflow;
- protected release/source governance, including `main` protection and required conformance checks.

Until those gates are accepted, `AUTHORITY_WRITES_DISABLED` remains mandatory.

## 5. Optional operational integrations

These are not prerequisites for read-only Core conformance unless a deployment selects them: external witness service/key deployment, federation transport, checkpoint-accelerated replay, Recall/search/vector storage, and human GUI. Their contracts/primitives must not be confused with enabled services.
