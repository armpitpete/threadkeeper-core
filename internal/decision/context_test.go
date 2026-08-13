package decision

import "testing"

func TestDecisionContextPreservesAlternativesAndReopening(t *testing.T) {
	c := Context{Alternatives: []Alternative{{ID: "b", Summary: "Option B", ReasonRejected: "higher migration risk"}}, Dissent: []Dissent{{Actor: "reviewer:x", Position: "prefer B"}}, ReopeningConditions: []string{"multi-region requirement appears"}}
	if err := c.Validate(); err != nil { t.Fatal(err) }
}
