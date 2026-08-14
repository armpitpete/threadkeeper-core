package ledger

import (
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/policy"
)

type FreshGenesisEvidence struct {
	StoragePath             string `json:"storage_path"`
	ProjectID               string `json:"project_id"`
	LedgerID                string `json:"ledger_id"`
	AuthoritativeRef        string `json:"authoritative_ref"`
	GenesisContentSHA256    string `json:"genesis_content_sha256"`
	ActorPolicyContentSHA256 string `json:"actor_policy_content_sha256"`
	GenesisCommit           string `json:"genesis_commit"`
	LedgerCommit            string `json:"ledger_commit"`
	GitObjectFormat         string `json:"git_object_format"`
	InitialSchemaCount      int    `json:"initial_schema_count"`
	InitialBindingCount     int    `json:"initial_binding_count"`
}

// InitializeFreshGenesis creates a brand-new dedicated authority ledger whose
// only commit is the supplied Genesis trust root plus explicitly allowed initial
// immutable configuration. The target is create-only; existing paths are never
// adopted or overwritten. Success is returned only after the new repository is
// reopened through the hardened reader and the complete Replay/FSCK path proves
// the exact Genesis and actor-policy identity.
func InitializeFreshGenesis(ctx context.Context, gitDir, ref string, rawGenesis []byte, seedFiles map[string][]byte) (*FreshGenesisEvidence, error) {
	root, err := genesis.Validate(rawGenesis)
	if err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_INVALID: %w", err)
	}
	if ref == "" {
		ref = gitledger.DefaultRef
	}
	files := make(map[string][]byte, len(seedFiles)+1)
	files[genesis.LedgerPath] = append([]byte(nil), rawGenesis...)
	for p, raw := range seedFiles {
		if err := validateFreshGenesisSeedPath(p); err != nil {
			return nil, err
		}
		if p == genesis.LedgerPath {
			return nil, fmt.Errorf("FRESH_GENESIS_INVALID: seed files must not supply %q", genesis.LedgerPath)
		}
		files[p] = append([]byte(nil), raw...)
	}
	policyRaw, ok := seedFiles[actorauth.LedgerPolicyPath]
	if !ok {
		return nil, fmt.Errorf("FRESH_GENESIS_INVALID: seed root must supply authoritative actor policy %q", actorauth.LedgerPolicyPath)
	}
	policyDoc, _, err := actorauth.ParsePolicyDocument(policyRaw, root.LedgerID)
	if err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_INVALID: %w", err)
	}
	if err := actorauth.ValidateInitialAuthorities(policyDoc, root.InitialAuthorities); err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_INVALID: %w", err)
	}

	createdRoot, err := gitledger.InitializeBareRoot(ctx, gitDir, ref, files)
	if err != nil {
		return nil, err
	}
	r, err := gitledger.New(gitDir, ref)
	if err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: reopen created ledger: %w", err)
	}
	defer r.Close()
	manifest, err := Replay(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: %w", err)
	}
	if manifest.GenesisCommit != createdRoot || manifest.LedgerCommit != createdRoot || manifest.HistoryCommitCount != 1 {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: created root=%s genesis=%s head=%s history=%d", createdRoot, manifest.GenesisCommit, manifest.LedgerCommit, manifest.HistoryCommitCount)
	}
	if manifest.GenesisRoot.ProjectID != root.ProjectID || manifest.GenesisRoot.LedgerID != root.LedgerID || manifest.GenesisRoot.ContentSHA256 != root.ContentSHA256 {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: replayed Genesis identity differs from validated input")
	}
	snapshot, err := LoadCurrentActorPolicy(ctx, r)
	if err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: actor policy: %w", err)
	}
	if snapshot.LedgerCommit != createdRoot || snapshot.SourceEventID != "" || snapshot.PolicyContentSHA != policyDoc.ContentSHA256 {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: initial actor-policy identity changed during bootstrap")
	}
	return &FreshGenesisEvidence{
		StoragePath:              r.GitDir(),
		ProjectID:                root.ProjectID,
		LedgerID:                 root.LedgerID,
		AuthoritativeRef:         manifest.AuthoritativeRef,
		GenesisContentSHA256:     root.ContentSHA256,
		ActorPolicyContentSHA256: policyDoc.ContentSHA256,
		GenesisCommit:            createdRoot,
		LedgerCommit:             manifest.LedgerCommit,
		GitObjectFormat:          manifest.GitObjectFormat,
		InitialSchemaCount:       len(root.InitialSchemas),
		InitialBindingCount:      manifest.ReducerBindingCount,
	}, nil
}

func validateFreshGenesisSeedPath(p string) error {
	if p == "" || strings.Contains(p, "\\") || strings.ContainsAny(p, "\x00\r\n") || path.IsAbs(p) || path.Clean(p) != p || !strings.HasSuffix(p, ".json") {
		return fmt.Errorf("FRESH_GENESIS_INVALID: unsafe seed path %q", p)
	}
	allowedSchema := strings.HasPrefix(p, "config/schemas/")
	allowedBinding := strings.HasPrefix(p, policy.ReducerBindingPrefix+"/")
	allowedActorPolicy := p == actorauth.LedgerPolicyPath
	if !allowedSchema && !allowedBinding && !allowedActorPolicy {
		return fmt.Errorf("FRESH_GENESIS_INVALID: seed path %q is outside initial schema/reducer-binding/actor-policy namespaces", p)
	}
	return nil
}
