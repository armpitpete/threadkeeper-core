package actorauth

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/armpitpete/threadkeeper-core/internal/digest"
)

func TestPolicyDocumentRejectsUnknownField(t *testing.T) {
	doc := validPolicyDocumentValue(t)
	doc["ambient_override"] = true
	_, _, err := ParsePolicyDocument(completePolicyJSON(t, doc), "ledger:test")
	if err == nil || !strings.Contains(err.Error(), "unknown field") {
		t.Fatalf("unknown policy field = %v", err)
	}
}

func TestPolicyDocumentRejectsUnsortedKeys(t *testing.T) {
	doc := validPolicyDocumentValue(t)
	keys := doc["keys"].([]map[string]any)
	second := clonePolicyKey(keys[0])
	second["key_id"] = "key:aaa"
	doc["keys"] = []map[string]any{keys[0], second}
	_, _, err := ParsePolicyDocument(completePolicyJSON(t, doc), "ledger:test")
	if err == nil || !strings.Contains(err.Error(), "keys must be sorted") {
		t.Fatalf("unsorted keys = %v", err)
	}
}

func TestPolicyDocumentRejectsDuplicateGrant(t *testing.T) {
	doc := validPolicyDocumentValue(t)
	grant := doc["grants"].([]map[string]any)[0]
	doc["grants"] = []map[string]any{grant, clonePolicyGrant(grant)}
	_, _, err := ParsePolicyDocument(completePolicyJSON(t, doc), "ledger:test")
	if err == nil || !strings.Contains(err.Error(), "duplicate actor grant") {
		t.Fatalf("duplicate grant = %v", err)
	}
}

func TestPolicyDocumentRejectsMalformedPublicKey(t *testing.T) {
	doc := validPolicyDocumentValue(t)
	doc["keys"].([]map[string]any)[0]["public_key"] = base64.RawStdEncoding.EncodeToString([]byte("too short"))
	_, _, err := ParsePolicyDocument(completePolicyJSON(t, doc), "ledger:test")
	if err == nil || !strings.Contains(err.Error(), "Ed25519 public key") {
		t.Fatalf("malformed public key = %v", err)
	}
}

func TestPolicyDocumentRejectsGrantWithoutActiveKey(t *testing.T) {
	doc := validPolicyDocumentValue(t)
	doc["keys"].([]map[string]any)[0]["revoked"] = true
	_, _, err := ParsePolicyDocument(completePolicyJSON(t, doc), "ledger:test")
	if err == nil || !strings.Contains(err.Error(), "no active trusted key") {
		t.Fatalf("grant without active key = %v", err)
	}
}

func TestPolicyDocumentInitialAuthoritiesMustMatchGrantedActors(t *testing.T) {
	raw := completePolicyJSON(t, validPolicyDocumentValue(t))
	doc, _, err := ParsePolicyDocument(raw, "ledger:test")
	if err != nil {
		t.Fatal(err)
	}
	if err := ValidateInitialAuthorities(doc, []string{"actor:other"}); err == nil || !strings.Contains(err.Error(), "initial_authorities") {
		t.Fatalf("authority mismatch = %v", err)
	}
}

func validPolicyDocumentValue(t *testing.T) map[string]any {
	t.Helper()
	seed := make([]byte, ed25519.SeedSize)
	for i := range seed {
		seed[i] = 7
	}
	publicKey := ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey)
	return map[string]any{
		"max_proof_lifetime_seconds": int64(300),
		"keys": []map[string]any{{
			"actor_id":   "actor:owner",
			"key_id":     "key:owner:v1",
			"public_key": base64.RawStdEncoding.EncodeToString(publicKey),
			"revoked":    false,
		}},
		"grants": []map[string]any{{
			"actor_id": "actor:owner",
			"action":   ActionOperate,
			"target":   "target:test",
		}},
	}
}

func completePolicyJSON(t *testing.T, value map[string]any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	completed, _, err := digest.Complete(raw)
	if err != nil {
		t.Fatal(err)
	}
	return completed
}

func clonePolicyKey(src map[string]any) map[string]any {
	out := make(map[string]any, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func clonePolicyGrant(src map[string]any) map[string]any {
	out := make(map[string]any, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}
