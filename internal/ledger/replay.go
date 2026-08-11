package ledger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/schema"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

type ReplayEntry struct {
	Sequence        int    `json:"sequence"`
	AcceptedCommit  string `json:"accepted_commit"`
	Path            string `json:"path"`
	EventID         string `json:"event_id"`
	EventType       string `json:"event_type"`
	SchemaVersion   string `json:"schema_version"`
	ContentSHA256   string `json:"content_sha256"`
	TargetCount     int    `json:"target_count"`
}

type ReplayManifest struct {
	LedgerCommit       string        `json:"ledger_commit"`
	AuthoritativeRef   string        `json:"authoritative_ref"`
	GitObjectFormat    string        `json:"git_object_format"`
	BareRepository     bool          `json:"bare_repository"`
	HistoryCommitCount int           `json:"history_commit_count"`
	EventCount         int           `json:"event_count"`
	ReplaySHA256       string        `json:"replay_sha256"`
	Events             []ReplayEntry `json:"events"`
}

// Replay validates the authoritative Git history and builds a deterministic
// audit projection. It does not apply event-type-specific current-state
// semantics; that requires separately accepted event semantics.
func Replay(ctx context.Context, r *gitledger.Reader) (*ReplayManifest, error) {
	head, err := r.Head(ctx)
	if err != nil {
		return nil, err
	}
	if err := r.FSCK(ctx); err != nil {
		return nil, err
	}
	objectFormat, err := r.ObjectFormat(ctx)
	if err != nil {
		return nil, err
	}
	bare, err := r.IsBare(ctx)
	if err != nil {
		return nil, err
	}
	history, err := r.History(ctx, head)
	if err != nil {
		return nil, err
	}

	manifest := &ReplayManifest{
		LedgerCommit:       head,
		AuthoritativeRef:   r.Ref(),
		GitObjectFormat:    objectFormat,
		BareRepository:     bare,
		HistoryCommitCount: len(history),
		Events:             []ReplayEntry{},
	}

	for _, commit := range history {
		additions, err := r.EventAdditions(ctx, commit.ID)
		if err != nil {
			return nil, err
		}
		if len(additions) == 0 {
			continue
		}
		registry, err := loadSchemasAt(ctx, r, commit.ID)
		if err != nil {
			return nil, fmt.Errorf("schema snapshot at %s: %w", commit.ID, err)
		}
		for _, addition := range additions {
			entry, err := validateEvent(ctx, r, registry, addition)
			if err != nil {
				return nil, err
			}
			entry.Sequence = len(manifest.Events) + 1
			manifest.Events = append(manifest.Events, entry)
		}
	}
	manifest.EventCount = len(manifest.Events)
	replayDigest, err := replayDigest(manifest)
	if err != nil {
		return nil, err
	}
	manifest.ReplaySHA256 = replayDigest
	return manifest, nil
}

func loadSchemasAt(ctx context.Context, r *gitledger.Reader, commit string) (*schema.Registry, error) {
	paths, err := r.ListJSON(ctx, commit, "config/schemas")
	if err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("UNKNOWN_SCHEMA: no accepted schemas at commit %s", commit)
	}
	registry := schema.NewRegistry()
	for _, path := range paths {
		raw, err := r.ReadFile(ctx, commit, path)
		if err != nil {
			return nil, err
		}
		value, err := strictjson.Decode(raw)
		if err != nil {
			return nil, fmt.Errorf("schema %s: %w", path, err)
		}
		obj, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("SCHEMA_INVALID: schema %s root must be object", path)
		}
		id, _ := obj["$id"].(string)
		if id == "" {
			return nil, fmt.Errorf("SCHEMA_INVALID: schema %s has no $id", path)
		}
		if err := registry.Add(id, raw); err != nil {
			return nil, fmt.Errorf("schema %s: %w", path, err)
		}
	}
	return registry, nil
}

func validateEvent(ctx context.Context, r *gitledger.Reader, registry *schema.Registry, addition gitledger.EventAddition) (ReplayEntry, error) {
	raw, err := r.ReadFile(ctx, addition.Commit, addition.Path)
	if err != nil {
		return ReplayEntry{}, err
	}
	if err := strictjson.Validate(raw); err != nil {
		return ReplayEntry{}, fmt.Errorf("event %s at %s: %w", addition.Path, addition.Commit, err)
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return ReplayEntry{}, fmt.Errorf("event %s at %s canonicalization: %w", addition.Path, addition.Commit, err)
	}
	if !bytes.Equal(raw, canonical) {
		return ReplayEntry{}, fmt.Errorf("INTEGRITY_FAILURE: event %s at %s is not stored as RFC 8785 canonical JSON", addition.Path, addition.Commit)
	}
	if err := digest.Verify(raw); err != nil {
		return ReplayEntry{}, fmt.Errorf("event %s at %s: %w", addition.Path, addition.Commit, err)
	}
	value, err := strictjson.Decode(raw)
	if err != nil {
		return ReplayEntry{}, err
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return ReplayEntry{}, fmt.Errorf("event %s root must be object", addition.Path)
	}
	schemaVersion, _ := obj["schema_version"].(string)
	if schemaVersion == "" {
		return ReplayEntry{}, fmt.Errorf("UNKNOWN_SCHEMA: event %s has no schema_version", addition.Path)
	}
	if err := registry.Validate(schemaVersion, raw); err != nil {
		return ReplayEntry{}, fmt.Errorf("event %s at %s: %w", addition.Path, addition.Commit, err)
	}
	eventID, _ := obj["event_id"].(string)
	eventType, _ := obj["event_type"].(string)
	contentSHA, _ := obj[digest.Field].(string)
	if eventID == "" || eventType == "" || contentSHA == "" {
		return ReplayEntry{}, fmt.Errorf("INTEGRITY_FAILURE: event %s lacks replay identity fields", addition.Path)
	}
	targets, ok := obj["targets"].([]any)
	if !ok {
		return ReplayEntry{}, fmt.Errorf("INTEGRITY_FAILURE: event %s targets must be an array", addition.Path)
	}
	return ReplayEntry{
		AcceptedCommit: addition.Commit,
		Path:           addition.Path,
		EventID:        eventID,
		EventType:      eventType,
		SchemaVersion:  schemaVersion,
		ContentSHA256:  contentSHA,
		TargetCount:    len(targets),
	}, nil
}

func replayDigest(manifest *ReplayManifest) (string, error) {
	payload := struct {
		LedgerCommit       string        `json:"ledger_commit"`
		AuthoritativeRef   string        `json:"authoritative_ref"`
		GitObjectFormat    string        `json:"git_object_format"`
		HistoryCommitCount int           `json:"history_commit_count"`
		Events             []ReplayEntry `json:"events"`
	}{
		LedgerCommit:       manifest.LedgerCommit,
		AuthoritativeRef:   manifest.AuthoritativeRef,
		GitObjectFormat:    manifest.GitObjectFormat,
		HistoryCommitCount: manifest.HistoryCommitCount,
		Events:             manifest.Events,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal replay projection: %w", err)
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return "", fmt.Errorf("canonicalize replay projection: %w", err)
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), nil
}
