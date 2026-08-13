package genesis

import (
	"encoding/json"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/digest"
)

func TestValidateGenesis(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"project_id": "project:test",
		"ledger_id": "ledger:test",
		"created_at": "2026-08-12T16:00:00Z",
		"initial_authority_policy": "policy:v1",
		"initial_schemas": []string{"schema:a", "schema:b"},
		"initial_authorities": []string{"owner:a", "owner:b"},
	})
	if err != nil { t.Fatal(err) }
	completed, _, err := digest.Complete(raw)
	if err != nil { t.Fatal(err) }
	root, err := Validate(completed)
	if err != nil { t.Fatal(err) }
	if root.LedgerID != "ledger:test" { t.Fatalf("unexpected root: %#v", root) }
}

func TestGenesisRejectsOrderingAmbiguity(t *testing.T) {
	raw, _ := json.Marshal(map[string]any{
		"project_id": "project:test", "ledger_id": "ledger:test", "created_at": "2026-08-12T16:00:00Z",
		"initial_authority_policy": "policy:v1", "initial_schemas": []string{"schema:b", "schema:a"}, "initial_authorities": []string{"owner:a"},
	})
	completed, _, _ := digest.Complete(raw)
	if _, err := Validate(completed); err == nil { t.Fatal("expected ordering rejection") }
}

func TestGenesisRejectsUnknownFields(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"project_id": "project:test",
		"ledger_id": "ledger:test",
		"created_at": "2026-08-12T16:00:00Z",
		"initial_authority_policy": "policy:v1",
		"initial_schemas": []string{"schema:a"},
		"initial_authorities": []string{"owner:a"},
		"unexpected": true,
	})
	if err != nil { t.Fatal(err) }
	completed, _, err := digest.Complete(raw)
	if err != nil { t.Fatal(err) }
	if _, err := Validate(completed); err == nil { t.Fatal("expected unknown-field rejection") }
}
