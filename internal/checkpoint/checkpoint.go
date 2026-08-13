package checkpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
)

type Checkpoint struct {
	LedgerCommit       string `json:"ledger_commit"`
	EventCount         uint64 `json:"event_count"`
	ProjectionSHA256   string `json:"projection_sha256"`
	SchemaSetSHA256    string `json:"schema_set_sha256"`
	BindingSetSHA256   string `json:"binding_set_sha256"`
}

func Build(ledgerCommit string, eventCount uint64, projection, schemas, bindings []byte) (Checkpoint, error) {
	if ledgerCommit == "" { return Checkpoint{}, fmt.Errorf("CHECKPOINT_INVALID: ledger commit required") }
	p, err := canonicalHash(projection); if err != nil { return Checkpoint{}, fmt.Errorf("projection: %w", err) }
	s, err := canonicalHash(schemas); if err != nil { return Checkpoint{}, fmt.Errorf("schemas: %w", err) }
	b, err := canonicalHash(bindings); if err != nil { return Checkpoint{}, fmt.Errorf("bindings: %w", err) }
	return Checkpoint{LedgerCommit: ledgerCommit, EventCount: eventCount, ProjectionSHA256:p, SchemaSetSHA256:s, BindingSetSHA256:b}, nil
}

func Verify(c Checkpoint, projection, schemas, bindings []byte) error {
	actual, err := Build(c.LedgerCommit, c.EventCount, projection, schemas, bindings); if err != nil { return err }
	if actual.ProjectionSHA256!=c.ProjectionSHA256 || actual.SchemaSetSHA256!=c.SchemaSetSHA256 || actual.BindingSetSHA256!=c.BindingSetSHA256 { return fmt.Errorf("CHECKPOINT_MISMATCH: derived state differs from checkpoint") }
	return nil
}

func canonicalHash(raw []byte) (string,error) { c,err:=canonicaljson.Canonicalize(raw); if err!=nil {return "",err}; sum:=sha256.Sum256(c); return hex.EncodeToString(sum[:]),nil }
