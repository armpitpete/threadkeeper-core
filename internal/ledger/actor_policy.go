package ledger

import (
	"context"
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
)

type ActorPolicySnapshot struct {
	LedgerCommit     string
	GenesisCommit    string
	LedgerID         string
	PolicyVersion    string
	PolicyContentSHA string
	SourceEventID    string
	Document         actorauth.PolicyDocument
	Policy           actorauth.Policy
}

func validateInitialActorPolicyHistory(ctx context.Context, r *gitledger.Reader, history []gitledger.Commit, root genesis.Root) (actorauth.PolicyDocument, error) {
	if len(history) == 0 {
		return actorauth.PolicyDocument{}, fmt.Errorf("AUTH_POLICY_INVALID: authoritative history is empty")
	}
	rootCommit := history[0].ID
	changes, err := r.ImmutableJSONAdditions(ctx, rootCommit, actorauth.LedgerPolicyPrefix)
	if err != nil {
		return actorauth.PolicyDocument{}, fmt.Errorf("AUTH_POLICY_INVALID: root actor-policy tree: %w", err)
	}
	if len(changes) != 1 || changes[0] != actorauth.LedgerPolicyPath {
		return actorauth.PolicyDocument{}, fmt.Errorf("AUTH_POLICY_INVALID: Genesis commit %s must add exactly %q, got %v", rootCommit, actorauth.LedgerPolicyPath, changes)
	}
	raw, err := r.ReadRegularFile(ctx, rootCommit, actorauth.LedgerPolicyPath)
	if err != nil {
		return actorauth.PolicyDocument{}, fmt.Errorf("AUTH_POLICY_INVALID: read root actor policy: %w", err)
	}
	doc, _, err := actorauth.ParsePolicyDocument(raw, root.LedgerID, root.InitialAuthorityPolicy)
	if err != nil {
		return actorauth.PolicyDocument{}, err
	}
	if err := actorauth.ValidateInitialAuthorities(doc, root.InitialAuthorities); err != nil {
		return actorauth.PolicyDocument{}, err
	}
	for _, commit := range history[1:] {
		later, err := r.ImmutableJSONAdditions(ctx, commit.ID, actorauth.LedgerPolicyPrefix)
		if err != nil {
			return actorauth.PolicyDocument{}, fmt.Errorf("AUTH_POLICY_IMMUTABLE: commit %s: %w", commit.ID, err)
		}
		if len(later) != 0 {
			return actorauth.PolicyDocument{}, fmt.Errorf("AUTH_POLICY_IMMUTABLE: commit %s changes root actor-policy material %v", commit.ID, later)
		}
	}
	return doc, nil
}

// LoadCurrentActorPolicy performs a complete authoritative replay and derives
// the trusted proof-verification policy from that exact snapshot. The immutable
// Genesis-root policy is the initial value. If the fixed governed policy target
// has later been created/replaced, its active value is authoritative instead.
// Revoking the governed target fails closed and does not fall back to Genesis.
func LoadCurrentActorPolicy(ctx context.Context, r *gitledger.Reader) (*ActorPolicySnapshot, error) {
	manifest, err := Replay(ctx, r)
	if err != nil {
		return nil, err
	}
	raw, err := r.ReadRegularFile(ctx, manifest.GenesisCommit, actorauth.LedgerPolicyPath)
	if err != nil {
		return nil, fmt.Errorf("AUTH_POLICY_INVALID: read root actor policy: %w", err)
	}
	doc, policy, err := actorauth.ParsePolicyDocument(raw, manifest.GenesisRoot.LedgerID, manifest.GenesisRoot.InitialAuthorityPolicy)
	if err != nil {
		return nil, err
	}
	sourceEvent := ""
	if state, exists := manifest.GovernedRecords[actorauth.PolicyTarget]; exists {
		if state.RecordKind != actorauth.PolicyRecordKind {
			return nil, fmt.Errorf("AUTH_POLICY_INVALID: governed target %q has record kind %q", actorauth.PolicyTarget, state.RecordKind)
		}
		if state.Status == reducer.StatusRevoked {
			return nil, fmt.Errorf("AUTH_POLICY_REVOKED: governed actor policy is revoked at event %s", state.CurrentEventID)
		}
		if state.Status != reducer.StatusActive {
			return nil, fmt.Errorf("AUTH_POLICY_INVALID: governed actor policy has status %q", state.Status)
		}
		doc, policy, err = actorauth.ParsePolicyDocument(state.Value, manifest.GenesisRoot.LedgerID, manifest.GenesisRoot.InitialAuthorityPolicy)
		if err != nil {
			return nil, err
		}
		sourceEvent = state.CurrentEventID
	}
	return &ActorPolicySnapshot{
		LedgerCommit:     manifest.LedgerCommit,
		GenesisCommit:    manifest.GenesisCommit,
		LedgerID:         manifest.GenesisRoot.LedgerID,
		PolicyVersion:    manifest.GenesisRoot.InitialAuthorityPolicy,
		PolicyContentSHA: doc.ContentSHA256,
		SourceEventID:    sourceEvent,
		Document:         doc,
		Policy:           policy,
	}, nil
}
