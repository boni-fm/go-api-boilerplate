// Package worker provides a bounded background worker pool.
//
// The pool holds a fixed number of goroutines (workers) and accepts
// background [Job]s through a bounded channel (capacity). When the channel is
// full, [Pool.Submit] sheds the incoming job immediately (returns false)
// instead of blocking, giving callers explicit backpressure signals at high
// throughput.
//
// # Lifecycle
//
//	pool := worker.New(4, 128)    // 4 workers, 128-job buffer
//	pool.Start(ctx)               // launch goroutines (once at startup)
//	...
//	pool.Stop()                   // drain queue, shut down gracefully
//
// # Usage in handlers
//
//	ok := pool.Submit(func(ctx context.Context) {
//	    // fire-and-forget: audit log, metric flush, email dispatch, etc.
//	    if err := sendAuditEvent(ctx, event); err != nil {
//	        log.Errorf("audit: %v", err)
//	    }
//	})
//	if !ok {
//	    log.Warn("worker pool saturated — audit event dropped")
//	}
package worker

import (
	"context"
	"sync"
	"sync/atomic"
)

// Job is a unit of work executed by the pool.
// The context carries the pool's shutdown signal; long-running jobs should
// honour ctx.Done() so they can exit cleanly during graceful shutdown.
type Job func(ctx context.Context)

// Pool dispatches submitted jobs to a fixed set of worker goroutines.
// The zero value is not usable; create instances with [New].
type Pool struct {
	workers int
	jobs    chan Job

	// cancel is set by Start; calling it cancels the context propagated to
	// every running job, enabling cooperative shutdown.
	cancel context.CancelFunc

	wg   sync.WaitGroup
	once sync.Once // makes Stop idempotent

	// mu guards the closed flag together with the channel-send in Submit so
	// that Submit can never race against Stop's close(p.jobs).
	mu     sync.Mutex
	closed bool

	processed atomic.Int64
	dropped   atomic.Int64
}

// New creates a new Pool with the given number of worker goroutines and
// job-queue capacity.
//
//   - workers:  number of concurrent goroutines; clamped to a minimum of 1.
//   - capacity: maximum number of queued (not-yet-executing) jobs; clamped to
//     a minimum of 1.
//
// Call [Pool.Start] once before submitting any jobs.
func New(workers, capacity int) *Pool {
	if workers < 1 {
		workers = 1
	}
	if capacity < 1 {
		capacity = workers * 2
	}
	return &Pool{
		workers: workers,
		jobs:    make(chan Job, capacity),
	}
}

// Start launches the worker goroutines. It must be called exactly once before
// any [Pool.Submit] calls.
//
// The provided context is used as the parent for the pool's internal
// cancellation context, which is forwarded to every running [Job]. Callers
// should prefer passing [context.Background] here; shutdown is coordinated
// through [Pool.Stop], not by cancelling this parent context.
func (p *Pool) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.work(ctx)
	}
}

// work is the goroutine body: drain the jobs channel until it is closed.
func (p *Pool) work(ctx context.Context) {
	defer p.wg.Done()
	for job := range p.jobs {
		job(ctx)
		p.processed.Add(1)
	}
}

// Submit enqueues a job for asynchronous execution. It is non-blocking:
//   - Returns true  — job was enqueued successfully.
//   - Returns false — job was dropped because the queue is full or [Pool.Stop]
//     has already been called. In either case the drop counter is incremented
//     so operators can alert on load shedding via [Pool.Stats].
func (p *Pool) Submit(job Job) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		p.dropped.Add(1)
		return false
	}
	select {
	case p.jobs <- job:
		return true
	default:
		p.dropped.Add(1)
		return false
	}
}

// Stop signals workers to finish draining the existing queue and then exit.
// In-flight jobs are not interrupted; they receive a cancelled context so that
// they can exit early if they check ctx.Done().
//
// Stop blocks until every worker goroutine has returned. It is safe to call
// multiple times (idempotent).
func (p *Pool) Stop() {
	p.once.Do(func() {
		p.mu.Lock()
		p.closed = true
		if p.cancel != nil {
			// Signal running jobs that shutdown is in progress.
			p.cancel()
		}
		close(p.jobs) // signal workers to drain and exit
		p.mu.Unlock()

		p.wg.Wait() // block until all goroutines have returned
	})
}

// Stats returns the cumulative count of jobs processed and dropped since
// [Pool.Start] was called. These counters are suitable for exposing as
// operational metrics or health-check fields.
func (p *Pool) Stats() (processed, dropped int64) {
	return p.processed.Load(), p.dropped.Load()
}
