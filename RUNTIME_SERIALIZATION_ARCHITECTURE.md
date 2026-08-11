# Runtime and Serialization Architecture v1

## Status

Candidate until merged to the default branch. When present on the default branch, this document defines the v1 Core implementation/runtime and durable serialization architecture under ADR-003.

## 1. Selected stack

```text
Threadkeeper Core service
  language/runtime: Go 1.26
  initial release toolchain: Go 1.26.5

Authority ledger
  implementation semantics: system Git CLI
  durable authority: exact Git commit

Durable record format
  syntax: JSON
  strict profile: Threadkeeper I-JSON/JCS profile
  canonicalization: RFC 8785 JCS
  content digest: SHA-256
  schema dialect: JSON Schema Draft 2020-12

Initial pinned libraries
  JCS: github.com/gowebpki/jcs v1.0.1
  JSON Schema: github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
```

The library names above are implementation choices. RFC 8785, Draft 2020-12, the Threadkeeper strict-input profile and accepted schemas are the semantic authority.

## 2. Process boundary

Threadkeeper Core v1 is one service process for the authority path.

Conceptually:

```text
clients
  │
  ▼
Core interface
  │
  ├─ request validation
  ├─ authority/policy validation
  ├─ strict JSON + schema validation
  ├─ JCS + SHA-256
  ├─ Git candidate commit construction
  └─ exact-head CAS ref advancement
         │
         ▼
   durable ledger.git
```

Recall, vector search and AI clients remain outside this process's authority semantics even if later deployed on the same machine.

## 3. Package boundary

The implementation should begin with explicit internal boundaries rather than one large service package:

```text
cmd/threadkeeper-core/
internal/
  authority/
  canonicaljson/
  digest/
  gitledger/
  ledger/
  provenance/
  schema/
  strictjson/
  projection/
  recovery/
  service/
```

Names may change without architectural significance, but the responsibilities must remain separated enough to test independently.

### `strictjson`

Validates raw candidate bytes before semantic decoding can erase ambiguity.

### `schema`

Owns the locally resolved Draft 2020-12 schema registry and validation rules.

### `canonicaljson`

Wraps the pinned RFC 8785 implementation and exposes no library-specific types to the rest of Core.

### `digest`

Owns the exact digest omission/canonicalization/SHA-256 procedure.

### `gitledger`

Owns all Git process invocation and exact-head CAS behavior.

### `ledger`

Coordinates logical durable events, stable IDs, accepted commit identity and replay.

### `projection`

Builds current/materialized state from durable history. It is rebuildable and not authority.

### `recovery`

Performs no-AI/no-Recall integrity validation and replay.

## 4. No shell authority path

The Core implementation must execute Git directly as a child process.

Forbidden on the authority path:

```text
sh -c "git ..."
bash -c "git ..."
cmd /c "git ..."
powershell -Command "git ..."
```

Git arguments are passed as an argument vector. User-controlled values are never interpolated into a shell command.

## 5. Git environment isolation

Each Git invocation must be given an explicit repository and controlled environment.

The implementation must prevent ambient user configuration from silently changing authority semantics. At minimum tests must cover effects from:

- global/system Git config;
- hooks;
- alternate object directories;
- replacement refs;
- environment-provided Git paths;
- signing defaults;
- pager/editor invocation;
- locale-sensitive output where parsing occurs.

Where Git provides machine-stable exit status or plumbing output, use that rather than parsing human-facing prose.

## 6. Strict JSON before ordinary decode

Authority JSON must not be decoded directly with a permissive object decoder before duplicate-name and syntax checks have run.

Required order:

```text
raw UTF-8 bytes
      │
      ▼
strict syntax/I-JSON checks
      │
      ▼
JSON value representation preserving number text where required
      │
      ▼
schema validation
      │
      ▼
canonicalization/digest
```

A strict-input failure is a governed write failure, not a warning.

## 7. Duplicate member names

Duplicate JSON object names are forbidden at every nesting depth.

This check must operate on raw parsed token structure before conversion into a Go `map`, because a map cannot preserve evidence that a duplicate appeared.

The duplicate check is case-sensitive at the JSON level. Schema-level field-name rules may impose additional restrictions.

## 8. UTF-8 and strings

Durable JSON is UTF-8.

Invalid UTF-8 fails closed.

JCS does not normalize Unicode. Therefore:

```text
"é" as U+00E9
```

and

```text
"e" + U+0301
```

remain different inputs and can produce different digests.

If a particular Threadkeeper field requires normalization, the schema/domain constructor must perform and test that rule before durable record construction. Canonicalization itself must not add hidden normalization.

## 9. Numeric policy

JCS number serialization follows its specified IEEE-754/ECMAScript model. Threadkeeper must not use that fact as permission to put semantically exact values into floating point fields casually.

For Threadkeeper-defined authority schemas:

- IDs are strings, not JSON numbers;
- SHA/digest values are strings;
- timestamps are RFC 3339-style strings unless a later schema explicitly chooses another exact representation;
- monetary/decimal values use domain-defined strings or integer minor units;
- counters may use JSON integers only inside schema-defined safe bounds;
- integers outside the interoperable exact range are represented as strings with explicit format semantics;
- `-0` is rejected;
- NaN and infinities are impossible/invalid JSON and rejected.

A schema that introduces a floating-point field into durable governance data requires explicit review and canonicalization fixtures for its edge cases.

## 10. Canonicalization boundary

The rest of Core must call an internal interface conceptually equivalent to:

