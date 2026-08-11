# Durable Storage Architecture v1

## Status

Candidate until merged to the default branch. When present on the default branch, this document defines the v1 durable Threadkeeper Core storage architecture under ADR-002.

## 1. Boundary

Threadkeeper Core durable storage is not the same thing as Threadkeeper Recall.

The durable layer exists to preserve the minimum information that must survive complete deletion of:

- Recall databases and indexes;
- embeddings and model-derived representations;
- AI/model installations;
- temporary work queues;
- client sessions;
- local caches.

The selected v1 durable architecture is a **Git-backed governance ledger containing canonical JSON records**.

## 2. Logical architecture

```text
Configured authoritative project sources/sinks
        │
        │ exact immutable versions / decisions
        ▼
┌───────────────────────────────────────────────┐
│ THREADKEEPER CORE DURABLE LEDGER              │
│                                               │
│ Git repository                               │
│ authoritative ref: refs/heads/main           │
│                                               │
│ config/                                      │
│   sources/                                   │
│   authority/                                 │
│   relationships/                             │
│   schemas/                                   │
│   migrations/                                │
│                                               │
│ events/                                      │
│   decisions/                                 │
│   supersessions/                             │
│   revocations/                               │
│   governance/                                │
│                                               │
│ checkpoints/                                 │
│   expected-invariants.json                   │
│                                               │
│ exports/  (generated, not committed by default)│
└──────────────────────┬────────────────────────┘
                       │
                       │ deterministic replay
                       ▼
              Materialized Core state
              (rebuildable projection)
                       │
                       ▼
                    Recall
              (separate decision later)
```

## 3. Authority root

The authority root is an exact Git commit, never merely the text `main`.

At runtime Threadkeeper may refer to the configured authoritative ref, initially:

```text
refs/heads/main
```

but every governed read or write that depends on current ledger state must resolve and expose the exact commit object ID.

The branch/ref is a locator. The commit is the exact state identity.

## 4. Canonical record layout

The exact directory names may evolve through governed migrations, but the v1 logical categories are fixed.

### `config/sources/`

Declares source identities, source types, immutable-version rules, and configured authoritative sinks.

### `config/authority/`

Contains authority policies defining which source or decision mechanism may establish which governed facts.

### `config/relationships/`

Defines governed relationship types and their semantics.

### `config/schemas/`

Contains schema/version declarations required to interpret durable records.

### `config/migrations/`

Contains migration declarations needed to interpret older ledger states under newer software.

### `events/decisions/`

Immutable accepted Threadkeeper-native decision events.

### `events/supersessions/`

Immutable events that replace the active effect of prior accepted records while preserving history.

### `events/revocations/`

Immutable events that explicitly revoke prior authority/effect.

### `events/governance/`

Other governance-significant transitions that must survive Recall loss.

### `checkpoints/expected-invariants.json`

Machine-checkable invariants used during recovery verification. This is not a cache of current state; it describes properties the recovered ledger must satisfy.

## 5. Record identity

Each accepted event has three separate identities:

1. **Logical event ID** — stable Threadkeeper identity used by clients and relationships.
2. **SHA-256 content digest** — implementation-independent integrity identity over canonical event content.
3. **Git state identity** — the commit containing the accepted event.

These must not be collapsed into one identifier.

A future Git hash-algorithm migration must not change the logical identity of a Threadkeeper event.

## 6. Canonical JSON requirements

Durable records must be deterministic enough to hash, export and compare reliably.

The v1 implementation must therefore define one canonical JSON serialization profile before authoritative writes are enabled.

At minimum it must define:

- UTF-8 encoding;
- deterministic object-key ordering;
- no semantically irrelevant whitespace in hashed form;
- exact numeric representation rules;
- timestamp format and timezone rules;
- treatment of absent versus explicit `null` values;
- normalization rules for strings where applicable;
- schema version field;
- digest field computation boundaries.

The specific canonicalization standard is a later implementation detail, but no authoritative event writing may ship before it is selected and tested.

## 7. Immutable event minimum envelope

Conceptually, each event must carry:

