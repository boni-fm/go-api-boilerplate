// Package worker itu worker pool yang bisa nampung job background.
//
// Dia punya sejumlah goroutine yang kerja paralel, dan queue buat nampung job.
// Kalau queue udah penuh, job langsung dibuang (return false) — gak nge-block.
//
// # Cara pakai
//
//	pool := worker.New(4, 128)    // 4 worker, buffer 128 job
//	pool.Start(ctx)               // jalanin goroutine-nya
//	...
//	pool.Stop()                   // drain queue, terus mati dengan elegan
//
// # Contoh di handler
//
//	ok := pool.Submit(func(ctx context.Context) {
//	    // fire-and-forget: audit log, flush metrik, kirim email, dll
//	    if err := sendAuditEvent(ctx, event); err != nil {
//	        log.Errorf("audit: %v", err)
//	    }
//	})
//	if !ok {
//	    log.Warn("worker pool penuh — job dibuang")
//	}
package worker

import (
	"context"
	"sync"
	"sync/atomic"
)

// Job adalah satu unit kerjaan yang dijalankan sama pool.
// Context-nya bawa sinyal shutdown — job yang lama-lama sebaiknya cek ctx.Done()
// biar bisa berhenti dengan bersih pas shutdown.
type Job func(ctx context.Context)

// Pool nyimpen dan ngedistribusiin job ke goroutine-goroutine yang udah siap kerja.
// Jangan pakai zero value — bikin lewat [New] aja.
type Pool struct {
	workers int
	jobs    chan Job

	// cancel dipasang pas Start dipanggil — buat cancel context yang dikasih ke tiap job.
	cancel context.CancelFunc

	wg   sync.WaitGroup
	once sync.Once // biar Stop bisa dipanggil berkali-kali tanpa meledak

	// mu jaga flag closed + pengiriman ke channel di Submit,
	// supaya Submit gak balapan sama close(p.jobs) di Stop.
	mu     sync.Mutex
	closed bool

	processed atomic.Int64
	dropped   atomic.Int64
}

// New bikin Pool baru.
//
//   - workers:  jumlah goroutine — minimal 1.
//   - capacity: kapasitas queue — minimal 1.
//
// Panggil [Pool.Start] dulu sebelum submit job apapun.
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

// Start jalanin goroutine-goroutine worker-nya. Dipanggil sekali aja pas startup.
//
// Context yang dikasih jadi parent buat context internal pool,
// yang nantinya diterusin ke tiap Job. Lebih baik kasih context.Background() di sini;
// urusan shutdown dihandle lewat [Pool.Stop].
func (p *Pool) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.work(ctx)
	}
}

// work itu isi goroutine-nya: terus ambil job dari channel sampai ditutup.
func (p *Pool) work(ctx context.Context) {
	defer p.wg.Done()
	for job := range p.jobs {
		job(ctx)
		p.processed.Add(1)
	}
}

// Submit masukin job ke antrian. Non-blocking:
//   - Return true  — job berhasil masuk antrian.
//   - Return false — job dibuang karena antrian penuh atau pool udah di-Stop.
//     Keduanya nambah counter dropped buat monitoring.
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

// Stop minta worker buat beresin sisa antrian terus berhenti.
// Job yang lagi jalan dapet context yang di-cancel sebagai tanda saatnya cabut.
// Job yang udah antri tapi belum jalan juga kena cancel context-nya —
// kalau job butuh tetap hidup pas shutdown, bikin context sendiri aja di dalamnya.
//
// Stop ngeblock sampe semua goroutine beres. Aman dipanggil berkali-kali.
func (p *Pool) Stop() {
	p.once.Do(func() {
		p.mu.Lock()
		p.closed = true
		if p.cancel != nil {
			p.cancel() // kasih tau job yang lagi jalan: waktunya shutdown
		}
		close(p.jobs) // kasih tau worker buat drain terus keluar
		p.mu.Unlock()

		p.wg.Wait() // tunggu sampe semua goroutine beneran beres
	})
}

// Stats balikin total job yang berhasil diproses dan yang dibuang
// sejak Start dipanggil. Cocok buat metrik atau health check.
func (p *Pool) Stats() (processed, dropped int64) {
	return p.processed.Load(), p.dropped.Load()
}
