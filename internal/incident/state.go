package incident

import "fmt"

type State string

const (
	Detected   State = "detected"
	Contained  State = "contained"
	Recovering State = "recovering"
	Resolved   State = "resolved"
)

func CanTransition(from,to State) bool {
	switch from { case Detected: return to==Contained; case Contained: return to==Recovering || to==Resolved; case Recovering: return to==Contained || to==Resolved; default: return false }
}

func RequireTransition(from,to State) error { if !CanTransition(from,to){return fmt.Errorf("INCIDENT_TRANSITION_INVALID: %s -> %s",from,to)}; return nil }
