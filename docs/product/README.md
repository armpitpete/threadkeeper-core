# Threadkeeper Product Planning

The user-facing Threadkeeper application is intentionally separated from Threadkeeper Core.

The normative planning boundary for current work is [`../../PRODUCT_BOUNDARY_V1.md`](../../PRODUCT_BOUNDARY_V1.md).

Core remains responsible for authority, integrity, provenance, deterministic replay, governed writes and recovery. Product-facing project semantics, actionable work state, integrations, AI continuation, portfolio and UI belong in the separate `threadkeeper` repository.

Until the application dogfood gate is proven, Core should receive only release-operational work or product-evidenced primitive changes.
