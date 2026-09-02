package worker

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/Kaveesha23dil/go-task-platform/internal/model"
	"github.com/Kaveesha23dil/go-task-platform/internal/queue"
)

type recordingUpdater struct {
	mu       sync.Mutex
	statuses map[string][]model.TaskStatus
	changed  chan struct{}
}

func newRecordingUpdater() *recordingUpdater {
	return &recordingUpdater{statuses: make(map[string][]model.TaskStatus), changed: make(chan struct{}, 100)}
}

func (u *recordingUpdater) UpdateStatus(_ context.Context, id string, status model.TaskStatus) error {
	u.mu.Lock()
	u.statuses[id] = append(u.statuses[id], status)
	u.mu.Unlock()
	u.changed <- struct{}{}
	return nil
}

func (u *recordingUpdater) statusCount() int {
	u.mu.Lock()
	defer u.mu.Unlock()
	total := 0
	for _, statuses := range u.statuses {
		total += len(statuses)
	}
	return total
}

func waitForStatuses(t *testing.T, updater *recordingUpdater, count int) {
	t.Helper()
	deadline := time.After(time.Second)
	for updater.statusCount() < count {
		select {
		case <-updater.changed:
		case <-deadline:
			t.Fatalf("timed out waiting for %d status changes; got %d", count, updater.statusCount())
		}
	}
}

func TestPoolChangesTaskFromRunningToCompleted(t *testing.T) {
	q := queue.NewLocalQueue(1)
	updater := newRecordingUpdater()
	pool := NewPool(1, q, updater, func(context.Context, queue.Job) error { return nil }, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err := pool.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := q.Submit(context.Background(), queue.Job{TaskID: "task-1"}); err != nil {
		t.Fatal(err)
	}
	waitForStatuses(t, updater, 2)
	statuses := updater.statuses["task-1"]
	if statuses[0] != model.TaskStatusRunning || statuses[1] != model.TaskStatusCompleted {
		t.Fatalf("unexpected lifecycle: %v", statuses)
	}
	if err := pool.Shutdown(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestPoolProcessesMultipleTasksConcurrently(t *testing.T) {
	q := queue.NewLocalQueue(3)
	updater := newRecordingUpdater()
	started := make(chan struct{}, 3)
	release := make(chan struct{})
	processor := func(ctx context.Context, _ queue.Job) error {
		started <- struct{}{}
		select {
		case <-release:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	pool := NewPool(3, q, updater, processor, slog.New(slog.NewTextHandler(io.Discard, nil)))
	_ = pool.Start(context.Background())
	for i := 0; i < 3; i++ {
		_ = q.Submit(context.Background(), queue.Job{TaskID: string(rune('a' + i))})
	}
	for i := 0; i < 3; i++ {
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("three jobs did not start concurrently")
		}
	}
	close(release)
	waitForStatuses(t, updater, 6)
	_ = pool.Shutdown(context.Background())
}

func TestPoolGracefulShutdownCancelsProcessing(t *testing.T) {
	q := queue.NewLocalQueue(1)
	updater := newRecordingUpdater()
	started := make(chan struct{})
	processor := func(ctx context.Context, _ queue.Job) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}
	pool := NewPool(1, q, updater, processor, slog.New(slog.NewTextHandler(io.Discard, nil)))
	_ = pool.Start(context.Background())
	_ = q.Submit(context.Background(), queue.Job{TaskID: "task-1"})
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("task did not start")
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := pool.Shutdown(ctx); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	waitForStatuses(t, updater, 2)
	if got := updater.statuses["task-1"][1]; got != model.TaskStatusFailed {
		t.Fatalf("expected failed after cancellation, got %s", got)
	}
}
