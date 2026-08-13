package simulation

import "testing"

func TestComparePolicySnapshots(t *testing.T) {
	r := Compare(Snapshot{AuthorityClasses: map[string]string{"x":"derived"}}, Snapshot{AuthorityClasses: map[string]string{"x":"authoritative"}, Conflicts: map[string]bool{"x":true}})
	if len(r.AuthorityChanged)!=1 || r.AuthorityChanged[0]!="x" || len(r.NewConflicts)!=1 { t.Fatalf("unexpected report %#v", r) }
}
