package catalog

import (
	"testing"
	"github.com/armpitpete/threadkeeper-core/internal/access"
	"github.com/armpitpete/threadkeeper-core/internal/conflict"
	"github.com/armpitpete/threadkeeper-core/internal/provenance"
	"github.com/armpitpete/threadkeeper-core/internal/relationship"
)

func TestCatalogEmitsEvidenceWithoutPromotingRetrieval(t *testing.T){p:=provenance.NewGraph();rels:=relationship.NewGraph();conf:=conflict.NewRegistry();if err:=conf.Add(conflict.Set{ID:"conflict:1",Records:[]string{"r1","r2"},Material:true,State:conflict.Open});err!=nil{t.Fatal(err)};c:=New(p,rels,conf);if err:=c.Add(Record{ID:"r1",Type:"claim",AuthorityClass:"derived",Classification:access.Internal});err!=nil{t.Fatal(err)};score:=1.0;env,err:=c.Envelope("r1",access.Restricted,&score);if err!=nil{t.Fatal(err)};if env.AuthorityClass!="derived"||len(env.Conflicts)!=1{t.Fatalf("unexpected envelope %#v",env)};if _,err:=c.Envelope("r1",access.Public,nil);err==nil{t.Fatal("expected access denial")}}
