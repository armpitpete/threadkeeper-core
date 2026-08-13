package evidence

import (
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/access"
	"github.com/armpitpete/threadkeeper-core/internal/coverage"
	"github.com/armpitpete/threadkeeper-core/internal/temporal"
)

type Envelope struct {
	RecordID          string                 `json:"record_id"`
	RecordType        string                 `json:"record_type"`
	AuthorityClass    string                 `json:"authority_class"`
	SourceID          string                 `json:"source_id,omitempty"`
	SourceVersions    []string               `json:"source_versions,omitempty"`
	Provenance        []string               `json:"provenance,omitempty"`
	Conflicts         []string               `json:"conflicts,omitempty"`
	Superseded        bool                   `json:"superseded"`
	ProjectionVersion string                 `json:"projection_version,omitempty"`
	RetrievalScore    *float64               `json:"retrieval_score,omitempty"`
	Coverage          *coverage.Report       `json:"coverage,omitempty"`
	Temporal          *temporal.Window       `json:"temporal,omitempty"`
	Classification    access.Classification  `json:"classification"`
}

func (e Envelope) Validate() error {
	if e.RecordID=="" || e.RecordType=="" || e.AuthorityClass=="" { return fmt.Errorf("EVIDENCE_ENVELOPE_INVALID: record identity/type/authority required") }
	if err:=e.Classification.Validate(); err!=nil{return err}
	if e.Coverage!=nil { if err:=e.Coverage.Validate(); err!=nil{return err} }
	if e.Temporal!=nil { if err:=e.Temporal.Validate(); err!=nil{return err} }
	return nil
}
