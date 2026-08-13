# Remaining Protected Gates

The assurance programme intentionally leaves a small number of boundaries unconsumed. They are not hidden implementation gaps.

## 1. CAS review status

PR #11 was owner-authorised and merged to `main` as `38ea7c28f2b0f5c5ff0ca38b8da94eff17bfec5b` after exact-head conformance passed. A genuinely independent full Issue #9 PASS was not recorded before that merge. The repository records that fact explicitly and must not represent the independent gate as having passed.

The merged CAS boundary remains subject to fresh independent hostile review before public authority-write enablement.

## 2. Assurance expansion integration

PR #12 must be replayed/rebased onto the accepted `main`, rerun full exact-head conformance, and receive review appropriate to its security/authority-adjacent scope before merge.

## 3. Genesis adoption

The development ledger predates Genesis v1. Adoption requires an explicit migration decision; history must not be rewritten to fabricate genesis.

## 4. Public write enablement

Before any public authority writer can exist, all of the following remain required:

- fresh independent hostile review of the then-exact CAS boundary;
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
