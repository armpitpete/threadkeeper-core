package schemas

import _ "embed"

// These are reference artifacts for conformance tests and ledger bootstrapping.
// Runtime replay never falls back to embedded schemas; accepted ledgers must
// carry their own schema snapshots under config/schemas.

//go:embed reducer-binding-v1.schema.json
var ReducerBindingV1 []byte

//go:embed exclusive-governed-record-event-v1.schema.json
var ExclusiveGovernedRecordEventV1 []byte
