package recovery

import (
	"fmt"
	"time"
)

type CaseStatus string

const (
	CaseOpen     CaseStatus = "recovery_fork"
	CaseResolved CaseStatus = "resolved"
)

type ForkCase struct {
	CaseID         string     `json:"case_id"`
	Status         CaseStatus `json:"status"`
	Kind           ForkKind   `json:"kind"`
	CommonAncestor string     `json:"common_ancestor,omitempty"`
	HeadA          string     `json:"head_a"`
	HeadB          string     `json:"head_b"`
}

type ResolutionCandidate struct {
	CaseID            string `json:"case_id"`
	SelectedHead      string `json:"selected_head"`
	RejectedHead      string `json:"rejected_head"`
	ActorID           string `json:"actor_id"`
	DecisionRef       string `json:"decision_ref"`
	Reason            string `json:"reason"`
	ResolvedAt        string `json:"resolved_at"`
	RejectedPreserved bool   `json:"rejected_preserved"`
}

type ResolutionResult struct {
	Case       ForkCase            `json:"case"`
	Resolution ResolutionCandidate `json:"resolution"`
}

// OpenForkCase classifies the supplied recovered histories itself rather than
// trusting caller-supplied fork metadata. Only histories for which neither side
// is already a strict continuation of the other enter operator selection.
func OpenForkCase(caseID string, historyA, historyB []string) (ForkCase, error) {
	if caseID == "" {
		return ForkCase{}, fmt.Errorf("RECOVERY_FORK_INVALID: case_id is required")
	}
	fork, err := Classify(historyA, historyB)
	if err != nil {
		return ForkCase{}, fmt.Errorf("RECOVERY_FORK_INVALID: classify histories: %w", err)
	}
	if fork.Kind != Divergent && fork.Kind != Unrelated {
		return ForkCase{}, fmt.Errorf("RECOVERY_FORK_NOT_REQUIRED: kind %q does not require operator selection", fork.Kind)
	}
	return ForkCase{
		CaseID:         caseID,
		Status:         CaseOpen,
		Kind:           fork.Kind,
		CommonAncestor: fork.CommonAncestor,
		HeadA:          fork.HeadA,
		HeadB:          fork.HeadB,
	}, nil
}

// ValidateResolution checks a proposed operator resolution. ActorID is a
// claimed identity only: this function does not authenticate it or move
// authority. The result must still pass the ordinary governed decision path.
func ValidateResolution(c ForkCase, candidate ResolutionCandidate) (ResolutionResult, error) {
	if c.Status != CaseOpen {
		return ResolutionResult{}, fmt.Errorf("RECOVERY_FORK_INVALID: case is not open")
	}
	if candidate.CaseID != c.CaseID {
		return ResolutionResult{}, fmt.Errorf("RECOVERY_FORK_INVALID: case identity mismatch")
	}
	if candidate.SelectedHead != c.HeadA && candidate.SelectedHead != c.HeadB {
		return ResolutionResult{}, fmt.Errorf("RECOVERY_FORK_INVALID: selected head is not a preserved fork head")
	}
	expectedRejected := c.HeadA
	if candidate.SelectedHead == c.HeadA {
		expectedRejected = c.HeadB
	}
	if candidate.RejectedHead != expectedRejected {
		return ResolutionResult{}, fmt.Errorf("RECOVERY_FORK_INVALID: rejected head must be the unselected preserved head")
	}
	if candidate.ActorID == "" || candidate.DecisionRef == "" || candidate.Reason == "" {
		return ResolutionResult{}, fmt.Errorf("RECOVERY_FORK_INVALID: actor_id, decision reference and reason are required")
	}
	if !candidate.RejectedPreserved {
		return ResolutionResult{}, fmt.Errorf("RECOVERY_FORK_INVALID: rejected history must remain preserved evidence")
	}
	if _, err := time.Parse(time.RFC3339, candidate.ResolvedAt); err != nil {
		return ResolutionResult{}, fmt.Errorf("RECOVERY_FORK_INVALID: resolved_at: %w", err)
	}

	resolved := c
	resolved.Status = CaseResolved
	return ResolutionResult{Case: resolved, Resolution: candidate}, nil
}
