package loadproof

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"runtime"
	"sync"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/strictjson"
)

var ErrEnvelopeExceeded = errors.New("LOAD_RESOURCE_ENVELOPE_EXCEEDED")

const sampleInterval = 5 * time.Millisecond

var requiredEnvelopeFields = []string{
	"name",
	"concurrent_workers",
	"iterations_per_worker",
	"max_peak_heap_growth_bytes",
	"max_settled_heap_growth_bytes",
	"max_peak_goroutine_growth",
	"max_settled_goroutine_growth",
	"max_peak_open_handle_growth",
	"max_settled_open_handle_growth",
	"require_open_handle_metric",
}

type Envelope struct {
	Name                       string `json:"name"`
	ConcurrentWorkers          int    `json:"concurrent_workers"`
	IterationsPerWorker        int    `json:"iterations_per_worker"`
	MaxPeakHeapGrowthBytes     int64  `json:"max_peak_heap_growth_bytes"`
	MaxSettledHeapGrowthBytes  int64  `json:"max_settled_heap_growth_bytes"`
	MaxPeakGoroutineGrowth     int64  `json:"max_peak_goroutine_growth"`
	MaxSettledGoroutineGrowth  int64  `json:"max_settled_goroutine_growth"`
	MaxPeakOpenHandleGrowth    int64  `json:"max_peak_open_handle_growth"`
	MaxSettledOpenHandleGrowth int64  `json:"max_settled_open_handle_growth"`
	RequireOpenHandleMetric    bool   `json:"require_open_handle_metric"`
}

func DecodeEnvelope(raw []byte) (Envelope, error) {
	if err := strictjson.Validate(raw); err != nil {
		return Envelope{}, fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: %w", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(raw, &fields); err != nil {
		return Envelope{}, fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: decode fields: %w", err)
	}
	for _, name := range requiredEnvelopeFields {
		value, ok := fields[name]
		if !ok {
			return Envelope{}, fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: required field %q is missing", name)
		}
		if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return Envelope{}, fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: required field %q must not be null", name)
		}
	}

	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var envelope Envelope
	if err := decoder.Decode(&envelope); err != nil {
		return Envelope{}, fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: decode: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Envelope{}, fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: trailing JSON value")
		}
		return Envelope{}, fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: trailing data: %w", err)
	}
	if err := envelope.Validate(); err != nil {
		return Envelope{}, err
	}
	return envelope, nil
}

func (e Envelope) Validate() error {
	if e.Name == "" {
		return fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: name is required")
	}
	if e.ConcurrentWorkers <= 0 || e.ConcurrentWorkers > 1024 {
		return fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: concurrent_workers must be 1..1024")
	}
	if e.IterationsPerWorker <= 0 || e.IterationsPerWorker > 1000000 {
		return fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: iterations_per_worker must be 1..1000000")
	}
	limits := []struct {
		name  string
		value int64
	}{
		{"max_peak_heap_growth_bytes", e.MaxPeakHeapGrowthBytes},
		{"max_settled_heap_growth_bytes", e.MaxSettledHeapGrowthBytes},
		{"max_peak_goroutine_growth", e.MaxPeakGoroutineGrowth},
		{"max_settled_goroutine_growth", e.MaxSettledGoroutineGrowth},
		{"max_peak_open_handle_growth", e.MaxPeakOpenHandleGrowth},
		{"max_settled_open_handle_growth", e.MaxSettledOpenHandleGrowth},
	}
	for _, limit := range limits {
		if limit.value < 0 {
			return fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: %s must be non-negative", limit.name)
		}
	}
	if e.MaxSettledHeapGrowthBytes > e.MaxPeakHeapGrowthBytes {
		return fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: settled heap ceiling exceeds peak ceiling")
	}
	if e.MaxSettledGoroutineGrowth > e.MaxPeakGoroutineGrowth {
		return fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: settled goroutine ceiling exceeds peak ceiling")
	}
	if e.MaxSettledOpenHandleGrowth > e.MaxPeakOpenHandleGrowth {
		return fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: settled open-handle ceiling exceeds peak ceiling")
	}
	return nil
}

