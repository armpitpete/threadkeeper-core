# ADR-006: MCP 2026-07-28 Is an Optional Interoperability Boundary

- **Status:** Accepted when present on the repository default branch; candidate otherwise
- **Date:** 2026-08-16
- **Decision scope:** Client/execution interoperability
- **Issue:** #68

## Context

Threadkeeper Core is deliberately protocol-neutral. It owns durable authority, provenance, historical records, deterministic projections and protected-write semantics; AI systems and other software are replaceable clients.

The Model Context Protocol (MCP) 2026-07-28 specification materially changes the shape of MCP integrations:

- the core transport is stateless and no longer depends on protocol-level sessions;
- application state is expected to use explicit handles where needed;
- long-running work moves into the `io.modelcontextprotocol/tasks` extension with `tasks/get`, `tasks/update` and cancellation;
- multi-round-trip requests can return `input_required` and resume with explicit input responses;
- authorization is hardened around OAuth/OIDC issuer and client metadata rules;
- list/resource responses can carry cache guidance;
- `server/discover`, header-based method/name routing and opt-in notification subscriptions support ordinary infrastructure;
- roots, sampling, logging and legacy HTTP+SSE are deprecated for new implementations.

These features make MCP a better fit for bounded agent execution, but they do not solve Threadkeeper's authority, provenance, acceptance, freshness or historical-correction problem.

## Decision

Threadkeeper Core accepts MCP 2026-07-28 as an **optional external interoperability protocol**, not as part of Core's authority model.

The preferred architecture is:

```text
MCP client / agent
      │
      ▼
Threadkeeper product/client adapter
      │
      ▼
Threadkeeper Core interface boundary
      │
      ├── evidence-rich reads
      ├── proposals / bounded action requests
      └── existing governed write path when separately enabled and authorised
```

Core does not become an MCP server framework and does not redesign its ledger, reducer, provenance or CAS writer around MCP.

No new Core runtime primitive is justified by MCP 2026-07-28 itself.

## Required semantic separation

### MCP task handle != Threadkeeper authority

An MCP task handle identifies execution lifecycle state. It may correlate work across calls, but it is not:

- an authority receipt;
- a project-state identity;
- an acceptance record;
- evidence that a protected transition may occur.

A task may be `completed` while the resulting project candidate remains unaccepted or while Threadkeeper revalidation fails.

### MCP authentication != project authority

OAuth/OIDC identity and authorization establish who/what may reach an MCP surface. Threadkeeper project authority remains separately determined by current project policy, exact state and attributable authority records.

A successfully authenticated client receives no implicit permission to merge, publish, accept, delete, spend, disclose or otherwise cross a protected boundary.

### MCP input_required != automatic approval

A multi-round-trip request may ask a user for missing data or confirmation. The returned input is application input. It counts as Threadkeeper authority only when the existing authority model recognises that exact attributable input as an authority event for the exact action/state in question and persists it through the governed path.

### MCP cache guidance != freshness proof

`ttlMs`/`cacheScope` may improve transport efficiency. They do not supersede Threadkeeper source-version identities, freshness requirements, dependency checks or fail-closed behaviour.

### MCP resources != authoritative records by default

Retrieved MCP resources are classified like any other source input. Retrieval relevance, transport origin or repeated exposure cannot promote them to accepted truth.

## Adopted interoperability points

Threadkeeper product/client adapters should prefer the following MCP 2026-07-28 features when an MCP surface is implemented:

1. stateless requests and explicit application handles;
2. the Tasks extension for genuinely long-running bounded execution;
3. `input_required` for non-authoritative missing input and for authority requests only when bridged into the existing exact authority-record path;
4. issuer-bound OAuth/OIDC client identity as transport authentication;
5. method/name routing headers for admission, rate-limit and audit infrastructure;
6. cache guidance for derived/read-only material where Threadkeeper freshness rules permit caching;
7. opt-in subscriptions for operational progress notifications;
8. `server/discover` when capability discovery is useful.

New Threadkeeper MCP work must not depend on deprecated roots, sampling, logging or legacy HTTP+SSE.

## Consequences

### Positive

- Agents can use a standard interoperability layer without becoming owners of project truth.
- Long-running work can expose explicit, inspectable execution handles instead of hidden transport sessions.
- Mid-operation input can be represented without weakening protected-action semantics.
- Transport authentication and infrastructure routing can improve independently of Core authority.
- Model/client providers remain replaceable.

### Costs

- Adapters must maintain two distinct concepts where MCP terminology appears similar: execution lifecycle versus governed project state/authority.
- Completion and authorization must be revalidated at the Threadkeeper boundary rather than inferred from MCP status.
- Live MCP integration requires version-specific conformance tests because the 2026-07-28 Tasks extension is not wire-compatible with the earlier experimental Tasks API.

## Rejected alternatives

### Alternative A — make MCP Tasks the Threadkeeper project-state machine

Rejected. MCP Tasks model execution lifecycle, not accepted project history, source provenance, corrections, authority or deterministic project-state projection.

### Alternative B — treat OAuth scopes as Threadkeeper authority

Rejected. Transport authorization cannot replace exact-state project authority and human/protected acceptance records.

### Alternative C — move Core storage/state into MCP session semantics

Rejected. MCP 2026-07-28 deliberately removes protocol session state, and Core's durable state already has a stronger explicit ownership model.

### Alternative D — implement MCP directly inside Core before product evidence

Rejected. Core is protocol-neutral and no missing primitive is demonstrated. A bounded adapter belongs in the product/client layer first.

## Reopening conditions

Reconsider only if real product integration proves that:

- a required interoperability property cannot be expressed through the current Core client boundary;
- the missing property is genuinely about authority/integrity/provenance rather than adapter convenience; and
- adding a Core primitive preserves AI independence, protocol neutrality, exact-state authority and fail-closed writes.

## Validation

A future MCP adapter must demonstrate all of the following:

1. removing MCP does not change authoritative state;
2. an MCP task can complete without manufacturing acceptance;
3. stale exact-state authority is rejected even when an MCP task or authenticated client says work may continue;
4. `input_required` cannot create protected authority outside the existing authority-record path;
5. cached MCP reads cannot suppress a required Threadkeeper freshness failure;
6. task cancellation/failure leaves project authority unchanged unless separately evidenced by an owning source;
7. deprecated MCP features are not required for the adapter.

## References

- MCP 2026-07-28 specification release: https://blog.modelcontextprotocol.io/posts/2026-07-28/
- MCP 2026-07-28 release-candidate Tasks summary / SEP-2663 context: https://blog.modelcontextprotocol.io/posts/2026-07-28-release-candidate/
