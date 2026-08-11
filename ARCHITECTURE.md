# Threadkeeper Core Architecture

## 1. Purpose

Threadkeeper Core is a model-independent infrastructure layer for project authority, provenance, relationships, state and retrieval.

It exists so that humans and software agents can ask **what is known, why it is known, which version it refers to, what conflicts with it, and whether it has actually been accepted** without treating an AI conversation or model context as project memory.

## 2. Non-goals

Threadkeeper Core is not:

- an LLM or model host;
- a chatbot memory feature;
- an autonomous authority;
- a substitute for authoritative repositories or records;
- a system that converts inference into truth by repetition;
- a requirement to use any particular database, protocol, search engine or model.

## 3. Conceptual planes

### 3.1 Authority plane

Contains the sources and decisions that are allowed to establish project truth. Authority is declared by policy and explicit acceptance, never inferred from confidence or convenience.

Examples may include version-controlled repositories, accepted documents, signed records, configured external systems and explicit owner decisions. The contract does not require any particular source technology.

### 3.2 Core plane

Maintains the structured knowledge needed to work with authority:

- source identities and versions;
- provenance chains;
- authority classifications;
- relationships between records;
- current-state projections;
- conflict and supersession information;
- proposal and decision metadata;
- rebuild instructions and source digests.

### 3.3 Recall plane

Contains derived retrieval structures used for speed and discovery. It may include lexical indexes, semantic representations, graph projections, caches or other derived structures.

**Everything in Recall must be reproducible from authoritative sources plus versioned Core configuration.**

### 3.4 Client plane

Humans, command-line tools, services and AI systems interact through a defined interface boundary. AI receives no implicit privilege over any other client.

## 4. Required lifecycle

```text
Discover source
  ↓
Identify exact source/version
  ↓
Classify authority
  ↓
Capture provenance
  ↓
Derive/index/relate
  ↓
Retrieve with evidence envelope
  ↓
Client reasons or proposes
  ↓
Explicit authorised decision if required
  ↓
Persist decision in an authoritative sink
  ↓
Re-ingest resulting authoritative state
```

A proposal is not an acceptance. A retrieval result is not a decision. A model output is not authority.

## 5. Core invariants

1. **AI independence** — Core remains usable if every AI component is removed.
2. **Recall disposability** — deleting Recall cannot erase accepted truth or human decisions.
3. **Provenance completeness** — every derived record must identify the source versions from which it was derived.
4. **No silent promotion** — derived, advisory or generated material cannot become authoritative without an authorised transition.
5. **Conflict preservation** — conflicting authoritative evidence must be represented, not silently overwritten.
6. **Version exactness** — project state must be attributable to exact source versions; labels such as `main`, `latest` or `current` are views, not immutable identities.
7. **Fail closed** — when authority, identity or expected state is materially ambiguous, protected writes must stop rather than guess.
8. **Model neutrality** — model provider, embedding system and reasoning engine are replaceable clients or derived-data producers.
9. **Technology neutrality** — the contract is independent of storage engine, transport protocol and deployment topology.
10. **Recoverability** — a documented rebuild process must reconstruct Recall from declared authoritative inputs without AI assistance.

## 6. Boundary between observation and decision

Threadkeeper may observe that a branch was merged, a file changed or a human recorded an acceptance. It may expose that event and derive current-state views from it.

It must not invent the acceptance itself. Where governance requires an explicit human act, the authorised act must be represented as an attributable event and preserved in an authoritative sink.

## 7. Component responsibilities

The eventual implementation may split these responsibilities into any number of processes or services, but the logical responsibilities must remain separable:

- **Source adapters** — obtain source objects and immutable versions.
- **Authority resolver** — applies declared authority policy.
- **Provenance recorder** — records derivation and source lineage.
- **Relationship engine** — records typed links without destroying source identity.
- **State projector** — computes current views from historical events.
- **Recall builder** — creates disposable retrieval structures.
- **Interface boundary** — exposes evidence-rich reads and controlled writes.
- **Decision recorder** — records authorised decisions and ensures durable authoritative persistence.

No component may bypass the authority model simply because it has direct storage access.

## 8. Acceptance test for the architecture

A conforming implementation must pass this thought experiment as an executable recovery test:

1. Stop every AI/model service.
2. Delete the entire Recall store.
3. Preserve only declared authoritative sources, versioned Core configuration and required credentials.
4. Start Threadkeeper Core without an AI dependency.
5. Rebuild Recall.
6. Retrieve previously accepted project state with the same authority classification and provenance.

If accepted truth, decisions or provenance disappear, the implementation does not conform.
