package provenance

import (
	"fmt"
	"sort"
)

type SourceVersion struct { SourceID string `json:"source_id"`; VersionID string `json:"version_id"` }

type Record struct {
	ID               string          `json:"id"`
	SourceVersions   []SourceVersion `json:"source_versions,omitempty"`
	DerivedFrom      []string        `json:"derived_from,omitempty"`
	ProducerID       string          `json:"producer_id,omitempty"`
	TransformationID string          `json:"transformation_id,omitempty"`
}

type Graph struct { records map[string]Record }
func NewGraph()*Graph{return &Graph{records:map[string]Record{}}}

func (g *Graph) Add(record Record) error {
	if record.ID==""{return fmt.Errorf("PROVENANCE_INVALID: record id required")}
	if _,exists:=g.records[record.ID];exists{return fmt.Errorf("PROVENANCE_EXISTS: %s",record.ID)}
	if len(record.DerivedFrom)>0 && (record.ProducerID=="" || record.TransformationID==""){return fmt.Errorf("PROVENANCE_INVALID: derived records require producer and transformation identity")}
	for _,ref:=range record.SourceVersions{if ref.SourceID==""||ref.VersionID==""{return fmt.Errorf("PROVENANCE_INVALID: exact source/version required")}}
	for _,parent:=range record.DerivedFrom { if parent==record.ID{return fmt.Errorf("PROVENANCE_CYCLE: self reference")}; if _,ok:=g.records[parent];!ok{return fmt.Errorf("PROVENANCE_PARENT_NOT_FOUND: %s",parent)} }
	g.records[record.ID]=clone(record)
	if g.hasCycle(record.ID,map[string]bool{},map[string]bool{}) { delete(g.records,record.ID); return fmt.Errorf("PROVENANCE_CYCLE: %s",record.ID) }
	return nil
}

func (g *Graph) Get(id string)(Record,bool){r,ok:=g.records[id];return clone(r),ok}

func (g *Graph) Lineage(id string)([]Record,error){ if _,ok:=g.records[id];!ok{return nil,fmt.Errorf("PROVENANCE_NOT_FOUND: %s",id)}; seen:=map[string]bool{}; out:=[]Record{}; var walk func(string); walk=func(cur string){if seen[cur]{return}; seen[cur]=true; r:=g.records[cur]; parents:=append([]string(nil),r.DerivedFrom...); sort.Strings(parents); for _,p:=range parents{walk(p)}; out=append(out,clone(r))}; walk(id); return out,nil }

func (g *Graph) hasCycle(id string,visiting,done map[string]bool)bool{if visiting[id]{return true};if done[id]{return false};visiting[id]=true;for _,p:=range g.records[id].DerivedFrom{if g.hasCycle(p,visiting,done){return true}};delete(visiting,id);done[id]=true;return false}
func clone(r Record)Record{r.SourceVersions=append([]SourceVersion(nil),r.SourceVersions...);r.DerivedFrom=append([]string(nil),r.DerivedFrom...);return r}
