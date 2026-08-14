package loadproof

import (
	"context"
	"encoding/json"
	"runtime"
	"strings"
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

func TestDecodeEnvelopeRejectsUnknownDuplicateAndTrailingInput(t *testing.T) {
	validRaw, err := json.Marshal(ReferenceEnvelope())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodeEnvelope(validRaw); err != nil {
		t.Fatalf("valid envelope rejected: %v", err)
	}

	unknown := append([]byte(nil), validRaw[:len(validRaw)-1]...)
	unknown = append(unknown, []byte(`,"ambient_override":true}`)...)
	if _, err := DecodeEnvelope(unknown); err == nil {
		t.Fatal("unknown envelope field was accepted")
	}

	duplicate := []byte(`{"name":"first","name":"second","concurrent_workers":1,"iterations_per_worker":1,"max_peak_heap_growth_bytes":1,"max_settled_heap_growth_bytes":1,"max_peak_goroutine_growth":1,"max_settled_goroutine_growth":1,"max_peak_open_handle_growth":1,"max_settled_open_handle_growth":1,"require_open_handle_metric":false}`)
	if _, err := DecodeEnvelope(duplicate); err == nil {
		t.Fatal("duplicate envelope field was accepted")
	}

	trailing := append(append([]byte(nil), validRaw...), []byte(` {}`)...)
	if _, err := DecodeEnvelope(trailing); err == nil {
		t.Fatal("trailing JSON value was accepted")
	}
}

func TestEvaluateRejectsRequiredOpenHandleMetricGap(t *testing.T) {
	envelope := ReferenceEnvelope()
	evidence := Evidence{
		Envelope:                           envelope,
		ResourceSamples:                    4,
		OpenHandleMetricUnavailableSamples: 1,
		Before:                             Snapshot{OpenHandlesAvailable: true},
		Peak:                               Snapshot{OpenHandlesAvailable: true},
		AfterSettled:                       Snapshot{OpenHandlesAvailable: true},
	}
	if err := evaluate(evidence); err == nil || !strings.Contains(err.Error(), "LOAD_RESOURCE_METRIC_UNAVAILABLE") {
		t.Fatalf("required sampled metric gap was accepted: %v", err)
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
	if evidence.SampleIntervalMillis != 5 {
		t.Fatalf("sample interval = %d ms want 5", evidence.SampleIntervalMillis)
	}
	if evidence.ResourceSamples < 3 {
		t.Fatalf("resource samples = %d want at least baseline, final monitor and settled samples", evidence.ResourceSamples)
	}
	if evidence.OpenHandleMetricUnavailableSamples != 0 {
		t.Fatalf("required open-handle metric had %d unavailable samples", evidence.OpenHandleMetricUnavailableSamples)
	}
	want := envelope.ConcurrentWorkers * envelope.IterationsPerWorker
	if evidence.CompletedOperations != want {
		t.Fatalf("completed operations = %d want %d", evidence.CompletedOperations, want)
	}
	if !evidence.Before.OpenHandlesAvailable || !evidence.AfterSettled.OpenHandlesAvailable {
		t.Fatal("required process open-handle metric was unavailable")
	}
}
