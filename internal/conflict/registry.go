package conflict

import (
	"fmt"
	"sort"
)

type State string
const(Open State="open";Resolved State="resolved")
type Set struct{ID string `json:"id"`;Records []string `json:"records"`;Material bool `json:"material"`;State State `json:"state"`;ResolutionEventID string `json:"resolution_event_id,omitempty"`}
type Registry struct{sets map[string]Set}
func NewRegistry()*Registry{return &Registry{sets:map[string]Set{}}}
func(r *Registry)Add(s Set)error{if s.ID==""||len(s.Records)<2{return fmt.Errorf("CONFLICT_INVALID: id and at least two records required")};switch s.State{case Open:if s.ResolutionEventID!=""{return fmt.Errorf("CONFLICT_INVALID: open conflict cannot have resolution event")};case Resolved:if s.ResolutionEventID==""{return fmt.Errorf("CONFLICT_INVALID: resolved conflict requires resolution event")};default:return fmt.Errorf("CONFLICT_INVALID: unknown state")};s.Records=append([]string(nil),s.Records...);sort.Strings(s.Records);for i,v:=range s.Records{if v==""||(i>0&&s.Records[i-1]==v){return fmt.Errorf("CONFLICT_INVALID: distinct record identities required")}};if _,ok:=r.sets[s.ID];ok{return fmt.Errorf("CONFLICT_EXISTS: %s",s.ID)};r.sets[s.ID]=s;return nil}
func(r *Registry)For(recordID string)[]Set{out:=[]Set{};for _,s:=range r.sets{for _,id:=range s.Records{if id==recordID{c:=s;c.Records=append([]string(nil),s.Records...);out=append(out,c);break}}};sort.Slice(out,func(i,j int)bool{return out[i].ID<out[j].ID});return out}
