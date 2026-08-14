package restoreproof

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
	contractschemas "github.com/armpitpete/threadkeeper-core/schemas"
)

const restoreTestPolicy = "authority-policy:restore-test:v1"

func TestVerifyRestoredLedgerProvesCoreEquivalenceButNotOperationalIndependence(t *testing.T) {
	originalPath, originalProof := newRestoreTestLedger(t, "ledger:restore-test")
	restoredPath := cloneRestoreLedger(t, originalPath)
	r, err := gitledger.New(restoredPath, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	provenance := provenanceForProof(t, originalProof)

	report, err := Verify(context.Background(), r, originalProof, provenance)
	if err != nil {
		t.Fatal(err)
	}
	if !report.CoreEquivalencePassed {
		t.Fatal("exact restored ledger did not pass Core equivalence")
	}
	if report.OperationalIndependenceStatus != OperationalIndependenceRequiresExternalReview {
		t.Fatalf("operational independence status = %q", report.OperationalIndependenceStatus)
	}
	if report.RestoredStoragePath == "" || report.ProvenanceContentSHA256 == "" || report.OriginalRecoveryProofSHA256 != report.RestoredRecoveryProofSHA256 {
		t.Fatalf("incomplete restore report: %#v", report)
	}
}

func TestVerifyRejectsAlteredAuthorityState(t *testing.T) {
	originalPath, originalProof := newRestoreTestLedger(t, "ledger:restore-test")
	restoredPath := cloneRestoreLedger(t, originalPath)
	r, err := gitledger.New(restoredPath, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()

	cases := []struct {
		name   string
		mutate func(*ledger.RecoveryProof)
	}{
		{"genesis", func(p *ledger.RecoveryProof) { p.GenesisContentSHA256 = strings.Repeat("1", 64) }},
		{"actor-policy", func(p *ledger.RecoveryProof) { p.ActorPolicyRootContentSHA256 = strings.Repeat("2", 64) }},
		{"head", func(p *ledger.RecoveryProof) { p.LedgerCommit = "1111111111111111111111111111111111111111" }},
		{"projection", func(p *ledger.RecoveryProof) { p.GovernedRecordsSHA256 = strings.Repeat("3", 64) }},
		{"replay", func(p *ledger.RecoveryProof) { p.ReplaySHA256 = strings.Repeat("4", 64) }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			altered := originalProof
			tc.mutate(&altered)
			provenance := provenanceForProof(t, altered)
			report, err := Verify(context.Background(), r, altered, provenance)
			if err == nil || !strings.Contains(err.Error(), "RESTORE_AUTHORITY_MISMATCH") {
				t.Fatalf("altered %s proof accepted: report=%#v err=%v", tc.name, report, err)
			}
			if report == nil || report.CoreEquivalencePassed {
				t.Fatalf("altered %s produced passing report: %#v", tc.name, report)
			}
			if report.OperationalIndependenceStatus != OperationalIndependenceRequiresExternalReview {
				t.Fatalf("altered %s changed independence status: %#v", tc.name, report)
			}
		})
	}
}

func TestVerifyRejectsProvenanceForDifferentOriginalProof(t *testing.T) {
	originalPath, originalProof := newRestoreTestLedger(t, "ledger:restore-test")
	restoredPath := cloneRestoreLedger(t, originalPath)
	r, err := gitledger.New(restoredPath, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	provenance := provenanceForProof(t, originalProof)
	provenance.OriginalRecoveryProofSHA256 = strings.Repeat("f", 64)
	provenance = rebindProvenance(t, provenance)
	if _, err := Verify(context.Background(), r, originalProof, provenance); err == nil || !strings.Contains(err.Error(), "RESTORE_PROVENANCE_MISMATCH") {
		t.Fatalf("wrong original-proof binding accepted: %v", err)
	}
}

func TestVerifyRevalidatesDirectlyConstructedProvenance(t *testing.T) {
	originalPath, originalProof := newRestoreTestLedger(t, "ledger:restore-test")
	restoredPath := cloneRestoreLedger(t, originalPath)
	r, err := gitledger.New(restoredPath, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	provenance := provenanceForProof(t, originalProof)
	provenance.SecondaryAuthorityDomainID = provenance.PrimaryAuthorityDomainID
	provenance = rebindProvenance(t, provenance)
	if _, err := Verify(context.Background(), r, originalProof, provenance); err == nil || !strings.Contains(err.Error(), "must differ") {
		t.Fatalf("directly constructed contradictory provenance accepted: %v", err)
	}
}

func TestDecodeRecoveryProofRejectsUnknownAndMissingFields(t *testing.T) {
	_, proof := newRestoreTestLedger(t, "ledger:restore-test")
	raw, err := json.Marshal(proof)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeRecoveryProof(raw); err != nil {
		t.Fatalf("valid recovery proof rejected: %v", err)
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	value["operational_independence_verified"] = true
	unknown, _ := json.Marshal(value)
	if _, err := DecodeRecoveryProof(unknown); err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown recovery proof field accepted: %v", err)
	}
	delete(value, "operational_independence_verified")
	delete(value, "ledger_id")
	missing, _ := json.Marshal(value)
	if _, err := DecodeRecoveryProof(missing); err == nil || !strings.Contains(err.Error(), "required field") {
		t.Fatalf("missing recovery proof field accepted: %v", err)
	}
}

func newRestoreTestLedger(t *testing.T, ledgerID string) (string, ledger.RecoveryProof) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for restore proof integration tests")
	}
	initialSchemas := []string{contracts.ExclusiveRecordEventSchemaV1, contracts.ReducerBindingSchemaV1}
	sort.Strings(initialSchemas)
	genesisRaw, err := json.Marshal(map[string]any{
		"project_id":               "project:restore-test",
		"ledger_id":                ledgerID,
		"created_at":               "2026-08-14T14:00:00Z",
		"initial_authority_policy": restoreTestPolicy,
		"initial_schemas":          initialSchemas,
		"initial_authorities":      []string{"owner:restore-test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	genesisBytes, _, err := digest.Complete(genesisRaw)
	if err != nil {
		t.Fatal(err)
	}

	seedBytes := make([]byte, ed25519.SeedSize)
	for i := range seedBytes {
		seedBytes[i] = 11
	}
	publicKey := ed25519.NewKeyFromSeed(seedBytes).Public().(ed25519.PublicKey)
	policyRaw, err := json.Marshal(map[string]any{
		"ledger_id":                  ledgerID,
		"authority_policy_version":   restoreTestPolicy,
		"max_proof_lifetime_seconds": int64(300),
		"keys": []map[string]any{{
			"actor_id":   "owner:restore-test",
			"key_id":     "key:restore-test:v1",
			"public_key": base64.RawStdEncoding.EncodeToString(publicKey),
			"revoked":    false,
		}},
		"grants": []map[string]any{{
			"actor_id": "owner:restore-test",
			"action":   actorauth.ActionOperate,
			"target":   "target:restore-test",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	policyBytes, _, err := digest.Complete(policyRaw)
	if err != nil {
		t.Fatal(err)
	}

	bindingRaw, err := json.Marshal(map[string]any{
		"schema_version":           contracts.ReducerBindingSchemaV1,
		"binding_id":               "binding:restore-actor-policy:v1",
		"record_kind":              actorauth.PolicyRecordKind,
		"state_model":              reducer.ModelExclusiveV1,
		"event_schema":             contracts.ExclusiveRecordEventSchemaV1,
		"authority_policy_version": restoreTestPolicy,
	})
	if err != nil {
		t.Fatal(err)
	}
	bindingBytes, _, err := digest.Complete(bindingRaw)
	if err != nil {
		t.Fatal(err)
	}

	seed := map[string][]byte{
		actorauth.LedgerPolicyPath:                                      policyBytes,
		"config/schemas/exclusive-record-event-v1.json":               contractschemas.ExclusiveGovernedRecordEventV1,
		"config/schemas/reducer-binding-v1.json":                      contractschemas.ReducerBindingV1,
		"config/authority/reducer-bindings/actor-auth-policy-v1.json": bindingBytes,
	}
	target := filepath.Join(t.TempDir(), "original.git")
	if _, err := ledger.InitializeFreshGenesis(context.Background(), target, gitledger.DefaultRef, genesisBytes, seed); err != nil {
		t.Fatal(err)
	}
	r, err := gitledger.New(target, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := ledger.ProveRecovery(context.Background(), r)
	r.Close()
	if err != nil {
		t.Fatal(err)
	}
	return target, *proof
}

func cloneRestoreLedger(t *testing.T, source string) string {
	t.Helper()
	target := filepath.Join(t.TempDir(), "restored.git")
	cmd := exec.Command("git", "clone", "--bare", "--no-hardlinks", source, target)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("clone restored ledger: %v\n%s", err, out)
	}
	return target
}

func provenanceForProof(t *testing.T, proof ledger.RecoveryProof) Provenance {
	t.Helper()
	proofSHA, err := RecoveryProofSHA256(proof)
	if err != nil {
		t.Fatal(err)
	}
	value := provenanceMap()
	value["original_recovery_proof_sha256"] = proofSHA
	raw := completeProvenance(t, value)
	p, err := DecodeProvenance(raw)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func rebindProvenance(t *testing.T, p Provenance) Provenance {
	t.Helper()
	p.ContentSHA256 = ""
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	_, contentSHA, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	p.ContentSHA256 = contentSHA
	return p
}
