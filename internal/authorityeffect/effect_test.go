package authorityeffect

import "testing"

func TestKnownAuthorityEffects(t *testing.T){ if err:=(Declaration{Component:"witness",Effect:IntegrityAttestation}).Validate(); err!=nil{t.Fatal(err)} }
