package service

import "testing"

func TestAuthorityWritesHardDisabled(t *testing.T) {
	if AuthorityWritesEnabled() { t.Fatal("authority writes must be disabled in skeleton") }
	if err := RequireAuthorityWritesEnabled(); err == nil { t.Fatal("expected disabled error") }
}
