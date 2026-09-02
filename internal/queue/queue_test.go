package queue

import (
	"context"
	"errors"
	"testing"
)

func TestLocalQueueSubmitAndFull(t *testing.T) {
	q := NewLocalQueue(1)
	if err := q.Submit(context.Background(), Job{TaskID: "one"}); err != nil {
		t.Fatalf("submit first job: %v", err)
	}
	if err := q.Submit(context.Background(), Job{TaskID: "two"}); !errors.Is(err, ErrFull) {
		t.Fatalf("expected ErrFull, got %v", err)
	}
	if job := <-q.Jobs(); job.TaskID != "one" {
		t.Fatalf("expected job one, got %q", job.TaskID)
	}
}

func TestLocalQueueRejectsWorkAfterClose(t *testing.T) {
	q := NewLocalQueue(1)
	q.Close()
	if err := q.Submit(context.Background(), Job{TaskID: "one"}); !errors.Is(err, ErrClosed) {
		t.Fatalf("expected ErrClosed, got %v", err)
	}
}
