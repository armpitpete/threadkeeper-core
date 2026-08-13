# Threat Model v1

## Protected assets

Threadkeeper protects authoritative history, exact record identity, provenance, policy bindings, current projections, recovery evidence, confidentiality labels and evidence needed to explain accepted state.

## In-scope adversaries

Core must fail closed against hostile or buggy clients, AI clients, malformed records, forged candidate handles, stale/concurrent writers, hostile repository metadata/configuration, compromised Recall data, malicious or buggy source adapters, unavailable remote sources, crashes and partial writes.

## Deployment assumptions

Before authority writes can be enabled, the executable and durable ledger directory must be service-owned and non-writable by untrusted local processes. Secrets and signing keys are outside the ledger and have their own lifecycle controls.

## Explicit non-claims

Threadkeeper v1 does not claim to survive an attacker that simultaneously controls the host kernel/root, the running Core binary, every durable backup and every independent witness; a cryptographic break of the selected primitives; or deliberate authorised acceptance of false information.

These are reopening conditions for the model, not excuses to weaken fail-closed checks inside the declared boundary.

## Review rule

Security findings are gate-blocking when they violate an in-scope property. New mechanisms outside this threat model require an explicit threat-model revision instead of silently expanding the security boundary forever.
