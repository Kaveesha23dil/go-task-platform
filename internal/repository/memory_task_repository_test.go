package repository

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Kaveesha23dil/go-task-platform/internal/model"
)

func TestMemoryTaskRepositoryConcurrentAccess(t *testing.T) {
	repo := NewMemoryTaskRepository()
	ctx := context.Background()
	const count = 100
	var wg sync.WaitGroup
	for i := 0; i < count; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			task := model.Task{ID: fmt.Sprintf("task-%d", index), CreatedAt: time.Now()}
			if _, err := repo.Create(ctx, task); err != nil {
				t.Errorf("create: %v", err)
			}
			if _, err := repo.GetByID(ctx, task.ID); err != nil {
				t.Errorf("get: %v", err)
			}
		}(i)
	}
	wg.Wait()
	tasks, err := repo.GetAll(ctx)
	if err != nil || len(tasks) != count {
		t.Fatalf("expected %d tasks, got %d (error: %v)", count, len(tasks), err)
	}
}
