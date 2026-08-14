package ledger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/policy"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
	"github.com/armpitpete/threadkeeper-core/internal/schema"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

type ReplayEntry struct {
	Sequence       int    `json:"sequence"`
	AcceptedCommit string `json:"accepted_commit"`
	Path           string `json:"path"`
	EventID        string `json:"event_id"`
	EventType      string `json:"event_type"`
	SchemaVersion  string `json:"schema_version"`
	ContentSHA256  string `json:"content_sha256"`
	TargetCount    int    `json:"target_count"`
}

type ReplayManifest struct {
	LedgerCommit          string             `json:"ledger_commit"`
	AuthoritativeRef      string             `json:"authoritative_ref"`
	GitObjectFormat       string             `json:"git_object_format"`
	BareRepository        bool               `json:"bare_repository"`
	GenesisCommit         string             `json:"genesis_commit"`
	GenesisRoot           genesis.Root       `json:"genesis_root"`
	HistoryCommitCount    int                `json:"history_commit_count"`
	EventCount            int                `json:"event_count"`
	ReducerBindingCount   int                `json:"reducer_binding_count"`
	GovernedRecordCount   int                `json:"governed_record_count"`
	GovernedRecordsSHA256 string             `json:"governed_records_sha256"`
	GovernedRecords       reducer.Projection `json:"governed_records"`
	ReplaySHA256          string             `json:"replay_sha256"`
	Events                []ReplayEntry      `json:"events"`
}

type eventDocument struct {
	SchemaVersion          string          `json:"schema_version"`
	EventID                string          `json:"event_id"`
	EventType              string          `json:"event_type"`
	ExpectedLedgerCommit   string          `json:"expected_ledger_commit"`
	AuthorityPolicyVersion string          `json:"authority_policy_version"`
	Targets                []string        `json:"targets"`
	RecordKind             string          `json:"record_kind"`
	Value                  json.RawMessage `json:"value"`
	PriorState             json.RawMessage `json:"prior_state"`
	ResultingState         json.RawMessage `json:"resulting_state"`
	IdempotencyKey         string          `json:"idempotency_key"`
}

type validatedEvent struct {
	Entry    ReplayEntry
	Document eventDocument
}

