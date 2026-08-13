package provenance

import "testing"

func TestLineagePreservesExactInputs(t *testing.T){g:=NewGraph();if err:=g.Add(Record{ID:"obs",SourceVersions:[]SourceVersion{{SourceID:"s",VersionID:"v1"}}});err!=nil{t.Fatal(err)};if err:=g.Add(Record{ID:"summary",DerivedFrom:[]string{"obs"},ProducerID:"tool:a",TransformationID:"summarise-v1"});err!=nil{t.Fatal(err)};lineage,err:=g.Lineage("summary");if err!=nil{t.Fatal(err)};if len(lineage)!=2||lineage[0].ID!="obs"{t.Fatalf("unexpected lineage %#v",lineage)}}
