# Core Assurance Programme v0.1

This programme implements the whole-project review recommendations without pretending that newly added primitives are already production-integrated.

| Capability | Primitive/contract | Integration state |
|---|---|---|
| Genesis trust root | `internal/genesis`, Genesis Contract | implemented; ledger genesis migration still required |
| Threat model | Threat Model v1 | adopted as review boundary |
| Source escrow | `internal/escrow` | digest/policy primitive implemented; source adapters/backends pending |
| Bitemporal time | `internal/temporal` | model implemented; broad record schemas pending migration |
| Coverage/completeness | `internal/coverage` | model implemented; source adapters must populate it |
| Confidentiality/retention | `internal/access` | classification/tombstone primitives implemented; authentication/storage enforcement pending |
| Decision dissent/reopening | `internal/decision` | context model implemented; decision schema integration pending |
| Fork recovery | `internal/recovery` | deterministic classification implemented; operator resolution workflow pending |
| Candidate quarantine | `internal/quarantine` | private storage primitive implemented; Git writer integration deliberately waits for CAS review |
| Policy simulation | `internal/simulation` | deterministic impact comparison implemented; policy loader integration pending |
| Replay checkpoints | `internal/checkpoint` | digest verification implemented; replay acceleration integration pending |
| External witness | `internal/witness` | Ed25519 statement signing/verification implemented; deployment/key service optional |
| Federation | `internal/federation` | exact reference/local-authority rule implemented; transport pending |
| Supply-chain integrity | `internal/buildprovenance` + CI evidence | implemented at conformance build level |
| Single authority effect | `internal/authorityeffect` | executable declaration vocabulary implemented |
| Operability | `internal/health` | domain model implemented; service endpoints/dashboard pending |
| Incident response | `internal/incident` + runbook | lifecycle primitive implemented |
| Key lifecycle | `internal/keylifecycle` | transition model implemented; secret backend pending |
| Reference client | Reference Client Contract | contract defined; executable client follows stable read API |
| Load safety | load acceptance matrix | required before public write enablement |

Public authority writes remain disabled throughout this programme.
