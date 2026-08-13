package actorauth

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"testing"
	"time"
)

func authFixture(t *testing.T) (Policy, Proof, RequestContext, time.Time) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil { t.Fatal(err) }
	now := time.Date(2026, 8, 13, 21, 30, 0, 0, time.UTC)
	ctx := RequestContext{
		LedgerID: "ledger:threadkeeper",
		Action: ActionDecide,
		Target: "project:alpha/status",
		ExpectedState: "0123456789abcdef0123456789abcdef01234567",
		IdempotencyKey: "decision-123",
	}
	proof := Proof{
		ActorID: "actor:owner",
		KeyID: "key:owner-1",
		IssuedAt: now.Add(-time.Minute).Format(time.RFC3339),
		ExpiresAt: now.Add(time.Minute).Format(time.RFC3339),
		Context: ctx,
	}
	message, err := signingBytes(proof)
	if err != nil { t.Fatal(err) }
	proof.Signature = base64.RawStdEncoding.EncodeToString(ed25519.Sign(priv, message))
	policy := Policy{
		Keys: []KeyBinding{{ActorID: proof.ActorID, KeyID: proof.KeyID, PublicKey: pub}},
		Grants: []Grant{{ActorID: proof.ActorID, LedgerID: ctx.LedgerID, Action: ActionDecide, Target: ctx.Target}},
		MaxProofLifetime: 5 * time.Minute,
	}
	return policy, proof, ctx, now
}

func TestAuthenticateAndAuthorizeValidDecision(t *testing.T) {
	policy, proof, ctx, now := authFixture(t)
	principal, err := AuthenticateAndAuthorize(policy, proof, ctx, now)
	if err != nil { t.Fatal(err) }
	if principal.ActorID != "actor:owner" || principal.KeyID != "key:owner-1" {
		t.Fatalf("unexpected principal: %#v", principal)
	}
}

func TestAuthenticationRejectsRequestSubstitution(t *testing.T) {
	policy, proof, ctx, now := authFixture(t)
	ctx.Target = "project:other/status"
	if _, err := AuthenticateAndAuthorize(policy, proof, ctx, now); err == nil {
		t.Fatal("expected signed-context mismatch rejection")
	}
}

func TestAuthenticationRejectsActorSubstitution(t *testing.T) {
	policy, proof, ctx, now := authFixture(t)
	proof.ActorID = "actor:attacker"
	policy.Keys = append(policy.Keys, KeyBinding{ActorID: "actor:attacker", KeyID: proof.KeyID, PublicKey: policy.Keys[0].PublicKey})
	policy.Grants = append(policy.Grants, Grant{ActorID: "actor:attacker", LedgerID: ctx.LedgerID, Action: ActionDecide, Target: ctx.Target})
	if _, err := AuthenticateAndAuthorize(policy, proof, ctx, now); err == nil {
		t.Fatal("expected signature to bind actor identity")
	}
}

func TestAuthorizationRequiresExactLedgerActionAndTargetGrant(t *testing.T) {
	policy, proof, ctx, now := authFixture(t)
	policy.Grants = []Grant{{ActorID: proof.ActorID, LedgerID: ctx.LedgerID, Action: ActionDecide, Target: "project:other/status"}}
	if _, err := AuthenticateAndAuthorize(policy, proof, ctx, now); err == nil {
		t.Fatal("expected authorization denial")
	}
}

func TestAuthenticationRejectsExpiredAndOverlongProofs(t *testing.T) {
	policy, proof, ctx, now := authFixture(t)
	if _, err := AuthenticateAndAuthorize(policy, proof, ctx, now.Add(2*time.Minute)); err == nil {
		t.Fatal("expected expired proof rejection")
	}
	policy.MaxProofLifetime = time.Second
	if _, err := AuthenticateAndAuthorize(policy, proof, ctx, now); err == nil {
		t.Fatal("expected proof lifetime policy rejection")
	}
}

func TestAuthenticationRejectsRevokedKey(t *testing.T) {
	policy, proof, ctx, now := authFixture(t)
	policy.Keys[0].Revoked = true
	if _, err := AuthenticateAndAuthorize(policy, proof, ctx, now); err == nil {
		t.Fatal("expected revoked key rejection")
	}
}
