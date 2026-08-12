package keylifecycle

import "fmt"

type State string

const (
	Generated   State = "generated"
	Active      State = "active"
	Rotating    State = "rotating"
	Compromised State = "compromised"
	Revoked     State = "revoked"
	Retired     State = "retired"
)

func CanTransition(from,to State) bool {
	switch from {
	case Generated: return to==Active || to==Revoked
	case Active: return to==Rotating || to==Compromised || to==Revoked || to==Retired
	case Rotating: return to==Active || to==Compromised || to==Revoked || to==Retired
	case Compromised: return to==Revoked
	default: return false
	}
}

func RequireTransition(from,to State) error { if !CanTransition(from,to){return fmt.Errorf("KEY_LIFECYCLE_INVALID: %s -> %s",from,to)}; return nil }
