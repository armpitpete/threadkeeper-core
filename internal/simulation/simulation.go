package simulation

import "sort"

type Snapshot struct {
	ProjectionDigests map[string]string `json:"projection_digests,omitempty"`
	AuthorityClasses  map[string]string `json:"authority_classes,omitempty"`
	Conflicts         map[string]bool   `json:"conflicts,omitempty"`
	Accessible        map[string]bool   `json:"accessible,omitempty"`
}

type Report struct {
	ProjectionChanged []string `json:"projection_changed,omitempty"`
	AuthorityChanged  []string `json:"authority_changed,omitempty"`
	NewConflicts      []string `json:"new_conflicts,omitempty"`
	AccessChanged     []string `json:"access_changed,omitempty"`
}

func Compare(before, after Snapshot) Report {
	return Report{
		ProjectionChanged: changed(before.ProjectionDigests, after.ProjectionDigests),
		AuthorityChanged: changed(before.AuthorityClasses, after.AuthorityClasses),
		NewConflicts: newTrue(before.Conflicts, after.Conflicts),
		AccessChanged: changedBool(before.Accessible, after.Accessible),
	}
}

func changed(a, b map[string]string) []string { keys := unionString(a,b); out:=[]string{}; for _,k:=range keys { if a[k]!=b[k] { out=append(out,k) } }; return out }
func changedBool(a,b map[string]bool) []string { keys:=unionBool(a,b); out:=[]string{}; for _,k:=range keys { av,aok:=a[k]; bv,bok:=b[k]; if aok!=bok || av!=bv { out=append(out,k) } }; return out }
func newTrue(a,b map[string]bool) []string { keys:=unionBool(a,b); out:=[]string{}; for _,k:=range keys { if !a[k] && b[k] { out=append(out,k) } }; return out }
func unionString(a,b map[string]string) []string { m:=map[string]struct{}{}; for k:=range a {m[k]=struct{}{}}; for k:=range b {m[k]=struct{}{}}; out:=make([]string,0,len(m)); for k:=range m {out=append(out,k)}; sort.Strings(out); return out }
func unionBool(a,b map[string]bool) []string { m:=map[string]struct{}{}; for k:=range a {m[k]=struct{}{}}; for k:=range b {m[k]=struct{}{}}; out:=make([]string,0,len(m)); for k:=range m {out=append(out,k)}; sort.Strings(out); return out }
