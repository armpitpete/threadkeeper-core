package relationship

import "testing"

func TestSymmetricRelationshipCanonicalisesWithoutMergingIdentity(t *testing.T){g:=NewGraph();if err:=g.Define(Definition{Type:"conflicts_with",Symmetric:true});err!=nil{t.Fatal(err)};if err:=g.Add(Edge{From:"b",Type:"conflicts_with",To:"a"});err!=nil{t.Fatal(err)};edges:=g.For("a");if len(edges)!=1||edges[0].From!="a"||edges[0].To!="b"{t.Fatalf("unexpected edges %#v",edges)}}
