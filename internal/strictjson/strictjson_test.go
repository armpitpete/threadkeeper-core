package strictjson

import "testing"

func TestRejectsDuplicateRootAndNested(t *testing.T) {
	for _, raw := range [][]byte{
		[]byte(`{"event_id":"A","event_id":"B"}`),
		[]byte(`{"outer":{"x":1,"x":2}}`),
	} {
		if err := Validate(raw); !IsCode(err, DuplicateMember) {
			t.Fatalf("expected DUPLICATE_MEMBER, got %v", err)
		}
	}
}

func TestRejectsInvalidUTF8(t *testing.T) {
	raw := []byte{'{', '"', 'x', '"', ':', '"', 0xff, '"', '}'}
	if err := Validate(raw); !IsCode(err, InvalidUTF8) {
		t.Fatalf("expected INVALID_UTF8, got %v", err)
	}
}

func TestRejectsNegativeZeroForms(t *testing.T) {
	for _, s := range []string{`-0`, `-0.0`, `-0e3`, `-0.000E-9`} {
		raw := []byte(`{"n":` + s + `}`)
		if err := Validate(raw); !IsCode(err, InvalidNumber) {
			t.Fatalf("%s: expected INVALID_NUMBER, got %v", s, err)
		}
	}
}

func TestAllowsNegativeNonZero(t *testing.T) {
	for _, s := range []string{`-1`, `-0.01`, `-1e-999`} {
		if err := Validate([]byte(`{"n":` + s + `}`)); err != nil {
			t.Fatalf("%s: unexpected error: %v", s, err)
		}
	}
}

func TestRejectsTrailingJSON(t *testing.T) {
	if err := Validate([]byte(`{} {}`)); !IsCode(err, InvalidJSON) {
		t.Fatalf("expected INVALID_JSON, got %v", err)
	}
}
