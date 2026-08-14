package ledger

import (
	"context"
	"encoding/json"
	"path/filepath"
	"sync"
	"testing"
)

func TestIssue25IdenticalConcurrentPreparesUsePrivateStages(t *testing.T) {
	r, head := candidateTestReader(t)
	defer r.Close()

	event := makeCreateCandidateEvent(t, head, "issue25-private-stages", "idem-issue25-private-stages", json.RawMessage(`{"enabled":true}`))
	req := CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/issue25-private-stages.json",
		Event:        event,
	}

	arrived := make(chan struct{}, 2)
	release := make(chan struct{})
	var releaseOnce sync.Once
	defer releaseOnce.Do(func() { close(release) })

	type result struct {
		candidate *WriteCandidate
		response  *WriteResponse
		err       error
	}
	done := make(chan result, 2)
	for i := 0; i < 2; i++ {
		go func() {
			candidate, response, err := prepareWriteCandidate(context.Background(), r, req, &prepareWriteHooks{
				beforeEventCommit: func() {
					arrived <- struct{}{}
					<-release
				},
			})
			done <- result{candidate: candidate, response: response, err: err}
		}()
	}

	<-arrived
	<-arrived
	stages, err := filepath.Glob(filepath.Join(r.CandidateQuarantineDir(), quarantineStagePrefix+"*.candidate"))
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) != 2 {
		t.Fatalf("identical concurrent prepares share staging material: got %v, want two private stage files", stages)
	}

	releaseOnce.Do(func() { close(release) })
	first := <-done
	second := <-done
	for i, got := range []result{first, second} {
		if got.err != nil {
			t.Fatalf("prepare %d failed: %v", i+1, got.err)
		}
		if got.candidate == nil || got.response != nil {
			t.Fatalf("prepare %d candidate=%#v response=%#v", i+1, got.candidate, got.response)
		}
	}
	if first.candidate.CandidateCommit != second.candidate.CandidateCommit {
		t.Fatalf("identical prepares produced different deterministic candidates: %s vs %s", first.candidate.CandidateCommit, second.candidate.CandidateCommit)
	}
	stages, err = filepath.Glob(filepath.Join(r.CandidateQuarantineDir(), quarantineStagePrefix+"*.candidate"))
	if err != nil {
		t.Fatal(err)
	}
	if len(stages) != 0 {
		t.Fatalf("private staging entries not cleaned after prepare: %v", stages)
	}
}
