# ADR-002: Git Is the v1 Durable Governance Ledger

- **Status:** Candidate until merged to the default branch; Accepted when present on the default branch
- **Date:** 2026-08-11
- **Decision scope:** Threadkeeper Core durable storage architecture
- **Depends on:** ADR-001 and Threadkeeper Core Contract Standard v0.1

## Context

Threadkeeper Core needs a durable storage architecture for its own governance-significant state before any Recall, vector, search, or AI technology is selected.

The storage contract requires that:

- accepted project truth and decision history survive complete Recall loss;
- Core can recover without an AI model;
- authoritative state changes are attributable and history-preserving;
- protected writes reject stale expected state;
- exact source versions and provenance remain inspectable;
- durable configuration is versioned;
- storage technology can later be replaced without changing authority semantics;
- incomplete writes fail closed.

The durable Core workload is intentionally not the same as the Recall workload. Durable Core storage contains relatively small, high-value governance records and configuration. Large search indexes, embeddings, cached projections, retrieval statistics, and other rebuildable material belong elsewhere.

## Decision

Threadkeeper Core v1 will use a **dedicated Git repository as its durable governance ledger**.

The authoritative state of the ledger is the exact commit referenced by its configured authoritative ref, initially `refs/heads/main`.

Git is used here as a versioned durable record system, not merely as a source-code collaboration tool.

The ledger will hold canonical, machine-readable records for:

- Core source registry and source identity policy;
- authority policy;
- relationship and type definitions;
- schema and contract versions;
- migration metadata;
- Threadkeeper-native decision events;
- supersession/revocation records;
- durable provenance required to recover governance state;
- rebuild configuration.

The ledger will **not** hold:

- embeddings;
- vector indexes;
- full-text indexes;
- model caches;
- raw AI sessions;
- temporary queues;
- retrieval statistics;
- large derived corpora that can be reconstructed from declared sources.

Those belong to Recall or transient storage.

## Authoritative write model

A governed ledger write is valid only when all of the following are true:

1. the client supplies the expected current ledger commit;
2. the proposed records pass schema and policy validation;
3. immutable event records are written without editing prior events;
4. a new Git commit is created with the expected commit as parent;
5. the authoritative ref is advanced only if it still equals the expected commit;
6. failure to advance the ref means the entire authority transition is rejected as stale;
7. the resulting commit identity is returned as the exact new ledger state.

The implementation MUST use Git compare-and-swap ref semantics, equivalent to `git update-ref <ref> <new-oid> <old-oid>`, for the authoritative ref update.

A working tree is never authority. An unreferenced commit is never accepted state. A branch name without its exact commit identity is never sufficient evidence of state.

## Record format

Durable Threadkeeper-native records will use **UTF-8 canonical JSON documents** with explicit schema versions.

Each immutable governance event will have a stable event identifier independent of Git object identity and will include at minimum:

- schema version;
- event identifier;
- event type;
- affected logical object(s);
- prior state/version where applicable;
- resulting state/version;
- actor or authorised mechanism identity;
- event time;
- reason/decision reference;
- exact source version references;
- authority policy version;
- expected prior ledger commit;
- idempotency key;
- cryptographic content digest using SHA-256.

Git object IDs provide repository identity and history linkage, but Threadkeeper record integrity MUST NOT depend solely on the repository's Git object-hash algorithm. Canonical records therefore carry an implementation-independent SHA-256 digest.

## Append and supersede semantics

Decision events and other governance events are immutable after acceptance.

Corrections do not edit prior accepted events. They append a new event that explicitly supersedes, revokes, corrects, or reopens the earlier state.

Current state is a projection over the accepted ledger history.

Configuration may be changed by later commits, but prior versions remain recoverable from Git history and any migration must preserve interpretation of historical events.

## Physical placement

The primary ledger will live on Threadkeeper Core storage that is independently manageable from AI/model storage.

The v1 design assumes:

- a dedicated Threadkeeper storage area/volume;
- a local primary Git ledger used by the Threadkeeper Manager/service;
- one authoritative writer path through Threadkeeper Core;
- zero or more read-only clients;
- separate backup/mirror targets.

A GitHub or other remote may mirror the ledger, but a hosted provider is not required for Core to recover or operate. Remote availability must not be confused with authority.

## Backup and recovery

The ledger backup design must preserve all reachable objects and authoritative refs.

Recovery must be testable from a clean machine with AI and Recall unavailable.

