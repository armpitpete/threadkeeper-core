package coverage

import "testing"

func TestAbsenceRequiresCompleteCoverage(t *testing.T) {
	partial := Report{Expected: 10, Available: 9, Checked: 9, KnownGaps: []string{"source:x"}}
	if partial.CanClaimAbsence() { t.Fatal("partial coverage must not support absence claim") }
	complete := Report{Expected: 10, Available: 10, Checked: 10, Complete: true}
	if !complete.CanClaimAbsence() { t.Fatal("complete coverage should support bounded absence claim") }
}
