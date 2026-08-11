package service

import "errors"

var ErrAuthorityWritesDisabled = errors.New("AUTHORITY_WRITES_DISABLED: authoritative event writes are disabled until the conformance gate is satisfied")

func AuthorityWritesEnabled() bool { return false }

func RequireAuthorityWritesEnabled() error { return ErrAuthorityWritesDisabled }
