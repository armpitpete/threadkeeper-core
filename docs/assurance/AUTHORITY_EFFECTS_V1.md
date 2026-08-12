# Single Authority Effect Principle v1

Every Core mechanism declares exactly one authority effect. Composition must not create authority implicitly.

Examples: escrow = `evidence_preservation`; witness = `integrity_attestation`; checkpoint/simulation/Recall = `derived_projection`; access policy = `access_control`; an explicitly authorised CAS event = `authority_transition`; ordinary validation = `none`.

A mechanism that needs more than one authority effect must be split or receive a new reviewed contract. In particular, preservation, retrieval, integrity proof and federation are not authority promotion.
