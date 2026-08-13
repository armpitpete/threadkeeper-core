package incident

import "testing"

func TestIncidentCannotSkipContainment(t *testing.T){ if CanTransition(Detected,Resolved){t.Fatal("incident must not skip containment")}; if !CanTransition(Detected,Contained){t.Fatal("expected containment transition")}}
