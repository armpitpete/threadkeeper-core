package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/policy"
	"github.com/armpitpete/threadkeeper-core/internal/schema"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

var (
	ErrIdempotencyConflict = errors.New("IDEMPOTENCY_CONFLICT")
	ErrCandidateInvalid    = errors.New("CANDIDATE_INVALID")
)

const (
	WriteStatusAccepted        = "accepted"
	WriteStatusAlreadyAccepted = "already_accepted"
)

type CandidateRequest struct {
	ExpectedHead string
	EventPath    string
	Event        []byte
}

type WriteCandidate struct {
	ExpectedHead    string `json:"expected_head"`
	CandidateCommit string `json:"candidate_commit"`
	EventPath       string `json:"event_path"`
	EventID         string `json:"event_id"`
	IdempotencyKey  string `json:"idempotency_key"`
	ContentSHA256   string `json:"content_sha256"`
}

type WriteResponse struct {
	Status         string `json:"status"`
	EventID        string `json:"event_id"`
	IdempotencyKey string `json:"idempotency_key"`
	ContentSHA256  string `json:"content_sha256"`
	EventPath      string `json:"event_path"`
	AcceptedCommit string `json:"accepted_commit"`
	LedgerCommit   string `json:"ledger_commit"`
}

type candidateDocument struct {
	EventID        string
	IdempotencyKey string
	ContentSHA256  string
	Document       eventDocument
}

// PrepareWriteCandidate validates a fully formed durable event against the
// exact current ledger state and creates unreachable Git candidate objects.
// It never updates the authoritative ref.
func PrepareWriteCandidate(ctx context.Context, r *gitledger.Reader, req CandidateRequest) (*WriteCandidate, *WriteResponse, error) {
	manifest, err := Replay(ctx, r)
	if err != nil {
		return nil, nil, err
	}
	doc, err := parseCandidateDocument(req.Event)
	if err != nil {
		return nil, nil, err
	}
	if doc.IdempotencyKey == "" {
		return nil, nil, fmt.Errorf("%w: idempotency_key is required", ErrCandidateInvalid)
	}

	accepted, err := findAcceptedIdempotency(ctx, r, doc.IdempotencyKey)
	if err != nil {
		return nil, nil, err
	}
	if accepted != nil {
		if accepted.Entry.ContentSHA256 != doc.ContentSHA256 || accepted.Document.EventID != doc.EventID {
			return nil, nil, fmt.Errorf("%w: key %q is already bound to event %q digest %s", ErrIdempotencyConflict, doc.IdempotencyKey, accepted.Document.EventID, accepted.Entry.ContentSHA256)
		}
		return nil, responseFromAccepted(WriteStatusAlreadyAccepted, manifest.LedgerCommit, accepted), nil
	}

	if strings.ToLower(req.ExpectedHead) != manifest.LedgerCommit {
		return nil, nil, fmt.Errorf("%w: expected %s current %s", gitledger.ErrStaleState, req.ExpectedHead, manifest.LedgerCommit)
	}
	if doc.Document.ExpectedLedgerCommit != manifest.LedgerCommit {
		return nil, nil, fmt.Errorf("EXPECTED_LEDGER_COMMIT_MISMATCH: event has %q request/current head is %q", doc.Document.ExpectedLedgerCommit, manifest.LedgerCommit)
	}
	if !strings.HasPrefix(doc.Document.EventType, "core.record.") {
		return nil, nil, fmt.Errorf("%w: event type %q has no accepted write semantics", ErrCandidateInvalid, doc.Document.EventType)
	}

	registry, err := loadSchemasAt(ctx, r, manifest.LedgerCommit)
	if err != nil {
		return nil, nil, err
	}
	if err := registry.Validate(doc.Document.SchemaVersion, req.Event); err != nil {
		return nil, nil, fmt.Errorf("%w: schema validation: %v", ErrCandidateInvalid, err)
	}
	bindings, err := policy.LoadReducerBindingsAt(ctx, r, registry, manifest.LedgerCommit)
	if err != nil {
		return nil, nil, err
	}
	validated := validatedEvent{
		Entry: ReplayEntry{
			EventID:       doc.EventID,
			EventType:     doc.Document.EventType,
			SchemaVersion: doc.Document.SchemaVersion,
			ContentSHA256: doc.ContentSHA256,
			TargetCount:   len(doc.Document.Targets),
		},
		Document: doc.Document,
	}
	if _, err := applyGovernedRecordEvent(manifest.GovernedRecords, bindings, gitledger.Commit{ID: "candidate", Parent: manifest.LedgerCommit}, validated); err != nil {
		return nil, nil, fmt.Errorf("%w: reducer validation: %v", ErrCandidateInvalid, err)
	}

	gitCandidate, err := r.PrepareEventCommit(ctx, manifest.LedgerCommit, req.EventPath, req.Event, doc.EventID)
	if err != nil {
		return nil, nil, err
	}
	addition, err := validatePreparedCandidate(ctx, r, registry, gitCandidate, req.Event)
	if err != nil {
		return nil, nil, err
	}
	if addition.Document.IdempotencyKey != doc.IdempotencyKey || addition.Entry.ContentSHA256 != doc.ContentSHA256 || addition.Document.EventID != doc.EventID {
		return nil, nil, fmt.Errorf("%w: prepared candidate identity changed", ErrCandidateInvalid)
	}

	return &WriteCandidate{
		ExpectedHead:    manifest.LedgerCommit,
		CandidateCommit: gitCandidate.Commit,
		EventPath:       req.EventPath,
		EventID:         doc.EventID,
		IdempotencyKey:  doc.IdempotencyKey,
		ContentSHA256:   doc.ContentSHA256,
	}, nil, nil
}

