package portable

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/conflict"
	"github.com/armpitpete/threadkeeper-core/internal/evidence"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/provenance"
	"github.com/armpitpete/threadkeeper-core/internal/relationship"
	"github.com/armpitpete/threadkeeper-core/internal/source"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

const FormatV1="urn:threadkeeper:portable-core:v1"

type Bundle struct {
	Format                  string                    `json:"format"`
	LedgerCommit            string                    `json:"ledger_commit"`
	RecoveryProof           ledger.RecoveryProof      `json:"recovery_proof"`
	Sources                 []source.Source           `json:"sources,omitempty"`
	Provenance              []provenance.Record       `json:"provenance,omitempty"`
	RelationshipDefinitions []relationship.Definition `json:"relationship_definitions,omitempty"`
	Relationships           []relationship.Edge       `json:"relationships,omitempty"`
	Conflicts               []conflict.Set            `json:"conflicts,omitempty"`
	Evidence                []evidence.Envelope       `json:"evidence,omitempty"`
}

func Encode(bundle Bundle)([]byte,error){normalized:=normalize(bundle);if err:=Validate(normalized);err!=nil{return nil,err};raw,err:=json.Marshal(normalized);if err!=nil{return nil,err};return canonicaljson.Canonicalize(raw)}

func Decode(raw []byte)(Bundle,error){
	if err:=strictjson.Validate(raw);err!=nil{return Bundle{},err}
	canonical,err:=canonicaljson.Canonicalize(raw);if err!=nil{return Bundle{},err}
	if !bytes.Equal(raw,canonical){return Bundle{},fmt.Errorf("PORTABLE_INVALID: export must be RFC 8785 canonical JSON")}
	var bundle Bundle
	dec:=json.NewDecoder(bytes.NewReader(raw));dec.DisallowUnknownFields()
	if err:=dec.Decode(&bundle);err!=nil{return Bundle{},fmt.Errorf("PORTABLE_INVALID: decode: %w",err)}
	if err:=Validate(bundle);err!=nil{return Bundle{},err}
	normalized,err:=Encode(bundle);if err!=nil{return Bundle{},err}
	if !bytes.Equal(raw,normalized){return Bundle{},fmt.Errorf("PORTABLE_INVALID: export is not in deterministic normalized form")}
	return bundle,nil
}

func Validate(bundle Bundle)error{
	if bundle.Format!=FormatV1{return fmt.Errorf("PORTABLE_INVALID: format %q",bundle.Format)};if bundle.LedgerCommit==""||bundle.RecoveryProof.LedgerCommit!=bundle.LedgerCommit{return fmt.Errorf("PORTABLE_INVALID: ledger/recovery proof identity mismatch")}
	sr:=source.NewRegistry();for _,s:=range bundle.Sources{if err:=sr.Register(s);err!=nil{return err}}
	if err:=validateProvenance(bundle.Provenance);err!=nil{return err}
	rg:=relationship.NewGraph();for _,d:=range bundle.RelationshipDefinitions{if err:=rg.Define(d);err!=nil{return err}};for _,e:=range bundle.Relationships{if err:=rg.Add(e);err!=nil{return err}}
	cr:=conflict.NewRegistry();for _,c:=range bundle.Conflicts{if err:=cr.Add(c);err!=nil{return err}}
	seenEvidence:=map[string]struct{}{};for _,env:=range bundle.Evidence{if err:=env.Validate();err!=nil{return err};if _,ok:=seenEvidence[env.RecordID];ok{return fmt.Errorf("PORTABLE_INVALID: duplicate evidence record %q",env.RecordID)};seenEvidence[env.RecordID]=struct{}{}}
	return nil
}

func Equivalent(a,b Bundle)(bool,error){ea,err:=Encode(a);if err!=nil{return false,err};eb,err:=Encode(b);if err!=nil{return false,err};return bytes.Equal(ea,eb),nil}

func validateProvenance(records []provenance.Record)error{remaining:=append([]provenance.Record(nil),records...);g:=provenance.NewGraph();for len(remaining)>0{progress:=false;next:=[]provenance.Record{};for _,r:=range remaining{if err:=g.Add(r);err==nil{progress=true}else{next=append(next,r)}};if !progress{return fmt.Errorf("PORTABLE_INVALID: provenance dependencies are missing or cyclic")};remaining=next};return nil}

func normalize(in Bundle)Bundle{raw,_:=json.Marshal(in);var out Bundle;_ = json.Unmarshal(raw,&out);sort.Slice(out.Sources,func(i,j int)bool{return out.Sources[i].ID<out.Sources[j].ID});for i:=range out.Provenance{sort.Slice(out.Provenance[i].SourceVersions,func(a,b int)bool{x,y:=out.Provenance[i].SourceVersions[a],out.Provenance[i].SourceVersions[b];if x.SourceID!=y.SourceID{return x.SourceID<y.SourceID};return x.VersionID<y.VersionID});sort.Strings(out.Provenance[i].DerivedFrom)};sort.Slice(out.Provenance,func(i,j int)bool{return out.Provenance[i].ID<out.Provenance[j].ID});sort.Slice(out.RelationshipDefinitions,func(i,j int)bool{return out.RelationshipDefinitions[i].Type<out.RelationshipDefinitions[j].Type});for i:=range out.Relationships{sort.Strings(out.Relationships[i].Evidence)};sort.Slice(out.Relationships,func(i,j int)bool{a,b:=out.Relationships[i],out.Relationships[j];if a.Type!=b.Type{return a.Type<b.Type};if a.From!=b.From{return a.From<b.From};return a.To<b.To});for i:=range out.Conflicts{sort.Strings(out.Conflicts[i].Records)};sort.Slice(out.Conflicts,func(i,j int)bool{return out.Conflicts[i].ID<out.Conflicts[j].ID});for i:=range out.Evidence{sort.Strings(out.Evidence[i].SourceVersions);sort.Strings(out.Evidence[i].Provenance);sort.Strings(out.Evidence[i].Conflicts);if out.Evidence[i].Coverage!=nil{sort.Strings(out.Evidence[i].Coverage.KnownGaps)}};sort.Slice(out.Evidence,func(i,j int)bool{return out.Evidence[i].RecordID<out.Evidence[j].RecordID});return out}
