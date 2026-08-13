package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
)

func TestAuthorityWriteAdmissionKillSwitchDominates(t *testing.T) {
	_, err := AdmitAuthorityWrite(actorauth.Policy{}, actorauth.Proof{}, actorauth.RequestContext{}, time.Now())
	if !errors.Is(err, ErrAuthorityWritesDisabled) {
		t.Fatalf("expected release gate to reject before auth processing, got %v", err)
	}
}

func TestAuthorityWriteAdmissionRequiresAuthenticationAfterReleaseGate(t *testing.T) {
	ctx := actorauth.RequestContext{
		LedgerID:       "ledger:test",
		Action:         actorauth.ActionDecide,
		Target:         "project:test/status",
		ExpectedState:  "0123456789abcdef0123456789abcdef01234567",
		IdempotencyKey: "decision:test",
	}
	policy := actorauth.Policy{MaxProofLifetime: 5 * time.Minute}
	_, err := admitAuthorityWrite(true, policy, actorauth.Proof{}, ctx, time.Now())
	if err == nil || !strings.Contains(err.Error(), "AUTHENTICATION_FAILED") {
		t.Fatalf("expected authentication failure after release gate, got %v", err)
	}
}
