package reducer

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

const testKind = "project.setting"

var testBindings = Bindings{testKind: ModelExclusiveV1}

func TestCreateReplaceRevokeLifecycle(t *testing.T) {
	projection := Projection{}

	created := State{
		TargetID:        "setting:one",
		RecordKind:      testKind,
		Status:          StatusActive,
		Revision:        1,
		CurrentEventID:  "event-1",
		PreviousEventID: nil,
		Value:           raw(`{"enabled":true}`),
	}
	var err error
	projection, err = Apply(projection, testBindings, Event{
		EventID:        "event-1",
		EventType:      EventCreated,
		Targets:        []string{"setting:one"},
		RecordKind:     testKind,
		Value:          raw(`{"enabled":true}`),
		PriorState:     assertion(t, map[string]any{"exists": false, "target_id": "setting:one"}),
		ResultingState: assertion(t, created),
	})
	if err != nil {
		t.Fatal(err)
	}
	state1 := projection["setting:one"]
	if state1.Revision != 1 || state1.Status != StatusActive || state1.PreviousEventID != nil {
		t.Fatalf("unexpected created state: %#v", state1)
	}

	previous1 := "event-1"
	replaced := State{
		TargetID:        "setting:one",
		RecordKind:      testKind,
		Status:          StatusActive,
		Revision:        2,
		CurrentEventID:  "event-2",
		PreviousEventID: &previous1,
		Value:           raw(`{"enabled":false}`),
	}
	projection, err = Apply(projection, testBindings, Event{
		EventID:        "event-2",
		EventType:      EventReplaced,
		Targets:        []string{"setting:one"},
		RecordKind:     testKind,
		Value:          raw(`{"enabled":false}`),
		PriorState:     assertion(t, state1),
		ResultingState: assertion(t, replaced),
	})
	if err != nil {
		t.Fatal(err)
	}
	state2 := projection["setting:one"]
	if state2.Revision != 2 || state2.PreviousEventID == nil || *state2.PreviousEventID != "event-1" {
		t.Fatalf("unexpected replaced state: %#v", state2)
	}

	previous2 := "event-2"
	revoked := State{
		TargetID:        "setting:one",
		RecordKind:      testKind,
		Status:          StatusRevoked,
		Revision:        3,
		CurrentEventID:  "event-3",
		PreviousEventID: &previous2,
	}
	projection, err = Apply(projection, testBindings, Event{
		EventID:        "event-3",
		EventType:      EventRevoked,
		Targets:        []string{"setting:one"},
		RecordKind:     testKind,
		PriorState:     assertion(t, state2),
		ResultingState: assertion(t, revoked),
	})
	if err != nil {
		t.Fatal(err)
	}
	state3 := projection["setting:one"]
	if state3.Revision != 3 || state3.Status != StatusRevoked || len(state3.Value) != 0 {
		t.Fatalf("unexpected revoked state: %#v", state3)
	}
}

func TestDuplicateCreateFails(t *testing.T) {
	projection, state := oneCreated(t, raw(`1`))
	_, err := Apply(projection, testBindings, Event{
		EventID:        "event-2",
		EventType:      EventCreated,
		Targets:        []string{state.TargetID},
		RecordKind:     testKind,
		Value:          raw(`2`),
		PriorState:     assertion(t, map[string]any{"exists": false, "target_id": state.TargetID}),
		ResultingState: assertion(t, state),
	})
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected already-exists, got %v", err)
	}
}

func TestMissingTargetTransitionsFail(t *testing.T) {
	for _, eventType := range []string{EventReplaced, EventRevoked} {
		_, err := Apply(Projection{}, testBindings, Event{EventID: "e", EventType: eventType, Targets: []string{"missing"}, RecordKind: testKind})
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("%s: expected not-found, got %v", eventType, err)
		}
	}
}

