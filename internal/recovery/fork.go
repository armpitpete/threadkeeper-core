package recovery

import "fmt"

type ForkKind string

const (
	Identical   ForkKind = "identical"
	ADescendsB  ForkKind = "a_descends_b"
	BDescendsA  ForkKind = "b_descends_a"
	Divergent   ForkKind = "divergent"
	Unrelated   ForkKind = "unrelated"
)

type ForkResult struct {
	Kind           ForkKind `json:"kind"`
	CommonAncestor string   `json:"common_ancestor,omitempty"`
	HeadA          string   `json:"head_a"`
	HeadB          string   `json:"head_b"`
}

func Classify(historyA, historyB []string) (ForkResult, error) {
	if err := validateHistory(historyA); err != nil { return ForkResult{}, fmt.Errorf("history A: %w", err) }
	if err := validateHistory(historyB); err != nil { return ForkResult{}, fmt.Errorf("history B: %w", err) }
	result := ForkResult{HeadA: historyA[len(historyA)-1], HeadB: historyB[len(historyB)-1]}
	common := 0
	for common < len(historyA) && common < len(historyB) && historyA[common] == historyB[common] { common++ }
	if common == 0 { result.Kind = Unrelated; return result, nil }
	result.CommonAncestor = historyA[common-1]
	switch {
	case common == len(historyA) && common == len(historyB): result.Kind = Identical
	case common == len(historyB): result.Kind = ADescendsB
	case common == len(historyA): result.Kind = BDescendsA
	default: result.Kind = Divergent
	}
	return result, nil
}

func validateHistory(history []string) error {
	if len(history) == 0 { return fmt.Errorf("RECOVERY_HISTORY_INVALID: empty history") }
	seen := map[string]struct{}{}
	for _, id := range history {
		if id == "" { return fmt.Errorf("RECOVERY_HISTORY_INVALID: empty commit identity") }
		if _, ok := seen[id]; ok { return fmt.Errorf("RECOVERY_HISTORY_INVALID: duplicate commit %q", id) }
		seen[id] = struct{}{}
	}
	return nil
}
