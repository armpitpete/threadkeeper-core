package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
)

func TestAuthorityWriteAdmissionKillSwitchDominates(t *testing.T) {
	_, err := AdmitAuthorityWrite(context.Background(), nil, actorauth.Proof{}, actorauth.RequestContext{}, time.Now())
	if !errors.Is(err, ErrAuthorityWritesDisabled) {
		t.Fatalf("expected release gate to reject before ledger/policy/auth processing, got %v", err)
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
	loader := func(context.Context, *gitledger.Reader) (*ledger.ActorPolicySnapshot, error) {
		return &ledger.ActorPolicySnapshot{
			LedgerCommit: ctx.ExpectedState,
			LedgerID:     ctx.LedgerID,
			Policy:       actorauth.Policy{MaxProofLifetime: 5 * time.Minute},
		}, nil
	}
	_, err := admitAuthorityWrite(true, loader, context.Background(), nil, actorauth.Proof{}, ctx, time.Now())
	if err == nil || !strings.Contains(err.Error(), "AUTHENTICATION_FAILED") {
		t.Fatalf("expected authentication failure after release gate, got %v", err)
	}
}

func TestAuthorityWriteAdmissionRejectsWrongLedgerBeforeAuth(t *testing.T) {
	expected := actorauth.RequestContext{LedgerID: "ledger:wrong", Action: actorauth.ActionOperate, Target: "target:test", ExpectedState: "head", IdempotencyKey: "idem"}
	loader := func(context.Context, *gitledger.Reader) (*ledger.ActorPolicySnapshot, error) {
		return &ledger.ActorPolicySnapshot{LedgerCommit: "head", LedgerID: "ledger:real", Policy: actorauth.Policy{MaxProofLifetime: time.Minute}}, nil
	}
	_, err := admitAuthorityWrite(true, loader, context.Background(), nil, actorauth.Proof{}, expected, time.Now())
	if err == nil || !strings.Contains(err.Error(), "AUTH_CONTEXT_INVALID") || !strings.Contains(err.Error(), "ledger_id") {
		t.Fatalf("wrong ledger reached auth path: %v", err)
	}
}

func TestAuthorityWriteAdmissionRejectsWrongExpectedStateBeforeAuth(t *testing.T) {
	expected := actorauth.RequestContext{LedgerID: "ledger:test", Action: actorauth.ActionOperate, Target: "target:test", ExpectedState: "old-head", IdempotencyKey: "idem"}
	loader := func(context.Context, *gitledger.Reader) (*ledger.ActorPolicySnapshot, error) {
		return &ledger.ActorPolicySnapshot{LedgerCommit: "new-head", LedgerID: expected.LedgerID, Policy: actorauth.Policy{MaxProofLifetime: time.Minute}}, nil
	}
	_, err := admitAuthorityWrite(true, loader, context.Background(), nil, actorauth.Proof{}, expected, time.Now())
	if err == nil || !strings.Contains(err.Error(), "AUTH_CONTEXT_INVALID") || !strings.Contains(err.Error(), "expected_state") {
		t.Fatalf("wrong expected state reached auth path: %v", err)
	}
}