// Replay validates the authoritative Git history, including the immutable
// first-commit Genesis trust root and actor-auth trust policy, builds a
// deterministic audit manifest, and applies only explicitly accepted
// current-state reducer semantics. It remains read-only and cannot advance the
// authoritative ref.
func Replay(ctx context.Context, r *gitledger.Reader) (*ReplayManifest, error) {
	if err := r.CheckHistorySafety(ctx); err != nil {
		return nil, err
	}
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
	genesisRoot, genesisCommit, err := validateGenesisHistory(ctx, r, history)
	if err != nil {
		return nil, err
	}

	manifest := &ReplayManifest{
		LedgerCommit:       head,
		AuthoritativeRef:   r.Ref(),
		GitObjectFormat:    objectFormat,
		BareRepository:     bare,
		GenesisCommit:      genesisCommit,
		GenesisRoot:        genesisRoot,
		HistoryCommitCount: len(history),
		Events:             []ReplayEntry{},
		GovernedRecords:    reducer.Projection{},
	}
	seenEventIDs := map[string]ReplayEntry{}
	seenIdempotencyKeys := map[string]ReplayEntry{}
	projection := reducer.Projection{}
	bindingCount := 0

	for _, commit := range history {
		schemaChanges, err := r.ImmutableJSONAdditions(ctx, commit.ID, "config/schemas")
		if err != nil {
			return nil, err
		}
		bindingChanges, err := r.ImmutableJSONAdditions(ctx, commit.ID, policy.ReducerBindingPrefix)
		if err != nil {
			return nil, err
		}
		additions, err := r.EventAdditions(ctx, commit.ID)
		if err != nil {
			return nil, err
		}
		if len(schemaChanges) == 0 && len(bindingChanges) == 0 && len(additions) == 0 {
			continue
		}

		registry, err := loadSchemasAt(ctx, r, commit.ID)
		if err != nil {
			return nil, fmt.Errorf("schema snapshot at %s: %w", commit.ID, err)
		}
		bindings, err := policy.LoadReducerBindingsAt(ctx, r, registry, commit.ID)
		if err != nil {
			return nil, fmt.Errorf("reducer binding snapshot at %s: %w", commit.ID, err)
		}
		if commit.ID == genesisCommit {
			for _, binding := range bindings.ByBindingID {
				if binding.AuthorityPolicyVersion != genesisRoot.InitialAuthorityPolicy {
					return nil, fmt.Errorf("GENESIS_POLICY_MISMATCH: root binding %q uses authority policy %q; Genesis declares %q", binding.BindingID, binding.AuthorityPolicyVersion, genesisRoot.InitialAuthorityPolicy)
				}
			}
		}
		bindingCount = len(bindings.ByRecordKind)

		for _, addition := range additions {
			validated, err := validateEvent(ctx, r, registry, addition)
			if err != nil {
				return nil, err
			}
			entry := validated.Entry
			if prior, exists := seenEventIDs[entry.EventID]; exists {
				return nil, fmt.Errorf("INTEGRITY_FAILURE: duplicate logical event_id %q at %s and %s", entry.EventID, prior.Path, entry.Path)
			}
			if validated.Document.IdempotencyKey != "" {
				if prior, exists := seenIdempotencyKeys[validated.Document.IdempotencyKey]; exists {
					return nil, fmt.Errorf("INTEGRITY_FAILURE: duplicate idempotency_key %q at %s and %s", validated.Document.IdempotencyKey, prior.Path, entry.Path)
				}
				seenIdempotencyKeys[validated.Document.IdempotencyKey] = entry
			}

			if strings.HasPrefix(entry.EventType, "core.record.") {
				projection, err = applyGovernedRecordEvent(projection, bindings, commit, validated)
				if err != nil {
					return nil, fmt.Errorf("event %s at %s reducer: %w", entry.Path, commit.ID, err)
				}
			}

			entry.Sequence = len(manifest.Events) + 1
			manifest.Events = append(manifest.Events, entry)
			seenEventIDs[entry.EventID] = entry
		}
	}

	manifest.EventCount = len(manifest.Events)
	manifest.ReducerBindingCount = bindingCount
	manifest.GovernedRecordCount = len(projection)
	manifest.GovernedRecords = projection
	projectionCanonical, err := reducer.CanonicalProjection(projection)
	if err != nil {
		return nil, err
	}
	projectionSum := sha256.Sum256(projectionCanonical)
	manifest.GovernedRecordsSHA256 = hex.EncodeToString(projectionSum[:])

	replayDigest, err := replayDigest(manifest)
	if err != nil {
		return nil, err
	}
	manifest.ReplaySHA256 = replayDigest
	return manifest, nil
}

