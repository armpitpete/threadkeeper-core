package escrow

import "testing"

func TestVerifyEscrowSnapshot(t *testing.T) {
	content := []byte("evidence")
	s := Snapshot{SourceID: "source:a", VersionID: "v1", ContentSHA256: HashContent(content), Size: int64(len(content))}
	if err := VerifyContent(s, content); err != nil { t.Fatal(err) }
	if err := VerifyContent(s, []byte("changed")); err == nil { t.Fatal("expected digest/size failure") }
}
