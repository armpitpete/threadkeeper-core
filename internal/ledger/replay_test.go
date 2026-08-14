package ledger

import (
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
)

const testSchemaID = "urn:threadkeeper:test:event:v1"

var testSchema = []byte(`{
  "$schema":"https://json-schema.org/draft/2020-12/schema",
  "$id":"urn:threadkeeper:test:event:v1",
  "type":"object",
  "required":["schema_version","event_id","event_type","targets","prior_state","resulting_state","content_sha256"],
  "properties":{
    "schema_version":{"const":"urn:threadkeeper:test:event:v1"},
    "event_id":{"type":"string","minLength":1},
    "event_type":{"type":"string","minLength":1},
    "targets":{"type":"array","items":{"type":"string"}},
    "prior_state":{"type":"object"},
    "resulting_state":{"type":"object"},
    "content_sha256":{"type":"string","pattern":"^[0-9a-f]{64}$"}
  },
  "additionalProperties":false
}`)

func TestReplayValidBareLedgerDeterministically(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "event-1", "decision.accepted", "target-a")
	commitAll(t, work, "accept event 1")
	writeEvent(t, work, "events/governance/002.json", "event-2", "governance.recorded", "target-b")
	commitAll(t, work, "accept event 2")

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
	if !first.BareRepository {
		t.Fatal("expected bare ledger repository")
	}
	if first.GenesisCommit == "" || first.GenesisRoot.ProjectID != "project:test" || first.GenesisRoot.LedgerID != "ledger:test" {
		t.Fatalf("replay did not expose test Genesis identity: %#v", first.GenesisRoot)
	}
	if first.ActorPolicyVersion != testPolicyV1 || first.ActorPolicyRootContentSHA256 == "" {
		t.Fatalf("replay did not expose root actor-policy identity: %#v", first)
	}
	if first.EventCount != 2 || len(first.Events) != 2 {
		t.Fatalf("event count = %d, want 2", first.EventCount)
	}
	if first.Events[0].EventID != "event-1" || first.Events[1].EventID != "event-2" {
		t.Fatalf("unexpected replay order: %#v", first.Events)
	}
	if first.ReplaySHA256 == "" || first.ReplaySHA256 != second.ReplaySHA256 {
		t.Fatalf("replay digest is not deterministic: %q / %q", first.ReplaySHA256, second.ReplaySHA256)
	}
}

func TestReplayRejectsMutationOfAcceptedEventPath(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	path := "events/decisions/001.json"
	writeEvent(t, work, path, "event-1", "decision.accepted", "target-a")
	commitAll(t, work, "accept event 1")
	writeEvent(t, work, path, "event-1-rewritten", "decision.accepted", "target-a")
	commitAll(t, work, "illegally rewrite event 1")

	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Replay(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), "event files are immutable") {
		t.Fatalf("expected immutable-event failure, got %v", err)
	}
}

