# Actor Authentication and Authorization v1

Threadkeeper does not grant trust from client type, model identity, UI presence, or a caller-supplied actor name.

## Authentication

An authenticated request is bound to a configured actor key. v1 uses Ed25519 proof verification. The signed proof covers:

- actor identity;
- key identity;
- ledger identity;
- requested action;
- exact target;
- expected prior state;
- idempotency identity;
- issue and expiry times.

Changing any bound request field invalidates the proof. Proof lifetime is bounded by policy. Revoked or unknown keys fail closed.

## Authorization

Authentication does not imply permission. After proof verification, Core requires an exact policy grant matching:

- authenticated actor;
- ledger identity;
- action;
- governed target.

The v1 action vocabulary is `propose`, `decide`, and `operate`. A grant on one ledger, action or target does not authorize another.

## Authority boundary

Authentication and authorization are necessary but not sufficient for an authority-changing write. Existing expected-state, reducer, policy, CAS, idempotency, persistence, quarantine, recovery, deployment and release gates remain independent requirements.

The global `AUTHORITY_WRITES_DISABLED` release boundary remains in force. This contract does not enable a public write transport and does not convert an authenticated client into an authority source by itself.

## Transport independence

The proof is protocol-neutral. HTTP, CLI, MCP, RPC or another transport may carry it later, but the transport must not weaken the signed context or substitute its own trust decision for Core policy.
