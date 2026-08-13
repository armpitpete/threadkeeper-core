package coverage

import "fmt"

type Report struct {
	Expected              int      `json:"sources_expected"`
	Available             int      `json:"sources_available"`
	Checked               int      `json:"sources_checked"`
	Complete              bool     `json:"complete"`
	LastCompleteIngestion string   `json:"last_complete_ingestion,omitempty"`
	KnownGaps             []string `json:"known_gaps,omitempty"`
}

func (r Report) Validate() error {
	if r.Expected < 0 || r.Available < 0 || r.Checked < 0 { return fmt.Errorf("COVERAGE_INVALID: counts must be non-negative") }
	if r.Checked > r.Available { return fmt.Errorf("COVERAGE_INVALID: checked exceeds available") }
	if r.Complete && (r.Expected == 0 || r.Available != r.Expected || r.Checked != r.Expected || len(r.KnownGaps) != 0) {
		return fmt.Errorf("COVERAGE_INVALID: complete coverage contradicts counts or known gaps")
	}
	return nil
}

func (r Report) CanClaimAbsence() bool {
	return r.Validate() == nil && r.Complete && r.Expected > 0 && r.Available == r.Expected && r.Checked == r.Expected && len(r.KnownGaps) == 0
}
