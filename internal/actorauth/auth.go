package actorauth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
)

type Action string

const (
	ActionPropose Action = "propose"
	ActionDecide  Action = "decide"
	ActionOperate Action = "operate"
)

type RequestContext struct {
	LedgerID       string `json:"ledger_id"`
	Action         Action `json:"action"`
	Target         string `json:"target"`
	ExpectedState  string `json:"expected_state"`
	IdempotencyKey string `json:"idempotency_key"`
}

type Proof struct {
	ActorID   string         `json:"actor_id"`
	KeyID     string         `json:"key_id"`
	IssuedAt  string         `json:"issued_at"`
	ExpiresAt string         `json:"expires_at"`
	Context   RequestContext `json:"context"`
	Signature string         `json:"signature"`
}

type KeyBinding struct {
	ActorID   string
	KeyID     string
	PublicKey ed25519.PublicKey
	Revoked   bool
}

type Grant struct {
	ActorID  string
	LedgerID string
	Action   Action
}

type Policy struct {
	Keys             []KeyBinding
	Grants           []Grant
	MaxProofLifetime time.Duration
}

type Principal struct {
	ActorID string `json:"actor_id"`
	KeyID   string `json:"key_id"`
}

func AuthenticateAndAuthorize(policy Policy, proof Proof, expected RequestContext, now time.Time) (Principal, error) {
	if err := validatePolicy(policy); err != nil {
		return Principal{}, err
	}
	if err := expected.Validate(); err != nil {
		return Principal{}, err
	}
	if proof.ActorID == "" || proof.KeyID == "" || proof.Signature == "" {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: actor_id, key_id and signature are required")
	}
	if proof.Context != expected {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: signed request context does not match the operation")
	}

	issued, err := time.Parse(time.RFC3339, proof.IssuedAt)
	if err != nil {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: issued_at: %w", err)
	}
	expires, err := time.Parse(time.RFC3339, proof.ExpiresAt)
	if err != nil {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: expires_at: %w", err)
	}
	if !expires.After(issued) {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: expires_at must be after issued_at")
	}
	if expires.Sub(issued) > policy.MaxProofLifetime {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: proof lifetime exceeds policy maximum")
	}
	if now.Before(issued) || !now.Before(expires) {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: proof is not currently valid")
	}

	key, ok := findKey(policy.Keys, proof.ActorID, proof.KeyID)
	if !ok || key.Revoked {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: actor key is not trusted")
	}
	sig, err := base64.RawStdEncoding.DecodeString(proof.Signature)
	if err != nil || len(sig) != ed25519.SignatureSize {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: signature encoding is invalid")
	}
	message, err := signingBytes(proof)
	if err != nil {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: canonical proof: %w", err)
	}
	if !ed25519.Verify(key.PublicKey, message, sig) {
		return Principal{}, fmt.Errorf("AUTHENTICATION_FAILED: signature verification failed")
	}

	if !hasGrant(policy.Grants, proof.ActorID, expected.LedgerID, expected.Action) {
		return Principal{}, fmt.Errorf("AUTHORIZATION_DENIED: actor %q lacks %q on ledger %q", proof.ActorID, expected.Action, expected.LedgerID)
	}
	return Principal{ActorID: proof.ActorID, KeyID: proof.KeyID}, nil
}

func (c RequestContext) Validate() error {
	if c.LedgerID == "" || c.Target == "" || c.ExpectedState == "" || c.IdempotencyKey == "" {
		return fmt.Errorf("AUTH_CONTEXT_INVALID: ledger_id, target, expected_state and idempotency_key are required")
	}
	if !validAction(c.Action) {
		return fmt.Errorf("AUTH_CONTEXT_INVALID: unsupported action %q", c.Action)
	}
	return nil
}

func validatePolicy(policy Policy) error {
	if policy.MaxProofLifetime <= 0 {
		return fmt.Errorf("AUTH_POLICY_INVALID: max proof lifetime must be positive")
	}
	seen := map[string]struct{}{}
	for _, key := range policy.Keys {
		if key.ActorID == "" || key.KeyID == "" || len(key.PublicKey) != ed25519.PublicKeySize {
			return fmt.Errorf("AUTH_POLICY_INVALID: every key requires actor_id, key_id and an Ed25519 public key")
		}
		identity := key.ActorID + "\x00" + key.KeyID
		if _, exists := seen[identity]; exists {
			return fmt.Errorf("AUTH_POLICY_INVALID: duplicate actor/key binding")
		}
		seen[identity] = struct{}{}
	}
	for _, grant := range policy.Grants {
		if grant.ActorID == "" || grant.LedgerID == "" || !validAction(grant.Action) {
			return fmt.Errorf("AUTH_POLICY_INVALID: every grant requires actor_id, ledger_id and a supported action")
		}
	}
	return nil
}

func validAction(action Action) bool {
	switch action {
	case ActionPropose, ActionDecide, ActionOperate:
		return true
	default:
		return false
	}
}

func findKey(keys []KeyBinding, actorID, keyID string) (KeyBinding, bool) {
	for _, key := range keys {
		if key.ActorID == actorID && key.KeyID == keyID {
			return key, true
		}
	}
	return KeyBinding{}, false
}

func hasGrant(grants []Grant, actorID, ledgerID string, action Action) bool {
	for _, grant := range grants {
		if grant.ActorID == actorID && grant.LedgerID == ledgerID && grant.Action == action {
			return true
		}
	}
	return false
}

func signingBytes(proof Proof) ([]byte, error) {
	proof.Signature = ""
	raw, err := json.Marshal(proof)
	if err != nil {
		return nil, err
	}
	return canonicaljson.Canonicalize(raw)
}
