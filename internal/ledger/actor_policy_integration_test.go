package ledger

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
)

func TestReplayRejectsDirectRootActorPolicyMutation(t *testing.T) {
	work := newWorkRepo(t)
	path := filepath.Join(work, filepath.FromSlash(actorauth.LedgerPolicyPath))
	if err := os.WriteFile(path, actorPolicyValue(t, 2, actorauth.ActionOperate, "target:test", false), 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "mutate root actor policy")
	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := Replay(context.Background(), r); err == nil || !strings.Contains(err.Error(), "AUTH_POLICY_IMMUTABLE") {
		t.Fatalf("mutated root actor policy replay = %v", err)
	}
}

func TestFreshGenesisRejectsActorPolicyAuthorityMismatchBeforeCreation(t *testing.T) {
	target := filepath.Join(t.TempDir(), "must-not-exist.git")
	seed := freshGenesisSeed(t, nil)
	seed[actorauth.LedgerPolicyPath] = actorPolicyValueForActor(t, 3, "other:actor", actorauth.ActionOperate, "target:test", false)
	_, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), seed)
	if err == nil || !strings.Contains(err.Error(), "initial_authorities") {
		t.Fatalf("authority mismatch = %v", err)
	}
	if _, statErr := os.Lstat(target); !os.IsNotExist(statErr) {
		t.Fatalf("authority mismatch created target: %v", statErr)
	}
}

func TestPrepareRejectsMalformedActorPolicyRotationBeforeCAS(t *testing.T) {
	r, head := freshActorPolicyReader(t)
	defer r.Close()
	malformed := malformedActorPolicyValue(t)
	event := actorPolicyCreateEvent(t, head, "actor-policy-malformed", "idem-actor-policy-malformed", malformed)
	candidate, response, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/actor-policy-malformed.json",
		Event:        event,
	})
	if err == nil || !strings.Contains(err.Error(), "AUTH_POLICY_INVALID") {
		t.Fatalf("malformed actor policy prepared candidate=%#v response=%#v err=%v", candidate, response, err)
	}
	got, headErr := r.Head(context.Background())
	if headErr != nil {
		t.Fatal(headErr)
	}
	if got != head {
		t.Fatalf("malformed actor policy changed authority: got %s want %s", got, head)
	}
}

func TestGovernedActorPolicyRotationIsLoadedAfterRestart(t *testing.T) {
	r, head := freshActorPolicyReader(t)
	defer r.Close()
	rotated := actorPolicyValue(t, 4, actorauth.ActionDecide, "project:test/status", false)
	eventID := "actor-policy-rotation"
	event := actorPolicyCreateEvent(t, head, eventID, "idem-actor-policy-rotation", rotated)
	candidate, response, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/actor-policy-rotation.json",
		Event:        event,
	})
	if err != nil || candidate == nil || response != nil {
		t.Fatalf("prepare rotation candidate=%#v response=%#v err=%v", candidate, response, err)
	}
	accepted, err := AcceptWriteCandidate(context.Background(), r, *candidate)
	if err != nil || accepted == nil || accepted.Status != WriteStatusAccepted {
		t.Fatalf("accept rotation response=%#v err=%v", accepted, err)
	}

	restarted, err := gitledger.New(r.GitDir(), gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Close()
	snapshot, err := LoadCurrentActorPolicy(context.Background(), restarted)
	if err != nil {
		t.Fatal(err)
	}
	if snapshot.SourceEventID != eventID || snapshot.PolicyContentSHA == "" || snapshot.LedgerCommit != accepted.LedgerCommit {
		t.Fatalf("restart did not load rotated policy: %#v", snapshot)
	}
	if len(snapshot.Policy.Grants) != 1 || snapshot.Policy.Grants[0].Action != actorauth.ActionDecide || snapshot.Policy.Grants[0].Target != "project:test/status" || snapshot.Policy.Grants[0].LedgerID != "ledger:fresh-test" {
		t.Fatalf("rotated exact grants = %#v", snapshot.Policy.Grants)
	}
}

func TestGovernedActorPolicyRevocationFailsClosed(t *testing.T) {
	r, head := freshActorPolicyReader(t)
	defer r.Close()
	value := actorPolicyValue(t, 5, actorauth.ActionOperate, "target:test", false)
	createID := "actor-policy-create-before-revoke"
	create := actorPolicyCreateEvent(t, head, createID, "idem-policy-create-before-revoke", value)
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: head, EventPath: "events/governance/policy-create-before-revoke.json", Event: create})
	if err != nil {
		t.Fatal(err)
	}
	accepted, err := AcceptWriteCandidate(context.Background(), r, *candidate)
	if err != nil {
		t.Fatal(err)
	}

	manifest, err := Replay(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	prior := manifest.GovernedRecords[actorauth.PolicyTarget]
	revokeID := "actor-policy-revoke"
	revoke := actorPolicyRevokeEvent(t, accepted.LedgerCommit, revokeID, "idem-actor-policy-revoke", prior)
	revokeCandidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{ExpectedHead: accepted.LedgerCommit, EventPath: "events/governance/actor-policy-revoke.json", Event: revoke})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := AcceptWriteCandidate(context.Background(), r, *revokeCandidate); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadCurrentActorPolicy(context.Background(), r); err == nil || !strings.Contains(err.Error(), "AUTH_POLICY_REVOKED") {
		t.Fatalf("revoked policy did not fail closed: %v", err)
	}
}

