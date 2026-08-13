package recovery

import "testing"

func TestClassifyRecoveryFork(t *testing.T) {
	r, err := Classify([]string{"g", "a1"}, []string{"g", "b1"})
	if err != nil { t.Fatal(err) }
	if r.Kind != Divergent || r.CommonAncestor != "g" { t.Fatalf("unexpected fork: %#v", r) }
	r, err = Classify([]string{"g", "a1", "a2"}, []string{"g", "a1"})
	if err != nil { t.Fatal(err) }
	if r.Kind != ADescendsB { t.Fatalf("expected A descendant, got %#v", r) }
}