func TestReplayRejectsDigestMismatch(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	path := filepath.Join(work, "events/decisions/001.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	bad := []byte(`{"content_sha256":"0000000000000000000000000000000000000000000000000000000000000000","event_id":"event-1","event_type":"decision.accepted","prior_state":{},"resulting_state":{"ok":true},"schema_version":"urn:threadkeeper:test:event:v1","targets":["target-a"]}`)
	if err := os.WriteFile(path, bad, 0o644); err != nil {
		t.Fatal(err)
	}
	commitAll(t, work, "accept corrupted event")

	bare := cloneBare(t, work)
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Replay(context.Background(), r)
	if err == nil || !strings.Contains(err.Error(), "DIGEST_MISMATCH") {
		t.Fatalf("expected digest mismatch, got %v", err)
	}
}

func TestFSCKDetectsMissingReachableObject(t *testing.T) {
	work := newWorkRepo(t)
	writeSchema(t, work)
	writeEvent(t, work, "events/decisions/001.json", "event-1", "decision.accepted", "target-a")
	commitAll(t, work, "accept event")
	bare := cloneBare(t, work)

	removed := false
	objects := filepath.Join(bare, "objects")
	err := filepath.WalkDir(objects, func(path string, d os.DirEntry, err error) error {
		if err != nil || removed || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(objects, path)
		if err != nil {
			return err
		}
		parts := strings.Split(filepath.ToSlash(rel), "/")
		if len(parts) == 2 && len(parts[0]) == 2 && len(parts[1]) == 38 {
			if err := os.Remove(path); err != nil {
				return err
			}
			removed = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !removed {
		t.Skip("bare clone contains no loose objects to corrupt")
	}
	r, err := gitledger.New(bare, gitledger.DefaultRef)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.FSCK(context.Background()); err == nil {
		t.Fatal("expected fsck failure after deleting reachable object")
	}
}

func newWorkRepo(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is required for ledger integration tests")
	}
	dir := filepath.Join(t.TempDir(), "work")
	runGit(t, "", "init", "-b", "main", dir)
	runGit(t, dir, "config", "user.name", "Threadkeeper Test")
	runGit(t, dir, "config", "user.email", "threadkeeper-test@example.invalid")
	writeTestGenesis(t, dir, nil)
	commitAll(t, dir, "create test Genesis")
	return dir
}

func writeTestGenesis(t *testing.T, work string, initialSchemas []string) []byte {
	t.Helper()
	if initialSchemas == nil {
		initialSchemas = []string{}
	}
	raw, err := json.Marshal(map[string]any{
		"project_id":               "project:test",
		"ledger_id":                "ledger:test",
		"created_at":               "2026-08-12T16:00:00Z",
		"initial_authority_policy": testPolicyV1,
		"initial_schemas":          initialSchemas,
		"initial_authorities":      []string{"owner:test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(work, filepath.FromSlash(genesis.LedgerPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, completed, 0o644); err != nil {
		t.Fatal(err)
	}
	writeTestActorPolicy(t, work)
	return completed
}

func makeTestActorPolicy(t *testing.T, ledgerID, policyVersion string) []byte {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = 1
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	publicKey := privateKey.Public().(ed25519.PublicKey)
	raw, err := json.Marshal(map[string]any{
		"ledger_id":                 ledgerID,
		"authority_policy_version": policyVersion,
		"max_proof_lifetime_seconds": int64(300),
		"keys": []map[string]any{{
			"actor_id":   "owner:test",
			"key_id":     "key:test:v1",
			"public_key": base64.RawStdEncoding.EncodeToString(publicKey),
			"revoked":    false,
		}},
		"grants": []map[string]any{{
			"actor_id": "owner:test",
			"action":   actorauth.ActionOperate,
			"target":   "target:test",
		}},
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

func writeTestActorPolicy(t *testing.T, work string) []byte {
	t.Helper()
	completed := makeTestActorPolicy(t, "ledger:test", testPolicyV1)
	path := filepath.Join(work, filepath.FromSlash(actorauth.LedgerPolicyPath))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, completed, 0o644); err != nil {
		t.Fatal(err)
	}
	return completed
}

func writeSchema(t *testing.T, work string) {
	t.Helper()
	path := filepath.Join(work, "config/schemas/event/test-v1.json")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, testSchema, 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeEvent(t *testing.T, work, rel, id, eventType, target string) {
	t.Helper()
	candidate := []byte(`{"schema_version":"` + testSchemaID + `","event_id":"` + id + `","event_type":"` + eventType + `","targets":["` + target + `"],"prior_state":{},"resulting_state":{"accepted":true}}`)
	completed, _, err := digest.Complete(candidate)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(work, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, completed, 0o644); err != nil {
		t.Fatal(err)
	}
}

func commitAll(t *testing.T, work, message string) {
	t.Helper()
	runGit(t, work, "add", "--all")
	runGit(t, work, "commit", "-m", message)
}

func cloneBare(t *testing.T, work string) string {
	t.Helper()
	bare := filepath.Join(t.TempDir(), "ledger.git")
	runGit(t, "", "clone", "--bare", "--no-hardlinks", work, bare)
	return bare
}

func runGit(t *testing.T, dir string, args ...string) string {
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
