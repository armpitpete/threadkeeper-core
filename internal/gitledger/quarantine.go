package gitledger

import "github.com/armpitpete/threadkeeper-core/internal/quarantine"

// CandidateQuarantineDir is derived from the canonical pinned ledger root.
// Callers do not choose a different quarantine path for prepare versus accept.
func (r *Reader) CandidateQuarantineDir() string {
	return r.gitDir + ".candidate-quarantine"
}

func (r *Reader) OpenCandidateQuarantine() (*quarantine.Store, error) {
	return quarantine.Open(r.CandidateQuarantineDir())
}

func (r *Reader) OpenExistingCandidateQuarantine() (*quarantine.Store, error) {
	return quarantine.OpenExisting(r.CandidateQuarantineDir())
}
