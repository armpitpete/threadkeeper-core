package federation

import "fmt"

type Reference struct {
	ProjectID            string `json:"project_id"`
	LedgerID             string `json:"ledger_id"`
	RecordID             string `json:"record_id"`
	VersionID            string `json:"version_id"`
	SourceAuthorityClass string `json:"source_authority_class,omitempty"`
	LocalAuthorityClass  string `json:"local_authority_class"`
}

func (r Reference) Validate() error {
	if r.ProjectID=="" || r.LedgerID=="" || r.RecordID=="" || r.VersionID=="" { return fmt.Errorf("FEDERATION_INVALID: exact source identity is required") }
	if r.LocalAuthorityClass=="" { return fmt.Errorf("FEDERATION_INVALID: local_authority_class is required; source authority is never imported implicitly") }
	return nil
}
