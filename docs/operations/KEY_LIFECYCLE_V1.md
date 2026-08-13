# Key Lifecycle v1

Keys used for witness signatures, release provenance or future actor authentication are operational authority dependencies and require explicit lifecycle state.

Supported states are `generated`, `active`, `rotating`, `compromised`, `revoked`, and `retired`. Compromise is never repaired by pretending the old key remained trustworthy; affected signatures are evaluated against the key's validity interval and incident evidence.

Private key material is never stored in the governance ledger. Core may preserve public-key identity, lifecycle events and the evidence needed to explain which key was trusted when.
