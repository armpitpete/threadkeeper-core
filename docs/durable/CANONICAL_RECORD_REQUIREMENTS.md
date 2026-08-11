# Canonical Durable Record Requirements v0.1

## Status

Candidate until merged to the default branch. When present on the default branch, these requirements are normative for the v1 durable ledger under ADR-002.

## Purpose

Threadkeeper durable events carry an implementation-independent SHA-256 digest in addition to Git object identity. The digest rule must not be circular or depend on incidental JSON formatting.

## Digest boundary

For a durable record containing a `content_sha256` member:

1. construct the logical record with all required fields except `content_sha256`;
2. serialize that digest-input record using the selected Threadkeeper canonical JSON profile;
3. compute SHA-256 over the exact resulting UTF-8 bytes;
4. encode the digest as lowercase hexadecimal;
5. insert that value as `content_sha256` in the stored record;
6. validate by removing `content_sha256`, repeating canonicalization and hashing, and comparing the result.

`content_sha256` is therefore **excluded from its own digest input**.

Changing only insignificant presentation whitespace must not change the logical digest after canonicalization. Changing any value covered by the canonical record must change the digest except for cryptographic collision.

## Canonicalization gate

Authoritative event writing MUST remain disabled until one canonical JSON profile is selected and contract-tested.

That later selection must define at minimum:

- UTF-8 byte encoding;
- object member ordering;
- number serialization;
- string escaping and normalization rules;
- timestamp representation;
- treatment of absent members versus explicit `null`;
- whether arrays preserve supplied order or have schema-specific ordering rules;
- rejection behavior for values not representable by the profile.

The profile must have implementations/tests sufficient for deterministic hashing and portable export. The profile may be replaced later only through a versioned migration that preserves interpretation of historical digests.

## Verification requirements

Tests must prove:

- serialize → hash → store → verify round trips deterministically;
- reformatting non-canonical JSON does not alter the verified logical digest after parsing and canonicalization;
- mutation of a covered field is detected;
- malformed, duplicate-key, non-conforming numeric or otherwise ambiguous JSON fails closed rather than being silently normalized into authority;
- historical records remain verifiable using the canonicalization profile identified by their schema/version metadata.