At minimum, recovery verification will:

1. restore or clone the durable ledger;
2. verify Git object integrity;
3. verify Threadkeeper record schemas and SHA-256 digests;
4. verify the configured authoritative ref;
5. replay accepted governance events into a clean projection;
6. compare the recovered current-state projection against declared expected invariants.

Portable export remains mandatory. Threadkeeper will later provide a deterministic logical export of durable records independent of Git's internal object representation.

## Concurrency model

Threadkeeper Core v1 uses a **single authoritative writer** for the durable ledger.

This is deliberate. Governance writes are low-volume and high-value. A single writer plus expected-head compare-and-swap is simpler to reason about than distributed multi-writer consensus and is sufficient for the initial system.

Multiple readers are permitted.

If a future workload requires multiple independent writers, that is an architecture reopening condition rather than a reason to weaken stale-write protection now.

## Why Git

Git maps unusually well to the contract:

- commits provide immutable exact snapshots linked to prior history;
- content is addressable by object identity;
- history is naturally preserved rather than overwritten;
- ref updates can verify an expected old object before accepting a new one;
- the complete ledger can operate locally without a database server or AI runtime;
- records remain inspectable with ordinary tools;
- backup and replication do not require redefining authority semantics.

Git documentation explicitly supports conditional ref updates using an expected old object ID. This gives Threadkeeper a direct implementation primitive for stale-write rejection.

## Alternatives considered

### SQLite as the sole durable authority store

**Not selected for v1.**

SQLite is a strong embedded transactional database and remains a valid future component. Its ACID transaction semantics and mature backup facilities satisfy many durability requirements.

It is not selected as the primary governance ledger because Threadkeeper's initial authoritative workload benefits more from inspectable immutable history and exact-head compare-and-swap than from database query performance. Selecting both Git and SQLite as co-authorities would also introduce avoidable cross-store consistency problems.

SQLite remains a leading candidate for non-authoritative materialized projections and Recall metadata in a later decision.

### PostgreSQL or another client/server relational database

**Not selected for v1.**

It would provide strong transactions and concurrency, but the initial workload does not justify a continuously administered database server, distributed availability concerns, or a broader operational failure surface.

This may be reconsidered if Threadkeeper becomes a genuine multi-writer service requiring transactional concurrency beyond the single-writer ledger model.

### Plain append-only files without Git

**Rejected.**

This would require Threadkeeper to reimplement version identities, history traversal, object integrity, compare-and-swap head movement, replication conventions, and recovery tooling that Git already supplies.

### Git plus SQLite as two authoritative stores

**Rejected.**

There must be one unambiguous durable authority transition. A design in which a decision is authoritative only after two independent stores commit creates a distributed transaction problem without enough benefit.

A future SQLite database may project the Git ledger, but that projection is rebuildable and non-authoritative.

## Consequences

### Positive

- authority maps to an exact Git commit;
- stale protected writes map to expected-head mismatch;
- accepted events are naturally history-preserving;
- durable state remains human-inspectable;
- AI removal has no effect on governance state;
- no database server is required;
- future SQLite/search/vector choices remain independent;
- the dedicated Threadkeeper storage boundary remains clean.

### Costs

- high-frequency mutation would be inefficient compared with a transactional database;
- very large binary or derived data does not belong in the ledger;
- the service must serialize authoritative writes;
- canonical JSON and schema validation must be implemented carefully;
- backup must protect the repository as a repository, not merely copy a working directory.

These costs are acceptable because the ledger is intentionally small, durable governance state rather than a general-purpose data lake.

## Reopening conditions

Reconsider the durable ledger technology if one or more of these becomes true:

- sustained governance-write volume makes commit-per-transition materially unsuitable;
- independently authenticated multi-writer operation is required;
- atomic transitions must span resources that cannot be represented as one ledger commit;
- measured recovery or integrity requirements cannot be met with the Git ledger;
- a replacement technology demonstrates equal or stronger provenance, stale-write, history, portability, and AI-independence properties with lower total complexity.

Do not reopen this decision merely because another database is faster at queries. Query speed belongs primarily to projections and Recall.

## Primary implementation references

- Git `update-ref` documentation: https://git-scm.com/docs/git-update-ref
- Git user manual, commit/tree/object model: https://git-scm.com/docs/user-manual
- SQLite transactional guarantees: https://www.sqlite.org/transactional.html
- SQLite backup API: https://www.sqlite.org/backup.html
