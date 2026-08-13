# Exact Source Adapter v1

Source adapters obtain bytes; they do not decide authority.

An exact-version adapter request contains a logical source ID plus an immutable expected content identity. The adapter must fail if the bytes currently available do not match that identity. Mutable location names are locators, not version identities.

The initial filesystem adapter is intentionally narrow: it reads only regular files beneath one configured root, rejects traversal and symlink indirection, and verifies the exact SHA-256 supplied by the caller. Content may then be copied into the escrow store according to source preservation policy.
