package reducer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

const (
	ModelExclusiveV1 = "exclusive-governed-record-v1"
	EventCreated     = "core.record.created"
	EventReplaced    = "core.record.replaced"
	EventRevoked     = "core.record.revoked"
	StatusActive     = "active"
	StatusRevoked    = "revoked"
)

var (
	ErrNotApplicable          = errors.New("REDUCER_NOT_APPLICABLE")
	ErrUnknownReducerEvent    = errors.New("UNKNOWN_REDUCER_EVENT")
	ErrPolicyUnbound          = errors.New("REDUCER_POLICY_UNBOUND")
	ErrTargetCardinality      = errors.New("REDUCER_TARGET_CARDINALITY")
	ErrAlreadyExists          = errors.New("REDUCER_ALREADY_EXISTS")
	ErrNotFound               = errors.New("REDUCER_NOT_FOUND")
	ErrTerminalState          = errors.New("REDUCER_TERMINAL_STATE")
	ErrRecordKindMismatch     = errors.New("REDUCER_RECORD_KIND_MISMATCH")
	ErrPriorStateMismatch     = errors.New("REDUCER_PRIOR_STATE_MISMATCH")
	ErrResultingStateMismatch = errors.New("REDUCER_RESULTING_STATE_MISMATCH")
	ErrTransitionInvalid      = errors.New("REDUCER_TRANSITION_INVALID")
)

// Event is the reducer-facing view of an already validated durable event.
// The ledger/event-schema layer is responsible for mapping a durable event
// into this structure. Reducer semantics remain independent of the writer.
type Event struct {
	EventID        string
	EventType      string
	Targets        []string
	RecordKind     string
	Value          json.RawMessage
	PriorState     json.RawMessage
	ResultingState json.RawMessage
}

type State struct {
	TargetID        string          `json:"target_id"`
	RecordKind      string          `json:"record_kind"`
	Status          string          `json:"status"`
	Revision        uint64          `json:"revision"`
	CurrentEventID  string          `json:"current_event_id"`
	PreviousEventID *string         `json:"previous_event_id"`
	Value           json.RawMessage `json:"value,omitempty"`
}

type Projection map[string]State

// Bindings maps record_kind -> accepted state model name.
type Bindings map[string]string

func Apply(current Projection, bindings Bindings, event Event) (Projection, error) {
	if !strings.HasPrefix(event.EventType, "core.record.") {
		return nil, ErrNotApplicable
	}
	if event.EventType != EventCreated && event.EventType != EventReplaced && event.EventType != EventRevoked {
		return nil, fmt.Errorf("%w: %s", ErrUnknownReducerEvent, event.EventType)
	}
	if event.EventID == "" {
		return nil, fmt.Errorf("%w: event_id is required", ErrTransitionInvalid)
	}
	if len(event.Targets) != 1 || event.Targets[0] == "" {
		return nil, fmt.Errorf("%w: expected exactly one non-empty target", ErrTargetCardinality)
	}
	target := event.Targets[0]

	if state, ok := current[target]; ok {
		if err := validateState(target, state); err != nil {
			return nil, err
		}
	}

	switch event.EventType {
	case EventCreated:
		return applyCreate(current, bindings, target, event)
	case EventReplaced:
		return applyReplace(current, bindings, target, event)
	case EventRevoked:
		return applyRevoke(current, bindings, target, event)
	default:
		panic("unreachable reducer event dispatch")
	}
}

func applyCreate(current Projection, bindings Bindings, target string, event Event) (Projection, error) {
	if _, exists := current[target]; exists {
		return nil, fmt.Errorf("%w: target %q", ErrAlreadyExists, target)
	}
	if event.RecordKind == "" {
		return nil, fmt.Errorf("%w: record_kind is required", ErrTransitionInvalid)
	}
	if err := requireBinding(bindings, event.RecordKind); err != nil {
		return nil, err
	}
	value, err := requireValue(event.Value)
	if err != nil {
		return nil, err
	}
	absent := struct {
		Exists   bool   `json:"exists"`
		TargetID string `json:"target_id"`
	}{false, target}
	if err := requireAssertion(event.PriorState, absent, ErrPriorStateMismatch); err != nil {
		return nil, err
	}
	result := State{
		TargetID:        target,
		RecordKind:      event.RecordKind,
		Status:          StatusActive,
		Revision:        1,
		CurrentEventID:  event.EventID,
		PreviousEventID: nil,
		Value:           value,
	}
	if err := requireAssertion(event.ResultingState, result, ErrResultingStateMismatch); err != nil {
		return nil, err
	}
	next := cloneProjection(current)
	next[target] = result
	return next, nil
}

func applyReplace(current Projection, bindings Bindings, target string, event Event) (Projection, error) {
	state, exists := current[target]
	if !exists {
		return nil, fmt.Errorf("%w: target %q", ErrNotFound, target)
	}
	if state.Status == StatusRevoked {
		return nil, fmt.Errorf("%w: target %q is revoked", ErrTerminalState, target)
	}
	if event.RecordKind != state.RecordKind {
		return nil, fmt.Errorf("%w: target %q has kind %q, event has %q", ErrRecordKindMismatch, target, state.RecordKind, event.RecordKind)
	}
	if err := requireBinding(bindings, state.RecordKind); err != nil {
		return nil, err
	}
	value, err := requireValue(event.Value)
	if err != nil {
		return nil, err
	}
	if err := requireAssertion(event.PriorState, state, ErrPriorStateMismatch); err != nil {
		return nil, err
	}
	previous := state.CurrentEventID
	result := State{
		TargetID:        target,
		RecordKind:      state.RecordKind,
		Status:          StatusActive,
		Revision:        state.Revision + 1,
		CurrentEventID:  event.EventID,
		PreviousEventID: &previous,
		Value:           value,
	}
	if err := requireAssertion(event.ResultingState, result, ErrResultingStateMismatch); err != nil {
		return nil, err
	}
	next := cloneProjection(current)
	next[target] = result
	return next, nil
}

