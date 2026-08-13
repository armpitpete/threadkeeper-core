# Source Escrow Contract v1

An immutable source identifier is not sufficient evidence preservation if the referenced source can disappear.

Every source class must declare one preservation mode: `reference_only`, `metadata_snapshot`, `content_snapshot`, `externally_durable`, or `preservation_prohibited`.

`content_snapshot` stores exact bytes under an independently verified SHA-256 when policy, licensing and confidentiality permit it. `externally_durable` records the independently managed preservation system and immutable version. `preservation_prohibited` explicitly records that Core must not retain content.

Escrow preserves evidence; it never promotes authority. Access and retention policy continue to apply to escrowed content.
