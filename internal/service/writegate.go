package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
	"github.com/armpitpete/threadkeeper-core/internal/gitledger"
	"github.com/armpitpete/threadkeeper-core/internal/ledger"
)

var ErrAuthorityWritesDisabled = errors.New("AUTHORITY_WRITES_DISABLED: authoritative event writes are disabled until the conformance gate is satisfied")

func AuthorityWritesEnabled() bool { return false }

func RequireAuthorityWritesEnabled() error { return ErrAuthorityWritesDisabled }

// AdmitAuthorityWrite is the service-level admission boundary for any future
// authority-changing transport. The hard release gate is evaluated before any
// ledger read, policy load, authentication or authorization work. Callers
// cannot inject an alternate actorauth.Policy.
func AdmitAuthorityWrite(ctx context.Context, r *gitledger.Reader, proof actorauth.Proof, expected actorauth.RequestContext, now time.Time) (actorauth.Principal, error) {
	if err := RequireAuthorityWritesEnabled(); err != nil {
		return actorauth.Principal{}, err
	}
	snapshot, err := ledger.LoadCurrentActorPolicy(ctx, r)
	if err != nil {
		return actorauth.Principal{}, err
	}
	return authenticateAuthoritativePolicySnapshot(snapshot, proof, expected, now)
}

// authenticateAuthoritativePolicySnapshot is a pure admission check over a
// snapshot already derived from authoritative ledger state. It is separated so
// ledger/head binding can be unit-tested while the global write kill-switch is
// hard false. It does not load policy, move authority, expose a transport or
// bypass RequireAuthorityWritesEnabled.
func authenticateAuthoritativePolicySnapshot(snapshot *ledger.ActorPolicySnapshot, proof actorauth.Proof, expected actorauth.RequestContext, now time.Time) (actorauth.Principal, error) {
	if snapshot == nil {
		return actorauth.Principal{}, fmt.Errorf("AUTH_POLICY_INVALID: actor policy snapshot is nil")
	}
	if expected.LedgerID != snapshot.LedgerID {
		return actorauth.Principal{}, fmt.Errorf("AUTH_CONTEXT_INVALID: request ledger_id %q does not match authoritative ledger %q", expected.LedgerID, snapshot.LedgerID)
	}
	if expected.ExpectedState != snapshot.LedgerCommit {
		return actorauth.Principal{}, fmt.Errorf("AUTH_CONTEXT_INVALID: request expected_state %q does not match authoritative ledger head %q", expected.ExpectedState, snapshot.LedgerCommit)
	}
	return actorauth.AuthenticateAndAuthorize(snapshot.Policy, proof, expected, now)
}
