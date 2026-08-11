package canonicaljson

import (
	"bytes"
	"os"
	"testing"
)

func TestGoldenObjectOrder(t *testing.T) {
	input, err := os.ReadFile("../../testdata/jcs/object-order.input.json")
	if err != nil { t.Fatal(err) }
	want, err := os.ReadFile("../../testdata/jcs/object-order.expected.json")
	if err != nil { t.Fatal(err) }
	want = bytes.TrimSpace(want)
	got, err := Canonicalize(input)
	if err != nil { t.Fatal(err) }
	if !bytes.Equal(got, want) {
		t.Fatalf("unexpected canonical bytes: got %q want %q", got, want)
	}
	second, err := Canonicalize([]byte(`{"a":2,"b":1}`))
	if err != nil { t.Fatal(err) }
	if !bytes.Equal(got, second) { t.Fatal("object order changed canonical result") }
}

func TestArrayOrderSignificant(t *testing.T) {
	a, err := Canonicalize([]byte(`[1,2]`)); if err != nil { t.Fatal(err) }
	b, err := Canonicalize([]byte(`[2,1]`)); if err != nil { t.Fatal(err) }
	if bytes.Equal(a, b) { t.Fatal("array order was lost") }
}

func TestUnicodeNotNormalized(t *testing.T) {
	a, err := Canonicalize([]byte(`{"x":"é"}`)); if err != nil { t.Fatal(err) }
	b, err := Canonicalize([]byte("{\"x\":\"e\\u0301\"}")); if err != nil { t.Fatal(err) }
	if bytes.Equal(a, b) { t.Fatal("distinct Unicode sequences were normalized") }
}

func TestNullAndOmissionRemainDistinct(t *testing.T) {
	a, err := Canonicalize([]byte(`{"x":null}`)); if err != nil { t.Fatal(err) }
	b, err := Canonicalize([]byte(`{}`)); if err != nil { t.Fatal(err) }
	if bytes.Equal(a, b) { t.Fatal("null collapsed into omission") }
}