func validateGenesisHistory(ctx context.Context, r *gitledger.Reader, history []gitledger.Commit) (genesis.Root, string, error) {
	if len(history) == 0 {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_INVALID: authoritative ledger has no root commit")
	}
	rootCommit := history[0].ID
	changes, err := r.ImmutableJSONAdditions(ctx, rootCommit, genesis.LedgerPrefix)
	if err != nil {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_INVALID: root Genesis tree: %w", err)
	}
	if len(changes) != 1 || changes[0] != genesis.LedgerPath {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_INVALID: root commit %s must add exactly %q, got %v", rootCommit, genesis.LedgerPath, changes)
	}
	rootEvents, err := r.EventAdditions(ctx, rootCommit)
	if err != nil {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_ROOT_INVALID: inspect root events: %w", err)
	}
	if len(rootEvents) != 0 {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_ROOT_INVALID: root commit must not contain durable events")
	}
	raw, err := r.ReadRegularFile(ctx, rootCommit, genesis.LedgerPath)
	if err != nil {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_INVALID: read root record: %w", err)
	}
	root, err := genesis.Validate(raw)
	if err != nil {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_INVALID: root record at %s: %w", rootCommit, err)
	}
	actualSchemas, err := schemaIDsAt(ctx, r, rootCommit)
	if err != nil {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_INVALID: inspect initial schemas: %w", err)
	}
	if len(actualSchemas) != len(root.InitialSchemas) {
		return genesis.Root{}, "", fmt.Errorf("GENESIS_SCHEMA_MISMATCH: Genesis declares %v; root contains %v", root.InitialSchemas, actualSchemas)
	}
	for i := range actualSchemas {
		if actualSchemas[i] != root.InitialSchemas[i] {
			return genesis.Root{}, "", fmt.Errorf("GENESIS_SCHEMA_MISMATCH: Genesis declares %v; root contains %v", root.InitialSchemas, actualSchemas)
		}
	}
	if err := validateInitialActorPolicyHistory(ctx, r, history, root); err != nil {
		return genesis.Root{}, "", err
	}

	for _, commit := range history[1:] {
		later, err := r.ImmutableJSONAdditions(ctx, commit.ID, genesis.LedgerPrefix)
		if err != nil {
			return genesis.Root{}, "", fmt.Errorf("GENESIS_IMMUTABLE: commit %s: %w", commit.ID, err)
		}
		if len(later) != 0 {
			return genesis.Root{}, "", fmt.Errorf("GENESIS_IMMUTABLE: commit %s adds later Genesis material %v", commit.ID, later)
		}
	}
	return root, rootCommit, nil
}

func schemaIDsAt(ctx context.Context, r *gitledger.Reader, commit string) ([]string, error) {
	paths, err := r.ListJSON(ctx, commit, "config/schemas")
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(paths))
	for _, schemaPath := range paths {
		raw, err := r.ReadFile(ctx, commit, schemaPath)
		if err != nil {
			return nil, err
		}
		value, err := strictjson.Decode(raw)
		if err != nil {
			return nil, fmt.Errorf("schema %s: %w", schemaPath, err)
		}
		obj, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("SCHEMA_INVALID: schema %s root must be object", schemaPath)
		}
		id, _ := obj["$id"].(string)
		if id == "" {
			return nil, fmt.Errorf("SCHEMA_INVALID: schema %s has no $id", schemaPath)
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids, nil
}

