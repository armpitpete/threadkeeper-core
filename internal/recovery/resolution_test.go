package recovery

import "testing"

func testFork() ForkResult {
	return ForkResult{Kind: Divergent, CommonAncestor: "a", HeadA: "b", HeadB: "c"}
}

func testResolution(c ForkCase) ResolutionCandidate {
	return ResolutionCandidate{
		CaseID: c.CaseID,
		SelectedHead: c.HeadA,
		RejectedHead: c.HeadB,
		ActorID: "actor:operator",
		DecisionRef: "decision:1",
		Reason: "recovery evidence supports head A",
		ResolvedAt: "2026-08-13T21:40:00Z",
		RejectedPreserved: true,
	}
}

func TestRecoveryForkResolution(t *testing.T) {
	c, err := OpenForkCase("case:1", testFork())
	if err != nil { t.Fatal(err) }
	got, err := ValidateResolution(c, testResolution(c))
	if err != nil { t.Fatal(err) }
	if got.Case.Status != CaseResolved || !got.Resolution.RejectedPreserved { t.Fatalf("unexpected result %#v", got) }
}

func TestRecoveryForkResolutionFailsClosed(t *testing.T) {
	c, err := OpenForkCase("case:1", testFork())
	if err != nil { t.Fatal(err) }
	cases := []func(*ResolutionCandidate){
		func(r *ResolutionCandidate){ r.SelectedHead = "x" },
		func(r *ResolutionCandidate){ r.RejectedHead = r.SelectedHead },
		func(r *ResolutionCandidate){ r.ActorID = "" },
		func(r *ResolutionCandidate){ r.DecisionRef = "" },
		func(r *ResolutionCandidate){ r.Reason = "" },
		func(r *ResolutionCandidate){ r.RejectedPreserved = false },
		func(r *ResolutionCandidate){ r.ResolvedAt = "invalid" },
	}
	for i, mutate := range cases {
		candidate := testResolution(c)
		mutate(&candidate)
		if _, err := ValidateResolution(c, candidate); err == nil { t.Fatalf("case %d unexpectedly accepted", i) }
	}
}

func TestRecoveryForkOnlyForGenuineChoice(t *testing.T) {
	for _, kind := range []ForkKind{Identical, ADescendsB, BDescendsA} {
		fork := testFork(); fork.Kind = kind
		if _, err := OpenForkCase("case:bad", fork); err == nil { t.Fatalf("kind %q unexpectedly opened", kind) }
	}
	fork := testFork(); fork.Kind = Unrelated; fork.CommonAncestor = ""
	if _, err := OpenForkCase("case:unrelated", fork); err != nil { t.Fatal(err) }
}
