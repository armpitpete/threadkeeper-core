package keylifecycle

import "testing"

func TestCompromisedKeyMustMoveToRevoked(t *testing.T){ if !CanTransition(Compromised,Revoked){t.Fatal("compromised key must be revocable")}; if CanTransition(Revoked,Active){t.Fatal("revoked key must not reactivate")}}
