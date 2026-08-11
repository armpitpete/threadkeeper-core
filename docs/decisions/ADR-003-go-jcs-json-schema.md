# ADR-003: Go Runtime, RFC 8785 JCS, and JSON Schema 2020-12

## Status

Candidate until merged to the default branch. When present on the default branch, this ADR is accepted.

## Date

2026-08-11

## Context

ADR-001 requires Threadkeeper Core to remain independent from AI. ADR-002 selects a Git-backed durable governance ledger with canonical JSON records, exact-head protected writes, and deterministic recovery.

The next dependency-order decision is the implementation runtime and the serialization/schema tooling used on the authority path.

The selection must optimise for correctness and operational simplicity rather than language novelty. In particular it must support:

- one long-running Core/Manager service;
- Linux deployment on Oracle with straightforward service supervision;
- portable test/build support on Windows and Linux;
- robust process control around the system Git executable;
- deterministic canonical JSON hashing;
- JSON Schema validation without network dependence at runtime;
- strict fail-closed handling of ambiguous JSON;
- a small dependency and deployment surface;
- reproducible builds and pinned authority-path dependencies.

## Decision

Threadkeeper Core v1 will be implemented in **Go**.

The initial implementation baseline is **Go 1.26**, with the release/CI toolchain pinned to the current accepted patch level through repository configuration. At the time of this decision the current stable patch is 1.26.5. Go 1.27 is not yet released and is not part of the baseline.

Durable Threadkeeper JSON will use:

- **RFC 8785 JSON Canonicalization Scheme (JCS)** for canonical bytes;
- **SHA-256** over the canonical digest payload defined by ADR-002;
- **JSON Schema Draft 2020-12** for durable record schemas;
- `github.com/gowebpki/jcs` **v1.0.1** as the initial Go JCS implementation, behind an internal canonicalization interface and pinned by `go.mod`/`go.sum`;
- `github.com/santhosh-tekuri/jsonschema/v6` **v6.0.2** as the initial JSON Schema validator, explicitly configured for Draft 2020-12 and pinned by `go.mod`/`go.sum`.

The system **Git CLI** remains the v1 authority-path Git implementation. Threadkeeper will not make a Go-native Git library the source of truth for commit/ref semantics in v1.

## Why Go

### Operational fit

Go produces a single compiled service binary with a small runtime footprint and straightforward Linux service deployment. The language/runtime does not require a separately managed application interpreter environment in production.

### Correctness fit

Go provides strong static typing, explicit error handling, memory safety for ordinary code, mature standard cryptographic primitives, good subprocess control, and simple test/fuzz tooling. These are more relevant to Threadkeeper's authority path than maximum language expressiveness.

### Git fit

ADR-002 depends on exact Git object/ref semantics, particularly compare-and-swap ref advancement. Calling the installed Git executable keeps those semantics with Git itself rather than reimplementing them through a second object model.

### Maintenance fit

Go's compatibility policy and conservative language/tooling model are a good match for a long-lived infrastructure service whose durable data may outlive many application versions.

## Alternatives considered

### Rust

Rust was a serious candidate. It provides stronger compile-time guarantees around ownership and memory and is suitable for a durable infrastructure service.

It is not selected for v1 because Threadkeeper's primary risk is semantic authority corruption rather than memory-unsafe systems programming. Rust's additional implementation and dependency complexity would not materially improve the Git CAS, canonicalization, schema, provenance, or recovery contracts. Rust remains a valid future implementation language if evidence shows Go is insufficient.

### Python

Python would reduce initial coding time and has strong JSON/schema libraries. It is not selected for the Core authority service because it adds interpreter/environment/package-management state to production and weakens the single-artifact deployment goal. Python remains appropriate for offline analysis, migration tooling, test or research utilities that do not own authority.

### Node.js / TypeScript

Node has excellent JSON ergonomics and a mature JCS ecosystem, but introduces a larger runtime/package surface for a service whose central duties are Git orchestration, validation, hashing and deterministic replay. It is not selected for the authority service.

## Go toolchain policy

1. The repository records the minimum Go language/toolchain family required by the source.
2. CI and release builds pin an exact accepted Go patch release.
3. Toolchain upgrades occur through reviewed PRs with the complete authority-path and recovery suites passing.
4. An unreleased Go version must not be required by authoritative production code.
5. `CGO_ENABLED=0` is the default release-build target unless a separately reviewed dependency proves cgo is necessary.
6. Production startup reports the build's Go version and dependency/build metadata in an inspectable diagnostic endpoint or command.

## Git integration rule

V1 uses the external Git executable as a required operational dependency.

Core must capability-check the required Git operations at startup or installation time rather than guessing from a version string alone. Required capabilities include at minimum:

- object/tree/commit creation needed by the implementation;
- exact object-ID resolution;
- compare-and-swap authoritative ref update;
- repository integrity checking;
- read-only historical inspection.

All Git command execution must:

- use argument arrays, never shell-concatenated command strings;
- set an explicit repository path/environment;
- capture exit status/stdout/stderr separately;
- enforce timeouts/cancellation;
- treat unexpected output or non-zero status as failure;
- avoid ambient user Git configuration where it can change authority semantics;
- never report an authority transition as successful before the exact ref advancement is verified.

