package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
)

func TestPrepareWriteCandidateDoesNotAdvanceAuthority(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-1", "idem-candidate-1", json.RawMessage(`{"enabled":true}`))

	candidate, accepted, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-1.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if accepted != nil || candidate == nil {
		t.Fatalf("unexpected prepare result candidate=%#v accepted=%#v", candidate, accepted)
	}
	if candidate.ExpectedHead != head || candidate.CandidateCommit == "" || candidate.CandidateCommit == head {
		t.Fatalf("bad candidate identity: %#v", candidate)
	}
	got, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if got != head {
		t.Fatalf("prepare advanced authority: got %s want %s", got, head)
	}
	manifest, err := Replay(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.LedgerCommit != head || manifest.EventCount != 0 {
		t.Fatalf("unreachable candidate affected replay: %#v", manifest)
	}
}

func TestAcceptWriteCandidateCASAndReplay(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-1", "idem-candidate-1", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-1.json", Event: event})
	if err != nil {
		t.Fatal(err)
	}
	response, err := AcceptWriteCandidate(context.Background(), r, *candidate)
	if err != nil {
		t.Fatal(err)
	}
	if response.Status != WriteStatusAccepted || response.AcceptedCommit != candidate.CandidateCommit || response.LedgerCommit != candidate.CandidateCommit {
		t.Fatalf("unexpected acceptance response: %#v", response)
	}
	manifest, err := Replay(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	state, ok := manifest.GovernedRecords[testTargetID]
	if !ok || state.Status != reducer.StatusActive || state.CurrentEventID != "candidate-1" || state.Revision != 1 {
		t.Fatalf("accepted candidate not recoverable through replay: %#v", state)
	}
}

func TestCompetingCandidatesRejectStaleSecond(t *testing.T) {
	r, head := candidateTestReader(t)
	firstEvent := makeCreateCandidateEvent(t, head, "candidate-1", "idem-1", json.RawMessage(`{"enabled":true}`))
	secondEvent := makeCreateCandidateEvent(t, head, "candidate-2", "idem-2", json.RawMessage(`{"enabled":false}`))
	first, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-1.json", Event: firstEvent})
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-2.json", Event: secondEvent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *first); err != nil {
		t.Fatal(err)
	}
	_, err = AcceptWriteCandidate(context.Background(), r, *second)
	if err == nil || !errors.Is(err, gitledger.ErrStaleState) {
		t.Fatalf("expected stale-state rejection, got %v", err)
	}
}

func TestRetryAfterCASReconstructsOriginalAcceptance(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-1", "idem-retry", json.RawMessage(`{"enabled":true}`))
	req := CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-1.json", Event: event}
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, req)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate a process crash immediately after the atomic ref update and
	// before a response can be returned to the caller.
	if err := r.CompareAndSwap(context.Background(), candidate.ExpectedHead, candidate.CandidateCommit); err != nil {
		t.Fatal(err)
	}

	restarted, err := gitledger.New(r.GitDir(), gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	retryCandidate, response, err := PrepareWriteCandidate(context.Background(), restarted, req)
	if err != nil {
		t.Fatal(err)
	}
	if retryCandidate != nil || response == nil || response.Status != WriteStatusAlreadyAccepted {
		t.Fatalf("retry was not reconstructed from durable ledger: candidate=%#v response=%#v", retryCandidate, response)
	}
	if response.AcceptedCommit != candidate.CandidateCommit || response.EventID != candidate.EventID || response.ContentSHA256 != candidate.ContentSHA256 {
		t.Fatalf("retry response changed accepted identity: %#v", response)
	}
}

func TestIdempotencyConflictWinsOverStaleRebase(t *testing.T) {
	r, head := candidateTestReader(t)
	firstEvent := makeCreateCandidateEvent(t, head, "candidate-1", "idem-conflict", json.RawMessage(`{"enabled":true}`))
	first, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-1.json", Event: firstEvent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *first); err != nil {
		t.Fatal(err)
	}

	conflicting := makeCreateCandidateEvent(t, head, "candidate-2", "idem-conflict", json.RawMessage(`{"enabled":false}`))
	_, _, err = PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-2.json", Event: conflicting})
	if err == nil || !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("expected idempotency conflict, got %v", err)
	}
}

func TestCASIgnoresRepositoryReferenceTransactionHook(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell hook fixture is POSIX-only")
	}
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-1", "idem-hook", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-1.json", Event: event})
	if err != nil {
		t.Fatal(err)
	}

	sentinel := filepath.Join(t.TempDir(), "hook-ran")
	hook := filepath.Join(r.GitDir(), "hooks", "reference-transaction")
	if err := os.MkdirAll(filepath.Dir(hook), 0o755); err != nil {
		t.Fatal(err)
	}
	script := "#!/bin/sh\necho ran > " + shellQuote(sentinel) + "\nexit 1\n"
	if err := os.WriteFile(hook, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *candidate); err != nil {
		t.Fatalf("repository hook affected CAS: %v", err)
	}
	if _, err := os.Stat(sentinel); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("repository hook unexpectedly executed: %v", err)
	}
}

