# External Witness Contract v1

An optional witness makes historical rewriting more detectable without creating a second authority system.

A witness signs a canonical statement containing ledger identity, exact head, projection digest and timestamp. Verification proves that the witness attested to those bytes at that time; it does not prove the project content was true and does not grant project authority.

Witness keys have their own lifecycle and compromise handling. Missing witness availability must not prevent read-only recovery from the authoritative ledger.
