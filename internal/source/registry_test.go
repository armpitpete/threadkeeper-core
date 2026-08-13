package source

import (
	"testing"
	"github.com/armpitpete/threadkeeper-core/internal/access"
	"github.com/armpitpete/threadkeeper-core/internal/escrow"
)

func TestSourceVersionsAreImmutableByIdentity(t *testing.T){
	r:=NewRegistry(); if err:=r.Register(Source{ID:"repo:a",Kind:"git",AuthorityClass:"authoritative",Classification:access.Internal,Preservation:escrow.Policy{Mode:escrow.ExternallyDurable}});err!=nil{t.Fatal(err)}
	v:=Version{ID:"abc",Locator:"git:abc"}; if err:=r.AddVersion("repo:a",v);err!=nil{t.Fatal(err)}; if err:=r.AddVersion("repo:a",v);err!=nil{t.Fatal("exact retry should be idempotent")}
	if err:=r.AddVersion("repo:a",Version{ID:"abc",Locator:"git:different"});err==nil{t.Fatal("expected immutable version conflict")}
}
