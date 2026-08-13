package checkpoint

import "testing"

func TestCheckpointVerification(t *testing.T) {
	c, err := Build("abc", 4, []byte(`{"x":1}`), []byte(`["s1"]`), []byte(`{"k":"v"}`)); if err != nil { t.Fatal(err) }
	if err := Verify(c, []byte(`{"x":1}`), []byte(`["s1"]`), []byte(`{"k":"v"}`)); err != nil { t.Fatal(err) }
	if err := Verify(c, []byte(`{"x":2}`), []byte(`["s1"]`), []byte(`{"k":"v"}`)); err == nil { t.Fatal("expected mismatch") }
}
