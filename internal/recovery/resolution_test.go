package recovery

import "testing"

func divergentFork() ForkResult {
	return ForkResult{
		Kind:           Divergent,
		CommonAncestor: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		HeadA:          "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		HeadB:          "cccccccccccccccccccccccccccccccccccccccc",
	}
}

func validResolution(c ForkCase) ResolutionCandidate {
	return ResolutionCandidate{
		CaseID:            c.CaseID,
		SelectedHead:      c.HeadA,
		RejectedHead:      c.HeadB,
		AuthorisedBy:      "actor:recovery-operator",
		DecisionRef:       "decision:recovery-001",
		Reason:            "Verified restored history A against independent recovery evidence.",
		ResolvedAt:        "2026-08-13T21:40:00Z",
		RejectedPreserved: true,
	}
}

func TestOpenForkCaseRequiresGenuineChoice(t *testing.T) {
	c, err := OpenForkCase("recovery:1", divergentFork())
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != CaseOpen || c.HeadA == c.HeadB {
		t.Fatalf("unexpected case %#v", c)
	}

	for _, kind := range []ForkKind{Identical, ADescendsB, BDescendsA} {
		fork := divergentFork()
		fork.Kind = kind
		if _, err := OpenForkCase("recovery:bad", fork); err == nil {
			t.Fatalf("expected %q to avoid manual fork selection", kind)
		}
	}
}

func TestUnrelatedHistoriesRequireOperatorSelection(t *testing.T) {
	fork := divergentFork()
	fork.Kind = Unrelated
	fork.CommonAncestor = ""
	if _, err := OpenForkCase("recovery:unrelated", fork); err != nil {
		t.Fatal(err)
	}
}

func TestResolutionSelectsExactlyOneAndPreservesOther(t *testing.T) {
	c, err := OpenForkCase("recovery:1", divergentFork())
	if err != nil {
		t.Fatal(err)
	}
	candidate := validResolution(c)
	got, err := ValidateResolution(c, candidate)
	if err != nil {
		t.Fatal(err)
	}
	if got.Case.Status != CaseResolved {
		t.Fatalf("case not resolved: %#v", got.Case)
	}
	if got.Resolution.RejectedHead != c.HeadB || !got.Resolution.RejectedPreserved {
		t.Fatalf("rejected history was not preserved: %#v", got.Resolution)
	}
}

func TestResolutionFailsClosedOnAmbiguousOrUnattributedChoice(t *testing.T) {
	c, err := OpenForkCase("recovery:1", divergentFork())
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name string
		mutate func(*ResolutionCandidate)
	}{
		{"unknown selected head", func(r *ResolutionCandidate) { r.SelectedHead = "dddddddddddddddddddddddddddddddddddddddd" }},
		{"wrong rejected head", func(r *ResolutionCandidate) { r.RejectedHead = r.SelectedHead }},
		{"missing authority", func(r *ResolutionCandidate) { r.AuthorisedBy = "" }},
		{"missing decision reference", func(r *ResolutionCandidate) { r.DecisionRef = "" }},
		{"missing reason", func(r *ResolutionCandidate) { r.Reason = "" }},
		{"rejected history discarded", func(r *ResolutionCandidate) { r.RejectedPreserved = false }},
		{"invalid timestamp", func(r *ResolutionCandidate) { r.ResolvedAt = "latest" }},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			candidate := validResolution(c)
			tc.mutate(&candidate)
			if _, err := ValidateResolution(c, candidate); err == nil {
				t.Fatal("expected fail-closed resolution rejection")
			}
		})
	}
}