func applyRevoke(current Projection, bindings Bindings, target string, event Event) (Projection, error) {
	state, exists := current[target]
	if !exists {
		return nil, fmt.Errorf("%w: target %q", ErrNotFound, target)
	}
	if state.Status == StatusRevoked {
		return nil, fmt.Errorf("%w: target %q is revoked", ErrTerminalState, target)
	}
	if event.RecordKind != state.RecordKind {
		return nil, fmt.Errorf("%w: target %q has kind %q, event has %q", ErrRecordKindMismatch, target, state.RecordKind, event.RecordKind)
	}
	if err := requireBinding(bindings, state.RecordKind); err != nil {
		return nil, err
	}
	if len(event.Value) != 0 {
		return nil, fmt.Errorf("%w: revoke must not carry a value", ErrTransitionInvalid)
	}
	if err := requireAssertion(event.PriorState, state, ErrPriorStateMismatch); err != nil {
		return nil, err
	}
	previous := state.CurrentEventID
	result := State{
		TargetID:        target,
		RecordKind:      state.RecordKind,
		Status:          StatusRevoked,
		Revision:        state.Revision + 1,
		CurrentEventID:  event.EventID,
		PreviousEventID: &previous,
	}
	if err := requireAssertion(event.ResultingState, result, ErrResultingStateMismatch); err != nil {
		return nil, err
	}
	next := cloneProjection(current)
	next[target] = result
	return next, nil
}

func requireBinding(bindings Bindings, recordKind string) error {
	if bindings == nil || bindings[recordKind] != ModelExclusiveV1 {
		return fmt.Errorf("%w: record_kind %q is not bound to %s", ErrPolicyUnbound, recordKind, ModelExclusiveV1)
	}
	return nil
}

func requireValue(raw json.RawMessage) (json.RawMessage, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("%w: transition value is required", ErrTransitionInvalid)
	}
	if err := strictjson.Validate(raw); err != nil {
		return nil, fmt.Errorf("%w: invalid transition value: %v", ErrTransitionInvalid, err)
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return nil, fmt.Errorf("%w: canonicalize transition value: %v", ErrTransitionInvalid, err)
	}
	return append(json.RawMessage(nil), canonical...), nil
}

func requireAssertion(raw json.RawMessage, expected any, mismatch error) error {
	if len(raw) == 0 {
		return fmt.Errorf("%w: assertion is required", mismatch)
	}
	if err := strictjson.Validate(raw); err != nil {
		return fmt.Errorf("%w: invalid assertion JSON: %v", mismatch, err)
	}
	got, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return fmt.Errorf("%w: canonicalize assertion: %v", mismatch, err)
	}
	expectedRaw, err := json.Marshal(expected)
	if err != nil {
		return fmt.Errorf("%w: marshal expected assertion: %v", ErrTransitionInvalid, err)
	}
	want, err := canonicaljson.Canonicalize(expectedRaw)
	if err != nil {
		return fmt.Errorf("%w: canonicalize expected assertion: %v", ErrTransitionInvalid, err)
	}
	if !bytes.Equal(got, want) {
		return fmt.Errorf("%w: got %s want %s", mismatch, got, want)
	}
	return nil
}

func validateState(key string, state State) error {
	if state.TargetID != key || state.TargetID == "" || state.RecordKind == "" || state.Revision == 0 || state.CurrentEventID == "" {
		return fmt.Errorf("%w: invalid existing state for target %q", ErrTransitionInvalid, key)
	}
	switch state.Status {
	case StatusActive:
		if len(state.Value) == 0 {
			return fmt.Errorf("%w: active target %q has no value", ErrTransitionInvalid, key)
		}
		if _, err := requireValue(state.Value); err != nil {
			return err
		}
	case StatusRevoked:
		if len(state.Value) != 0 {
			return fmt.Errorf("%w: revoked target %q retains a value", ErrTransitionInvalid, key)
		}
	default:
		return fmt.Errorf("%w: target %q has unknown status %q", ErrTransitionInvalid, key, state.Status)
	}
	return nil
}

func cloneProjection(src Projection) Projection {
	dst := make(Projection, len(src))
	for key, state := range src {
		copyState := state
		if state.PreviousEventID != nil {
			previous := *state.PreviousEventID
			copyState.PreviousEventID = &previous
		}
		copyState.Value = append(json.RawMessage(nil), state.Value...)
		dst[key] = copyState
	}
	return dst
}

func CanonicalProjection(projection Projection) ([]byte, error) {
	raw, err := json.Marshal(projection)
	if err != nil {
		return nil, fmt.Errorf("marshal reducer projection: %w", err)
	}
	return canonicaljson.Canonicalize(raw)
}
