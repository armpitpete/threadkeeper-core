package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/ledger"
	"github.com/armpitpete/threadkeeper-core/internal/restoreproof"
)

func TestRecoveryCompareCLIRejectsIncompleteProofs(t *testing.T) {
	a := writeCLIProof(t, []byte(`{}`))
	b := writeCLIProof(t, []byte(`{"a":1}`))
	out, err := runCLIHelper(t, "recovery-compare", a, b)
	if err == nil {
		t.Fatalf("incomplete proofs unexpectedly succeeded: %s", out)
	}
	if !strings.Contains(out, "RECOVERY_PROOF_INVALID") {
		t.Fatalf("unexpected incomplete-proof error: %s", out)
	}

	a = writeCLIProof(t, []byte(`{}`))
	b = writeCLIProof(t, []byte(`{}`))
	out, err = runCLIHelper(t, "recovery-compare", a, b)
	if err == nil || !strings.Contains(out, "RECOVERY_PROOF_INVALID") {
		t.Fatalf("two incomplete proofs unexpectedly compared: err=%v out=%s", err, out)
	}
}

func TestRecoveryCompareCLIValidEqualAndMismatch(t *testing.T) {
	proof := cliValidRecoveryProof()
	raw, err := json.Marshal(proof)
	if err != nil {
		t.Fatal(err)
	}
	a := writeCLIProof(t, raw)
	b := writeCLIProof(t, raw)
	out, err := runCLIHelper(t, "recovery-compare", a, b)
	if err != nil {
		t.Fatalf("equal valid proofs failed: %v out=%s", err, out)
	}
	if !strings.Contains(out, `"equivalent": true`) {
		t.Fatalf("equal valid proofs did not report equivalence: %s", out)
	}

	mismatch := proof
	mismatch.LedgerCommit = strings.Repeat("9", 40)
	mismatchRaw, _ := json.Marshal(mismatch)
	c := writeCLIProof(t, mismatchRaw)
	out, err = runCLIHelper(t, "recovery-compare", a, c)
	if err == nil {
		t.Fatalf("mismatched valid proofs unexpectedly succeeded: %s", out)
	}
	if !strings.Contains(out, "RECOVERY_PROOF_MISMATCH") {
		t.Fatalf("unexpected mismatch error: %s", out)
	}
}

func TestDigestCLIEmitsCanonicalProvenanceBytes(t *testing.T) {
	raw, err := json.Marshal(map[string]any{
		"schema_version":                 restoreproof.ProvenanceSchemaV1,
		"primary_authority_domain_id":    "authority:primary",
		"secondary_authority_domain_id":  "authority:secondary",
		"secondary_location_id":          "location:secondary-a",
		"secondary_operator_id":          "operator:secondary-a",
		"backup_set_id":                  "backup:set-001",
		"backup_artifact_id":             "artifact:ledger-001",
		"backup_artifact_sha256":         strings.Repeat("a", 64),
		"original_recovery_proof_sha256": strings.Repeat("b", 64),
		"captured_at":                    "2026-08-14T14:00:00Z",
		"restored_at":                    "2026-08-14T14:30:00Z",
		"external_evidence_refs":         []string{"evidence:provider-receipt", "evidence:restore-log"},
	})
	if err != nil {
		t.Fatal(err)
	}
	input := writeCLIProof(t, raw)
	out, err := runCLIHelper(t, "digest", input)
	if err != nil {
		t.Fatalf("digest failed: %v out=%s", err, out)
	}
	if strings.HasSuffix(out, "\n") || strings.HasSuffix(out, "\r") {
		t.Fatalf("digest output contains trailing whitespace: %q", out)
	}
	if _, err := restoreproof.DecodeProvenance([]byte(out)); err != nil {
		t.Fatalf("digest output is not directly usable as strict provenance: %v", err)
	}
	withNewline := append([]byte(out), '\n')
	if _, err := restoreproof.DecodeProvenance(withNewline); err == nil || !strings.Contains(err.Error(), "canonical JSON") {
		t.Fatalf("strict provenance decoder accepted newline-contaminated bytes: %v", err)
	}
}

func TestThreadkeeperCoreCLIHelperProcess(t *testing.T) {
	if os.Getenv("THREADKEEPER_CORE_TEST_HELPER") != "1" {
		return
	}
	separator := -1
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator < 0 || separator+1 >= len(os.Args) {
		os.Exit(125)
	}
	os.Args = append([]string{"threadkeeper-core"}, os.Args[separator+1:]...)
	main()
	os.Exit(0)
}

func runCLIHelper(t *testing.T, args ...string) (string, error) {
	t.Helper()
	commandArgs := []string{"-test.run=TestThreadkeeperCoreCLIHelperProcess", "--"}
	commandArgs = append(commandArgs, args...)
	cmd := exec.Command(os.Args[0], commandArgs...)
	cmd.Env = append(os.Environ(), "THREADKEEPER_CORE_TEST_HELPER=1")
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func writeCLIProof(t *testing.T, raw []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "proof.json")
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func cliValidRecoveryProof() ledger.RecoveryProof {
	return ledger.RecoveryProof{
		LedgerCommit:                 strings.Repeat("1", 40),
		AuthoritativeRef:             "refs/heads/main",
		GitObjectFormat:              "sha1",
		GenesisCommit:                strings.Repeat("2", 40),
		ProjectID:                    "project:test",
		LedgerID:                     "ledger:test",
		GenesisContentSHA256:         strings.Repeat("3", 64),
		ActorPolicyVersion:           "authority-policy:test:v1",
		ActorPolicyRootContentSHA256: strings.Repeat("4", 64),
		HistoryCommitCount:           1,
		EventCount:                   0,
		ReducerBindingCount:          1,
		GovernedRecordCount:          0,
		GovernedRecordsSHA256:        strings.Repeat("5", 64),
		ReplaySHA256:                 strings.Repeat("6", 64),
	}
}
