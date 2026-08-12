# Implementation Status

## Installed authority kernel

Implemented and exercised by conformance tests:

- Go CLI/service skeleton and machine-readable build metadata;
- strict raw JSON validation, duplicate-member/UTF-8/negative-zero rejection;
- RFC 8785 canonicalisation and SHA-256 content identity;
- local-only JSON Schema Draft 2020-12 registry;
- hardened bare Git ledger opening, repository safety checks and strict fsck;
- linear history inspection and immutable semantic JSON tree rules;
- Git ledger replay with event/schema/policy validation;
- exclusive governed-record reducer and accepted reducer bindings;
- internal candidate-write construction, exact-head Git compare-and-swap, durable idempotency and post-CAS replay verification;
- adversarial repository-isolation regressions for alternates, `commondir`, partial clones/lazy fetch, symlinked authority stores, reftable/worktree config and non-regular JSON tree modes;
- explicit hard public/service gate: `AUTHORITY_WRITES_DISABLED`.

## Assurance expansion implemented as primitives/contracts

- genesis trust-root validation;
- explicit threat model;
- source escrow preservation modes and content verification;
- bitemporal effective/knowledge time;
- coverage/completeness and bounded absence claims;
- confidentiality clearance and governed redaction tombstones;
- decision alternatives, dissent and reopening conditions;
- deterministic fork recovery classification;
- private candidate quarantine storage primitive;
- policy impact simulation comparison;
- verified derived replay checkpoints;
- Ed25519 external witness signing/verification;
- federated exact references with mandatory local authority disposition;
- Core build provenance model and conformance artifact evidence;
- single declared authority-effect vocabulary;
- five-domain health model, incident lifecycle, key lifecycle and evidence envelope.

## Deliberately not yet enabled / still requiring integration

- PR #11's current repaired CAS head still requires a fresh independent full Issue #9 PASS before merge;
- public authority-write transport and actor authentication/authorisation remain disabled;
- quarantine is not yet wired into Git candidate materialisation because that would change the CAS boundary before its independent acceptance;
- source-adapter escrow backends and broad temporal/coverage schema migration are not yet installed;
- checkpoint-accelerated replay has not replaced full replay; only checkpoint build/verification is installed;
- external witness deployment/key service is optional and not configured;
- recovery fork operator resolution and destructive secondary-backup restore drill remain to be proved;
- federation transport and executable reference client are not yet built;
- Recall/search/vector storage remains deliberately separate and unimplemented.

## Write status

`AUTHORITY_WRITES_DISABLED`

This remains true until the enlarged recovery, authentication/authorisation, deployment-ownership and load-safety gates are separately accepted.
