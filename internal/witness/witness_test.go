package witness

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestWitnessSignVerify(t *testing.T) {
	pub,priv,err:=ed25519.GenerateKey(rand.Reader); if err!=nil{t.Fatal(err)}
	e,err:=Sign(priv,Statement{LedgerID:"ledger:a",HeadCommit:"head",ProjectionSHA256:"digest",WitnessedAt:"2026-08-12T16:00:00Z"}); if err!=nil{t.Fatal(err)}
	if err:=Verify(pub,e); err!=nil{t.Fatal(err)}
	e.Statement.HeadCommit="changed"; if err:=Verify(pub,e); err==nil{t.Fatal("expected signature failure")}
}
