package witness

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
)

type Statement struct {
	LedgerID         string `json:"ledger_id"`
	HeadCommit       string `json:"head_commit"`
	ProjectionSHA256 string `json:"projection_sha256"`
	WitnessedAt      string `json:"witnessed_at"`
}

type Envelope struct {
	Statement Statement `json:"statement"`
	Signature string    `json:"signature"`
}

func Sign(privateKey ed25519.PrivateKey, statement Statement) (Envelope, error) {
	payload, err := canonicalStatement(statement); if err != nil { return Envelope{}, err }
	sig := ed25519.Sign(privateKey, payload)
	return Envelope{Statement: statement, Signature: base64.StdEncoding.EncodeToString(sig)}, nil
}

func Verify(publicKey ed25519.PublicKey, envelope Envelope) error {
	payload, err := canonicalStatement(envelope.Statement); if err != nil { return err }
	sig, err := base64.StdEncoding.DecodeString(envelope.Signature); if err != nil { return fmt.Errorf("WITNESS_INVALID: signature encoding: %w", err) }
	if !ed25519.Verify(publicKey, payload, sig) { return fmt.Errorf("WITNESS_INVALID: signature verification failed") }
	return nil
}

func canonicalStatement(s Statement) ([]byte,error) {
	if s.LedgerID=="" || s.HeadCommit=="" || s.ProjectionSHA256=="" { return nil, fmt.Errorf("WITNESS_INVALID: ledger, head and projection digest required") }
	if _,err:=time.Parse(time.RFC3339,s.WitnessedAt); err!=nil { return nil, fmt.Errorf("WITNESS_INVALID: witnessed_at: %w",err) }
	raw,err:=json.Marshal(s); if err!=nil{return nil,err}; return canonicaljson.Canonicalize(raw)
}
