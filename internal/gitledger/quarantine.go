package gitledger

import (
	"errors"
	"io/fs"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/quarantine"
)

const CandidateQuarantineRetention = 24 * time.Hour

// CandidateQuarantineDir is derived from the canonical pinned ledger root.
// Callers do not choose a different quarantine path for prepare versus accept.
func (r *Reader) CandidateQuarantineDir() string {
	return r.gitDir + ".candidate-quarantine"
}

func (r *Reader) OpenCandidateQuarantine() (*quarantine.Store, error) {
	q, err := quarantine.Open(r.CandidateQuarantineDir())
	if err != nil {
		return nil, err
	}
	if _, err := q.PruneBefore(time.Now().UTC().Add(-CandidateQuarantineRetention)); err != nil {
		q.Close()
		return nil, err
	}
	return q, nil
}

func (r *Reader) OpenExistingCandidateQuarantine() (*quarantine.Store, error) {
	q, err := quarantine.OpenExisting(r.CandidateQuarantineDir())
	if err != nil {
		return nil, err
	}
	if _, err := q.PruneBefore(time.Now().UTC().Add(-CandidateQuarantineRetention)); err != nil {
		q.Close()
		return nil, err
	}
	return q, nil
}

// PruneCandidateQuarantine is the maintenance hook for the eventual service
// loop. It is safe to call when no quarantine directory has been created yet.
func (r *Reader) PruneCandidateQuarantine(now time.Time) (int, error) {
	q, err := quarantine.OpenExisting(r.CandidateQuarantineDir())
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	defer q.Close()
	return q.PruneBefore(now.UTC().Add(-CandidateQuarantineRetention))
}