```text
Canonicalize(validatedJSONBytes) -> canonicalUTF8Bytes
```

No caller is allowed to rely on package-specific internal types or ordering behavior.

The initial implementation delegates RFC 8785 transformation to the pinned `gowebpki/jcs` module.

Golden tests define accepted output independently of that module. If the module is replaced, the same golden corpus must pass byte-for-byte.

## 11. Schema registry

The durable ledger is the runtime source of accepted Threadkeeper schemas.

Conceptual layout under ADR-002:

```text
config/schemas/
  event/
  source/
  authority/
  relationship/
  migration/
```

Each accepted schema has:

- an immutable schema version/identity;
- `$schema` identifying Draft 2020-12;
- a stable `$id`;
- committed bytes in the ledger;
- tests/fixtures proving intended positive and negative cases.

Historical records continue to validate against the schema version they name.

## 12. No network schema resolution

Runtime authority validation must be deterministic and offline-capable.

The schema compiler must not retrieve `$ref`, `$dynamicRef`, meta-schema or vocabulary resources from arbitrary network locations during a governed write or recovery operation.

Required schema/meta-schema resources are bundled, registered locally, or otherwise made part of the accepted deployment/ledger inputs.

An unresolved reference fails closed.

## 13. Validation order for accepted events

For a candidate event:

```text
1. receive candidate bytes/object
2. strict JSON validation
3. resolve named schema from exact ledger state H0
4. validate digest-free candidate semantics
5. validate authority/provenance/policy against H0
6. derive digest payload by schema-defined omission rule
7. RFC 8785 canonicalize digest payload
8. SHA-256 digest canonical bytes
9. insert content_sha256
10. validate completed event again
11. RFC 8785 canonicalize completed stored event
12. construct Git candidate commit H1
13. CAS refs/heads/main H0 -> H1
14. re-resolve ref and verify exact H1
15. return success
```

Any failure before step 13 leaves H0 authoritative.

Any candidate Git objects created before a failed CAS are not authority.

## 14. Go serialization rules

Go structs may be used as typed in-memory domain models, but durable bytes are governed by schema + JCS, not by the incidental output of struct marshaling.

Rules:

- `encoding/json.Marshal` output is not a durable canonical format;
- struct tags are implementation mapping, not schema authority;
- omitted versus explicit `null` must be intentional and schema-tested;
- unknown durable fields must not be silently discarded before validation/replay decisions;
- number handling must avoid conversion to `float64` where exact lexical/value identity matters;
- round-trip tests must prove parse -> validate -> canonicalize stability.

## 15. Error model

Authority-path errors are typed/machine-inspectable.

At minimum distinguish:

```text
INVALID_JSON
DUPLICATE_MEMBER
INVALID_UTF8
INVALID_NUMBER
UNKNOWN_SCHEMA
SCHEMA_INVALID
SCHEMA_VALIDATION_FAILED
AUTHORITY_DENIED
PROVENANCE_INCOMPLETE
DIGEST_MISMATCH
GIT_FAILURE
STALE_STATE
IDEMPOTENCY_CONFLICT
INTEGRITY_FAILURE
RECOVERY_REQUIRED
```

Natural-language messages may accompany these codes but cannot replace them.

## 16. Build/release baseline

Initial baseline:

```text
Go family: 1.26
accepted patch at decision time: 1.26.5
release target: linux/amd64 first
CGO_ENABLED=0 by default
```

Additional targets may be added after tests prove equivalence.

The repository must capture dependency checksums and build metadata sufficient to identify:

- source commit;
- Go version;
- dependency versions;
- build mode/target;
- Threadkeeper contract/schema version.

## 17. Dependency surface

The authority path starts with two deliberate third-party Go dependencies:

```text
github.com/gowebpki/jcs v1.0.1
github.com/santhosh-tekuri/jsonschema/v6 v6.0.2
```

Other dependencies require justification. Convenience libraries should not be added where the standard library is sufficient.

A production dependency update is a governed code change, not an automatic floating update.

## 18. Fuzzing targets

Go fuzz tests should target at least:

- strict JSON token handling;
- nested duplicate-name detection;
- canonicalization wrapper invariants;
- digest omission/insertion logic;
- schema registry resolution;
- event replay;
- Git command argument construction;
- parser handling of hostile Unicode/number inputs.

Fuzz findings that could alter durable interpretation block release until resolved or explicitly governed.

## 19. Initial implementation sequence

After this architecture is accepted, implement in this order:

1. Go module/toolchain pin and minimal CLI/service skeleton.
2. strict JSON validator.
3. local Draft 2020-12 schema registry.
4. RFC 8785 canonicalizer wrapper + golden corpus.
5. SHA-256 digest procedure.
6. read-only Git ledger open/inspect/fsck.
7. deterministic event replay/projector.
8. candidate Git commit construction.
9. exact-head CAS write path.
10. idempotency and stale-state tests.
11. destructive recovery suite from ADR-002.
12. only then expose a client write interface.

## 20. Explicitly deferred

Still not selected by this decision:

- event-ID format;
- public API transport (HTTP/gRPC/MCP/etc.);
- authentication/identity provider;
- Recall database;
- FTS implementation;
- vector index;
- embedding provider/model;
- backup product/provider;
- signing/authenticity scheme.

These remain downstream decisions.

## 21. Next gate after acceptance

Once this runtime/serialization architecture is accepted, the next substantive lane is **the minimal executable Core skeleton plus strict JSON/JCS/schema conformance harness**.

That implementation must prove serialization correctness before it is allowed to write authoritative Git events.
