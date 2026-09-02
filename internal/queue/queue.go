package queue

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrFull   = errors.New("task queue is full")
	ErrClosed = errors.New("task queue is closed")
)

type Job struct {
	TaskID string
}

type Submitter interface {
	Submit(ctx context.Context, job Job) error
}

type LocalQueue struct {
	jobs      chan Job
	closed    chan struct{}
	closeOnce sync.Once
	mu        sync.RWMutex
}

func NewLocalQueue(size int) *LocalQueue {
	return &LocalQueue{
		jobs:   make(chan Job, size),
		closed: make(chan struct{}),
	}
}

func (q *LocalQueue) Submit(ctx context.Context, job Job) error {
	q.mu.RLock()
	defer q.mu.RUnlock()
	select {
	case <-q.closed:
		return ErrClosed
	default:
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	select {
	case q.jobs <- job:
		return nil
	default:
		return ErrFull
	}
}

func (q *LocalQueue) Jobs() <-chan Job {
	return q.jobs
}

func (q *LocalQueue) Close() {
	q.closeOnce.Do(func() {
		q.mu.Lock()
		defer q.mu.Unlock()
		close(q.closed)
	})
}
