package temporal

import (
	"fmt"
	"time"
)

type Window struct {
	EffectiveFrom  string `json:"effective_from,omitempty"`
	EffectiveUntil string `json:"effective_until,omitempty"`
	ObservedAt     string `json:"observed_at"`
	AcceptedAt     string `json:"accepted_at"`
}

func (w Window) Validate() error {
	observed, err := time.Parse(time.RFC3339, w.ObservedAt)
	if err != nil { return fmt.Errorf("TEMPORAL_INVALID: observed_at: %w", err) }
	accepted, err := time.Parse(time.RFC3339, w.AcceptedAt)
	if err != nil { return fmt.Errorf("TEMPORAL_INVALID: accepted_at: %w", err) }
	if accepted.Before(observed) { return fmt.Errorf("TEMPORAL_INVALID: accepted_at precedes observed_at") }
	var from, until time.Time
	if w.EffectiveFrom != "" {
		from, err = time.Parse(time.RFC3339, w.EffectiveFrom)
		if err != nil { return fmt.Errorf("TEMPORAL_INVALID: effective_from: %w", err) }
	}
	if w.EffectiveUntil != "" {
		until, err = time.Parse(time.RFC3339, w.EffectiveUntil)
		if err != nil { return fmt.Errorf("TEMPORAL_INVALID: effective_until: %w", err) }
	}
	if !from.IsZero() && !until.IsZero() && until.Before(from) { return fmt.Errorf("TEMPORAL_INVALID: effective_until precedes effective_from") }
	return nil
}

func (w Window) EffectiveAt(at time.Time) bool {
	if w.EffectiveFrom != "" { from, err := time.Parse(time.RFC3339, w.EffectiveFrom); if err != nil || at.Before(from) { return false } }
	if w.EffectiveUntil != "" { until, err := time.Parse(time.RFC3339, w.EffectiveUntil); if err != nil || !at.Before(until) { return false } }
	return true
}

func (w Window) KnownAt(at time.Time) bool {
	accepted, err := time.Parse(time.RFC3339, w.AcceptedAt)
	return err == nil && !accepted.After(at)
}
