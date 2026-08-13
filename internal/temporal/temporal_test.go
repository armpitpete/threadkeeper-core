package temporal

import (
	"testing"
	"time"
)

func TestEffectiveTimeDiffersFromKnowledgeTime(t *testing.T) {
	w := Window{EffectiveFrom: "2026-07-01T00:00:00Z", ObservedAt: "2026-08-12T09:00:00Z", AcceptedAt: "2026-08-12T10:00:00Z"}
	if err := w.Validate(); err != nil { t.Fatal(err) }
	july := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	if !w.EffectiveAt(july) { t.Fatal("expected effective in July") }
	if w.KnownAt(july) { t.Fatal("knowledge must not travel backwards in acceptance time") }
}
