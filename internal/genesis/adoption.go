package genesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

// Adoption records the explicit prospective adoption of Genesis v1 by a
// pre-existing Git governance ledger. It does not rewrite or reinterpret the
// pre-adoption commit history.
type Adoption struct {
	ProjectID             string `json:"project_id"`
	LedgerID              string `json:"ledger_id"`
	PreAdoptionLedgerHead string `json:"pre_adoption_ledger_head"`
	AuthorityPolicy       string `json:"authority_policy"`
	AdoptionAuthority     string `json:"adoption_authority"`
	GenesisContractVersion string `json:"genesis_contract_version"`
	AdoptedAt             string `json:"adopted_at"`
	ContentSHA256         string `json:"content_sha256"`
}

// ValidateAdoption verifies the durable legacy-ledger Genesis adoption record.
// The record is strict, RFC 8785 canonical JSON with Threadkeeper's ordinary
// content digest. The pre-adoption head is an immutable full Git object ID.
func ValidateAdoption(raw []byte) (Adoption, error) {
	if err := strictjson.Validate(raw); err != nil {
		return Adoption{}, err
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return Adoption{}, err
	}
	if !bytes.Equal(raw, canonical) {
		return Adoption{}, fmt.Errorf("GENESIS_ADOPTION_INVALID: record must be RFC 8785 canonical JSON")
	}
	if err := digest.Verify(raw); err != nil {
		return Adoption{}, fmt.Errorf("GENESIS_ADOPTION_INVALID: %w", err)
	}

	var adoption Adoption
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&adoption); err != nil {
		return Adoption{}, fmt.Errorf("GENESIS_ADOPTION_INVALID: decode: %w", err)
	}

	if adoption.ProjectID == "" || adoption.LedgerID == "" || adoption.AuthorityPolicy == "" || adoption.AdoptionAuthority == "" || adoption.GenesisContractVersion == "" {
		return Adoption{}, fmt.Errorf("GENESIS_ADOPTION_INVALID: project_id, ledger_id, authority_policy, adoption_authority and genesis_contract_version are required")
	}
	if !fullGitObjectID(adoption.PreAdoptionLedgerHead) {
		return Adoption{}, fmt.Errorf("GENESIS_ADOPTION_INVALID: pre_adoption_ledger_head must be a full SHA-1 or SHA-256 Git object ID")
	}
	if _, err := time.Parse(time.RFC3339, adoption.AdoptedAt); err != nil {
		return Adoption{}, fmt.Errorf("GENESIS_ADOPTION_INVALID: adopted_at: %w", err)
	}
	return adoption, nil
}

func fullGitObjectID(value string) bool {
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for _, c := range value {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