func TestRevocationIsTerminalAndTargetCannotBeRecycled(t *testing.T) {
	projection, state1 := oneCreated(t, raw(`1`))
	previous := state1.CurrentEventID
	revoked := State{TargetID: state1.TargetID, RecordKind: testKind, Status: StatusRevoked, Revision: 2, CurrentEventID: "event-2", PreviousEventID: &previous}
	projection, err := Apply(projection, testBindings, Event{
		EventID: "event-2", EventType: EventRevoked, Targets: []string{state1.TargetID}, RecordKind: testKind,
		PriorState: assertion(t, state1), ResultingState: assertion(t, revoked),
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, event := range []Event{
		{EventID: "event-3", EventType: EventReplaced, Targets: []string{state1.TargetID}, RecordKind: testKind, Value: raw(`2`)},
		{EventID: "event-3", EventType: EventRevoked, Targets: []string{state1.TargetID}, RecordKind: testKind},
	} {
		_, err := Apply(projection, testBindings, event)
		if !errors.Is(err, ErrTerminalState) {
			t.Fatalf("%s: expected terminal-state, got %v", event.EventType, err)
		}
	}

	_, err = Apply(projection, testBindings, Event{
		EventID: "event-3", EventType: EventCreated, Targets: []string{state1.TargetID}, RecordKind: testKind, Value: raw(`2`),
		PriorState: assertion(t, map[string]any{"exists": false, "target_id": state1.TargetID}), ResultingState: assertion(t, state1),
	})
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected target-recycle rejection, got %v", err)
	}
}

func TestRecordKindCannotChange(t *testing.T) {
	projection, state := oneCreated(t, raw(`1`))
	_, err := Apply(projection, Bindings{testKind: ModelExclusiveV1, "other": ModelExclusiveV1}, Event{
		EventID: "event-2", EventType: EventReplaced, Targets: []string{state.TargetID}, RecordKind: "other", Value: raw(`2`),
		PriorState: assertion(t, state), ResultingState: assertion(t, state),
	})
	if !errors.Is(err, ErrRecordKindMismatch) {
		t.Fatalf("expected kind mismatch, got %v", err)
	}
}

func TestPriorAndResultingAssertionsAreChecked(t *testing.T) {
	projection, state := oneCreated(t, raw(`1`))
	previous := state.CurrentEventID
	result := State{TargetID: state.TargetID, RecordKind: testKind, Status: StatusActive, Revision: 2, CurrentEventID: "event-2", PreviousEventID: &previous, Value: raw(`2`)}

	_, err := Apply(projection, testBindings, Event{
		EventID: "event-2", EventType: EventReplaced, Targets: []string{state.TargetID}, RecordKind: testKind, Value: raw(`2`),
		PriorState: assertion(t, map[string]any{"wrong": true}), ResultingState: assertion(t, result),
	})
	if !errors.Is(err, ErrPriorStateMismatch) {
		t.Fatalf("expected prior mismatch, got %v", err)
	}

	_, err = Apply(projection, testBindings, Event{
		EventID: "event-2", EventType: EventReplaced, Targets: []string{state.TargetID}, RecordKind: testKind, Value: raw(`2`),
		PriorState: assertion(t, state), ResultingState: assertion(t, state),
	})
	if !errors.Is(err, ErrResultingStateMismatch) {
		t.Fatalf("expected resulting mismatch, got %v", err)
	}
}

func TestTargetCardinality(t *testing.T) {
	for _, targets := range [][]string{nil, {}, {"one", "two"}, {""}} {
		_, err := Apply(Projection{}, testBindings, Event{EventID: "event-1", EventType: EventCreated, Targets: targets, RecordKind: testKind, Value: raw(`1`)})
		if !errors.Is(err, ErrTargetCardinality) {
			t.Fatalf("targets %#v: expected cardinality failure, got %v", targets, err)
		}
	}
}

func TestPolicyBindingRequired(t *testing.T) {
	_, err := Apply(Projection{}, Bindings{}, Event{
		EventID: "event-1", EventType: EventCreated, Targets: []string{"one"}, RecordKind: testKind, Value: raw(`1`),
		PriorState: assertion(t, map[string]any{"exists": false, "target_id": "one"}),
	})
	if !errors.Is(err, ErrPolicyUnbound) {
		t.Fatalf("expected policy-unbound, got %v", err)
	}
}

func TestUnknownAndUnrelatedEvents(t *testing.T) {
	_, err := Apply(Projection{}, testBindings, Event{EventID: "event-1", EventType: "core.record.reopened", Targets: []string{"one"}})
	if !errors.Is(err, ErrUnknownReducerEvent) {
		t.Fatalf("expected unknown reducer event, got %v", err)
	}
	_, err = Apply(Projection{}, testBindings, Event{EventID: "event-1", EventType: "evidence.observed", Targets: []string{"one"}})
	if !errors.Is(err, ErrNotApplicable) {
		t.Fatalf("expected not-applicable, got %v", err)
	}
}

func TestExplicitNullIsActiveValue(t *testing.T) {
	projection, state := oneCreated(t, raw(`null`))
	got := projection[state.TargetID]
	if got.Status != StatusActive || string(got.Value) != "null" {
		t.Fatalf("explicit null was not preserved as active value: %#v", got)
	}
	encoded := assertion(t, got)
	if !bytes.Contains(encoded, []byte(`"value":null`)) {
		t.Fatalf("active null value missing from canonical state: %s", encoded)
	}
}

func TestSameValueReplacementIsAllowed(t *testing.T) {
	projection, state1 := oneCreated(t, raw(`{"x":1}`))
	previous := state1.CurrentEventID
	state2 := State{TargetID: state1.TargetID, RecordKind: testKind, Status: StatusActive, Revision: 2, CurrentEventID: "event-2", PreviousEventID: &previous, Value: raw(`{"x":1}`)}
	projection, err := Apply(projection, testBindings, Event{
		EventID: "event-2", EventType: EventReplaced, Targets: []string{state1.TargetID}, RecordKind: testKind, Value: raw(`{"x":1}`),
		PriorState: assertion(t, state1), ResultingState: assertion(t, state2),
	})
	if err != nil {
		t.Fatal(err)
	}
	if projection[state1.TargetID].Revision != 2 {
		t.Fatalf("same-value replacement did not advance revision")
	}
}

func TestFailureDoesNotMutateProjection(t *testing.T) {
	projection, state := oneCreated(t, raw(`1`))
	before, err := CanonicalProjection(projection)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Apply(projection, testBindings, Event{
		EventID: "event-2", EventType: EventReplaced, Targets: []string{state.TargetID}, RecordKind: testKind, Value: raw(`2`),
		PriorState: assertion(t, map[string]any{"wrong": true}), ResultingState: assertion(t, state),
	})
	if !errors.Is(err, ErrPriorStateMismatch) {
		t.Fatalf("expected prior mismatch, got %v", err)
	}
	after, err := CanonicalProjection(projection)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, after) {
		t.Fatalf("failed transition mutated input projection\nbefore: %s\nafter: %s", before, after)
	}
}

