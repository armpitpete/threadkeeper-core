# Runtime and Serialization Acceptance Gates

## Status

Candidate until merged to the default branch. These gates become normative for the v1 implementation when present on the default branch.

## Purpose

These tests prove that the selected Go/JCS/JSON Schema stack preserves Threadkeeper's authority semantics before the service is permitted to write accepted governance events.

A test implementation may use temporary repositories and fixtures. It must not weaken expected production behavior.

## RS-01 — Toolchain identity

**Given** a release build  
**When** build metadata is inspected  
**Then** the exact source commit, Go toolchain version, dependency versions and target platform are recoverable.

## RS-02 — No unreleased toolchain dependency

The accepted baseline must build and test using an officially released Go 1.26 patch toolchain without requiring Go 1.27 or experimental language/library features.

## RS-03 — CGO-free baseline

The default Linux release build succeeds with `CGO_ENABLED=0`.

If this later becomes false, the exception requires a separately reviewed architecture change.

## RS-04 — Duplicate root member rejection

Input:

```json
{"event_id":"A","event_id":"B"}
```

must fail with `DUPLICATE_MEMBER` before conversion into a map/struct or hash computation.

## RS-05 — Duplicate nested member rejection

A duplicate member at arbitrary nesting depth must fail identically.

## RS-06 — Invalid UTF-8 rejection

Any durable candidate containing invalid UTF-8 fails closed before schema validation or hashing.

## RS-07 — Negative zero rejection

Candidate JSON containing `-0` in an authority-path number position must fail with `INVALID_NUMBER`.

## RS-08 — Non-finite value impossibility

NaN, Infinity and equivalent non-JSON forms must be rejected as invalid JSON and can never reach JCS hashing.

## RS-09 — Unicode non-normalization

Fixtures containing canonically equivalent but byte/code-point-distinct NFC and NFD strings must remain distinct unless the field's own schema explicitly requires normalization.

JCS must not silently normalize them.

## RS-10 — Object order invariance

Two semantically identical objects differing only in object member order must produce byte-identical JCS output and identical digest payload hashes.

## RS-11 — Array order significance

Two arrays containing the same members in different order must produce different canonical bytes unless the arrays themselves happen to be byte-identical.

Threadkeeper must never sort arrays as part of generic canonicalization.

## RS-12 — Null versus omission significance

A field explicitly present as `null` and the same field omitted entirely must remain distinct inputs and produce distinct canonical records where the schema permits both.

## RS-13 — RFC 8785 numeric corpus

The committed canonicalization fixtures must cover the RFC 8785 number serialization examples relevant to the implementation and match expected canonical bytes exactly.

## RS-14 — Unicode property ordering

Committed fixtures must cover RFC 8785/JCS property-name sorting, including non-ASCII and supplementary-plane cases, and match expected bytes exactly.

## RS-15 — JCS wrapper replacement equivalence

The golden corpus is tested through Threadkeeper's internal canonicalization interface.

Replacing `gowebpki/jcs` with another implementation without changing fixtures must either pass byte-for-byte or be rejected as a semantic change.

## RS-16 — Digest boundary

For an event containing `content_sha256`:

1. remove exactly that member as defined by the schema;
2. JCS-canonicalize the remaining digest payload;
3. SHA-256 the resulting bytes;
4. encode lowercase hexadecimal;
5. compare to the stored field.

The test must prove that including `content_sha256` in its own hash input fails the fixture.

## RS-17 — Stored record is canonical

The completed durable record, including `content_sha256`, is stored in RFC 8785 canonical form.

Recanonicalizing stored bytes must be byte-identical.

## RS-18 — Exact schema dialect

Every v1 durable schema explicitly declares JSON Schema Draft 2020-12.

A schema with an unknown or unsupported dialect fails closed.

## RS-19 — Schema resource is local

Validation and recovery pass with all external network access disabled.

No accepted schema or `$ref` requires runtime internet retrieval.

## RS-20 — Missing schema fails closed

A durable record naming an unavailable schema version cannot be accepted or replayed as authoritative current state.

## RS-21 — Historical schema immutability

After a newer schema version is introduced, a historical accepted record still validates/replays according to the exact schema version it originally named.

## RS-22 — Positive/negative fixtures

Every durable event schema has at least:

- one minimum valid fixture;
- one representative full valid fixture;
- missing-required-field failure;
- unknown/disallowed-field behavior fixture where relevant;
- wrong-type failure;
- malformed authority/provenance failure where schema-expressible.

## RS-23 — Unknown-field preservation boundary

No authority path may parse a durable record into a typed Go struct and silently discard unknown members before schema/version handling has determined whether those members are valid.

A fixture must prove unknown data cannot disappear unnoticed.

## RS-24 — Git command injection resistance

A user-controlled value containing shell metacharacters, spaces, quotes and newline characters must remain a literal argument/data value and must not cause shell execution.

The authority Git path must invoke Git without a shell.

## RS-25 — Ambient Git config isolation

Tests deliberately configure hostile or surprising global Git settings/hooks/replacement refs and prove the controlled ledger operation either ignores them or fails closed according to the declared Git environment policy.

## RS-26 — Git capability check

Startup/install validation proves the configured Git executable supports the exact operations Threadkeeper requires before authoritative writes are enabled.

## RS-27 — Git timeout/cancellation

A deliberately blocked child Git process is cancelled or timed out without reporting a successful authority transition.

## RS-28 — Git stderr is not success

A failed Git command with plausible-looking stdout/stderr never becomes a successful machine state solely because text output contains words such as `success`, `updated` or a commit-looking hash.

## RS-29 — CAS remains the acceptance boundary

Creating candidate Git objects/commit must not alter authoritative state.

Only a successful exact expected-old -> new authoritative ref update can accept H1.

## RS-30 — Post-CAS verification

After a successful CAS, Core re-resolves the authoritative ref and verifies it equals H1 before returning a successful governed result.

## RS-31 — Canonicalization fuzzing

Fuzz input must not cause panic, uncontrolled resource growth, silent duplicate collapse or non-deterministic canonical bytes.

## RS-32 — Strict JSON fuzzing

Fuzzing targets nesting, malformed strings/escapes, Unicode edge cases, duplicate names and number forms.

Every accepted input must have one unambiguous parsed meaning.

## RS-33 — Schema resolver fuzz/hostility

Malicious or cyclic references must not trigger unbounded network retrieval or reinterpret a different schema as the requested accepted version.

## RS-34 — Dependency pinning

CI fails if authority-path Go dependencies are added without being represented in `go.mod` and integrity-covered by `go.sum`.

Release builds must not resolve floating `latest` dependencies.

## RS-35 — Dependency upgrade regression

Any change to the JCS or JSON Schema validator dependency must run the entire runtime/serialization corpus plus ADR-002 destructive recovery tests.

## RS-36 — Cross-platform semantic equivalence

Where a second supported OS/architecture is added, the same committed durable fixtures must produce identical JCS bytes and SHA-256 values.

Git commit object IDs may differ if commit metadata differs, but Threadkeeper logical event IDs/content digests and replayed governed meaning must not.

## Write-enablement gate

Threadkeeper Core **MUST NOT enable authoritative event writes** until RS-01 through RS-30 pass in CI and the implementation also passes the relevant ADR-002 durable-ledger tests.

RS-31 through RS-36 are release-quality gates and must be active before a v1 production release claims conformance.
