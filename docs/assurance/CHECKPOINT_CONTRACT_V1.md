# Verified Replay Checkpoint v1

Checkpoints accelerate replay; they are never authority.

A checkpoint binds an exact ledger commit and event count to SHA-256 digests of the canonical projection, accepted schema set and reducer-binding set. Core may start incremental replay from a verified checkpoint, but periodic full replay from genesis must reproduce the same checkpoint digest.

A missing or invalid checkpoint falls back to full replay. It must never change accepted state.