func freshActorPolicyReader(t *testing.T) (*gitledger.Reader, string) {
	t.Helper()
	target := filepath.Join(t.TempDir(), "actor-policy-ledger.git")
	if _, err := InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, freshGenesisFixture(t, nil), freshGenesisSeed(t, nil)); err != nil {
		t.Fatal(err)
	}
	r, err := gitledger.New(target, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	head, err := r.Head(context.Background())
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	return r, head
}

func actorPolicyValue(t *testing.T, seedByte byte, action actorauth.Action, target string, revoked bool) json.RawMessage {
	return actorPolicyValueForActor(t, seedByte, "owner:test", action, target, revoked)
}

func actorPolicyValueForActor(t *testing.T, seedByte byte, actorID string, action actorauth.Action, target string, revoked bool) json.RawMessage {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = seedByte
	}
	publicKey := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	raw, err := json.Marshal(map[string]any{
		"max_proof_lifetime_seconds": int64(300),
		"keys": []map[string]any{{"actor_id": actorID, "key_id": "key:rotated:v1", "public_key": base64.RawStdEncoding.EncodeToString(publicKey), "revoked": revoked}},
		"grants": []map[string]any{{"actor_id": actorID, "action": action, "target": target}},
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

func malformedActorPolicyValue(t *testing.T) json.RawMessage {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	publicKey := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	raw, err := json.Marshal(map[string]any{
		"max_proof_lifetime_seconds": int64(300),
		"keys": []map[string]any{{"actor_id": "owner:test", "key_id": "revoked-only", "public_key": base64.RawStdEncoding.EncodeToString(publicKey), "revoked": true}},
		"grants": []map[string]any{{"actor_id": "owner:test", "action": actorauth.ActionOperate, "target": "target:test"}},
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

func actorPolicyCreateEvent(t *testing.T, expectedHead, eventID, idempotencyKey string, value json.RawMessage) []byte {
	t.Helper()
	state := reducer.State{TargetID: actorauth.PolicyTarget, RecordKind: actorauth.PolicyRecordKind, Status: reducer.StatusActive, Revision: 1, CurrentEventID: eventID, Value: value}
	raw, err := json.Marshal(map[string]any{
		"schema_version": contracts.ExclusiveRecordEventSchemaV1, "event_id": eventID, "event_type": reducer.EventCreated,
		"occurred_at": "2026-08-14T12:00:00Z", "actor": map[string]any{"type": "human", "id": "owner:test"},
		"expected_ledger_commit": expectedHead, "authority_policy_version": testPolicyV1,
		"targets": []string{actorauth.PolicyTarget}, "source_versions": []string{"source:actor-policy@1"},
		"record_kind": actorauth.PolicyRecordKind, "value": value,
		"prior_state": map[string]any{"exists": false, "target_id": actorauth.PolicyTarget}, "resulting_state": state,
		"reason": "rotate authoritative actor policy", "idempotency_key": idempotencyKey,
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

func actorPolicyRevokeEvent(t *testing.T, expectedHead, eventID, idempotencyKey string, prior reducer.State) []byte {
	t.Helper()
	previous := prior.CurrentEventID
	result := reducer.State{TargetID: actorauth.PolicyTarget, RecordKind: actorauth.PolicyRecordKind, Status: reducer.StatusRevoked, Revision: prior.Revision + 1, CurrentEventID: eventID, PreviousEventID: &previous}
	raw, err := json.Marshal(map[string]any{
		"schema_version": contracts.ExclusiveRecordEventSchemaV1, "event_id": eventID, "event_type": reducer.EventRevoked,
		"occurred_at": "2026-08-14T12:01:00Z", "actor": map[string]any{"type": "human", "id": "owner:test"},
		"expected_ledger_commit": expectedHead, "authority_policy_version": testPolicyV1,
		"targets": []string{actorauth.PolicyTarget}, "source_versions": []string{"source:actor-policy@2"},
		"record_kind": actorauth.PolicyRecordKind, "prior_state": prior, "resulting_state": result,
		"reason": "emergency revoke authoritative actor policy", "idempotency_key": idempotencyKey,
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