```json
{
  "schema_version": "...",
  "event_id": "...",
  "event_type": "...",
  "occurred_at": "...",
  "actor": { "type": "...", "id": "..." },
  "expected_ledger_commit": "...",
  "authority_policy_version": "...",
  "targets": [],
  "source_versions": [],
  "prior_state": {},
  "resulting_state": {},
  "reason": "...",
  "idempotency_key": "...",
  "content_sha256": "..."
}
```

This is illustrative, not yet the final JSON Schema.

## 8. Protected write algorithm

A protected write must behave as one logical authority transition:

```text
1. Resolve authoritative ref -> H0
2. Require client's expected head == H0
3. Validate actor/authorisation
4. Validate source-version evidence
5. Validate proposal against active schemas/policy at H0
6. Check idempotency key against accepted ledger history/projection
7. Create immutable canonical event record(s)
8. Verify content SHA-256
9. Create tree + commit H1 with parent H0
10. Atomically advance authoritative ref H0 -> H1 only if ref still == H0
11. If compare-and-swap fails: reject as STALE_STATE
12. If it succeeds: return H1 as accepted ledger state
```

No client-facing success may be returned before step 10 succeeds.

A commit created in step 9 but not referenced by the authoritative ref is harmless candidate data and may later be garbage-collected. It is not accepted state.

## 9. Idempotency

Every authority-changing request must carry a stable idempotency key.

If the same accepted request is retried:

- Threadkeeper must return the existing accepted decision identity/state;
- it must not create a second logical decision;
- conflicting reuse of the same key with different content must fail closed.

An idempotency projection may be cached, but the ability to reconstruct it must come from the durable ledger.

## 10. Single-writer rule

The initial durable ledger has one authoritative writer: Threadkeeper Core/Manager.

Clients do not receive direct filesystem or Git write authority to the ledger.

Clients submit proposals/decision requests through the Core interface. Core validates them and performs the protected ledger update.

Read-only inspection of the ledger may be exposed separately.

This rule reduces the concurrency model to serial authoritative transitions while retaining stale-write protection against delayed or duplicated clients.

## 11. Dedicated storage area

The deployment target must maintain an independently manageable Threadkeeper storage root, conceptually:

```text
/threadkeeper/
├── durable/
│   ├── ledger.git/          # primary authoritative Git repository
│   └── recovery/            # local recovery metadata/tools, if required
├── projections/             # rebuildable Core materialized views
├── recall/                  # later technology decision
├── transient/               # queues/scratch/session state
├── backups/                 # optional local staging, not sole backup
└── logs/                    # operational logs, non-authoritative unless explicitly ingested
```

The actual platform path may differ on Windows/Linux. The separation is semantic and operational, not dependent on one pathname.

AI model files, Ollama/model caches, provider SDK caches and AI session storage must live outside this Threadkeeper storage root or be independently isolated such that AI cleanup cannot remove it.

## 12. Repository form

The runtime ledger should normally be maintained as a **bare Git repository** or equivalently protected service-owned repository rather than a user-edited working tree.

Reasons:

- the authoritative object/ref database is what matters;
- accidental working-tree edits cannot be mistaken for accepted state;
- service ownership and permissions are clearer;
- backup can target the repository itself;
- clients need not manipulate a checkout.

Temporary trees/indexes used to construct commits are implementation details and never authority.

## 13. Remote mirror rule

A remote mirror may be used for backup, audit convenience or disaster recovery.

A hosted remote is not automatically the authority root.

The primary v1 authority rule is configured explicitly. If the local ledger is the primary authoritative sink, a remote outage must not prevent local governance operation unless policy intentionally requires remote durability before acceptance.

If policy later requires synchronous remote durability, that is a separate governance/availability decision because it changes when an authority transition is considered complete.

## 14. Backup requirements

The production implementation must have at least two independent recoverable copies of durable ledger data, with at least one copy outside the primary Threadkeeper storage device.

Backup procedure must preserve:

- authoritative refs;
- all reachable Git objects;
- repository configuration required for interpretation;
- Threadkeeper schemas/configuration inside the ledger;
- integrity-verification metadata.

A mere copy of a checked-out working tree is not a complete ledger backup.