// AcceptWriteCandidate performs the sole authority-changing primitive in this
// package: exact-head Git compare-and-swap. It is intentionally not exposed by
// the CLI while the service write gate remains disabled.
func AcceptWriteCandidate(ctx context.Context, r *gitledger.Reader, candidate WriteCandidate) (*WriteResponse, error) {
	manifest, err := Replay(ctx, r)
	if err != nil {
		return nil, err
	}
	accepted, err := findAcceptedIdempotency(ctx, r, candidate.IdempotencyKey)
	if err != nil {
		return nil, err
	}
	if accepted != nil {
		if accepted.Entry.ContentSHA256 != candidate.ContentSHA256 || accepted.Document.EventID != candidate.EventID {
			return nil, fmt.Errorf("%w: key %q is already bound to another accepted request", ErrIdempotencyConflict, candidate.IdempotencyKey)
		}
		return responseFromAccepted(WriteStatusAlreadyAccepted, manifest.LedgerCommit, accepted), nil
	}
	if manifest.LedgerCommit != strings.ToLower(candidate.ExpectedHead) {
		return nil, fmt.Errorf("%w: expected %s current %s", gitledger.ErrStaleState, candidate.ExpectedHead, manifest.LedgerCommit)
	}
	if err := validateCandidateForAcceptance(ctx, r, manifest, candidate); err != nil {
		return nil, err
	}
	if err := r.CompareAndSwap(ctx, candidate.ExpectedHead, candidate.CandidateCommit); err != nil {
		if errors.Is(err, gitledger.ErrStaleState) {
			afterRace, replayErr := Replay(ctx, r)
			if replayErr == nil {
				raced, lookupErr := findAcceptedIdempotency(ctx, r, candidate.IdempotencyKey)
				if lookupErr == nil && raced != nil && raced.Entry.ContentSHA256 == candidate.ContentSHA256 && raced.Document.EventID == candidate.EventID {
					return responseFromAccepted(WriteStatusAlreadyAccepted, afterRace.LedgerCommit, raced), nil
				}
			}
		}
		return nil, err
	}

	post, err := Replay(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("POST_ACCEPTANCE_VERIFICATION_FAILED: accepted commit %s: %w", candidate.CandidateCommit, err)
	}
	accepted, err = findAcceptedIdempotency(ctx, r, candidate.IdempotencyKey)
	if err != nil {
		return nil, fmt.Errorf("POST_ACCEPTANCE_VERIFICATION_FAILED: %w", err)
	}
	if accepted == nil || accepted.Entry.AcceptedCommit != candidate.CandidateCommit || accepted.Entry.ContentSHA256 != candidate.ContentSHA256 || accepted.Document.EventID != candidate.EventID {
		return nil, fmt.Errorf("POST_ACCEPTANCE_VERIFICATION_FAILED: accepted event identity not recoverable from ledger")
	}
	return responseFromAccepted(WriteStatusAccepted, post.LedgerCommit, accepted), nil
}

