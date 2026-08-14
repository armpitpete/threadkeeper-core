package service

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
	contractschemas "github.com/armpitpete/threadkeeper-core/schemas"
)

const (
	e2eProjectID     = "project:e2e-release"
	e2eLedgerID      = "ledger:e2e-release"
	e2eActorID       = "owner:e2e-release"
	e2eKeyID         = "key:e2e-release:v1"
	e2ePolicyVersion = "authority-policy:e2e-release:v1"
	e2eRecordKind    = "core.e2e-release-record-v1"
	e2eTarget        = "e2e:release-record"
	e2eIdempotency   = "e2e-release-idempotency-v1"
)

func TestWriteDisabledCoreV1EndToEndReleaseAcceptance(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for release acceptance integration test")
	}
	ctx := context.Background()
	publicKey, privateKey := deterministicE2EKey()
	root, seed := e2eGenesisMaterial(t, publicKey)

	rootDir := t.TempDir()
	ledgerDir := filepath.Join(rootDir, "authority.git")
	evidence, err := ledger.InitializeFreshGenesis(ctx, ledgerDir, gitledger.DefaultRef, root, seed)
	if err != nil {
		t.Fatal(err)
	}
	if evidence.GenesisCommit == "" || evidence.GenesisCommit != evidence.LedgerCommit || evidence.LedgerID != e2eLedgerID {
		t.Fatalf("fresh Genesis evidence is not a one-root authority ledger: %#v", evidence)
	}

	r, err := gitledger.New(ledgerDir, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	h0, err := r.Head(ctx)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	if h0 != evidence.GenesisCommit {
		r.Close()
		t.Fatalf("initial head %s != Genesis %s", h0, evidence.GenesisCommit)
	}
	policySnapshot, err := ledger.LoadCurrentActorPolicy(ctx, r)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	if policySnapshot.LedgerCommit != h0 || policySnapshot.LedgerID != e2eLedgerID || policySnapshot.PolicyContentSHA != evidence.ActorPolicyContentSHA256 {
		r.Close()
		t.Fatalf("initial actor policy does not bind to Genesis snapshot: %#v", policySnapshot)
	}

	now := time.Date(2026, 8, 14, 14, 0, 30, 0, time.UTC)
	requestContext := actorauth.RequestContext{
		LedgerID:       e2eLedgerID,
		Action:         actorauth.ActionOperate,
		Target:         e2eTarget,
		ExpectedState:  h0,
		IdempotencyKey: e2eIdempotency,
	}
	proof := signE2EProof(t, privateKey, requestContext, now.Add(-30*time.Second), now.Add(2*time.Minute))

	// The exported service authority path must remain closed even for a fully
	// valid signed proof. A nil Reader proves the release gate dominates before
	// any ledger/policy/authentication work.
	if _, err := AdmitAuthorityWrite(ctx, nil, proof, requestContext, now); !errors.Is(err, ErrAuthorityWritesDisabled) {
		r.Close()
		t.Fatalf("exported authority admission escaped hard kill-switch: %v", err)
	}
	principal, err := authenticateAuthoritativePolicySnapshot(policySnapshot, proof, requestContext, now)
	if err != nil {
		r.Close()
		t.Fatalf("ledger-derived proof authentication failed: %v", err)
	}
	if principal.ActorID != e2eActorID || principal.KeyID != e2eKeyID {
		r.Close()
		t.Fatalf("unexpected authenticated principal: %#v", principal)
	}

	event := e2eCreateEvent(t, h0, "e2e-release-event-1", e2eIdempotency, json.RawMessage(`{"ready":true}`))
	candidate, response, err := ledger.PrepareWriteCandidate(ctx, r, ledger.CandidateRequest{
		ExpectedHead: h0,
		EventPath:    "events/governance/e2e-release-event-1.json",
		Event:        event,
	})
	if err != nil || candidate == nil || response != nil {
		r.Close()
		t.Fatalf("prepare H0->H1 candidate=%#v response=%#v err=%v", candidate, response, err)
	}
	accepted, err := ledger.AcceptWriteCandidate(ctx, r, *candidate)
	if err != nil || accepted == nil || accepted.Status != ledger.WriteStatusAccepted {
		r.Close()
		t.Fatalf("accept H0->H1 response=%#v err=%v", accepted, err)
	}
	h1 := accepted.LedgerCommit
	if h1 == "" || h1 == h0 {
		r.Close()
		t.Fatalf("accepted write did not advance authority: H0=%s H1=%s", h0, h1)
	}
	originalProof, err := ledger.ProveRecovery(ctx, r)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	originalManifest, err := ledger.Replay(ctx, r)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	if originalManifest.LedgerCommit != h1 || originalManifest.GovernedRecordCount == 0 {
		r.Close()
		t.Fatalf("accepted projection not authoritative at H1: %#v", originalManifest)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	// Restart from the real on-disk ledger and require exact trust/recovery
	// identity before exercising idempotent and conflict outcomes.
	restarted, err := gitledger.New(ledgerDir, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	restartedPolicy, err := ledger.LoadCurrentActorPolicy(ctx, restarted)
	if err != nil {
		restarted.Close()
		t.Fatal(err)
	}
	if restartedPolicy.GenesisCommit != evidence.GenesisCommit || restartedPolicy.LedgerCommit != h1 || restartedPolicy.PolicyContentSHA != evidence.ActorPolicyContentSHA256 {
		restarted.Close()
		t.Fatalf("restart changed Genesis/policy/head identity: %#v", restartedPolicy)
	}
	restartedProof, err := ledger.ProveRecovery(ctx, restarted)
	if err != nil {
		restarted.Close()
		t.Fatal(err)
	}
	if err := ledger.CompareRecoveryProofs(*originalProof, *restartedProof); err != nil {
		restarted.Close()
		t.Fatalf("restart recovery proof changed: %v", err)
	}

	retry, err := ledger.AcceptWriteCandidate(ctx, restarted, *candidate)
	if err != nil || retry == nil || retry.Status != ledger.WriteStatusAlreadyAccepted || retry.AcceptedCommit != h1 {
		restarted.Close()
		t.Fatalf("idempotent retry response=%#v err=%v want already_accepted %s", retry, err, h1)
	}
	conflictingEvent := e2eCreateEvent(t, h1, "e2e-release-event-conflict", e2eIdempotency, json.RawMessage(`{"ready":false}`))
	conflictCandidate, conflictResponse, conflictErr := ledger.PrepareWriteCandidate(ctx, restarted, ledger.CandidateRequest{
		ExpectedHead: h1,
		EventPath:    "events/governance/e2e-release-event-conflict.json",
		Event:        conflictingEvent,
	})
	if conflictErr == nil || !strings.Contains(conflictErr.Error(), "IDEMPOTENCY_CONFLICT") || conflictCandidate != nil || conflictResponse != nil {
		restarted.Close()
		t.Fatalf("same-key conflict candidate=%#v response=%#v err=%v", conflictCandidate, conflictResponse, conflictErr)
	}
	current, err := restarted.Head(ctx)
	if err != nil {
		restarted.Close()
		t.Fatal(err)
	}
	if current != h1 {
		restarted.Close()
		t.Fatalf("same-key conflict moved authority: got %s want %s", current, h1)
	}

	backup := filepath.Join(t.TempDir(), "secondary-local.git")
	runE2EGit(t, "", "clone", "--bare", "--no-hardlinks", restarted.GitDir(), backup)
	if err := restarted.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(ledgerDir); err != nil {
		t.Fatal(err)
	}
	restoredDir := filepath.Join(t.TempDir(), "restored-authority.git")
	runE2EGit(t, "", "clone", "--bare", "--no-hardlinks", backup, restoredDir)
	restored, err := gitledger.New(restoredDir, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()
	restoredProof, err := ledger.ProveRecovery(ctx, restored)
	if err != nil {
		t.Fatal(err)
	}
	if err := ledger.CompareRecoveryProofs(*originalProof, *restoredProof); err != nil {
		t.Fatalf("destructive local restore changed recovery proof: %v", err)
	}
	restoredManifest, err := ledger.Replay(ctx, restored)
	if err != nil {
		t.Fatal(err)
	}
	if restoredManifest.GenesisCommit != originalManifest.GenesisCommit || restoredManifest.LedgerCommit != h1 || restoredManifest.GovernedRecordsSHA256 != originalManifest.GovernedRecordsSHA256 || restoredManifest.ReplaySHA256 != originalManifest.ReplaySHA256 {
		t.Fatalf("restored replay differs: original=%#v restored=%#v", originalManifest, restoredManifest)
	}

	// The complete code-side acceptance run must end with the public release gate
	// still closed. No test helper may convert this proof into write enablement.
	if AuthorityWritesEnabled() {
		t.Fatal("end-to-end harness enabled authority writes")
	}
}

func deterministicE2EKey() (ed25519.PublicKey, ed25519.PrivateKey) {
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = 0x47
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	return privateKey.Public().(ed25519.PublicKey), privateKey
}

func e2eGenesisMaterial(t *testing.T, publicKey ed25519.PublicKey) ([]byte, map[string][]byte) {
	t.Helper()
	actorPolicy := completeE2EJSON(t, map[string]any{
		"max_proof_lifetime_seconds": int64(300),
		"keys": []map[string]any{{
			"actor_id":   e2eActorID,
			"key_id":     e2eKeyID,
			"public_key": base64.RawStdEncoding.EncodeToString(publicKey),
			"revoked":    false,
		}},
		"grants": []map[string]any{
			{"actor_id": e2eActorID, "action": actorauth.ActionOperate, "target": actorauth.PolicyTarget},
			{"actor_id": e2eActorID, "action": actorauth.ActionOperate, "target": e2eTarget},
		},
	})
	actorBinding := e2eBinding(t, "binding:e2e-actor-policy:v1", actorauth.PolicyRecordKind)
	recordBinding := e2eBinding(t, "binding:e2e-release-record:v1", e2eRecordKind)
	initialSchemas := []string{contracts.ExclusiveRecordEventSchemaV1, contracts.ReducerBindingSchemaV1}
	sort.Strings(initialSchemas)
	genesis := completeE2EJSON(t, map[string]any{
		"project_id":               e2eProjectID,
		"ledger_id":                e2eLedgerID,
		"created_at":               "2026-08-14T14:00:00Z",
		"initial_authority_policy": e2ePolicyVersion,
		"initial_schemas":          initialSchemas,
		"initial_authorities":      []string{e2eActorID},
	})
	seed := map[string][]byte{
		actorauth.LedgerPolicyPath: actorPolicy,
		"config/schemas/exclusive-record-event-v1.json": contractschemas.ExclusiveGovernedRecordEventV1,
		"config/schemas/reducer-binding-v1.json":        contractschemas.ReducerBindingV1,
		"config/authority/reducer-bindings/e2e-actor-policy-v1.json": actorBinding,
		"config/authority/reducer-bindings/e2e-release-record-v1.json": recordBinding,
	}
	return genesis, seed
}

func e2eBinding(t *testing.T, bindingID, recordKind string) []byte {
	t.Helper()
	return completeE2EJSON(t, map[string]any{
		"schema_version":           contracts.ReducerBindingSchemaV1,
		"binding_id":               bindingID,
		"record_kind":              recordKind,
		"state_model":              reducer.ModelExclusiveV1,
		"event_schema":             contracts.ExclusiveRecordEventSchemaV1,
		"authority_policy_version": e2ePolicyVersion,
	})
}

func e2eCreateEvent(t *testing.T, expectedHead, eventID, idempotencyKey string, value json.RawMessage) []byte {
	t.Helper()
	state := reducer.State{
		TargetID:       e2eTarget,
		RecordKind:     e2eRecordKind,
		Status:         reducer.StatusActive,
		Revision:       1,
		CurrentEventID: eventID,
		Value:          value,
	}
	return completeE2EJSON(t, map[string]any{
		"schema_version":           contracts.ExclusiveRecordEventSchemaV1,
		"event_id":                 eventID,
		"event_type":               reducer.EventCreated,
		"occurred_at":              "2026-08-14T14:01:00Z",
		"actor":                    map[string]any{"type": "human", "id": e2eActorID},
		"expected_ledger_commit":   expectedHead,
		"authority_policy_version": e2ePolicyVersion,
		"targets":                  []string{e2eTarget},
		"source_versions":          []string{"source:e2e-release@1"},
		"record_kind":              e2eRecordKind,
		"value":                    value,
		"prior_state":              map[string]any{"exists": false, "target_id": e2eTarget},
		"resulting_state":          state,
		"reason":                   "Core v1 write-disabled release acceptance",
		"idempotency_key":          idempotencyKey,
	})
}

func signE2EProof(t *testing.T, privateKey ed25519.PrivateKey, requestContext actorauth.RequestContext, issuedAt, expiresAt time.Time) actorauth.Proof {
	t.Helper()
	proof := actorauth.Proof{
		ActorID:   e2eActorID,
		KeyID:     e2eKeyID,
		IssuedAt:  issuedAt.Format(time.RFC3339),
		ExpiresAt: expiresAt.Format(time.RFC3339),
		Context:   requestContext,
	}
	raw, err := json.Marshal(proof)
	if err != nil {
		t.Fatal(err)
	}
	message, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		t.Fatal(err)
	}
	proof.Signature = base64.RawStdEncoding.EncodeToString(ed25519.Sign(privateKey, message))
	return proof
}

func completeE2EJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	return completed
}

func runE2EGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	var cmd *exec.Cmd
	if dir == "" {
		cmd = exec.Command("git", args...)
	} else {
		cmdArgs := append([]string{"-C", dir}, args...)
		cmd = exec.Command("git", cmdArgs...)
	}
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
