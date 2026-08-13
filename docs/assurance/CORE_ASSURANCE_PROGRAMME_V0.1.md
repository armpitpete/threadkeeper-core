# Core Assurance Programme v0.1

This programme implements the whole-project review recommendations while distinguishing installed primitives from protected production integration.

| Capability | Implementation | State |
|---|---|---|
| Genesis trust root | `internal/genesis` + adoption contract | validator installed; legacy adoption decision pending |
| Threat model | Threat Model v1 | installed review boundary |
| Source escrow | `internal/escrow` content-addressed store | installed |
| Exact source ingestion | `internal/sourceadapter`, `internal/source` | installed filesystem adapter + immutable registry |
| Provenance | `internal/provenance` | installed acyclic lineage graph |
| Relationships/conflicts | `internal/relationship`, `internal/conflict` | installed |
| Evidence reads | `internal/catalog`, `internal/evidence` | installed read projection/envelope |
| Bitemporal time | `internal/temporal` | installed model; universal schema migration pending |
| Coverage/completeness | `internal/coverage` | installed model; adapters/records populate as applicable |
| Confidentiality/retention | `internal/access` | installed model/catalog enforcement; actor auth/storage enforcement pending |
| Decision dissent/reopening | `internal/decision`, `internal/proposal`, `internal/reviewbundle` | installed non-authoritative review path |
| Fork recovery | `internal/recovery` | classifier installed; operator resolution pending |
| Destructive restore proof | `internal/ledger/recovery_proof.go` + tests/CLI | installed; independent remote-backup drill pending |
| Candidate quarantine | `internal/quarantine` | private store installed; writer integration remains a separate CAS-changing review gate |
| Policy simulation | `internal/simulation` | deterministic impact comparison installed |
| Replay checkpoints | `internal/checkpoint` | digest verification installed; replay acceleration optional/pending |
| External witness | `internal/witness` | signing/verification installed; deployment/key service optional |
| Federation | `internal/federation` | exact reference/local-authority rule installed; transport optional |
| Portable export | `internal/portable` | deterministic canonical export/import installed |
| Supply-chain integrity | `internal/buildprovenance` + CI artifacts | installed at conformance build level |
| Single authority effect | `internal/authorityeffect` | installed vocabulary/contract |
| Operability | `internal/health` + CLI | installed model/reference output; dashboard optional |
| Incident response | `internal/incident` + runbook | installed lifecycle |
| Key lifecycle | `internal/keylifecycle` | installed lifecycle; secret backend pending |
| Reference client | CLI commands + contract | executable read/review client installed |
| Load safety | concurrent governance tests + `internal/service.Limiter` | core semantic tests installed; final resource/performance envelope pending |

Public authority writes remain disabled throughout this programme. Optional Recall remains a separate later layer.
