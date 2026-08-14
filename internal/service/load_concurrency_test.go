package service

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/armpitpete/threadkeeper-core/internal/actorauth"
)

func TestLimiterNeverAdmitsBeyondCapacityUnderBurst(t *testing.T) {
	const capacity = 8
	const requests = 64
	limiter := NewLimiter(capacity)

	start := make(chan struct{})
	release := make(chan struct{})
	var attempted sync.WaitGroup
	var workers sync.WaitGroup
	attempted.Add(requests)
	workers.Add(requests)
	var admitted atomic.Int64
	var overloaded atomic.Int64
	var current atomic.Int64
	var maximum atomic.Int64

	for i := 0; i < requests; i++ {
		go func() {
			defer workers.Done()
			<-start
			done, err := limiter.TryAcquire()
			if err != nil {
				if !errors.Is(err, ErrOverloaded) {
					t.Errorf("unexpected limiter error: %v", err)
				}
				overloaded.Add(1)
				attempted.Done()
				return
			}
			admitted.Add(1)
			now := current.Add(1)
			for {
				old := maximum.Load()
				if now <= old || maximum.CompareAndSwap(old, now) {
					break
				}
			}
			attempted.Done()
			<-release
			current.Add(-1)
			done()
		}()
	}
	close(start)
	attempted.Wait()

	if got := admitted.Load(); got != capacity {
		t.Fatalf("admitted = %d want exactly capacity %d", got, capacity)
	}
	if got := overloaded.Load(); got != requests-capacity {
		t.Fatalf("overloaded = %d want %d", got, requests-capacity)
	}
	if got := maximum.Load(); got > capacity {
		t.Fatalf("maximum concurrent admission = %d exceeds capacity %d", got, capacity)
	}
	close(release)
	workers.Wait()
}

func TestAuthorityWriteKillSwitchDominatesUnderConcurrency(t *testing.T) {
	const requests = 128
	start := make(chan struct{})
	var workers sync.WaitGroup
	workers.Add(requests)
	errs := make(chan error, requests)
	for i := 0; i < requests; i++ {
		go func() {
			defer workers.Done()
			<-start
			_, err := AdmitAuthorityWrite(context.Background(), nil, actorauth.Proof{}, actorauth.RequestContext{}, time.Now())
			errs <- err
		}()
	}
	close(start)
	workers.Wait()
	close(errs)
	for err := range errs {
		if !errors.Is(err, ErrAuthorityWritesDisabled) {
			t.Fatalf("authority admission escaped hard kill-switch: %v", err)
		}
	}
}
