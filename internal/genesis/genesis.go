package genesis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

type Root struct {
	ProjectID              string   `json:"project_id"`
	LedgerID               string   `json:"ledger_id"`
	CreatedAt              string   `json:"created_at"`
	InitialAuthorityPolicy string   `json:"initial_authority_policy"`
	InitialSchemas         []string `json:"initial_schemas"`
	InitialAuthorities     []string `json:"initial_authorities"`
	ContentSHA256          string   `json:"content_sha256"`
}

func Validate(raw []byte) (Root, error) {
	if err := strictjson.Validate(raw); err != nil {
		return Root{}, err
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return Root{}, err
	}
	if !bytes.Equal(raw, canonical) {
		return Root{}, fmt.Errorf("GENESIS_INVALID: record must be RFC 8785 canonical JSON")
	}
	if err := digest.Verify(raw); err != nil {
		return Root{}, fmt.Errorf("GENESIS_INVALID: %w", err)
	}
	var root Root
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&root); err != nil {
		return Root{}, fmt.Errorf("GENESIS_INVALID: decode: %w", err)
	}
	if root.ProjectID == "" || root.LedgerID == "" || root.InitialAuthorityPolicy == "" || len(root.InitialAuthorities) == 0 {
		return Root{}, fmt.Errorf("GENESIS_INVALID: project_id, ledger_id, initial_authority_policy and initial_authorities are required")
	}
	if _, err := time.Parse(time.RFC3339, root.CreatedAt); err != nil {
		return Root{}, fmt.Errorf("GENESIS_INVALID: created_at: %w", err)
	}
	if err := sortedUnique(root.InitialSchemas); err != nil {
		return Root{}, fmt.Errorf("GENESIS_INVALID: initial_schemas: %w", err)
	}
	if err := sortedUnique(root.InitialAuthorities); err != nil {
		return Root{}, fmt.Errorf("GENESIS_INVALID: initial_authorities: %w", err)
	}
	return root, nil
}

func sortedUnique(values []string) error {
	if !sort.StringsAreSorted(values) {
		return fmt.Errorf("values must be lexicographically sorted")
	}
	for i, value := range values {
		if value == "" {
			return fmt.Errorf("values must not be empty")
		}
		if i > 0 && values[i-1] == value {
			return fmt.Errorf("duplicate value %q", value)
		}
	}
	return nil
}
