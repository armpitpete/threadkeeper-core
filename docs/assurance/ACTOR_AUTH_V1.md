# Actor Authentication and Authorization v1

Threadkeeper does not grant trust from client type, model identity, UI presence, transport configuration or a caller-supplied actor name/policy.

## Authentication

An authenticated request is bound to an authoritative actor key. v1 uses Ed25519 proof verification. The signed proof covers:

- actor identity;
- key identity;
- ledger identity;
- requested action;
- exact target;
- expected prior state;
- idempotency identity;
- issue and expiry times.

Changing any bound request field invalidates the proof. Proof lifetime is bounded by authoritative policy. Revoked or unknown keys fail closed.

## Authorization

Authentication does not imply permission. After proof verification, Core requires an exact authoritative grant matching:

- authenticated actor;
- ledger identity;
- action;
- governed target.

The v1 action vocabulary is `propose`, `decide`, and `operate`. A grant on one ledger, action or target does not authorize another.

## Authoritative policy source

Trusted keys and grants are not caller/runtime authority. The initial policy is stored in the Genesis commit at:

`config/authority/actor-policy/root.json`

The policy document is strict RFC 8785 canonical JSON, digest-bound, sorted and duplicate-free. Public keys are canonical raw-base64 Ed25519 keys. Each grant actor must have at least one active trusted key. The root policy's granted actor set must exactly match Genesis `initial_authorities`.

The root actor-policy path is immutable. Direct later file replacement/removal/rename is rejected during replay.

Current policy is derived from the exact authoritative ledger snapshot:

1. the immutable root policy is the initial policy;
2. a governed `core.record.created` at fixed target `authority:actor-policy` may establish an updated active policy;
3. later `core.record.replaced` events rotate keys/grants;
4. `core.record.revoked` makes actor admission fail closed and does not fall back to Genesis.

Actor-policy governed events use record kind `core.actor-auth-policy-v1`. Their policy value is semantically validated by the reducer before candidate acceptance/CAS, so malformed key/grant material cannot become authoritative merely because an outer event is schema-valid.

The supported Fresh Genesis bootstrap requires the initial reducer binding for this record kind, so production trust policy is rotatable/revocable rather than permanently frozen.

## Service admission boundary

The exported service admission API does not accept an `actorauth.Policy` argument. A future transport supplies the ledger Reader, signed proof and exact request context only.

`AUTHORITY_WRITES_DISABLED` is checked **before** any ledger/policy/authentication work. Only if that release gate is deliberately opened does admission:

1. replay the exact authoritative ledger;
2. derive the current actor policy from that snapshot;
3. require request/proof `ledger_id` to match Genesis ledger identity;
4. require `expected_state` to match that exact authoritative head;
5. authenticate the Ed25519 proof and require an exact grant.

A runtime-supplied substitute policy cannot authorize through the exported boundary.

## Authority boundary

Authentication and authorization are necessary but not sufficient for an authority-changing write. Existing expected-state, reducer, CAS, idempotency, persistence, quarantine, recovery, deployment and release gates remain independent requirements.

The global `AUTHORITY_WRITES_DISABLED` release boundary remains in force. This contract does not enable a public write transport and does not convert an authenticated client into an authority source by itself.

## Transport independence

The proof is protocol-neutral. HTTP, CLI, MCP, RPC or another transport may carry it later, but the transport must not weaken the signed context, provide a substitute policy, or make its own trust decision outside authoritative Core state.