func TestPrepareRejectsNonCanonicalEventBeforeGitMutation(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-1", "idem-noncanonical", json.RawMessage(`{"enabled":true}`))
	event = append([]byte(" \n"), event...)
	_, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-1.json", Event: event})
	if err == nil || !errors.Is(err, ErrCandidateInvalid) {
		t.Fatalf("expected canonical-input rejection, got %v", err)
	}
	got, headErr := r.Head(context.Background())
	if headErr != nil {
		t.Fatal(headErr)
	}
	if got != head {
		t.Fatalf("invalid prepare changed authority: got %s want %s", got, head)
	}
}

func TestPrepareRejectsExistingEventPath(t *testing.T) {
	r, head := candidateTestReader(t)
	value1 := json.RawMessage(`{"enabled":true}`)
	firstEvent := makeCreateCandidateEvent(t, head, "candidate-1", "idem-path-1", value1)
	first, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/fixed.json", Event: firstEvent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *first); err != nil {
		t.Fatal(err)
	}
	newHead, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	replacement := makeReplaceCandidateEvent(t, newHead, "candidate-2", "idem-path-2", value1, json.RawMessage(`{"enabled":false}`), "candidate-1")
	_, _, err = PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: newHead, EventPath: "events/governance/fixed.json", Event: replacement})
	if err == nil || !strings.Contains(err.Error(), "EVENT_PATH_EXISTS") {
		t.Fatalf("expected add-only path rejection, got %v", err)
	}
}

func TestPrepareRejectsUnsafeEventPath(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-1", "idem-unsafe-path", json.RawMessage(`{"enabled":true}`))
	_, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/../candidate-1.json", Event: event})
	if err == nil || !strings.Contains(err.Error(), "invalid durable event path") {
		t.Fatalf("expected safe-path rejection, got %v", err)
	}
}

func TestCASRejectsCandidateThatIsNotExactChild(t *testing.T) {
	r, head := candidateTestReader(t)
	firstEvent := makeCreateCandidateEvent(t, head, "candidate-1", "idem-child-1", json.RawMessage(`{"enabled":true}`))
	secondEvent := makeCreateCandidateEvent(t, head, "candidate-2", "idem-child-2", json.RawMessage(`{"enabled":false}`))
	first, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-1.json", Event: firstEvent})
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/candidate-2.json", Event: secondEvent})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *first); err != nil {
		t.Fatal(err)
	}
	err = r.CompareAndSwap(context.Background(), first.CandidateCommit, second.CandidateCommit)
	if err == nil || !errors.Is(err, gitledger.ErrCandidateNotChild) {
		t.Fatalf("expected exact-child rejection, got %v", err)
	}
}

func candidateTestReader(t *testing.T) (*gitledger.Reader, string) {
	t.Helper()
	work := setupReducerLedger(t, true)
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	head, err := r.Head(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	return r, head
}

func makeCreateCandidateEvent(t *testing.T, expectedHead, eventID, idempotencyKey string, value json.RawMessage) []byte {
	t.Helper()
	state := reducer.State{
		TargetID:       testTargetID,
		RecordKind:     testRecordKind,
		Status:         reducer.StatusActive,
		Revision:       1,
		CurrentEventID: eventID,
		Value:          value,
	}
	raw, err := json.Marshal(map[string]any{
		"schema_version":           contracts.ExclusiveRecordEventSchemaV1,
		"event_id":                 eventID,
		"event_type":               reducer.EventCreated,
		"occurred_at":              "2026-08-11T20:00:00Z",
		"actor":                    map[string]any{"type": "human", "id": "owner:test"},
		"expected_ledger_commit":   expectedHead,
		"authority_policy_version": testPolicyV1,
		"targets":                  []string{testTargetID},
		"source_versions":          []string{"source:fixture@1"},
		"record_kind":              testRecordKind,
		"value":                    value,
		"prior_state":              map[string]any{"exists": false, "target_id": testTargetID},
		"resulting_state":          state,
		"reason":                   "candidate-write conformance fixture",
		"idempotency_key":          idempotencyKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	return completed
}

func makeReplaceCandidateEvent(t *testing.T, expectedHead, eventID, idempotencyKey string, previousValue, value json.RawMessage, previousEventID string) []byte {
	t.Helper()
	prior := reducer.State{
		TargetID:       testTargetID,
		RecordKind:     testRecordKind,
		Status:         reducer.StatusActive,
		Revision:       1,
		CurrentEventID: previousEventID,
		Value:          previousValue,
	}
	previous := previousEventID
	result := reducer.State{
		TargetID:        testTargetID,
		RecordKind:      testRecordKind,
		Status:          reducer.StatusActive,
		Revision:        2,
		CurrentEventID:  eventID,
		PreviousEventID: &previous,
		Value:           value,
	}
	raw, err := json.Marshal(map[string]any{
		"schema_version":           contracts.ExclusiveRecordEventSchemaV1,
		"event_id":                 eventID,
		"event_type":               reducer.EventReplaced,
		"occurred_at":              "2026-08-11T20:01:00Z",
		"actor":                    map[string]any{"type": "human", "id": "owner:test"},
		"expected_ledger_commit":   expectedHead,
		"authority_policy_version": testPolicyV1,
		"targets":                  []string{testTargetID},
		"source_versions":          []string{"source:fixture@2"},
		"record_kind":              testRecordKind,
		"value":                    value,
		"prior_state":              prior,
		"resulting_state":          result,
		"reason":                   "candidate replacement fixture",
		"idempotency_key":          idempotencyKey,
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	return completed
}

func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
