package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/ledger"
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
