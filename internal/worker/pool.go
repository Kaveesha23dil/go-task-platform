package worker

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Kaveesha23dil/go-task-platform/internal/model"
	"github.com/Kaveesha23dil/go-task-platform/internal/queue"
)

var ErrAlreadyStarted = errors.New("worker pool already started")

type JobSource interface {
	Jobs() <-chan queue.Job
	Close()
}

type StatusUpdater interface {
	UpdateStatus(ctx context.Context, taskID string, status model.TaskStatus) error
}

type Processor func(ctx context.Context, job queue.Job) error

type Pool struct {
	workerCount int
	queue       JobSource
	updater     StatusUpdater
	processor   Processor
	logger      *slog.Logger

	mu      sync.Mutex
	started bool
	cancel  context.CancelFunc
	wg      sync.WaitGroup
}

func NewPool(workerCount int, taskQueue JobSource, updater StatusUpdater, processor Processor, logger *slog.Logger) *Pool {
	if logger == nil {
		logger = slog.Default()
	}
	return &Pool{workerCount: workerCount, queue: taskQueue, updater: updater, processor: processor, logger: logger}
}

func (p *Pool) Start(parent context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.started {
		return ErrAlreadyStarted
	}
	ctx, cancel := context.WithCancel(parent)
	p.cancel = cancel
	p.started = true
	for i := 1; i <= p.workerCount; i++ {
		p.wg.Add(1)
		go p.runWorker(ctx, fmt.Sprintf("worker-%d", i))
	}
	return nil
}

func (p *Pool) runWorker(ctx context.Context, workerID string) {
	defer p.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-p.queue.Jobs():
			p.process(ctx, workerID, job)
		}
	}
}

func (p *Pool) process(ctx context.Context, workerID string, job queue.Job) {
	if err := p.updater.UpdateStatus(ctx, job.TaskID, model.TaskStatusRunning); err != nil {
		p.logger.Error("worker could not start task", "worker", workerID, "task_id", job.TaskID, "error", err)
		return
	}
	p.logger.Info("worker started task", "worker", workerID, "task_id", job.TaskID)

	status := model.TaskStatusCompleted
	if err := p.processor(ctx, job); err != nil {
		status = model.TaskStatusFailed
		p.logger.Warn("worker failed task", "worker", workerID, "task_id", job.TaskID, "error", err)
	}
	// A short-lived context lets the final status be stored even when shutdown
	// cancelled the processing context.
	if err := p.updater.UpdateStatus(context.WithoutCancel(ctx), job.TaskID, status); err != nil {
		p.logger.Error("worker could not update task", "worker", workerID, "task_id", job.TaskID, "error", err)
		return
	}
	if status == model.TaskStatusCompleted {
		p.logger.Info("worker completed task", "worker", workerID, "task_id", job.TaskID)
	}
}

func (p *Pool) Shutdown(ctx context.Context) error {
	p.queue.Close()
	p.mu.Lock()
	if p.cancel != nil {
		p.cancel()
	}
	p.mu.Unlock()

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
