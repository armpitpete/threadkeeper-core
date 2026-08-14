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

type actorPolicyLoader func(context.Context, *gitledger.Reader) (*ledger.ActorPolicySnapshot, error)

// AdmitAuthorityWrite is the service-level admission boundary for any future
// authority-changing transport. The global release gate is evaluated first.
// Only after that gate is deliberately opened does Core replay the supplied
// authoritative ledger, derive the current trusted actor policy from that exact
// snapshot, bind the proof to its ledger/head, and authenticate/authorise.
// Callers cannot inject an alternate actorauth.Policy.
func AdmitAuthorityWrite(ctx context.Context, r *gitledger.Reader, proof actorauth.Proof, expected actorauth.RequestContext, now time.Time) (actorauth.Principal, error) {
	return admitAuthorityWrite(AuthorityWritesEnabled(), ledger.LoadCurrentActorPolicy, ctx, r, proof, expected, now)
}

func admitAuthorityWrite(enabled bool, load actorPolicyLoader, ctx context.Context, r *gitledger.Reader, proof actorauth.Proof, expected actorauth.RequestContext, now time.Time) (actorauth.Principal, error) {
	if !enabled {
		return actorauth.Principal{}, ErrAuthorityWritesDisabled
	}
	if load == nil {
		return actorauth.Principal{}, fmt.Errorf("AUTH_POLICY_INVALID: actor policy loader is required")
	}
	snapshot, err := load(ctx, r)
	if err != nil {
		return actorauth.Principal{}, err
	}
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
