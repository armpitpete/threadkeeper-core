package decision

import "fmt"

type Alternative struct {
	ID             string `json:"id"`
	Summary        string `json:"summary"`
	ReasonRejected string `json:"reason_rejected"`
}

type Dissent struct {
	Actor    string `json:"actor"`
	Position string `json:"position"`
}

type Context struct {
	Alternatives        []Alternative `json:"alternatives_considered,omitempty"`
	Dissent             []Dissent     `json:"dissent,omitempty"`
	Uncertainties       []string      `json:"decision_uncertainties,omitempty"`
	ReopeningConditions []string      `json:"reopening_conditions,omitempty"`
}

func (c Context) Validate() error {
	seen := map[string]struct{}{}
	for _, a := range c.Alternatives {
		if a.ID == "" || a.Summary == "" || a.ReasonRejected == "" { return fmt.Errorf("DECISION_CONTEXT_INVALID: alternatives require id, summary and rejection reason") }
		if _, ok := seen[a.ID]; ok { return fmt.Errorf("DECISION_CONTEXT_INVALID: duplicate alternative %q", a.ID) }
		seen[a.ID] = struct{}{}
	}
	for _, d := range c.Dissent { if d.Actor == "" || d.Position == "" { return fmt.Errorf("DECISION_CONTEXT_INVALID: dissent requires actor and position") } }
	for _, r := range c.ReopeningConditions { if r == "" { return fmt.Errorf("DECISION_CONTEXT_INVALID: reopening conditions must not be empty") } }
	return nil
}
