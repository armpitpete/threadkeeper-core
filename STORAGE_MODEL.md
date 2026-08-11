# Storage Model

## 1. Principle

Threadkeeper storage is divided by **meaning and recoverability**, not by whichever database or disk is convenient.

The architecture requires a physical and logical boundary between AI/model storage and Threadkeeper Core storage.

## 2. Storage classes

### 2.1 Authoritative sources and sinks

These preserve project truth that cannot be reconstructed from Threadkeeper-derived data alone.

They may live outside Threadkeeper Core. Threadkeeper references them by immutable identity and policy.

Examples may include source repositories, accepted documents, explicit decision ledgers or other configured systems of record. No particular technology is required.

A Threadkeeper-native governance record may be treated as authoritative only after it is persisted to a configured authoritative sink.

### 2.2 Core durable configuration

This is the minimum versioned information required to reconstruct how Threadkeeper interprets sources. It includes concepts such as:

- source registry and source identity rules;
- authority policy;
- relationship/type definitions;
- schema and contract versions;
- rebuild configuration;
- migration metadata needed to interpret historical records.

This configuration must not exist only inside disposable Recall storage.

### 2.3 Recall storage

Recall is derived storage optimised for retrieval and navigation.

It may contain:

- text/search indexes;
- semantic representations;
- relationship projections;
- derived metadata;
- source digests;
- incremental-index checkpoints;
- cached projections;
- retrieval statistics.

Recall is **disposable**. A complete loss of Recall must not erase accepted truth, decisions or required provenance.

### 2.4 Transient working storage

Temporary files, queues, scratch state and session data are transient. They must not be the sole location of any record needed for recovery or governance.

### 2.5 Secrets

Credentials, private keys and tokens are operational secrets, not project knowledge. They must be stored through an appropriate secret mechanism and must not be embedded in Recall, logs, generated prompts or versioned repository content.

## 3. Physical separation requirement

Threadkeeper Core storage must be independently manageable from AI/model storage.

An implementation may place Threadkeeper on a dedicated volume, service or machine, but the contractual requirement is stronger than hardware layout:

- AI model installation, replacement or deletion must not mutate Threadkeeper authority;
- AI cache/model cleanup must not delete Threadkeeper data;
- Threadkeeper backup and restore must not require restoration of an AI model environment;
- Threadkeeper can operate in a no-model mode for ingestion, provenance, inspection and rebuild.

## 4. Recall rebuild contract

A compliant Recall store must be rebuildable from:

1. declared authoritative sources/sinks;
2. exact source identities/versions available from those systems;
3. versioned Core durable configuration;
4. the selected implementation's reproducible transformation/indexing logic.

AI output must not be required to recover accepted authority. If an optional model-derived index cannot be recreated because a model is unavailable, Threadkeeper must still recover authoritative records, provenance and non-model-dependent state.

## 5. Data identity

Storage must preserve stable logical identities independent of physical location.

At minimum, source-backed records must be identifiable by:

- logical source identity;
- immutable version identity and/or content digest;
- record identity within that version where applicable.

Moving a cache, changing a database engine or rebuilding an index must not create a new authoritative fact merely because storage identifiers changed.

## 6. Mutation rules

Derived storage may be replaced or rebuilt.

Authoritative records and authority-changing decisions must use append, supersede, revoke or equivalent history-preserving semantics. Destructive replacement must not erase the evidence that a prior authoritative state existed.

## 7. Crash and partial-write behaviour

The eventual implementation must provide a way to detect incomplete writes and must not expose partially committed authority transitions as valid state.

A write that cannot establish its complete provenance or durable authoritative destination must fail closed.

## 8. Portability

Threadkeeper must provide a deterministic export path for its durable configuration, authority metadata, provenance and relationship definitions using an implementation-independent logical model.

Changing storage technology must not require changing the meaning of authority classes, provenance, decisions or relationships.

## 9. Backup principle

Backups are copies, not automatically new authorities.

The backup policy must separately cover:

- authoritative sources/sinks according to their owning system;
- Core durable configuration;
- Recall only where rebuilding cost justifies backup;
- secrets through their own secure mechanism.

## 10. Destructive recovery test

A conforming implementation must periodically prove that the Recall store can be destroyed and rebuilt without loss of:

- accepted project state;
- authority classification;
- decision history;
- source provenance;
- conflict/supersession history required to explain current state.
