package loadproof

import (
	"context"
	"runtime"
	"testing"
)

func TestEnvelopeRejectsIncompleteOrInvertedLimits(t *testing.T) {
	invalid := ReferenceEnvelope()
	invalid.Name = ""
	if err := invalid.Validate(); err == nil {
		t.Fatal("missing envelope name was accepted")
	}

	invalid = ReferenceEnvelope()
	invalid.MaxSettledHeapGrowthBytes = invalid.MaxPeakHeapGrowthBytes + 1
	if err := invalid.Validate(); err == nil {
		t.Fatal("settled heap ceiling above peak was accepted")
	}
}

func TestReferenceEnvelopeMeasuresSimpleWorkload(t *testing.T) {
	if runtime.GOOS != "linux" && runtime.GOOS != "windows" {
		t.Skip("reference envelope requires a process open-handle metric")
	}
	envelope := ReferenceEnvelope()
	evidence, err := Run(context.Background(), envelope, func(context.Context, int, int) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	if !evidence.Passed {
		t.Fatal("successful reference workload did not produce passing evidence")
	}
	want := envelope.ConcurrentWorkers * envelope.IterationsPerWorker
	if evidence.CompletedOperations != want {
		t.Fatalf("completed operations = %d want %d", evidence.CompletedOperations, want)
	}
	if !evidence.Before.OpenHandlesAvailable || !evidence.AfterSettled.OpenHandlesAvailable {
		t.Fatal("required process open-handle metric was unavailable")
	}
}
