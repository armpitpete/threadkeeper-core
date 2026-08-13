package catalog

import (
	"fmt"
	"sort"

	"github.com/armpitpete/threadkeeper-core/internal/access"
	"github.com/armpitpete/threadkeeper-core/internal/conflict"
	"github.com/armpitpete/threadkeeper-core/internal/coverage"
	"github.com/armpitpete/threadkeeper-core/internal/evidence"
	"github.com/armpitpete/threadkeeper-core/internal/provenance"
	"github.com/armpitpete/threadkeeper-core/internal/relationship"
	"github.com/armpitpete/threadkeeper-core/internal/temporal"
)

type Record struct {
	ID                string                `json:"id"`
	Type              string                `json:"type"`
	AuthorityClass    string                `json:"authority_class"`
	SourceID          string                `json:"source_id,omitempty"`
	SourceVersions    []string              `json:"source_versions,omitempty"`
	ProvenanceIDs     []string              `json:"provenance,omitempty"`
	Superseded        bool                  `json:"superseded"`
	ProjectionVersion string                `json:"projection_version,omitempty"`
	Coverage          *coverage.Report      `json:"coverage,omitempty"`
	Temporal          *temporal.Window      `json:"temporal,omitempty"`
	Classification    access.Classification `json:"classification"`
}

type Catalog struct { records map[string]Record; provenance *provenance.Graph; relationships *relationship.Graph; conflicts *conflict.Registry }
func New(p *provenance.Graph,r *relationship.Graph,c *conflict.Registry)*Catalog{if p==nil{p=provenance.NewGraph()};if r==nil{r=relationship.NewGraph()};if c==nil{c=conflict.NewRegistry()};return &Catalog{records:map[string]Record{},provenance:p,relationships:r,conflicts:c}}
func(c *Catalog)Add(record Record)error{if record.ID==""||record.Type==""||record.AuthorityClass==""{return fmt.Errorf("CATALOG_INVALID: id/type/authority required")};if err:=record.Classification.Validate();err!=nil{return err};if record.Coverage!=nil{if err:=record.Coverage.Validate();err!=nil{return err}};if record.Temporal!=nil{if err:=record.Temporal.Validate();err!=nil{return err}};if _,ok:=c.records[record.ID];ok{return fmt.Errorf("CATALOG_EXISTS: %s",record.ID)};record.SourceVersions=append([]string(nil),record.SourceVersions...);record.ProvenanceIDs=append([]string(nil),record.ProvenanceIDs...);sort.Strings(record.SourceVersions);sort.Strings(record.ProvenanceIDs);c.records[record.ID]=record;return nil}
func(c *Catalog)Envelope(id string,clearance access.Classification,retrievalScore *float64)(evidence.Envelope,error){r,ok:=c.records[id];if !ok{return evidence.Envelope{},fmt.Errorf("CATALOG_NOT_FOUND: %s",id)};if !access.CanRead(clearance,r.Classification){return evidence.Envelope{},fmt.Errorf("ACCESS_DENIED: %s",id)};conflictIDs:=[]string{};for _,set:=range c.conflicts.For(id){conflictIDs=append(conflictIDs,set.ID)};sort.Strings(conflictIDs);env:=evidence.Envelope{RecordID:r.ID,RecordType:r.Type,AuthorityClass:r.AuthorityClass,SourceID:r.SourceID,SourceVersions:append([]string(nil),r.SourceVersions...),Provenance:append([]string(nil),r.ProvenanceIDs...),Conflicts:conflictIDs,Superseded:r.Superseded,ProjectionVersion:r.ProjectionVersion,RetrievalScore:retrievalScore,Coverage:r.Coverage,Temporal:r.Temporal,Classification:r.Classification};if err:=env.Validate();err!=nil{return evidence.Envelope{},err};return env,nil}
func(c *Catalog)Relationships(id string)[]relationship.Edge{return c.relationships.For(id)}
