# Production Load / Resource Proof v1

This runbook uses the same read-only proof machinery exercised in repository conformance. It does not choose a production envelope automatically and it does not enable authority writes.

## Required deployment facts

Do not claim production load acceptance until the actual deployment has identified:

- exact Threadkeeper Core build/source commit;
- production host/environment;
- service identity;
- dedicated authoritative ledger path and ref;
- selected service concurrency/backpressure limit;
- expected production-shaped replay/read workload;
- explicit resource-growth ceilings justified for that host/service envelope.

The envelope must be an explicit reviewed JSON document. Do not substitute `loadproof.ReferenceEnvelope()` for production merely because CI passes.

For production on Linux or Windows, set `require_open_handle_metric: true`. A missing required metric is a proof failure rather than permission to ignore descriptor/handle growth.

## Envelope fields

```json
{
  "name": "production-envelope-id",
  "concurrent_workers": 8,
  "iterations_per_worker": 10,
  "max_peak_heap_growth_bytes": 134217728,
  "max_settled_heap_growth_bytes": 67108864,
  "max_peak_goroutine_growth": 32,
  "max_settled_goroutine_growth": 8,
  "max_peak_open_handle_growth": 128,
  "max_settled_open_handle_growth": 16,
  "require_open_handle_metric": true
}
```

The numbers above illustrate the file shape only. They are **not** recommended production limits.

Envelope decoding is strict: duplicate members, unknown members, invalid JSON, trailing JSON values, zero/negative workload dimensions, negative resource ceilings and settled ceilings above their corresponding peak ceiling are rejected.

## Measurement

First capture build identity and confirm the hard release gate is still closed:

```text
threadkeeper-core version
```

Require `authority_writes_enabled: false`.

Then run:

```text
threadkeeper-core ledger-load-proof <ledger.git> <envelope.json> [authoritative-ref]
```

The command:

1. opens the ledger through the hardened Reader;
2. establishes a baseline full RecoveryProof;
3. runs the exact declared worker/iteration workload;
4. performs complete `ProveRecovery` / authoritative replay on every operation;
5. requires every operation to equal the exact baseline RecoveryProof;
6. samples Go live heap, goroutine count and process descriptor/handle count every 5 ms;
7. forces a settled post-work GC snapshot;
8. checks sampled peak and settled growth against the supplied ceilings;
9. emits machine-readable RecoveryProof + resource evidence.

The command is read-only. A proof failure must not trigger a write or an automatic relaxation of the envelope.

## Interpretation

Require all of the following:

- the RecoveryProof identifies the expected production ledger/Genesis/actor-policy state;
- `completed_operations == concurrent_workers * iterations_per_worker`;
- `passed == true`;
- required descriptor/handle metrics are available before, during and after the run;
- sampled peak growth and settled growth are within every supplied ceiling;
- no replay divergence occurs.

`heap_alloc_bytes` is Go live heap allocation, not process RSS or total system memory. The 5 ms monitor provides a sampled peak rather than a mathematically continuous maximum. Production host telemetry may be retained alongside this proof when RSS/CPU/IO/latency capacity evidence is desired, but it must not replace the Core proof of replay identity and process-resource boundedness.

## Repeat and failure rules

Run the proof from a quiet, production-shaped service environment. Unrelated process activity can conservatively increase process-global descriptor/handle counts and cause failure. If a run fails, preserve its output and determine the cause; do not simply increase limits until green.

If deployment code, Go/runtime version, Git version, storage model, service concurrency, ledger shape or production resource assumptions change materially, rerun the production envelope proof.

## Relationship to other gates

Passing this run closes the actual production load/resource envelope only for the exact measured deployment and envelope. It does not substitute for:

- production Fresh Genesis/filesystem ownership proof;
- independently operated secondary backup/restore proof;
- full production-shaped end-to-end release acceptance;
- final separately reviewed write-enablement decision.

`AUTHORITY_WRITES_DISABLED` remains mandatory throughout this run.
