# Implementation Status

## Core Skeleton v0

This implementation is intentionally **not an authority writer**.

Implemented in this lane:

- Go CLI/service skeleton;
- machine-readable build/dependency metadata;
- strict raw JSON validation before ordinary decoding;
- duplicate-member detection at arbitrary nesting depth;
- invalid UTF-8 rejection;
- negative-zero rejection;
- RFC 8785 canonicalization behind an internal boundary;
- SHA-256 digest omission/insertion/verification boundary;
- local-only JSON Schema Draft 2020-12 registry;
- initial conformance fixtures and CI;
- an explicit hard gate that rejects every authority-write attempt.

Not yet implemented:

- durable Git ledger open/fsck;
- event replay/projections;
- Git candidate commit construction;
- exact-head compare-and-swap writes;
- idempotency ledger;
- recovery/destructive tests;
- public client transport or authentication;
- Recall/search/vector storage.

## Write status

`AUTHORITY_WRITES_DISABLED`

This remains true until the normative write-enablement gates are satisfied and separately reviewed.
