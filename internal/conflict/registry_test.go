package conflict

import "testing"

func TestResolvedConflictRemainsInspectable(t *testing.T){r:=NewRegistry();if err:=r.Add(Set{ID:"c1",Records:[]string{"r2","r1"},Material:true,State:Resolved,ResolutionEventID:"decision:1"});err!=nil{t.Fatal(err)};sets:=r.For("r1");if len(sets)!=1||sets[0].State!=Resolved{t.Fatalf("unexpected sets %#v",sets)}}
