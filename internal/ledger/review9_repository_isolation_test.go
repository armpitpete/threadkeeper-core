package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

func TestReplayRejectsGitCommonDirIndirection(t *testing.T) {
	r, _ := candidateTestReader(t)
	facade := filepath.Join(t.TempDir(), "facade.git")
	if err := os.MkdirAll(facade, 0o755); err != nil {
		t.Fatal(err)
	}
	rel, err := filepath.Rel(facade, r.GitDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(facade, "commondir"), []byte(rel+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(facade, "HEAD"), []byte("ref: refs/heads/main\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(facade, "config"), []byte("[core]\n\trepositoryformatversion = 0\n\tbare = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	facadeReader, err := gitledger.New(facade, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Replay(context.Background(), facadeReader)
	if err == nil || !strings.Contains(err.Error(), "common-dir") {
		t.Fatalf("expected common-dir rejection, got %v", err)
	}
}

func TestReplayRejectsPromisorPartialCloneConfiguration(t *testing.T) {
	r, _ := candidateTestReader(t)
	configPath := filepath.Join(r.GitDir(), "config")
	f, err := os.OpenFile(configPath, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("\n[remote \"origin\"]\n\turl = file:///tmp/threadkeeper-promisor-fixture\n\tpromisor = true\n\tpartialclonefilter = blob:none\n"); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = Replay(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), "promisor/partial-clone") {
		t.Fatalf("expected promisor/partial-clone rejection, got %v", err)
	}
}

func TestReplayRejectsSymlinkedAuthorityStores(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink fixture requires Unix-like symlink semantics")
	}
	for _, rel := range []string{"objects", "refs"} {
		t.Run(rel, func(t *testing.T) {
			r, _ := candidateTestReader(t)
			original := filepath.Join(r.GitDir(), rel)
			external := filepath.Join(t.TempDir(), rel)
			if err := os.Rename(original, external); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(external, original); err != nil {
				t.Fatal(err)
			}

			_, err := Replay(context.Background(), r)
			if err == nil || !strings.Contains(err.Error(), "symlinked Git repository path") {
				t.Fatalf("expected symlinked %s rejection, got %v", rel, err)
			}
		})
	}
}

func TestAcceptRejectsNonRegularEventTreeModes(t *testing.T) {
	for _, mode := range []string{"100755", "120000"} {
		t.Run(mode, func(t *testing.T) {
			r, head := candidateTestReader(t)
			eventID := "candidate-mode-" + mode
			idem := "idem-mode-" + mode
			eventPath := "events/governance/candidate-mode-" + mode + ".json"
			event := makeCreateCandidateEvent(t, head, eventID, idem, json.RawMessage(`{"enabled":true}`))

			eventFile := filepath.Join(t.TempDir(), "event.json")
			if err := os.WriteFile(eventFile, event, 0o644); err != nil {
				t.Fatal(err)
			}
			blob := runGitInDirWithEnv(t, r.GitDir(), nil, nil, "hash-object", "-w", eventFile)
			index := filepath.Join(t.TempDir(), "index")
			env := []string{"GIT_INDEX_FILE=" + index}
			runGitInDirWithEnv(t, r.GitDir(), env, nil, "read-tree", head)
			runGitInDirWithEnv(t, r.GitDir(), env, nil, "update-index", "--add", "--cacheinfo", mode, blob, eventPath)
			tree := runGitInDirWithEnv(t, r.GitDir(), env, nil, "write-tree")
			commit := runGitInDirWithEnv(t, r.GitDir(), []string{
				"GIT_AUTHOR_NAME=Threadkeeper Test",
				"GIT_AUTHOR_EMAIL=threadkeeper-test@example.invalid",
				"GIT_COMMITTER_NAME=Threadkeeper Test",
				"GIT_COMMITTER_EMAIL=threadkeeper-test@example.invalid",
			}, []byte("forged non-regular event mode\n"), "commit-tree", tree, "-p", head)

			contentSHA, _, err := digest.Compute(event)
			if err != nil {
				t.Fatal(err)
			}
			forged := WriteCandidate{
				ExpectedHead:    head,
				CandidateCommit: commit,
				EventPath:       eventPath,
				EventID:         eventID,
				IdempotencyKey:  idem,
				ContentSHA256:   contentSHA,
			}
			_, err = AcceptWriteCandidate(context.Background(), r, forged)
			if err == nil || !strings.Contains(err.Error(), "100644 regular blob") {
				t.Fatalf("expected non-regular event mode rejection, got %v", err)
			}
			got, headErr := r.Head(context.Background())
			if headErr != nil {
				t.Fatal(headErr)
			}
			if got != head {
				t.Fatalf("non-regular event mode changed authority: got %s want %s", got, head)
			}
		})
	}
}

func runGitInDirWithEnv(t *testing.T, gitDir string, extraEnv []string, stdin []byte, args ...string) string {
	t.Helper()
	cmdArgs := append([]string{"--git-dir=" + gitDir}, args...)
	cmd := exec.Command("git", cmdArgs...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0")
	cmd.Env = append(cmd.Env, extraEnv...)
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v failed: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}
