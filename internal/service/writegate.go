package service

import (
	"errors"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
)

var ErrAuthorityWritesDisabled = errors.New("AUTHORITY_WRITES_DISABLED: authoritative event writes are disabled until the conformance gate is satisfied")

func AuthorityWritesEnabled() bool { return false }

func RequireAuthorityWritesEnabled() error { return ErrAuthorityWritesDisabled }

// AdmitAuthorityWrite is the service-level admission boundary for any future
// authority-changing transport. The global release gate is evaluated first.
// Only after that gate is deliberately opened may an authenticated and exactly
// authorised actor proceed to the existing candidate/reducer/CAS machinery.
func AdmitAuthorityWrite(policy actorauth.Policy, proof actorauth.Proof, expected actorauth.RequestContext, now time.Time) (actorauth.Principal, error) {
	return admitAuthorityWrite(AuthorityWritesEnabled(), policy, proof, expected, now)
}

func admitAuthorityWrite(enabled bool, policy actorauth.Policy, proof actorauth.Proof, expected actorauth.RequestContext, now time.Time) (actorauth.Principal, error) {
	if !enabled {
		return actorauth.Principal{}, ErrAuthorityWritesDisabled
	}
	return actorauth.AuthenticateAndAuthorize(policy, proof, expected, now)
}
