# Threadkeeper Core Product Boundary

## Purpose

Threadkeeper Core is the protected authority, integrity, provenance, deterministic-state and recovery kernel beneath the Threadkeeper product.

The user-facing Threadkeeper product is a separate application layer. Core MUST NOT expand merely to absorb application concerns that can be implemented above the kernel.

## Core owns

- authoritative ledger and immutable history;
- strict validation and canonical identity;
- deterministic replay and current-state reduction primitives;
- provenance, evidence relationships and conflict representation;
- Genesis and authority policy;
- actor/key/grant authentication and authorization;
- candidate preparation, quarantine, exact-head CAS and idempotency;
- integrity, recovery, restore-verification and conformance machinery;
- protocol-neutral interfaces needed by clients.

## Product layer owns

The separate `armpitpete/threadkeeper` product repository should own:

- project model and project lifecycle;
- goals and definitions of done;
- work states such as Now, Next, Blocked, Later and Done;
- project-level questions and blockers;
- project decisions and user-facing evidence references;
- project gates and understandable authority presentation;
- next-action selection;
- CLI and application API;
- GitHub and later external integrations;
- AI continuation/execution orchestration;
- portfolio view;
- web interface;
- onboarding and other product UX.

## Change rule

A new application requirement MUST first be attempted in the product layer.

Core changes are justified only when product evidence demonstrates that a missing primitive belongs below the authority boundary or cannot be implemented safely and deterministically above it.

## Immediate product programme

The first product milestone is intentionally narrow:

> A user can return to a long-running project after an interruption and immediately determine what is true, what is blocked, what can happen now, what comes next, and where the system must stop for human authority.

The product should prove this against three materially different real projects before broad UI, SaaS, team, billing, marketplace or mobile work begins.

## Core release boundary

Before long-running production authority writes are considered, the remaining protected operational gates stay separate from product development:

1. reconcile Core release/status documentation after the merged Core v1 E2E harness;
2. complete the accepted production read-only load proof;
3. establish and evidence a genuinely independent secondary custody boundary;
4. separately authorize and verify any destructive production restore/replacement;
5. separately review production service activation;
6. keep `AUTHORITY_WRITES_DISABLED` closed until an explicit later write-enable decision.

Product development MUST NOT treat a future write-enable decision as implied authority.
