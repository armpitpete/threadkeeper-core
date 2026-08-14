package actorauth

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/canonicaljson"
	"github.com/armpitpete/threadkeeper-core/internal/digest"
	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

const (
	LedgerPolicyPrefix = "config/authority/actor-policy"
	LedgerPolicyPath   = LedgerPolicyPrefix + "/root.json"
	PolicyTarget       = "authority:actor-policy"
	PolicyRecordKind   = "core.actor-auth-policy-v1"
)

type PolicyKeyDocument struct {
	ActorID   string `json:"actor_id"`
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
	Revoked   bool   `json:"revoked"`
}

type PolicyGrantDocument struct {
	ActorID string `json:"actor_id"`
	Action  Action `json:"action"`
	Target  string `json:"target"`
}

// PolicyDocument is the authoritative policy value. Ledger and policy-version
// identity are deliberately not duplicated here: they are supplied by the
// immutable Genesis root and accepted reducer binding. That prevents a mutable
// policy value from claiming a different trust domain while retaining the same
// governed-record target.
type PolicyDocument struct {
	MaxProofLifetimeSeconds int64                 `json:"max_proof_lifetime_seconds"`
	Keys                    []PolicyKeyDocument   `json:"keys"`
	Grants                  []PolicyGrantDocument `json:"grants"`
	ContentSHA256           string                `json:"content_sha256"`
}

// ParsePolicyDocument validates a canonical, digest-bound authoritative actor
// policy and converts it to the exact in-memory structure consumed by proof
// verification. ledgerID is supplied by the authoritative Genesis snapshot and
// is copied into every exact grant; it is never trusted from the policy value.
func ParsePolicyDocument(raw []byte, ledgerID string) (PolicyDocument, Policy, error) {
	if ledgerID == "" {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: authoritative ledger_id is required")
	}
	if err := strictjson.Validate(raw); err != nil {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: %w", err)
	}
	canonical, err := canonicaljson.Canonicalize(raw)
	if err != nil {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: canonicalize: %w", err)
	}
	if !bytes.Equal(raw, canonical) {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: document must be RFC 8785 canonical JSON")
	}
	if err := digest.Verify(raw); err != nil {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: %w", err)
	}
	var doc PolicyDocument
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: decode: %w", err)
	}
	if doc.MaxProofLifetimeSeconds <= 0 || doc.MaxProofLifetimeSeconds > math.MaxInt64/int64(time.Second) {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: max_proof_lifetime_seconds must be positive and representable")
	}
	if len(doc.Keys) == 0 || len(doc.Grants) == 0 {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: at least one key and one grant are required")
	}
	if !sort.SliceIsSorted(doc.Keys, func(i, j int) bool {
		if doc.Keys[i].ActorID != doc.Keys[j].ActorID {
			return doc.Keys[i].ActorID < doc.Keys[j].ActorID
		}
		return doc.Keys[i].KeyID < doc.Keys[j].KeyID
	}) {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: keys must be sorted by actor_id then key_id")
	}
	if !sort.SliceIsSorted(doc.Grants, func(i, j int) bool {
		if doc.Grants[i].ActorID != doc.Grants[j].ActorID {
			return doc.Grants[i].ActorID < doc.Grants[j].ActorID
		}
		if doc.Grants[i].Action != doc.Grants[j].Action {
			return doc.Grants[i].Action < doc.Grants[j].Action
		}
		return doc.Grants[i].Target < doc.Grants[j].Target
	}) {
		return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: grants must be sorted by actor_id, action then target")
	}

	policy := Policy{MaxProofLifetime: time.Duration(doc.MaxProofLifetimeSeconds) * time.Second}
	activeKeys := map[string]bool{}
	seenKeys := map[string]struct{}{}
	for _, keyDoc := range doc.Keys {
		if keyDoc.ActorID == "" || keyDoc.KeyID == "" || strings.ContainsAny(keyDoc.ActorID+keyDoc.KeyID, "\x00\r\n") {
			return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: every key requires safe actor_id and key_id")
		}
		identity := keyDoc.ActorID + "\x00" + keyDoc.KeyID
		if _, exists := seenKeys[identity]; exists {
			return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: duplicate actor/key binding")
		}
		seenKeys[identity] = struct{}{}
		decoded, err := base64.RawStdEncoding.DecodeString(keyDoc.PublicKey)
		if err != nil || len(decoded) != ed25519.PublicKeySize || base64.RawStdEncoding.EncodeToString(decoded) != keyDoc.PublicKey {
			return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: key %q/%q must contain canonical raw-base64 Ed25519 public key bytes", keyDoc.ActorID, keyDoc.KeyID)
		}
		pub := append(ed25519.PublicKey(nil), decoded...)
		policy.Keys = append(policy.Keys, KeyBinding{ActorID: keyDoc.ActorID, KeyID: keyDoc.KeyID, PublicKey: pub, Revoked: keyDoc.Revoked})
		if !keyDoc.Revoked {
			activeKeys[keyDoc.ActorID] = true
		}
	}

	seenGrants := map[string]struct{}{}
	for _, grantDoc := range doc.Grants {
		if grantDoc.ActorID == "" || grantDoc.Target == "" || strings.ContainsAny(grantDoc.ActorID+grantDoc.Target, "\x00\r\n") || !validAction(grantDoc.Action) {
			return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: every grant requires safe actor_id, target and a supported action")
		}
		identity := grantDoc.ActorID + "\x00" + string(grantDoc.Action) + "\x00" + grantDoc.Target
		if _, exists := seenGrants[identity]; exists {
			return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: duplicate actor grant")
		}
		seenGrants[identity] = struct{}{}
		if !activeKeys[grantDoc.ActorID] {
			return PolicyDocument{}, Policy{}, fmt.Errorf("AUTH_POLICY_INVALID: granted actor %q has no active trusted key", grantDoc.ActorID)
		}
		policy.Grants = append(policy.Grants, Grant{ActorID: grantDoc.ActorID, LedgerID: ledgerID, Action: grantDoc.Action, Target: grantDoc.Target})
	}
	if err := validatePolicy(policy); err != nil {
		return PolicyDocument{}, Policy{}, err
	}
	return doc, policy, nil
}

// ValidatePolicyValue performs the same semantic validation used by service
// admission without assigning authority to a real ledger. It is used by the
// reducer to reject malformed policy rotations before they can reach CAS.
func ValidatePolicyValue(raw []byte) error {
	_, _, err := ParsePolicyDocument(raw, "validation:ledger")
	return err
}

// ValidateInitialAuthorities requires Genesis's initial_authorities to describe
// exactly the actors initially granted authority by the root actor policy.
func ValidateInitialAuthorities(doc PolicyDocument, expected []string) error {
	actors := make([]string, 0, len(doc.Grants))
	seen := map[string]struct{}{}
	for _, grant := range doc.Grants {
		if _, exists := seen[grant.ActorID]; !exists {
			seen[grant.ActorID] = struct{}{}
			actors = append(actors, grant.ActorID)
		}
	}
	sort.Strings(actors)
	if len(actors) != len(expected) {
		return fmt.Errorf("AUTH_POLICY_INVALID: Genesis initial_authorities %v do not match granted actors %v", expected, actors)
	}
	for i := range actors {
		if actors[i] != expected[i] {
			return fmt.Errorf("AUTH_POLICY_INVALID: Genesis initial_authorities %v do not match granted actors %v", expected, actors)
		}
	}
	return nil
}