func applyGovernedRecordEvent(current reducer.Projection, bindings *policy.ReducerBindingSnapshot, commit gitledger.Commit, validated validatedEvent) (reducer.Projection, error) {
	doc := validated.Document
	binding, exists := bindings.ByRecordKind[doc.RecordKind]
	if !exists {
		return nil, fmt.Errorf("%w: record_kind %q", reducer.ErrPolicyUnbound, doc.RecordKind)
	}
	if validated.Entry.SchemaVersion != binding.EventSchema {
		return nil, fmt.Errorf("REDUCER_EVENT_SCHEMA_MISMATCH: event schema %q binding requires %q", validated.Entry.SchemaVersion, binding.EventSchema)
	}
	if doc.AuthorityPolicyVersion != binding.AuthorityPolicyVersion {
		return nil, fmt.Errorf("AUTHORITY_POLICY_VERSION_MISMATCH: event has %q binding requires %q", doc.AuthorityPolicyVersion, binding.AuthorityPolicyVersion)
	}
	if commit.Parent == "" {
		return nil, fmt.Errorf("EXPECTED_LEDGER_COMMIT_MISMATCH: reducer event cannot be accepted in root commit")
	}
	if doc.ExpectedLedgerCommit != commit.Parent {
		return nil, fmt.Errorf("EXPECTED_LEDGER_COMMIT_MISMATCH: event has %q accepting parent is %q", doc.ExpectedLedgerCommit, commit.Parent)
	}
	return reducer.Apply(current, bindings.ReducerBindings(), reducer.Event{
		EventID:        doc.EventID,
		EventType:      doc.EventType,
		Targets:        doc.Targets,
		RecordKind:     doc.RecordKind,
		Value:          doc.Value,
		PriorState:     doc.PriorState,
		ResultingState: doc.ResultingState,
	})
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

func validateEvent(ctx context.Context, r *gitledger.Reader, registry *schema.Registry, addition gitledger.EventAddition) (validatedEvent, error) {
	raw, err := r.ReadFile(ctx, addition.Commit, addition.Path)
	if err != nil {
		return validatedEvent{}, err
	}
	if err := strictjson.Validate(raw); err != nil {
		return validatedEvent{}, fmt.Errorf("event %s at %s: %w", addition.Path, addition.Commit, err)
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return validatedEvent{}, fmt.Errorf("event %s at %s canonicalization: %w", addition.Path, addition.Commit, err)
	}
	if !bytes.Equal(raw, canonical) {
		return validatedEvent{}, fmt.Errorf("INTEGRITY_FAILURE: event %s at %s is not stored as RFC 8785 canonical JSON", addition.Path, addition.Commit)
	}
	if err := digest.Verify(raw); err != nil {
		return validatedEvent{}, fmt.Errorf("event %s at %s: %w", addition.Path, addition.Commit, err)
	}

	var doc eventDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return validatedEvent{}, fmt.Errorf("event %s decode: %w", addition.Path, err)
	}
	if doc.SchemaVersion == "" {
		return validatedEvent{}, fmt.Errorf("UNKNOWN_SCHEMA: event %s has no schema_version", addition.Path)
	}
	if err := registry.Validate(doc.SchemaVersion, raw); err != nil {
		return validatedEvent{}, fmt.Errorf("event %s at %s: %w", addition.Path, addition.Commit, err)
	}
	contentSHA := ""
	value, err := strictjson.Decode(raw)
	if err != nil {
		return validatedEvent{}, err
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return validatedEvent{}, fmt.Errorf("event %s root must be object", addition.Path)
	}
	contentSHA, _ = obj[digest.Field].(string)
	if doc.EventID == "" || doc.EventType == "" || contentSHA == "" {
		return validatedEvent{}, fmt.Errorf("INTEGRITY_FAILURE: event %s lacks replay identity fields", addition.Path)
	}
	entry := ReplayEntry{
		AcceptedCommit: addition.Commit,
		Path:           addition.Path,
		EventID:        doc.EventID,
		EventType:      doc.EventType,
		SchemaVersion:  doc.SchemaVersion,
		ContentSHA256:  contentSHA,
		TargetCount:    len(doc.Targets),
	}
	return validatedEvent{Entry: entry, Document: doc}, nil
}

func replayDigest(manifest *ReplayManifest) (string, error) {
	payload := struct {
		LedgerCommit          string        `json:"ledger_commit"`
		AuthoritativeRef      string        `json:"authoritative_ref"`
		GitObjectFormat       string        `json:"git_object_format"`
		GenesisCommit         string        `json:"genesis_commit"`
		GenesisRoot           genesis.Root  `json:"genesis_root"`
		HistoryCommitCount    int           `json:"history_commit_count"`
		ReducerBindingCount   int           `json:"reducer_binding_count"`
		GovernedRecordCount   int           `json:"governed_record_count"`
		GovernedRecordsSHA256 string        `json:"governed_records_sha256"`
		Events                []ReplayEntry `json:"events"`
	}{
		LedgerCommit:          manifest.LedgerCommit,
		AuthoritativeRef:      manifest.AuthoritativeRef,
		GitObjectFormat:       manifest.GitObjectFormat,
		GenesisCommit:         manifest.GenesisCommit,
		GenesisRoot:           manifest.GenesisRoot,
		HistoryCommitCount:    manifest.HistoryCommitCount,
		ReducerBindingCount:   manifest.ReducerBindingCount,
		GovernedRecordCount:   manifest.GovernedRecordCount,
		GovernedRecordsSHA256: manifest.GovernedRecordsSHA256,
		Events:                manifest.Events,
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
