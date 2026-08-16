# MCP 2026-07-28 Interoperability Profile v0.1

Status: candidate until merged with ADR-006.

This profile translates MCP 2026-07-28 concepts into Threadkeeper boundaries. It is an adapter contract, not a new Core protocol or authority model.

## Feature mapping

| MCP 2026-07-28 feature | Threadkeeper use | Required boundary | Disposition |
|---|---|---|---|
| Stateless request core | Ordinary scalable transport for clients/adapters | No project truth may live only in transport/session state | ADOPT IN ADAPTER |
| Explicit application handles | Correlate long-running client work and Threadkeeper action-contract identifiers | Handle possession grants no authority | ADOPT IN ADAPTER |
| `io.modelcontextprotocol/tasks` | Represent long-running bounded execution lifecycle | Task state is execution state, not project acceptance/current-state authority | ADOPT WHEN NEEDED |
| `tasks/get` | Poll task progress/result | Result must still be validated/re-projected by Threadkeeper | ADOPT WHEN TASKS USED |
| `tasks/update` | Supply requested task input / resume work | Input becomes authority only through the existing exact authority-record path | ADOPT WITH GUARD |
| task cancellation | Stop execution | Cancellation does not rewrite accepted project state by itself | ADOPT WHEN TASKS USED |
| MRTR `input_required` | Ask for missing values or explicit human input without hidden session state | Confirmation text is not automatically an authority receipt | ADOPT WITH GUARD |
| OAuth/OIDC hardening | Client authentication and transport access control | Authentication is not Threadkeeper project authority | ADOPT AT TRANSPORT |
| `Mcp-Method` / `Mcp-Name` headers | Gateway routing, admission, metering and audit hints | Header route cannot bypass Threadkeeper action admission | ADOPT AT EDGE |
| `ttlMs` / `cacheScope` | Reduce repeated list/resource reads | Never override source identity/freshness/dependency checks | ADOPT FOR DERIVED READS |
| `server/discover` | Optional capability discovery | Discovery result is capability metadata, not project authority | OPTIONAL |
| `subscriptions/listen` | Operational progress/event notifications | Notifications are observations until classified against owning sources | OPTIONAL |
| roots | None for new adapter | Deprecated | DO NOT ADOPT |
| sampling | None for Core; model invocation remains client/product concern | Deprecated and would blur client/Core responsibility | DO NOT ADOPT |
| logging | Use ordinary application observability instead | Deprecated | DO NOT ADOPT |
| legacy HTTP+SSE | None for new implementation | Deprecated | DO NOT ADOPT |

## Recommended identifier separation

An implementation must keep these identifiers distinct even if one object stores links between them:

```text
mcp_task_id
    execution lifecycle only

threadkeeper_action_contract_id
    exact authorised bounded action request

threadkeeper_state_identity
    exact verified project/source state used to derive the action

authority_receipt_id
    separately attributable authority for protected scope when required

execution_receipt_id
    evidence of what the worker/tool actually did
```

No identifier may substitute for another merely because they refer to the same episode of work.

## Recommended task lifecycle bridge

```text
Threadkeeper verifies current state
  ↓
Threadkeeper derives exact bounded action contract
  ↓
Adapter calls MCP tool
  ↓
ordinary result OR MCP task handle
  ↓
working
  ├─→ input_required
  │       ↓
  │   classify requested input
  │       ├─ ordinary parameter → tasks/update
  │       └─ protected authority → existing Threadkeeper authority path first
  │
  ├─→ cancelled / failed
  │       ↓
  │   record execution outcome; project authority unchanged unless owning evidence says otherwise
  │
  └─→ completed
          ↓
      record execution receipt
          ↓
      revalidate affected owning sources
          ↓
      re-project Threadkeeper state
          ↓
      continue only if a new/current action contract authorises it
```

## Required adapter assertions

A conforming Threadkeeper MCP adapter must be able to assert and test:

- `mcp_task_id` is never accepted as `authority_receipt_id`;
- authenticated client identity never implies protected project scope;
- a completed task cannot mark a candidate accepted without owning acceptance evidence;
- stale exact state blocks continuation even when an MCP task reports success;
- `input_required` for a protected action routes through Threadkeeper's existing authority mechanism before resume;
- cached resource/list data is discarded or qualified when Threadkeeper freshness rules require newer proof;
- hostile instruction-like text retrieved through MCP remains data unless a separately recognised current instruction-authority surface applies;
- adapter removal leaves Core authoritative state unchanged.

## Implementation location

First implementation, if/when product evidence warrants it, belongs in `armpitpete/threadkeeper` as a client/adapter surface. `threadkeeper-core` should change only if that implementation demonstrates a genuinely missing integrity/authority/provenance primitive.

## Current conclusion

**No Core runtime implementation is required by MCP 2026-07-28.** The useful adoption is a strict interoperability profile plus a future bounded product-layer adapter.