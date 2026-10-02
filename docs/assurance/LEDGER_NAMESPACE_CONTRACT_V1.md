# Core v1 Authoritative Ledger Namespace Contract

## Status

This contract defines the committed-tree namespace accepted by Threadkeeper Core v1 authoritative replay.

Core v1 is **closed-world**: every committed tree path encountered anywhere in authoritative history MUST belong to a recognised v1 ledger path class. Unknown committed content is an integrity failure, not an inert extension point.

## Recognised committed path classes

Core v1 recognises only:

1. exact Genesis root:
   - `config/genesis/root.json`;
2. exact initial actor-policy root:
   - `config/authority/actor-policy/root.json`;
3. versioned JSON schemas beneath:
   - `config/schemas/`;
4. versioned JSON reducer bindings beneath:
   - `config/authority/reducer-bindings/`;
5. durable JSON events beneath:
   - `events/`, using the same event-path grammar as candidate construction.

The two singleton roots above do not permit sibling files merely because those files share the same parent directory.

Schema and reducer-binding resources must remain JSON paths and retain their existing append-only/immutable semantics.

Durable events retain the existing immutable event-path, regular-blob, canonical JSON, digest, schema and replay rules.

## Whole-history rule

Namespace validation is historical, not only a check of the current tree.

Core MUST inspect every commit reachable from the configured authoritative history before semantic replay. An unknown path fails closed even when:

- no currently recognised semantic object changed in that commit;
- the unknown path is later deleted;
- the commit also contains a valid event or configuration addition;
- the unknown path appeared in the Genesis/root commit.

This prevents arbitrary repository content from bypassing replay through a "no semantic changes" fast path.

## Failure

An unrecognised committed path fails with `INTEGRITY_FAILURE` and identifies the offending path and commit.

Core MUST NOT silently ignore the path or continue with a guessed projection.

## Git repository metadata is a separate boundary

This contract governs **committed tree content**.

Git repository internals such as objects, refs, repository config, hooks, alternates, shallow/graft metadata, ref backends and filesystem indirection remain governed by the existing hardened Git repository/environment isolation rules. They are not ledger tree namespaces.

## Future namespace extensions

A path family does not become valid merely because a future binary has code that could interpret it.

Adding a new authoritative namespace requires a separately reviewed format/migration contract that defines:

- the new namespace identity;
- activation/migration semantics;
- how historical state becomes valid, if at all;
- replay and recovery effects;
- compatibility with older binaries.

Core v1 binaries continue to fail closed on namespaces they do not recognise.

Conceptual paths appearing in older architecture documents are not accepted v1 paths unless included explicitly in this contract.

## Authority boundary

Namespace validation is read-only integrity enforcement. It does not:

- mutate accepted history;
- advance the authoritative ref;
- create a service or listener;
- enable a transport;
- enable authority writes.

`AUTHORITY_WRITES_DISABLED` remains mandatory.
