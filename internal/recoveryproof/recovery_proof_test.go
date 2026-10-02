package recoveryproof

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/ledger"
)

func validRecoveryProof() ledger.RecoveryProof {
	return ledger.RecoveryProof{
		LedgerCommit:                 strings.Repeat("1", 40),
		AuthoritativeRef:             "refs/heads/main",
		GitObjectFormat:              "sha1",
		GenesisCommit:                strings.Repeat("2", 40),
		ProjectID:                    "project:test",
		LedgerID:                     "ledger:test",
		GenesisContentSHA256:         strings.Repeat("3", 64),
		ActorPolicyVersion:           "authority-policy:test:v1",
		ActorPolicyRootContentSHA256: strings.Repeat("4", 64),
		HistoryCommitCount:           1,
		EventCount:                   0,
		ReducerBindingCount:          1,
		GovernedRecordCount:          0,
		GovernedRecordsSHA256:        strings.Repeat("5", 64),
		ReplaySHA256:                 strings.Repeat("6", 64),
	}
}

func TestRequiredFieldsMatchRecoveryProofJSONShape(t *testing.T) {
	typeOfProof := reflect.TypeOf(ledger.RecoveryProof{})
	want := make([]string, 0, typeOfProof.NumField())
	for i := 0; i < typeOfProof.NumField(); i++ {
		name := strings.Split(typeOfProof.Field(i).Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			t.Fatalf("RecoveryProof field %s has no required JSON name", typeOfProof.Field(i).Name)
		}
		want = append(want, name)
	}
	got := append([]string(nil), requiredFields...)
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("required RecoveryProof fields drifted: decoder=%v struct=%v", got, want)
	}
}

func TestDecodeAcceptsCompleteRecoveryProof(t *testing.T) {
	raw, err := json.Marshal(validRecoveryProof())
	if err != nil {
		t.Fatal(err)
	}
	got, err := Decode(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got != validRecoveryProof() {
		t.Fatalf("decoded proof changed: %#v", got)
	}
}

func TestDecodeRejectsEveryMissingOrNullRequiredField(t *testing.T) {
	raw, err := json.Marshal(validRecoveryProof())
	if err != nil {
		t.Fatal(err)
	}
	var base map[string]any
	if err := json.Unmarshal(raw, &base); err != nil {
		t.Fatal(err)
	}
	for _, field := range requiredFields {
		t.Run("missing_"+field, func(t *testing.T) {
			value := cloneMap(base)
			delete(value, field)
			candidate, _ := json.Marshal(value)
			if _, err := Decode(candidate); err == nil || !strings.Contains(err.Error(), "RECOVERY_PROOF_INVALID") {
				t.Fatalf("missing %s accepted: %v", field, err)
			}
		})
		t.Run("null_"+field, func(t *testing.T) {
			value := cloneMap(base)
			value[field] = nil
			candidate, _ := json.Marshal(value)
			if _, err := Decode(candidate); err == nil || !strings.Contains(err.Error(), "RECOVERY_PROOF_INVALID") {
				t.Fatalf("null %s accepted: %v", field, err)
			}
		})
	}
}

func TestDecodeRejectsUnknownDuplicateTrailingAndInvalidCounts(t *testing.T) {
	raw, err := json.Marshal(validRecoveryProof())
	if err != nil {
		t.Fatal(err)
	}
	var base map[string]any
	if err := json.Unmarshal(raw, &base); err != nil {
		t.Fatal(err)
	}

	unknown := cloneMap(base)
	unknown["unexpected"] = true
	unknownRaw, _ := json.Marshal(unknown)

	duplicate := []byte(strings.TrimSuffix(string(raw), "}") + `,"ledger_id":"ledger:other"}`)
	trailing := append(append([]byte{}, raw...), []byte(` {}`)...)

	invalidHistory := cloneMap(base)
	invalidHistory["history_commit_count"] = float64(0)
	invalidHistoryRaw, _ := json.Marshal(invalidHistory)

	invalidEvents := cloneMap(base)
	invalidEvents["event_count"] = float64(-1)
	invalidEventsRaw, _ := json.Marshal(invalidEvents)

	cases := map[string][]byte{
		"unknown":         unknownRaw,
		"duplicate":       duplicate,
		"trailing":        trailing,
		"zero_history":    invalidHistoryRaw,
		"negative_events": invalidEventsRaw,
	}
	for name, candidate := range cases {
		t.Run(name, func(t *testing.T) {
			if _, err := Decode(candidate); err == nil || !strings.Contains(err.Error(), "RECOVERY_PROOF_INVALID") {
				t.Fatalf("%s input accepted: %v", name, err)
			}
		})
	}
}

func cloneMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}