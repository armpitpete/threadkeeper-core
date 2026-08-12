package ledger

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestOpenReaderRejectsOrdinaryRepositoryRootReplacement(t *testing.T) {
	r, _ := candidateTestReader(t)
	root := r.GitDir()
	original := root + "-original"
	if err := os.Rename(root, original); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	// Make the replacement look superficially like a normal bare repository.
	// The identity check must reject it before trusting any of these children.
	if err := os.WriteFile(filepath.Join(root, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "config"), []byte("[core]\n\trepositoryformatversion = 0\n\tbare = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "objects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "refs", "heads"), 0o755); err != nil {
		t.Fatal(err)
	}

	_, err := r.Head(context.Background())
	if err == nil || !strings.Contains(err.Error(), "filesystem identity changed") {
		t.Fatalf("expected ordinary same-path repository replacement rejection, got %v", err)
	}
}

func TestSymbolicAuthoritativeRefCannotRedirectCAS(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "candidate-symref", "idem-symref", json.RawMessage(`{"enabled":true}`))
	candidate, _, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/candidate-symref.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}

	// Create a different direct ref at H0, then turn the configured authority
	// ref into an ordinary Git symbolic-ref file that points at it. This is not
	// an OS symlink and was the attack class missed by earlier filesystem checks.
	runGitInDirWithEnv(t, r.GitDir(), nil, nil, "update-ref", "refs/heads/redirect-target", head)
	mainPath := filepath.Join(r.GitDir(), "refs", "heads", "main")
	if err := os.WriteFile(mainPath, []byte("ref: refs/heads/redirect-target\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err = AcceptWriteCandidate(context.Background(), r, *candidate)
	if err == nil || !strings.Contains(err.Error(), "symbolic authoritative refs are forbidden") {
		t.Fatalf("expected symbolic authoritative ref rejection, got %v", err)
	}

	// The target ref must remain H0: the failed authority attempt must not have
	// followed the symbolic ref and mutated the redirect target.
	gotTarget := runGitInDirWithEnv(t, r.GitDir(), nil, nil, "rev-parse", "--verify", "refs/heads/redirect-target^{commit}")
	if gotTarget != head {
		t.Fatalf("symbolic ref redirected authority movement: target=%s want=%s", gotTarget, head)
	}

	// A newly opened Reader must reject the same static symbolic authority ref.
	if _, err := gitledger.New(r.GitDir(), gitledger.DefaultRef); err == nil || !strings.Contains(err.Error(), "symbolic authoritative refs are forbidden") {
		t.Fatalf("new Reader accepted symbolic authoritative ref: %v", err)
	}
}
