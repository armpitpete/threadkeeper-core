# Portable Core Export v1

A portable export is a deterministic logical representation of Core evidence state, independent of Git object layout, database IDs, Recall indexes or a particular transport.

The v1 bundle may contain the exact ledger/recovery proof, source registry entries, provenance records, relationship definitions/edges, conflict sets and evidence envelopes. Arrays are normalised into deterministic order and the whole bundle is RFC 8785 canonical JSON.

Import performs strict JSON validation plus logical validation of source identities, provenance dependencies, relationship definitions, conflict records, evidence envelopes and recovery-proof/head agreement. Re-encoding an imported bundle must reproduce identical canonical bytes.

Portable export is a copy of existing identities and authority metadata; exporting or importing does not create new authority.
