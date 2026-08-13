# Core Read Model v1

Threadkeeper's read side combines exact source/version identity, provenance, typed relationships, preserved conflicts, temporal/coverage state and confidentiality into an evidence envelope. It does not depend on Recall and it does not change authority.

## Source registry

A source has a stable logical identity, a declared authority class, confidentiality classification and preservation policy. Immutable source versions are registered under that source. Registering the same version identity with different metadata is an integrity conflict rather than an update.

## Provenance

A derived record names exact source versions and/or earlier records plus producer/transformation identity. Provenance must remain acyclic. Lineage can be explained without a model.

## Relationships

Relationships are typed edges. A relationship does not merge the identities of its endpoints. Relationship types may declare symmetry, but Core stores one canonical edge identity and can expose both directions where appropriate.

## Conflicts

Material conflict sets are durable explanatory records. Resolution changes their state but does not erase the records that conflicted.

## Catalog

The Core catalog is a deterministic read projection. Given a record identity and caller clearance it emits the evidence envelope defined by the interface contract. Retrieval score, if supplied by Recall or another client, remains separate from authority class.
