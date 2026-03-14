package scheduler

import "sync"

// WorkerStatus indicates whether a worker is available.
type WorkerStatus string

const (
	WorkerIdle WorkerStatus = "idle"
	WorkerBusy WorkerStatus = "busy"
)

// Worker is a named slot in a per-repo worker pool.
type Worker struct {
	Name     string
	RepoName string
	Status   WorkerStatus
	BeadID   string // non-empty when busy
	Step     string // current step when busy
}

// WorkerPool manages named workers for a single repository.
// Workers don't cross repo boundaries.
type WorkerPool struct {
	mu      sync.Mutex
	repo    string
	workers []*Worker
}

// NewWorkerPool creates a pool with the given named workers.
func NewWorkerPool(repo string, names []string) *WorkerPool {
	workers := make([]*Worker, len(names))
	for i, name := range names {
		workers[i] = &Worker{
			Name:     name,
			RepoName: repo,
			Status:   WorkerIdle,
		}
	}
	return &WorkerPool{repo: repo, workers: workers}
}

// IdleWorker returns the first idle worker, or nil if all are busy.
func (p *WorkerPool) IdleWorker() *Worker {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, w := range p.workers {
		if w.Status == WorkerIdle {
			return w
		}
	}
	return nil
}

// BusyCount returns the number of workers currently assigned work.
func (p *WorkerPool) BusyCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	n := 0
	for _, w := range p.workers {
		if w.Status == WorkerBusy {
			n++
		}
	}
	return n
}

// Assign marks a worker as busy with the given bead and step.
func (p *WorkerPool) Assign(w *Worker, beadID, step string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	w.Status = WorkerBusy
	w.BeadID = beadID
	w.Step = step
}

// Release marks a worker as idle and clears its assignment.
func (p *WorkerPool) Release(w *Worker) {
	p.mu.Lock()
	defer p.mu.Unlock()
	w.Status = WorkerIdle
	w.BeadID = ""
	w.Step = ""
}
