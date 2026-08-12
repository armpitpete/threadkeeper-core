package relationship

import (
	"fmt"
	"sort"
)

type Definition struct { Type string `json:"type"`; Symmetric bool `json:"symmetric"` }
type Edge struct { From string `json:"from"`; Type string `json:"type"`; To string `json:"to"`; Evidence []string `json:"evidence,omitempty"` }

type Graph struct { defs map[string]Definition; edges map[string]Edge }
func NewGraph()*Graph{return &Graph{defs:map[string]Definition{},edges:map[string]Edge{}}}
func (g *Graph) Define(d Definition)error{if d.Type==""{return fmt.Errorf("RELATIONSHIP_INVALID: type required")};if _,ok:=g.defs[d.Type];ok{return fmt.Errorf("RELATIONSHIP_TYPE_EXISTS: %s",d.Type)};g.defs[d.Type]=d;return nil}
func (g *Graph) Add(e Edge)error{d,ok:=g.defs[e.Type];if !ok{return fmt.Errorf("RELATIONSHIP_TYPE_UNKNOWN: %s",e.Type)};if e.From==""||e.To==""||e.From==e.To{return fmt.Errorf("RELATIONSHIP_INVALID: distinct endpoints required")};if d.Symmetric&&e.To<e.From{e.From,e.To=e.To,e.From};key:=e.From+"\x00"+e.Type+"\x00"+e.To;if _,exists:=g.edges[key];exists{return fmt.Errorf("RELATIONSHIP_EXISTS")};e.Evidence=append([]string(nil),e.Evidence...);sort.Strings(e.Evidence);g.edges[key]=e;return nil}
func (g *Graph) For(recordID string)[]Edge{out:=[]Edge{};for _,e:=range g.edges{if e.From==recordID||e.To==recordID{c:=e;c.Evidence=append([]string(nil),e.Evidence...);out=append(out,c)}};sort.Slice(out,func(i,j int)bool{if out[i].Type!=out[j].Type{return out[i].Type<out[j].Type};if out[i].From!=out[j].From{return out[i].From<out[j].From};return out[i].To<out[j].To});return out}