Backup success must be tested by restoration, not inferred from file-copy success.

## 15. Recovery sequence

A clean recovery must be possible without AI and without Recall:

```text
restore ledger repository
        ↓
Git integrity check
        ↓
resolve authoritative ref + exact commit
        ↓
validate durable JSON schemas/digests
        ↓
replay accepted events/config
        ↓
verify expected invariants
        ↓
rebuild materialized Core projections
        ↓
(optional later) rebuild Recall
```

If any stage cannot establish integrity or interpretation, Core must enter a read-only/recovery-failed state rather than inventing current truth.

## 16. Corruption and tamper detection

The initial integrity stack is layered:

- Git object/graph integrity;
- expected parent/ref history;
- Threadkeeper schema validation;
- canonical-record SHA-256 validation;
- provenance/source-version validation;
- replay/invariant checks.

Cryptographic commit signing is not required for v1 architecture conformance. It may be added later as an authenticity hardening layer without changing ledger semantics.

## 17. Filesystem and network constraints

The authoritative ledger must reside on storage with normal local filesystem semantics suitable for Git repository locking and atomic ref updates.

Do not place the primary authoritative Git repository on consumer sync folders or network filesystems whose locking/atomic-rename behavior is not explicitly verified.

Replication to network/cloud storage occurs through supported Git/backup mechanisms rather than by treating a live synced folder as the database.

## 18. Relationship to SQLite

SQLite is deliberately **not** the v1 authority root.

A later projection/Recall decision may select SQLite for:

- materialized current-state tables;
- source metadata projections;
- relationship traversal indexes;
- FTS search;
- incremental indexing checkpoints;
- idempotency acceleration;
- local query acceleration.

Any such SQLite database must remain rebuildable from the durable Git ledger plus declared authoritative project sources where required.

SQLite's transaction guarantees make it well suited to those roles, but the binary database must never become an undocumented second authority.

## 19. Acceptance tests for durable storage implementation

The first implementation is not complete until automated tests prove:

### DS-01 — AI removal
Delete/disable all AI/model services. Read and validate the durable ledger successfully.

### DS-02 — Recall destruction
Delete Recall completely. Recover accepted governance state from durable sources.

### DS-03 — Expected-head rejection
Prepare a write against H0, advance the ledger independently to H1, then attempt the H0 write. It must fail with stale-state semantics.

### DS-04 — Crash before ref advance
Create candidate objects/commit but interrupt before authoritative ref update. Recovery must still expose H0 as authority.

### DS-05 — Crash after ref advance
Interrupt immediately after successful ref advancement. Recovery must expose the new exact commit and complete accepted event.

### DS-06 — Replay protection
Submit the same idempotent decision twice. Exactly one logical event must exist.

### DS-07 — Conflict preservation
Record conflicting authoritative evidence. Both records must remain inspectable until a governed resolution event exists.

### DS-08 — Supersession history
Supersede an accepted record. Both prior and replacement states plus the transition must remain recoverable.

### DS-09 — Digest corruption
Alter canonical event content outside normal accepted history and verify integrity checks detect mismatch/tampering.

### DS-10 — Bare restore
Restore the ledger to a clean environment and reconstruct state with no original working directory.

### DS-11 — Portable export
Export durable logical records in deterministic implementation-independent form and validate equivalence after re-import into a clean test projection.

### DS-12 — Backup restoration
Restore from the secondary backup copy and pass DS-01, DS-02, schema/digest validation and invariant checks.

## 20. Explicitly deferred decisions

This architecture does not yet choose:

- programming language;
- service framework;
- canonical JSON standard/profile;
- JSON Schema tooling;
- event-ID format;
- SQLite version or use;
- FTS engine;
- vector database/index;
- embedding model;
- HTTP/gRPC/MCP transport;
- authentication provider;
- cryptographic signing scheme;
- backup product/provider.

Those decisions must be made in dependency order against the accepted contracts.

## 21. Next implementation gate

After this durable storage architecture is accepted, the next technology decision is:

> **Select the Core implementation language/runtime and canonical durable-record serialization/schema tooling needed to implement the ledger safely.**

Recall/search/vector technology remains later.
