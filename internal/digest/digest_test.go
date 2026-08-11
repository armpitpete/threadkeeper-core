package digest

import (
	"bytes"
	"testing"
)

func TestDigestBoundaryAndCanonicalStoredRecord(t *testing.T) {
	raw := []byte(`{"z":2,"event_id":"E1","content_sha256":"wrong","a":1}`)
	completed, d, err := Complete(raw); if err != nil { t.Fatal(err) }
	if d == "" { t.Fatal("empty digest") }
	if bytes.Contains(completed, []byte(`"content_sha256":"wrong"`)) { t.Fatal("old digest survived") }
	if err := Verify(completed); err != nil { t.Fatalf("verify: %v", err) }
	completed2, _, err := Complete(completed); if err != nil { t.Fatal(err) }
	if !bytes.Equal(completed, completed2) { t.Fatalf("stored record not stable:\n%s\n%s", completed, completed2) }
}

func TestIncludingDigestChangesPayload(t *testing.T) {
	raw := []byte(`{"event_id":"E1"}`)
	completed, _, err := Complete(raw); if err != nil { t.Fatal(err) }
	without, _, err := Compute(completed); if err != nil { t.Fatal(err) }
	if without == "" { t.Fatal("empty digest") }
}
