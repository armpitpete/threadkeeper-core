package ledger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/genesis"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/policy"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
	"github.com/armpitpete/threadkeeper-core/internal/schema"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

type FreshGenesisEvidence struct {
	StoragePath              string `json:"storage_path"`
	ProjectID                string `json:"project_id"`
	LedgerID                 string `json:"ledger_id"`
	AuthoritativeRef         string `json:"authoritative_ref"`
	GenesisContentSHA256     string `json:"genesis_content_sha256"`
	ActorPolicyContentSHA256 string `json:"actor_policy_content_sha256"`
	GenesisCommit            string `json:"genesis_commit"`
	LedgerCommit             string `json:"ledger_commit"`
	GitObjectFormat          string `json:"git_object_format"`
	InitialSchemaCount       int    `json:"initial_schema_count"`
	InitialBindingCount      int    `json:"initial_binding_count"`
}

// InitializeFreshGenesis creates a brand-new dedicated authority ledger whose
// only commit is the supplied Genesis trust root plus explicitly allowed initial
// immutable configuration. All semantic seed validation runs before the target
// path is created. Success is returned only after the new repository is reopened
// through the hardened reader and the complete Replay/FSCK path proves the exact
// Genesis and actor-policy identity.
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
	policyDoc, _, err := actorauth.ParsePolicyDocument(policyRaw, root.LedgerID, root.InitialAuthorityPolicy)
	if err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_INVALID: %w", err)
	}
	if err := actorauth.ValidateInitialAuthorities(policyDoc, root.InitialAuthorities); err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_INVALID: %w", err)
	}
	if err := validateFreshGenesisSeedContracts(root, seedFiles); err != nil {
		return nil, err
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
	registry, err := loadSchemasAt(ctx, r, createdRoot)
	if err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: initial schema registry: %w", err)
	}
	bindings, err := policy.LoadReducerBindingsAt(ctx, r, registry, createdRoot)
	if err != nil {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: initial reducer bindings: %w", err)
	}
	if _, ok := bindings.ByRecordKind[actorauth.PolicyRecordKind]; !ok {
		return nil, fmt.Errorf("FRESH_GENESIS_VERIFICATION_FAILED: no reducer binding for actor policy record kind %q", actorauth.PolicyRecordKind)
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

func validateFreshGenesisSeedContracts(root genesis.Root, seedFiles map[string][]byte) error {
	registry := schema.NewRegistry()
	schemaIDs := make([]string, 0)
	for p, raw := range seedFiles {
		if !strings.HasPrefix(p, "config/schemas/") {
			continue
		}
		value, err := strictjson.Decode(raw)
		if err != nil {
			return fmt.Errorf("FRESH_GENESIS_INVALID: schema %s: %w", p, err)
		}
		obj, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("FRESH_GENESIS_INVALID: schema %s root must be object", p)
		}
		id, _ := obj["$id"].(string)
		if id == "" {
			return fmt.Errorf("FRESH_GENESIS_INVALID: schema %s has no $id", p)
		}
		if err := registry.Add(id, raw); err != nil {
			return fmt.Errorf("FRESH_GENESIS_INVALID: schema %s: %w", p, err)
		}
		schemaIDs = append(schemaIDs, id)
	}
	sort.Strings(schemaIDs)
	if len(schemaIDs) != len(root.InitialSchemas) {
		return fmt.Errorf("FRESH_GENESIS_INVALID: Genesis initial_schemas %v do not match seed schema IDs %v", root.InitialSchemas, schemaIDs)
	}
	for i := range schemaIDs {
		if schemaIDs[i] != root.InitialSchemas[i] {
			return fmt.Errorf("FRESH_GENESIS_INVALID: Genesis initial_schemas %v do not match seed schema IDs %v", root.InitialSchemas, schemaIDs)
		}
	}

	actorBindingFound := false
	seenIDs := map[string]struct{}{}
	seenKinds := map[string]struct{}{}
	for p, raw := range seedFiles {
		if !strings.HasPrefix(p, policy.ReducerBindingPrefix+"/") {
			continue
		}
		if err := strictjson.Validate(raw); err != nil {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %s: %w", p, err)
		}
		canonical, err := canonicaljson.Canonicalize(raw)
		if err != nil {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %s canonicalization: %w", p, err)
		}
		if !bytes.Equal(raw, canonical) {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %s is not RFC 8785 canonical JSON", p)
		}
		if err := digest.Verify(raw); err != nil {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %s: %w", p, err)
		}
		var binding policy.ReducerBinding
		if err := json.Unmarshal(raw, &binding); err != nil {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %s decode: %w", p, err)
		}
		if binding.SchemaVersion != contracts.ReducerBindingSchemaV1 {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %s uses unknown schema %q", p, binding.SchemaVersion)
		}
		if err := registry.Validate(binding.SchemaVersion, raw); err != nil {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %s: %w", p, err)
		}
		if binding.StateModel != reducer.ModelExclusiveV1 {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %q uses model %q", binding.BindingID, binding.StateModel)
		}
		if binding.EventSchema != contracts.ExclusiveRecordEventSchemaV1 {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %q uses event schema %q", binding.BindingID, binding.EventSchema)
		}
		if _, err := registry.Compile(binding.EventSchema); err != nil {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %q event schema: %w", binding.BindingID, err)
		}
		if binding.AuthorityPolicyVersion != root.InitialAuthorityPolicy {
			return fmt.Errorf("FRESH_GENESIS_INVALID: reducer binding %q policy %q does not match Genesis %q", binding.BindingID, binding.AuthorityPolicyVersion, root.InitialAuthorityPolicy)
		}
		if _, exists := seenIDs[binding.BindingID]; exists {
			return fmt.Errorf("FRESH_GENESIS_INVALID: duplicate reducer binding_id %q", binding.BindingID)
		}
		if _, exists := seenKinds[binding.RecordKind]; exists {
			return fmt.Errorf("FRESH_GENESIS_INVALID: duplicate reducer record_kind %q", binding.RecordKind)
		}
		seenIDs[binding.BindingID] = struct{}{}
		seenKinds[binding.RecordKind] = struct{}{}
		if binding.RecordKind == actorauth.PolicyRecordKind {
			actorBindingFound = true
		}
	}
	if !actorBindingFound {
		return fmt.Errorf("FRESH_GENESIS_INVALID: no reducer binding for actor policy record kind %q", actorauth.PolicyRecordKind)
	}
	return nil
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
