package worker

import (
	"context"
	"time"

	"github.com/Kaveesha23dil/go-task-platform/internal/queue"
)

func SimulatedProcessor(duration time.Duration) Processor {
	return func(ctx context.Context, _ queue.Job) error {
		timer := time.NewTimer(duration)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			return nil
		}
	}
}