func validateCandidateForAcceptance(ctx context.Context, r *gitledger.Reader, manifest *ReplayManifest, candidate WriteCandidate) error {
	if candidate.IdempotencyKey == "" || candidate.EventID == "" || candidate.ContentSHA256 == "" {
		return fmt.Errorf("%w: candidate handle lacks durable identity", ErrCandidateInvalid)
	}
	if err := r.VerifyEventCandidate(ctx, manifest.LedgerCommit, candidate.CandidateCommit, candidate.EventPath); err != nil {
		return fmt.Errorf("%w: Git candidate verification: %v", ErrCandidateInvalid, err)
	}
	registry, err := loadSchemasAt(ctx, r, manifest.LedgerCommit)
	if err != nil {
		return err
	}
	additions, err := r.EventAdditions(ctx, candidate.CandidateCommit)
	if err != nil {
		return err
	}
	if len(additions) != 1 || additions[0].Path != candidate.EventPath {
		return fmt.Errorf("%w: candidate event path mismatch", ErrCandidateInvalid)
	}
	validated, err := validateEvent(ctx, r, registry, additions[0])
	if err != nil {
		return fmt.Errorf("%w: candidate event validation: %v", ErrCandidateInvalid, err)
	}
	if validated.Document.EventID != candidate.EventID || validated.Document.IdempotencyKey != candidate.IdempotencyKey || validated.Entry.ContentSHA256 != candidate.ContentSHA256 {
		return fmt.Errorf("%w: candidate handle does not match candidate commit event identity", ErrCandidateInvalid)
	}
	bindings, err := policy.LoadReducerBindingsAt(ctx, r, registry, manifest.LedgerCommit)
	if err != nil {
		return err
	}
	if _, err := applyGovernedRecordEvent(manifest.GovernedRecords, bindings, gitledger.Commit{ID: candidate.CandidateCommit, Parent: manifest.LedgerCommit}, validated); err != nil {
		return fmt.Errorf("%w: candidate reducer validation: %v", ErrCandidateInvalid, err)
	}
	return nil
}

func parseCandidateDocument(raw []byte) (candidateDocument, error) {
	if err := strictjson.Validate(raw); err != nil {
		return candidateDocument{}, fmt.Errorf("%w: %v", ErrCandidateInvalid, err)
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return candidateDocument{}, fmt.Errorf("%w: canonicalization: %v", ErrCandidateInvalid, err)
	}
	if !bytes.Equal(raw, canonical) {
		return candidateDocument{}, fmt.Errorf("%w: durable event must already be stored as RFC 8785 canonical JSON", ErrCandidateInvalid)
	}
	if err := digest.Verify(raw); err != nil {
		return candidateDocument{}, fmt.Errorf("%w: %v", ErrCandidateInvalid, err)
	}
	var doc eventDocument
	if err := json.Unmarshal(raw, &doc); err != nil {
		return candidateDocument{}, fmt.Errorf("%w: decode event: %v", ErrCandidateInvalid, err)
	}
	value, err := strictjson.Decode(raw)
	if err != nil {
		return candidateDocument{}, err
	}
	obj, ok := value.(map[string]any)
	if !ok {
		return candidateDocument{}, fmt.Errorf("%w: event root must be object", ErrCandidateInvalid)
	}
	contentSHA, _ := obj[digest.Field].(string)
	if doc.EventID == "" || doc.EventType == "" || doc.SchemaVersion == "" || contentSHA == "" {
		return candidateDocument{}, fmt.Errorf("%w: event lacks identity fields", ErrCandidateInvalid)
	}
	return candidateDocument{EventID: doc.EventID, IdempotencyKey: doc.IdempotencyKey, ContentSHA256: contentSHA, Document: doc}, nil
}

func validatePreparedCandidate(ctx context.Context, r *gitledger.Reader, registry *schema.Registry, candidate *gitledger.CandidateCommit, expected []byte) (validatedEvent, error) {
	additions, err := r.EventAdditions(ctx, candidate.Commit)
	if err != nil {
		return validatedEvent{}, err
	}
	if len(additions) != 1 || additions[0].Path != candidate.EventPath {
		return validatedEvent{}, fmt.Errorf("%w: candidate must add exactly event path %q", ErrCandidateInvalid, candidate.EventPath)
	}
	stored, err := r.ReadFile(ctx, candidate.Commit, candidate.EventPath)
	if err != nil {
		return validatedEvent{}, err
	}
	if !bytes.Equal(stored, expected) {
		return validatedEvent{}, fmt.Errorf("%w: candidate event bytes differ from validated request", ErrCandidateInvalid)
	}
	validated, err := validateEvent(ctx, r, registry, additions[0])
	if err != nil {
		return validatedEvent{}, err
	}
	return validated, nil
}

func findAcceptedIdempotency(ctx context.Context, r *gitledger.Reader, key string) (*validatedEvent, error) {
	if key == "" {
		return nil, nil
	}
	head, err := r.Head(ctx)
	if err != nil {
		return nil, err
	}
	history, err := r.History(ctx, head)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		for _, addition := range additions {
			validated, err := validateEvent(ctx, r, registry, addition)
			if err != nil {
				return nil, err
			}
			if validated.Document.IdempotencyKey == key {
				copy := validated
				return &copy, nil
			}
		}
	}
	return nil, nil
}

func responseFromAccepted(status, ledgerCommit string, accepted *validatedEvent) *WriteResponse {
	return &WriteResponse{
		Status:         status,
		EventID:        accepted.Document.EventID,
		IdempotencyKey: accepted.Document.IdempotencyKey,
		ContentSHA256:  accepted.Entry.ContentSHA256,
		EventPath:      accepted.Entry.Path,
		AcceptedCommit: accepted.Entry.AcceptedCommit,
		LedgerCommit:   ledgerCommit,
	}
}
