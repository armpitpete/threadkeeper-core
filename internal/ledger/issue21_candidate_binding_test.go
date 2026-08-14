package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestIssue21RejectsSameQuarantineAtDifferentSafePath(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "issue21-path", "idem-issue21-path", json.RawMessage(`{"enabled":true}`))
	prepared, accepted, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/issue21-path-a.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared == nil || accepted != nil {
		t.Fatalf("unexpected prepare result candidate=%#v accepted=%#v", prepared, accepted)
	}

	alternate, err := r.PrepareEventCommit(
		context.Background(),
		head,
		"events/governance/issue21-path-b.json",
		event,
		"issue21-path",
	)
	if err != nil {
		t.Fatal(err)
	}
	forged := *prepared
	forged.CandidateCommit = alternate.Commit
	forged.EventPath = alternate.EventPath

	_, err = AcceptWriteCandidate(context.Background(), r, forged)
	if err == nil || !errors.Is(err, ErrCandidateInvalid) || !strings.Contains(err.Error(), "not bound to exact prepared candidate identity") {
		t.Fatalf("expected exact prepared-path binding rejection, got %v", err)
	}
	got, headErr := r.Head(context.Background())
	if headErr != nil {
		t.Fatal(headErr)
	}
	if got != head {
		t.Fatalf("path-substituted candidate moved authority: got %s want %s", got, head)
	}
}

func TestIssue21RejectsSamePathAndBytesWithDifferentCommitIdentity(t *testing.T) {
	r, head := candidateTestReader(t)
	event := makeCreateCandidateEvent(t, head, "issue21-commit", "idem-issue21-commit", json.RawMessage(`{"enabled":true}`))
	prepared, accepted, err := PrepareWriteCandidate(context.Background(), r, CandidateRequest{
		ExpectedHead: head,
		EventPath:    "events/governance/issue21-commit.json",
		Event:        event,
	})
	if err != nil {
		t.Fatal(err)
	}
	if prepared == nil || accepted != nil {
		t.Fatalf("unexpected prepare result candidate=%#v accepted=%#v", prepared, accepted)
	}

	tree := runGitInDirWithEnv(t, r.GitDir(), nil, nil, "rev-parse", prepared.CandidateCommit+"^{tree}")
	alternateCommit := runGitInDirWithEnv(t, r.GitDir(), []string{
		"GIT_AUTHOR_NAME=Hostile Reviewer",
		"GIT_AUTHOR_EMAIL=hostile-review@example.invalid",
		"GIT_AUTHOR_DATE=2001-02-03T04:05:06Z",
		"GIT_COMMITTER_NAME=Hostile Reviewer",
		"GIT_COMMITTER_EMAIL=hostile-review@example.invalid",
		"GIT_COMMITTER_DATE=2001-02-03T04:05:06Z",
	}, []byte("same tree, attacker-selected commit metadata\n"), "commit-tree", tree, "-p", head)
	if alternateCommit == prepared.CandidateCommit {
		t.Fatal("hostile commit unexpectedly reproduced prepared commit identity")
	}

	forged := *prepared
	forged.CandidateCommit = alternateCommit
	_, err = AcceptWriteCandidate(context.Background(), r, forged)
	if err == nil || !errors.Is(err, ErrCandidateInvalid) || !strings.Contains(err.Error(), "not bound to exact prepared candidate identity") {
		t.Fatalf("expected exact prepared-commit binding rejection, got %v", err)
	}
	got, headErr := r.Head(context.Background())
	if headErr != nil {
		t.Fatal(headErr)
	}
	if got != head {
		t.Fatalf("commit-substituted candidate moved authority: got %s want %s", got, head)
	}
}
