package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
	contractschemas "github.com/armpitpete/threadkeeper-core/schemas"
)

const (
	testRecordKind = "project.setting"
	testTargetID   = "setting:release-mode"
	testPolicyV1   = "authority-policy:v1"
)

func TestReplayAppliesBoundGovernedRecordLifecycleDeterministically(t *testing.T) {
	work := setupReducerLedger(t, true)
	configHead := runGit(t, work, "rev-parse", "HEAD")

	value1 := json.RawMessage(`{"enabled":true}`)
	state1 := reducer.State{
		TargetID: testTargetID, RecordKind: testRecordKind, Status: reducer.StatusActive,
		Revision: 1, CurrentEventID: "record-1", Value: value1,
	}
	writeReducerEvent(t, work, configHead, "record-1", reducer.EventCreated, "idem-1",
		map[string]any{"exists": false, "target_id": testTargetID}, state1, value1, true, testPolicyV1)
	commitAll(t, work, "accept record create")
	createHead := runGit(t, work, "rev-parse", "HEAD")

	previous1 := "record-1"
	value2 := json.RawMessage(`{"enabled":false}`)
	state2 := reducer.State{
		TargetID: testTargetID, RecordKind: testRecordKind, Status: reducer.StatusActive,
		Revision: 2, CurrentEventID: "record-2", PreviousEventID: &previous1, Value: value2,
	}
	writeReducerEvent(t, work, createHead, "record-2", reducer.EventReplaced, "idem-2",
		state1, state2, value2, true, testPolicyV1)
	commitAll(t, work, "accept record replacement")
	replaceHead := runGit(t, work, "rev-parse", "HEAD")

	previous2 := "record-2"
	state3 := reducer.State{
		TargetID: testTargetID, RecordKind: testRecordKind, Status: reducer.StatusRevoked,
		Revision: 3, CurrentEventID: "record-3", PreviousEventID: &previous2,
	}
	writeReducerEvent(t, work, replaceHead, "record-3", reducer.EventRevoked, "idem-3",
		state2, state3, nil, false, testPolicyV1)
	commitAll(t, work, "accept record revocation")

	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Replay(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	second, err := Replay(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	if first.ReducerBindingCount != 1 {
		t.Fatalf("binding count = %d, want 1", first.ReducerBindingCount)
	}
	if first.GovernedRecordCount != 1 {
		t.Fatalf("governed record count = %d, want 1", first.GovernedRecordCount)
	}
	got, ok := first.GovernedRecords[testTargetID]
	if !ok {
		t.Fatalf("missing governed record %q", testTargetID)
	}
	if got.Status != reducer.StatusRevoked || got.Revision != 3 || got.CurrentEventID != "record-3" || got.Value != nil {
		t.Fatalf("unexpected final state: %#v", got)
	}
	if first.GovernedRecordsSHA256 == "" || first.GovernedRecordsSHA256 != second.GovernedRecordsSHA256 {
		t.Fatalf("governed projection digest is not deterministic: %q / %q", first.GovernedRecordsSHA256, second.GovernedRecordsSHA256)
	}
	if first.ReplaySHA256 == "" || first.ReplaySHA256 != second.ReplaySHA256 {
		t.Fatalf("replay digest is not deterministic: %q / %q", first.ReplaySHA256, second.ReplaySHA256)
	}
}

func TestReplayRejectsUnboundCoreRecordEvent(t *testing.T) {
	work := setupReducerLedger(t, false)
	head := runGit(t, work, "rev-parse", "HEAD")
	value := json.RawMessage(`{"enabled":true}`)
	result := reducer.State{TargetID: testTargetID, RecordKind: testRecordKind, Status: reducer.StatusActive, Revision: 1, CurrentEventID: "record-1", Value: value}
	writeReducerEvent(t, work, head, "record-1", reducer.EventCreated, "idem-1",
		map[string]any{"exists": false, "target_id": testTargetID}, result, value, true, testPolicyV1)
	commitAll(t, work, "accept unbound event")

	_, err := replayWorkRepo(t, work)
	if err == nil || !strings.Contains(err.Error(), "REDUCER_POLICY_UNBOUND") {
		t.Fatalf("expected unbound reducer failure, got %v", err)
	}
}

func TestReplayRejectsExpectedLedgerCommitMismatch(t *testing.T) {
	work := setupReducerLedger(t, true)
	value := json.RawMessage(`{"enabled":true}`)
	result := reducer.State{TargetID: testTargetID, RecordKind: testRecordKind, Status: reducer.StatusActive, Revision: 1, CurrentEventID: "record-1", Value: value}
	writeReducerEvent(t, work, strings.Repeat("0", 40), "record-1", reducer.EventCreated, "idem-1",
		map[string]any{"exists": false, "target_id": testTargetID}, result, value, true, testPolicyV1)
	commitAll(t, work, "accept stale-head event")

	_, err := replayWorkRepo(t, work)
	if err == nil || !strings.Contains(err.Error(), "EXPECTED_LEDGER_COMMIT_MISMATCH") {
		t.Fatalf("expected ledger-head mismatch, got %v", err)
	}
}

func TestReplayRejectsAuthorityPolicyVersionMismatch(t *testing.T) {
	work := setupReducerLedger(t, true)
	head := runGit(t, work, "rev-parse", "HEAD")
	value := json.RawMessage(`{"enabled":true}`)
	result := reducer.State{TargetID: testTargetID, RecordKind: testRecordKind, Status: reducer.StatusActive, Revision: 1, CurrentEventID: "record-1", Value: value}
	writeReducerEvent(t, work, head, "record-1", reducer.EventCreated, "idem-1",
		map[string]any{"exists": false, "target_id": testTargetID}, result, value, true, "authority-policy:v2")
	commitAll(t, work, "accept wrong-policy event")

	_, err := replayWorkRepo(t, work)
	if err == nil || !strings.Contains(err.Error(), "AUTHORITY_POLICY_VERSION_MISMATCH") {
		t.Fatalf("expected authority-policy mismatch, got %v", err)
	}
}

func TestReplayRejectsReducerBindingMutation(t *testing.T) {
	work := setupReducerLedger(t, true)
	writeReducerBinding(t, work, "authority-policy:v2")
	commitAll(t, work, "illegally mutate reducer binding")

	_, err := replayWorkRepo(t, work)
	if err == nil || !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("expected immutable binding failure, got %v", err)
	}
}

func TestReplayRejectsSchemaMutation(t *testing.T) {
	work := setupReducerLedger(t, true)
	path := filepath.Join(work, "config/schemas/exclusive-record-event-v1.json")
	mutated := append(append([]byte(nil), contractschemas.ExclusiveGovernedRecordEventV1...), ' ')
	if err := os.WriteFile(path, mutated, 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "illegally mutate accepted schema")

	_, err := replayWorkRepo(t, work)
	if err == nil || !strings.Contains(err.Error(), "append-only") {
		t.Fatalf("expected immutable schema failure, got %v", err)
	}
}

func TestReplayRejectsDuplicateIdempotencyKey(t *testing.T) {
	work := setupReducerLedger(t, true)
	configHead := runGit(t, work, "rev-parse", "HEAD")
	value1 := json.RawMessage(`{"enabled":true}`)
	state1 := reducer.State{TargetID: testTargetID, RecordKind: testRecordKind, Status: reducer.StatusActive, Revision: 1, CurrentEventID: "record-1", Value: value1}
	writeReducerEvent(t, work, configHead, "record-1", reducer.EventCreated, "idem-same",
		map[string]any{"exists": false, "target_id": testTargetID}, state1, value1, true, testPolicyV1)
	commitAll(t, work, "accept record create")
	createHead := runGit(t, work, "rev-parse", "HEAD")

	previous := "record-1"
	value2 := json.RawMessage(`{"enabled":false}`)
	state2 := reducer.State{TargetID: testTargetID, RecordKind: testRecordKind, Status: reducer.StatusActive, Revision: 2, CurrentEventID: "record-2", PreviousEventID: &previous, Value: value2}
	writeReducerEvent(t, work, createHead, "record-2", reducer.EventReplaced, "idem-same",
		state1, state2, value2, true, testPolicyV1)
	commitAll(t, work, "accept duplicate idempotency key")

	_, err := replayWorkRepo(t, work)
	if err == nil || !strings.Contains(err.Error(), "duplicate idempotency_key") {
		t.Fatalf("expected duplicate idempotency failure, got %v", err)
	}
}

func setupReducerLedger(t *testing.T, withBinding bool) string {
	t.Helper()
	work := newWorkRepo(t)
	writeFixtureFile(t, work, "config/schemas/reducer-binding-v1.json", contractschemas.ReducerBindingV1)
	writeFixtureFile(t, work, "config/schemas/exclusive-record-event-v1.json", contractschemas.ExclusiveGovernedRecordEventV1)
	if withBinding {
		writeReducerBinding(t, work, testPolicyV1)
	}
	commitAll(t, work, "configure reducer contracts")
	return work
}

func writeReducerBinding(t *testing.T, work, policyVersion string) {
	t.Helper()
	candidate, err := json.Marshal(map[string]any{
		"schema_version":           contracts.ReducerBindingSchemaV1,
		"binding_id":               "binding:project-setting:v1",
		"record_kind":              testRecordKind,
		"state_model":              reducer.ModelExclusiveV1,
		"event_schema":             contracts.ExclusiveRecordEventSchemaV1,
		"authority_policy_version": policyVersion,
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(candidate)
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, work, "config/authority/reducer-bindings/project-setting-v1.json", completed)
}

func writeReducerEvent(t *testing.T, work, expectedHead, eventID, eventType, idempotencyKey string, prior, resulting any, value json.RawMessage, includeValue bool, policyVersion string) {
	t.Helper()
	candidate := map[string]any{
		"schema_version":           contracts.ExclusiveRecordEventSchemaV1,
		"event_id":                 eventID,
		"event_type":               eventType,
		"occurred_at":              "2026-08-11T19:00:00Z",
		"actor":                    map[string]any{"type": "human", "id": "owner:test"},
		"expected_ledger_commit":   expectedHead,
		"authority_policy_version": policyVersion,
		"targets":                  []string{testTargetID},
		"source_versions":          []string{"source:fixture@1"},
		"record_kind":              testRecordKind,
		"prior_state":              prior,
		"resulting_state":          resulting,
		"reason":                   "conformance fixture",
		"idempotency_key":          idempotencyKey,
	}
	if includeValue {
		candidate["value"] = value
	}
	raw, err := json.Marshal(candidate)
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	writeFixtureFile(t, work, "events/governance/"+eventID+".json", completed)
}

func writeFixtureFile(t *testing.T, work, rel string, data []byte) {
	t.Helper()
	path := filepath.Join(work, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func replayWorkRepo(t *testing.T, work string) (*ReplayManifest, error) {
	t.Helper()
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		return nil, err
	}
	return Replay(context.Background(), r)
}