// ReferenceEnvelope is deliberately a repository/CI conformance profile, not a
// production capacity claim. Production must supply and measure its own exact
// envelope on the selected deployment.
func ReferenceEnvelope() Envelope {
	return Envelope{
		Name:                       "threadkeeper-core-ci-reference-v1",
		ConcurrentWorkers:          8,
		IterationsPerWorker:        4,
		MaxPeakHeapGrowthBytes:     128 << 20,
		MaxSettledHeapGrowthBytes:  64 << 20,
		MaxPeakGoroutineGrowth:     64,
		MaxSettledGoroutineGrowth:  8,
		MaxPeakOpenHandleGrowth:    128,
		MaxSettledOpenHandleGrowth: 16,
		RequireOpenHandleMetric:    true,
	}
}

type Snapshot struct {
	HeapAllocBytes       uint64 `json:"heap_alloc_bytes"`
	Goroutines           int    `json:"goroutines"`
	OpenHandles          uint64 `json:"open_handles"`
	OpenHandlesAvailable bool   `json:"open_handles_available"`
}

type Growth struct {
	HeapBytes   int64 `json:"heap_bytes"`
	Goroutines  int64 `json:"goroutines"`
	OpenHandles int64 `json:"open_handles"`
}

type Evidence struct {
	Envelope                 Envelope `json:"envelope"`
	SampleIntervalMillis     int64    `json:"sample_interval_millis"`
	OpenHandleMetricComplete bool     `json:"open_handle_metric_complete"`
	Before                   Snapshot `json:"before"`
	Peak                     Snapshot `json:"peak"`
	AfterSettled             Snapshot `json:"after_settled"`
	PeakGrowth               Growth   `json:"peak_growth"`
	SettledGrowth            Growth   `json:"settled_growth"`
	CompletedOperations      int      `json:"completed_operations"`
	Passed                   bool     `json:"passed"`
}

type Workload func(context.Context, int, int) error

// Run executes exactly the declared worker/iteration envelope, samples process
// resources during execution, then forces a settled GC snapshot. Both transient
// sampled peak growth and post-work settled growth are checked. The evidence is
// pure observation; it does not grant authority or infer production capacity.
func Run(ctx context.Context, envelope Envelope, workload Workload) (Evidence, error) {
	if err := envelope.Validate(); err != nil {
		return Evidence{}, err
	}
	if workload == nil {
		return Evidence{}, fmt.Errorf("LOAD_RESOURCE_ENVELOPE_INVALID: workload is required")
	}

	before := settledSnapshot()
	if envelope.RequireOpenHandleMetric && !before.OpenHandlesAvailable {
		return Evidence{}, fmt.Errorf("LOAD_RESOURCE_METRIC_UNAVAILABLE: open handle metric is required")
	}
	peak := before
	metricComplete := before.OpenHandlesAvailable
	var peakMu sync.Mutex
	updatePeak := func(s Snapshot) {
		peakMu.Lock()
		defer peakMu.Unlock()
		if s.HeapAllocBytes > peak.HeapAllocBytes {
			peak.HeapAllocBytes = s.HeapAllocBytes
		}
		if s.Goroutines > peak.Goroutines {
			peak.Goroutines = s.Goroutines
		}
		if !s.OpenHandlesAvailable {
			metricComplete = false
			return
		}
		if !peak.OpenHandlesAvailable || s.OpenHandles > peak.OpenHandles {
			peak.OpenHandles = s.OpenHandles
		}
		peak.OpenHandlesAvailable = true
	}

	monitorDone := make(chan struct{})
	var monitorWG sync.WaitGroup
	monitorWG.Add(1)
	go func() {
		defer monitorWG.Done()
		ticker := time.NewTicker(sampleInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				updatePeak(snapshot())
			case <-monitorDone:
				updatePeak(snapshot())
				return
			}
		}
	}()

	var workers sync.WaitGroup
	var resultMu sync.Mutex
	completed := 0
	var firstErr error
	for worker := 0; worker < envelope.ConcurrentWorkers; worker++ {
		worker := worker
		workers.Add(1)
		go func() {
			defer workers.Done()
			for iteration := 0; iteration < envelope.IterationsPerWorker; iteration++ {
				if err := ctx.Err(); err != nil {
					resultMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					resultMu.Unlock()
					return
				}
				if err := workload(ctx, worker, iteration); err != nil {
					resultMu.Lock()
					if firstErr == nil {
						firstErr = err
					}
					resultMu.Unlock()
					return
				}
				resultMu.Lock()
				completed++
				resultMu.Unlock()
			}
		}()
	}
	workers.Wait()
	close(monitorDone)
	monitorWG.Wait()

	after := settledSnapshot()
	updatePeak(after)
	peakMu.Lock()
	finalPeak := peak
	finalMetricComplete := metricComplete
	peakMu.Unlock()

	evidence := Evidence{
		Envelope:                 envelope,
		SampleIntervalMillis:     sampleInterval.Milliseconds(),
		OpenHandleMetricComplete: finalMetricComplete,
		Before:                   before,
		Peak:                     finalPeak,
		AfterSettled:             after,
		PeakGrowth:               growth(before, finalPeak),
		SettledGrowth:            growth(before, after),
		CompletedOperations:      completed,
	}
	if firstErr != nil {
		return evidence, fmt.Errorf("LOAD_RESOURCE_WORKLOAD_FAILED: %w", firstErr)
	}
	want := envelope.ConcurrentWorkers * envelope.IterationsPerWorker
	if completed != want {
		return evidence, fmt.Errorf("LOAD_RESOURCE_WORKLOAD_FAILED: completed %d operations want %d", completed, want)
	}
	if err := evaluate(evidence); err != nil {
		return evidence, err
	}
	evidence.Passed = true
	return evidence, nil
}

