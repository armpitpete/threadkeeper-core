package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/restoreproof"
)

// TestCoreV1ConsolidatedReleaseAcceptanceEvidence is the Issue #50 reference
// acceptance sequence. It is intentionally disposable and test-only: the
// exported service gate remains hard closed, while the already-reviewed pure
// authentication primitive and internal candidate/quarantine/CAS machinery are
// exercised directly against a fresh temporary authority ledger.
func TestCoreV1ConsolidatedReleaseAcceptanceEvidence(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for release acceptance integration test")
	}
	ctx := context.Background()
	publicKey, privateKey := deterministicE2EKey()
	root, seed := e2eGenesisMaterial(t, publicKey)

	rootDir := t.TempDir()
	ledgerDir := filepath.Join(rootDir, "authority.git")
	genesisEvidence, err := ledger.InitializeFreshGenesis(ctx, ledgerDir, gitledger.DefaultRef, root, seed)
	if err != nil {
		t.Fatal(err)
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
	if h0 != genesisEvidence.GenesisCommit {
		r.Close()
		t.Fatalf("initial head %s != Genesis %s", h0, genesisEvidence.GenesisCommit)
	}
	policySnapshot, err := ledger.LoadCurrentActorPolicy(ctx, r)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	if policySnapshot.GenesisCommit != genesisEvidence.GenesisCommit || policySnapshot.LedgerCommit != h0 || policySnapshot.PolicyContentSHA != genesisEvidence.ActorPolicyContentSHA256 {
		r.Close()
		t.Fatalf("initial actor policy identity mismatch: %#v", policySnapshot)
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
	if _, err := AdmitAuthorityWrite(ctx, nil, proof, requestContext, now); !errors.Is(err, ErrAuthorityWritesDisabled) {
		r.Close()
		t.Fatalf("exported service gate escaped AUTHORITY_WRITES_DISABLED: %v", err)
	}
	principal, err := authenticateAuthoritativePolicySnapshot(policySnapshot, proof, requestContext, now)
	if err != nil {
		r.Close()
		t.Fatalf("ledger-derived authentication failed: %v", err)
	}
	if principal.ActorID != e2eActorID || principal.KeyID != e2eKeyID {
		r.Close()
		t.Fatalf("unexpected authenticated principal: %#v", principal)
	}

	winnerEvent := e2eCreateEvent(t, h0, "e2e-release-event-1", e2eIdempotency, json.RawMessage(`{"ready":true}`))
	winner, response, err := ledger.PrepareWriteCandidate(ctx, r, ledger.CandidateRequest{
		ExpectedHead: h0,
		EventPath:    "events/governance/e2e-release-event-1.json",
		Event:        winnerEvent,
	})
	if err != nil || winner == nil || response != nil {
		r.Close()
		t.Fatalf("prepare winning candidate=%#v response=%#v err=%v", winner, response, err)
	}

	// Prepare a second valid H0 candidate before the winner moves authority. It
	// models a concurrent writer that loses the exact-head race and must never be
	// silently rebased onto H1.
	staleEvent := e2eCreateEvent(t, h0, "e2e-release-event-stale", "e2e-release-idempotency-stale-v1", json.RawMessage(`{"ready":"stale"}`))
	staleCandidate, stalePrepareResponse, err := ledger.PrepareWriteCandidate(ctx, r, ledger.CandidateRequest{
		ExpectedHead: h0,
		EventPath:    "events/governance/e2e-release-event-stale.json",
		Event:        staleEvent,
	})
	if err != nil || staleCandidate == nil || stalePrepareResponse != nil {
		r.Close()
		t.Fatalf("prepare stale competitor=%#v response=%#v err=%v", staleCandidate, stalePrepareResponse, err)
	}

	accepted, err := ledger.AcceptWriteCandidate(ctx, r, *winner)
	if err != nil || accepted == nil || accepted.Status != ledger.WriteStatusAccepted {
		r.Close()
		t.Fatalf("accept winner response=%#v err=%v", accepted, err)
	}
	h1 := accepted.LedgerCommit
	if h1 == "" || h1 == h0 {
		r.Close()
		t.Fatalf("accepted write did not advance authority: H0=%s H1=%s", h0, h1)
	}
	staleOutcome, staleErr := ledger.AcceptWriteCandidate(ctx, r, *staleCandidate)
	if staleErr == nil || !errors.Is(staleErr, gitledger.ErrStaleState) || staleOutcome != nil {
		r.Close()
		t.Fatalf("stale competitor was not rejected without rebase: response=%#v err=%v", staleOutcome, staleErr)
	}
	if current, err := r.Head(ctx); err != nil || current != h1 {
		r.Close()
		t.Fatalf("stale competitor moved authority: head=%s err=%v want=%s", current, err, h1)
	}

	originalProof, err := ledger.ProveRecovery(ctx, r)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	originalProofSHA, err := restoreproof.RecoveryProofSHA256(*originalProof)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	originalManifest, err := ledger.Replay(ctx, r)
	if err != nil {
		r.Close()
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}

	restarted, err := gitledger.New(ledgerDir, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	restartedPolicy, err := ledger.LoadCurrentActorPolicy(ctx, restarted)
	if err != nil {
		restarted.Close()
		t.Fatal(err)
	}
	restartedProof, err := ledger.ProveRecovery(ctx, restarted)
	if err != nil {
		restarted.Close()
		t.Fatal(err)
	}
	if err := ledger.CompareRecoveryProofs(*originalProof, *restartedProof); err != nil {
		restarted.Close()
		t.Fatalf("restart changed RecoveryProof: %v", err)
	}
	restartedProofSHA, err := restoreproof.RecoveryProofSHA256(*restartedProof)
	if err != nil {
		restarted.Close()
		t.Fatal(err)
	}

	retry, err := ledger.AcceptWriteCandidate(ctx, restarted, *winner)
	if err != nil || retry == nil || retry.Status != ledger.WriteStatusAlreadyAccepted || retry.AcceptedCommit != h1 {
		restarted.Close()
		t.Fatalf("idempotent retry response=%#v err=%v", retry, err)
	}
	conflictingEvent := e2eCreateEvent(t, h1, "e2e-release-event-conflict", e2eIdempotency, json.RawMessage(`{"ready":false}`))
	conflictCandidate, conflictResponse, conflictErr := ledger.PrepareWriteCandidate(ctx, restarted, ledger.CandidateRequest{
		ExpectedHead: h1,
		EventPath:    "events/governance/e2e-release-event-conflict.json",
		Event:        conflictingEvent,
	})
	if conflictErr == nil || !errors.Is(conflictErr, ledger.ErrIdempotencyConflict) || conflictCandidate != nil || conflictResponse != nil {
		restarted.Close()
		t.Fatalf("same-key conflict candidate=%#v response=%#v err=%v", conflictCandidate, conflictResponse, conflictErr)
	}
	if current, err := restarted.Head(ctx); err != nil || current != h1 {
		restarted.Close()
		t.Fatalf("same-key conflict moved authority: head=%s err=%v want=%s", current, err, h1)
	}

	// Make one real file artifact rather than claiming a synthetic hash. The
	// bundle is disposable/local test evidence and is still not operational
	// independence evidence.
	backupBundle := filepath.Join(t.TempDir(), "secondary-local.bundle")
	runE2EGit(t, restarted.GitDir(), "bundle", "create", backupBundle, "--all")
	backupBytes, err := os.ReadFile(backupBundle)
	if err != nil {
		restarted.Close()
		t.Fatal(err)
	}
	backupSum := sha256.Sum256(backupBytes)
	backupSHA := fmt.Sprintf("%x", backupSum[:])
	if err := restarted.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.RemoveAll(ledgerDir); err != nil {
		t.Fatal(err)
	}
	restoredDir := filepath.Join(t.TempDir(), "restored-authority.git")
	runE2EGit(t, "", "clone", "--bare", backupBundle, restoredDir)
	restored, err := gitledger.New(restoredDir, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer restored.Close()

	provenanceRaw := completeE2EJSON(t, map[string]any{
		"schema_version":                 restoreproof.ProvenanceSchemaV1,
		"primary_authority_domain_id":    "authority:e2e-primary",
		"secondary_authority_domain_id":  "authority:e2e-local-secondary",
		"secondary_location_id":          "location:test-tempdir",
		"secondary_operator_id":          "operator:test-harness",
		"backup_set_id":                  "backup:e2e-set-v1",
		"backup_artifact_id":             "artifact:e2e-git-bundle-v1",
		"backup_artifact_sha256":         backupSHA,
		"original_recovery_proof_sha256": originalProofSHA,
		"captured_at":                    "2026-08-14T14:02:00Z",
		"restored_at":                    "2026-08-14T14:03:00Z",
		"external_evidence_refs": []string{
			"evidence:local-test-only",
			"evidence:not-operational-independence",
		},
	})
	provenance, err := restoreproof.DecodeProvenance(provenanceRaw)
	if err != nil {
		t.Fatal(err)
	}
	restoreReport, err := restoreproof.Verify(ctx, restored, *originalProof, provenance)
	if err != nil {
		t.Fatalf("restore verification failed: %v", err)
	}
	if !restoreReport.CoreEquivalencePassed || restoreReport.OperationalIndependenceStatus != restoreproof.OperationalIndependenceRequiresExternalReview {
		t.Fatalf("unexpected restore report: %#v", restoreReport)
	}
	if restoreReport.BackupArtifactSHA256 != backupSHA || restoreReport.RestoredRecoveryProofSHA256 != originalProofSHA {
		t.Fatalf("restore report lost artifact/proof binding: %#v", restoreReport)
	}
	restoredManifest, err := ledger.Replay(ctx, restored)
	if err != nil {
		t.Fatal(err)
	}
	if restoredManifest.GenesisCommit != originalManifest.GenesisCommit || restoredManifest.ActorPolicyRootContentSHA256 != originalManifest.ActorPolicyRootContentSHA256 || restoredManifest.LedgerCommit != originalManifest.LedgerCommit || restoredManifest.GovernedRecordsSHA256 != originalManifest.GovernedRecordsSHA256 || restoredManifest.ReplaySHA256 != originalManifest.ReplaySHA256 {
		t.Fatalf("restored authoritative identity differs: original=%#v restored=%#v", originalManifest, restoredManifest)
	}
	if AuthorityWritesEnabled() {
		t.Fatal("consolidated E2E harness enabled authority writes")
	}

	acceptanceEvidence := map[string]any{
		"schema_version":                       "threadkeeper.core-v1-e2e-acceptance.v1",
		"genesis_commit":                       genesisEvidence.GenesisCommit,
		"genesis_content_sha256":               genesisEvidence.GenesisContentSHA256,
		"actor_policy_version":                 policySnapshot.Policy.Version,
		"actor_policy_root_content_sha256":     genesisEvidence.ActorPolicyContentSHA256,
		"actor_policy_current_content_sha256":  restartedPolicy.PolicyContentSHA,
		"authenticated_actor_id":               principal.ActorID,
		"authenticated_key_id":                 principal.KeyID,
		"request_context":                      requestContext,
		"accepted_event_id":                    accepted.EventID,
		"accepted_idempotency_key":             accepted.IdempotencyKey,
		"accepted_content_sha256":              accepted.ContentSHA256,
		"accepted_commit":                      h1,
		"restart_recovery_proof_sha256":        restartedProofSHA,
		"restart_retry_status":                 retry.Status,
		"competing_write_disposition":          "stale_state_no_rebase",
		"same_key_conflict_disposition":        "idempotency_conflict",
		"backup_artifact_sha256":               backupSHA,
		"pre_restore_recovery_proof_sha256":    originalProofSHA,
		"restored_recovery_proof_sha256":       restoreReport.RestoredRecoveryProofSHA256,
		"restored_core_equivalence":             restoreReport.CoreEquivalencePassed,
		"operational_independence_status":       restoreReport.OperationalIndependenceStatus,
		"authority_writes_enabled":              AuthorityWritesEnabled(),
	}
	rawEvidence, err := json.Marshal(acceptanceEvidence)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("CORE_V1_E2E_ACCEPTANCE %s", rawEvidence)
}
