package genesis

import (
	"encoding/json"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/digest"
)

func adoptionFixture(t *testing.T, head string) []byte {
	t.Helper()
	raw, err := json.Marshal(map[string]any{
		"project_id": "project:test",
		"ledger_id": "ledger:test",
		"pre_adoption_ledger_head": head,
		"authority_policy": "policy:v1",
		"adoption_authority": "owner:test",
		"genesis_contract_version": "genesis-v1",
		"adopted_at": "2026-08-13T21:00:00Z",
	})
	if err != nil { t.Fatal(err) }
	completed, _, err := digest.Complete(raw)
	if err != nil { t.Fatal(err) }
	return completed
}

func TestValidateLegacyGenesisAdoption(t *testing.T) {
	raw := adoptionFixture(t, "0123456789abcdef0123456789abcdef01234567")
	got, err := ValidateAdoption(raw)
	if err != nil { t.Fatal(err) }
	if got.LedgerID != "ledger:test" { t.Fatalf("unexpected adoption: %#v", got) }
}

func TestLegacyGenesisAdoptionRejectsAbbreviatedHead(t *testing.T) {
	raw := adoptionFixture(t, "0123456")
	if _, err := ValidateAdoption(raw); err == nil { t.Fatal("expected abbreviated ledger head rejection") }
}

func TestLegacyGenesisAdoptionRejectsNonCanonicalJSON(t *testing.T) {
	raw := adoptionFixture(t, "0123456789abcdef0123456789abcdef01234567")
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil { t.Fatal(err) }
	pretty, err := json.MarshalIndent(value, "", "  ")
	if err != nil { t.Fatal(err) }
	if _, err := ValidateAdoption(pretty); err == nil { t.Fatal("expected non-canonical adoption rejection") }
}