A Go Git library may later be used for read-only acceleration or replaced into the write path only through a separately reviewed equivalence decision and conformance evidence.

## Canonical JSON decision

Threadkeeper adopts RFC 8785 JCS rather than inventing a custom key-order/whitespace convention.

The canonicalization library is an implementation dependency, not authority. The normative behavior is RFC 8785 plus Threadkeeper's stricter input profile and committed conformance vectors.

### Strict Threadkeeper JSON profile

Before a durable record is accepted or hashed, the strict decoder must reject at minimum:

- invalid UTF-8;
- duplicate object member names at any nesting depth;
- invalid JSON syntax;
- non-finite numeric forms;
- negative zero (`-0`) on the authority path;
- values that violate the active JSON Schema;
- any value outside Threadkeeper's declared numeric representation rules.

Threadkeeper-defined durable schemas should avoid floating-point values where exact semantic identity matters. Exact decimal quantities, large integers, money, hashes, IDs and timestamps must use schema-defined integer or string representations that cannot change meaning through IEEE-754 conversion.

Unicode strings are **not normalized** as part of JCS. NFC and NFD are distinct byte/semantic inputs unless a particular field schema explicitly defines a normalization rule before record construction.

## Digest procedure

For a durable event containing `content_sha256`:

1. Strictly parse the candidate JSON and reject ambiguous/invalid input.
2. Validate against the exact locally available Draft 2020-12 schema version.
3. Construct the digest payload by removing the `content_sha256` member exactly as defined by the event schema.
4. Canonicalize that payload with RFC 8785 JCS.
5. Compute SHA-256 over the canonical UTF-8 bytes.
6. Encode the digest in the schema-defined lowercase hexadecimal form.
7. Insert `content_sha256`.
8. Validate the completed record again.
9. JCS-canonicalize the completed record for durable storage.
10. During verification, repeat steps 1-6 and compare the stored digest in constant-time where practical.

No pretty-printed or ordinary `encoding/json` serialization is ever used as hash input.

## JSON Schema decision

Durable Threadkeeper schemas use **JSON Schema Draft 2020-12**.

Rules:

- every durable schema has a stable `$id` and explicit `$schema` dialect;
- schema versions are immutable once used by accepted ledger records;
- a breaking semantic change creates a new schema version and migration path rather than mutating historical interpretation;
- runtime authority validation resolves schemas from the durable ledger/bundled schema registry, not from the public network;
- network retrieval of `$ref` targets is disabled on the authority path;
- unknown/unavailable schema versions fail closed;
- schema compilation errors fail startup or the governed operation rather than degrading to permissive validation;
- `format` assertions are enabled only where Threadkeeper schemas deliberately depend on them and are covered by tests.

## Dependency policy

Authority-path third-party modules are pinned to exact accepted versions in `go.mod`/`go.sum`.

Upgrades require:

1. dependency diff/release review;
2. unit and integration tests;
3. RFC 8785 golden-vector conformance where canonicalization is affected;
4. JSON Schema conformance tests where validation is affected;
5. durable replay/recovery tests;
6. an explicit PR rather than floating `latest` production builds.

A dependency is replaceable. Durable record semantics are not.

## Conformance requirements

The implementation must commit golden canonicalization fixtures so production correctness does not depend on a live third-party service or mutable web documentation.

At minimum the canonicalization suite covers:

- RFC 8785 numeric examples;
- recursive object-key ordering;
- array-order preservation;
- escaped/control characters;
- Unicode supplementary-plane sorting behavior;
- NFC versus NFD distinction;
- null versus absent field distinction;
- duplicate-name rejection by Threadkeeper's strict layer;
- negative-zero rejection;
- digest omission/insertion boundary.

The schema suite must validate Threadkeeper schemas under Draft 2020-12 and include positive and negative fixtures for every governance event type before authoritative writes are enabled.

## Consequences

### Positive

- one primary implementation language and service artifact;
- no AI/runtime coupling;
- canonical durable bytes are governed by a published specification;
- schema behavior is explicit and versioned;
- Git remains the source of Git write semantics;
- JCS and schema libraries can be replaced without changing durable meaning;
- dependency upgrades become reviewable events.

### Costs

- Go 1.26's standard `encoding/json` does not itself provide all strict JSON checks required by Threadkeeper, so a small audited strict-input layer is required;
- the JCS library is security/correctness critical and therefore needs golden-vector tests independent of the package's own tests;
- the system Git executable remains an operational dependency;
- exact numeric policies require discipline in schema design.

These costs are accepted because they make ambiguity explicit instead of hiding it behind convenient parsers.

## Reopening conditions

Reconsider the runtime/tooling choice if evidence shows one or more of the following:

- Go cannot meet a required Threadkeeper contract without unsafe or fragile machinery;
- Git CLI process orchestration cannot provide the required crash/concurrency semantics;
- the selected JCS implementation cannot match the committed RFC 8785 conformance corpus;
- the selected JSON Schema validator materially diverges from required Draft 2020-12 behavior;
- deployment constraints make Go materially less reliable than an alternative;
- a different implementation demonstrably lowers authority risk without weakening portability/recovery.

Developer preference alone is not a reopening condition.
