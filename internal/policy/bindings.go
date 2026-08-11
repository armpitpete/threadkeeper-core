package policy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/contracts"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/reducer"
	"github.com/armpitpete/threadkeeper-core/internal/schema"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

const ReducerBindingPrefix = "config/authority/reducer-bindings"

type ReducerBinding struct {
	SchemaVersion          string `json:"schema_version"`
	BindingID              string `json:"binding_id"`
	RecordKind             string `json:"record_kind"`
	StateModel             string `json:"state_model"`
	EventSchema            string `json:"event_schema"`
	AuthorityPolicyVersion string `json:"authority_policy_version"`
	ContentSHA256          string `json:"content_sha256"`
}

type ReducerBindingSnapshot struct {
	ByRecordKind map[string]ReducerBinding
	ByBindingID  map[string]ReducerBinding
}

func LoadReducerBindingsAt(ctx context.Context, r *gitledger.Reader, registry *schema.Registry, commit string) (*ReducerBindingSnapshot, error) {
	paths, err := r.ListJSON(ctx, commit, ReducerBindingPrefix)
	if err != nil {
		return nil, err
	}
	snapshot := &ReducerBindingSnapshot{
		ByRecordKind: make(map[string]ReducerBinding),
		ByBindingID:  make(map[string]ReducerBinding),
	}
	for _, path := range paths {
		raw, err := r.ReadFile(ctx, commit, path)
		if err != nil {
			return nil, err
		}
		if err := strictjson.Validate(raw); err != nil {
			return nil, fmt.Errorf("reducer binding %s: %w", path, err)
		}
		canonical, err := canonicaljson.Canonicalize(raw)
		if err != nil {
			return nil, fmt.Errorf("reducer binding %s canonicalization: %w", path, err)
		}
		if !bytes.Equal(raw, canonical) {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: reducer binding %s is not stored as RFC 8785 canonical JSON", path)
		}
		if err := digest.Verify(raw); err != nil {
			return nil, fmt.Errorf("reducer binding %s: %w", path, err)
		}

		var binding ReducerBinding
		if err := json.Unmarshal(raw, &binding); err != nil {
			return nil, fmt.Errorf("reducer binding %s decode: %w", path, err)
		}
		if binding.SchemaVersion != contracts.ReducerBindingSchemaV1 {
			return nil, fmt.Errorf("UNKNOWN_SCHEMA: reducer binding %s uses %q", path, binding.SchemaVersion)
		}
		if err := registry.Validate(binding.SchemaVersion, raw); err != nil {
			return nil, fmt.Errorf("reducer binding %s: %w", path, err)
		}
		if binding.StateModel != reducer.ModelExclusiveV1 {
			return nil, fmt.Errorf("UNKNOWN_REDUCER_MODEL: binding %q uses %q", binding.BindingID, binding.StateModel)
		}
		if binding.EventSchema != contracts.ExclusiveRecordEventSchemaV1 {
			return nil, fmt.Errorf("REDUCER_EVENT_SCHEMA_MISMATCH: binding %q uses %q", binding.BindingID, binding.EventSchema)
		}
		if _, err := registry.Compile(binding.EventSchema); err != nil {
			return nil, fmt.Errorf("reducer binding %q event schema: %w", binding.BindingID, err)
		}
		if prior, exists := snapshot.ByBindingID[binding.BindingID]; exists {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: duplicate reducer binding_id %q for record kinds %q and %q", binding.BindingID, prior.RecordKind, binding.RecordKind)
		}
		if prior, exists := snapshot.ByRecordKind[binding.RecordKind]; exists {
			return nil, fmt.Errorf("INTEGRITY_FAILURE: record_kind %q has multiple reducer bindings %q and %q", binding.RecordKind, prior.BindingID, binding.BindingID)
		}
		snapshot.ByBindingID[binding.BindingID] = binding
		snapshot.ByRecordKind[binding.RecordKind] = binding
	}
	return snapshot, nil
}

func (s *ReducerBindingSnapshot) ReducerBindings() reducer.Bindings {
	out := make(reducer.Bindings, len(s.ByRecordKind))
	for kind, binding := range s.ByRecordKind {
		out[kind] = binding.StateModel
	}
	return out
}