func settledSnapshot() Snapshot {
	runtime.GC()
	runtime.GC()
	return snapshot()
}

func snapshot() Snapshot {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	handles, available := openHandleCount()
	return Snapshot{
		HeapAllocBytes:       mem.HeapAlloc,
		Goroutines:           runtime.NumGoroutine(),
		OpenHandles:          handles,
		OpenHandlesAvailable: available,
	}
}

func growth(before, after Snapshot) Growth {
	return Growth{
		HeapBytes:   signedDelta(before.HeapAllocBytes, after.HeapAllocBytes),
		Goroutines:  int64(after.Goroutines - before.Goroutines),
		OpenHandles: signedDelta(before.OpenHandles, after.OpenHandles),
	}
}

func signedDelta(before, after uint64) int64 {
	if after >= before {
		delta := after - before
		if delta > math.MaxInt64 {
			return math.MaxInt64
		}
		return int64(delta)
	}
	delta := before - after
	if delta > math.MaxInt64 {
		return math.MinInt64
	}
	return -int64(delta)
}

func evaluate(e Evidence) error {
	if e.Envelope.RequireOpenHandleMetric && !e.OpenHandleMetricComplete {
		return fmt.Errorf("LOAD_RESOURCE_METRIC_UNAVAILABLE: required open-handle metric was unavailable during measurement")
	}
	if e.Envelope.RequireOpenHandleMetric && (!e.Before.OpenHandlesAvailable || !e.Peak.OpenHandlesAvailable || !e.AfterSettled.OpenHandlesAvailable) {
		return fmt.Errorf("LOAD_RESOURCE_METRIC_UNAVAILABLE: open handle metric is incomplete")
	}
	checks := []struct {
		name  string
		got   int64
		limit int64
	}{
		{"peak heap growth", positive(e.PeakGrowth.HeapBytes), e.Envelope.MaxPeakHeapGrowthBytes},
		{"settled heap growth", positive(e.SettledGrowth.HeapBytes), e.Envelope.MaxSettledHeapGrowthBytes},
		{"peak goroutine growth", positive(e.PeakGrowth.Goroutines), e.Envelope.MaxPeakGoroutineGrowth},
		{"settled goroutine growth", positive(e.SettledGrowth.Goroutines), e.Envelope.MaxSettledGoroutineGrowth},
	}
	if e.Before.OpenHandlesAvailable && e.Peak.OpenHandlesAvailable && e.AfterSettled.OpenHandlesAvailable {
		checks = append(checks,
			struct {
				name  string
				got   int64
				limit int64
			}{"peak open-handle growth", positive(e.PeakGrowth.OpenHandles), e.Envelope.MaxPeakOpenHandleGrowth},
			struct {
				name  string
				got   int64
				limit int64
			}{"settled open-handle growth", positive(e.SettledGrowth.OpenHandles), e.Envelope.MaxSettledOpenHandleGrowth},
		)
	}
	for _, check := range checks {
		if check.got > check.limit {
			return fmt.Errorf("%w: %s=%d limit=%d", ErrEnvelopeExceeded, check.name, check.got, check.limit)
		}
	}
	return nil
}

func positive(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}
