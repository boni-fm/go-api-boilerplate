package worker_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"go-api-boilerplate/internal/worker"
)

// ── helpers ──────────────────────────────────────────────────────────────────

// waitOrTimeout blocks until done is closed or the deadline elapses.
// It calls t.Fatal on timeout so callers don't need to repeat the pattern.
func waitOrTimeout(t *testing.T, done <-chan struct{}, deadline time.Duration) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(deadline):
		t.Fatalf("timed out after %v", deadline)
	}
}

// ── New() argument clamping ───────────────────────────────────────────────────

func TestNew_ClampNonPositiveArgs(t *testing.T) {
	// New(0, 0) must not panic and must produce a functional pool.
	p := worker.New(0, 0)
	if p == nil {
		t.Fatal("New(0, 0) returned nil")
	}
	p.Start(context.Background())
	p.Stop() // must not block or panic
}

// ── Basic job execution ───────────────────────────────────────────────────────

var jobExecTests = []struct {
	name     string
	workers  int
	capacity int
	numJobs  int
	wantAll  bool // expect all jobs to succeed (no queue saturation)
}{
	{"single worker", 1, 16, 5, true},
	{"multi worker", 4, 32, 20, true},
	{"exact capacity", 2, 5, 5, true},
}

func TestPool_JobsAreExecuted(t *testing.T) {
	for _, tc := range jobExecTests {
		t.Run(tc.name, func(t *testing.T) {
			p := worker.New(tc.workers, tc.capacity)
			p.Start(context.Background())
			defer p.Stop()

			var executed atomic.Int64
			var wg sync.WaitGroup
			wg.Add(tc.numJobs)

			for i := 0; i < tc.numJobs; i++ {
				ok := p.Submit(func(_ context.Context) {
					executed.Add(1)
					wg.Done()
				})
				if !ok {
					wg.Done() // compensate for dropped job
					if tc.wantAll {
						t.Errorf("Submit returned false unexpectedly for job %d", i)
					}
				}
			}

			done := make(chan struct{})
			go func() { wg.Wait(); close(done) }()
			waitOrTimeout(t, done, 5*time.Second)

			if tc.wantAll {
				if got := executed.Load(); got != int64(tc.numJobs) {
					t.Errorf("executed: got %d, want %d", got, tc.numJobs)
				}
			}
		})
	}
}

// ── Stats ─────────────────────────────────────────────────────────────────────

func TestPool_StatsProcessedCount(t *testing.T) {
	const n = 10
	p := worker.New(2, n)
	p.Start(context.Background())

	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		p.Submit(func(_ context.Context) { wg.Done() })
	}
	wg.Wait()
	p.Stop()

	processed, _ := p.Stats()
	if processed != n {
		t.Errorf("processed: got %d, want %d", processed, n)
	}
}

func TestPool_StatsDroppedCountWhenQueueFull(t *testing.T) {
	// capacity=1 with 1 worker blocked → third Submit must be dropped.
	p := worker.New(1, 1)
	p.Start(context.Background())
	defer p.Stop()

	block := make(chan struct{})

	// First job occupies the single worker goroutine.
	p.Submit(func(_ context.Context) { <-block })
	// Second job fills the single-slot queue.
	p.Submit(func(_ context.Context) {})
	// Third job: queue full, must be dropped.
	ok := p.Submit(func(_ context.Context) {})

	if ok {
		t.Error("Submit should return false (drop) when queue is full")
	}
	_, dropped := p.Stats()
	if dropped < 1 {
		t.Errorf("dropped: got %d, want >= 1", dropped)
	}

	close(block) // unblock so defer Stop() can complete cleanly
}

// ── Idempotent Stop ───────────────────────────────────────────────────────────

func TestPool_StopIsIdempotent(t *testing.T) {
	p := worker.New(2, 4)
	p.Start(context.Background())
	// Multiple consecutive calls must not panic or deadlock.
	p.Stop()
	p.Stop()
	p.Stop()
}

// ── Submit after Stop ─────────────────────────────────────────────────────────

func TestPool_SubmitAfterStopReturnsFalse(t *testing.T) {
	p := worker.New(1, 4)
	p.Start(context.Background())
	p.Stop()

	ok := p.Submit(func(_ context.Context) {})
	if ok {
		t.Error("Submit after Stop must return false")
	}
	_, dropped := p.Stats()
	if dropped < 1 {
		t.Errorf("dropped counter after post-Stop Submit: got %d, want >= 1", dropped)
	}
}

// ── Context propagation ───────────────────────────────────────────────────────

// TestPool_JobContextCancelledOnStop verifies that a long-running job receives
// a cancelled context when Stop is called, allowing it to exit early.
func TestPool_JobContextCancelledOnStop(t *testing.T) {
	p := worker.New(1, 1)
	p.Start(context.Background())

	ctxCancelled := make(chan struct{})
	p.Submit(func(ctx context.Context) {
		select {
		case <-ctx.Done():
			close(ctxCancelled)
		case <-time.After(5 * time.Second):
			// If Stop never cancels ctx, the job waits the full 5 s.
			// The test will then time out after 3 s via waitOrTimeout.
		}
	})

	// Stop cancels the context; the job should unblock quickly.
	go p.Stop()

	waitOrTimeout(t, ctxCancelled, 3*time.Second)
}

// ── Concurrency safety (run with: go test -race) ──────────────────────────────

// TestPool_ConcurrentSubmitAndStop fires many submitter goroutines while Stop
// is called concurrently. The test must complete without data races or panics.
func TestPool_ConcurrentSubmitAndStop(t *testing.T) {
	const (
		numSubmitters = 50
		jobsEach      = 40
	)

	p := worker.New(4, 32)
	p.Start(context.Background())

	var submittersDone sync.WaitGroup
	for i := 0; i < numSubmitters; i++ {
		submittersDone.Add(1)
		go func() {
			defer submittersDone.Done()
			for j := 0; j < jobsEach; j++ {
				p.Submit(func(_ context.Context) {
					// lightweight computation — exercises the hot path
					_ = j * j
				})
			}
		}()
	}

	// Stop concurrently while submitters are active.
	go p.Stop()

	done := make(chan struct{})
	go func() { submittersDone.Wait(); close(done) }()
	waitOrTimeout(t, done, 10*time.Second)
}

// TestPool_HighConcurrencyNoRace simulates 100 goroutines each submitting 20
// jobs against a running pool. Run with -race to confirm zero data races.
func TestPool_HighConcurrencyNoRace(t *testing.T) {
	p := worker.New(8, 64)
	p.Start(context.Background())

	const goroutines = 100
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				p.Submit(func(_ context.Context) {
					time.Sleep(time.Microsecond) // simulate minimal I/O wait
				})
			}
		}()
	}

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	waitOrTimeout(t, done, 15*time.Second)

	// Stop drains the queue: after it returns every job that was enqueued
	// before the pool closed has been processed.
	p.Stop()

	processed, dropped := p.Stats()
	total := processed + dropped
	if total != goroutines*20 {
		t.Errorf("processed+dropped = %d, want %d", total, goroutines*20)
	}
}