func TestEquivalentReductionsAreDeterministic(t *testing.T) {
	makeProjection := func() Projection {
		projection, _ := oneCreated(t, raw(`{"b":2,"a":1}`))
		return projection
	}
	first, err := CanonicalProjection(makeProjection())
	if err != nil {
		t.Fatal(err)
	}
	second, err := CanonicalProjection(makeProjection())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Fatalf("equivalent reductions differ\n%s\n%s", first, second)
	}
}

func oneCreated(t *testing.T, value json.RawMessage) (Projection, State) {
	t.Helper()
	canonicalValue := canonical(t, value)
	state := State{
		TargetID: "target:one", RecordKind: testKind, Status: StatusActive, Revision: 1,
		CurrentEventID: "event-1", PreviousEventID: nil, Value: canonicalValue,
	}
	projection, err := Apply(Projection{}, testBindings, Event{
		EventID: "event-1", EventType: EventCreated, Targets: []string{"target:one"}, RecordKind: testKind, Value: value,
		PriorState: assertion(t, map[string]any{"exists": false, "target_id": "target:one"}), ResultingState: assertion(t, state),
	})
	if err != nil {
		t.Fatal(err)
	}
	return projection, projection["target:one"]
}

func assertion(t *testing.T, value any) json.RawMessage {
	t.Helper()
	rawBytes, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return canonical(t, rawBytes)
}

func canonical(t *testing.T, value json.RawMessage) json.RawMessage {
	t.Helper()
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		t.Fatal(err)
	}
	rawBytes, err := json.Marshal(decoded)
	if err != nil {
		t.Fatal(err)
	}
	return rawBytes
}

func raw(value string) json.RawMessage { return json.RawMessage(value) }
