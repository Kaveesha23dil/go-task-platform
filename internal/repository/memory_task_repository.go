package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/Kaveesha23dil/go-task-platform/internal/model"
)

type MemoryTaskRepository struct {
	mu    sync.RWMutex
	tasks map[string]model.Task
}

func NewMemoryTaskRepository() *MemoryTaskRepository {
	return &MemoryTaskRepository{tasks: make(map[string]model.Task)}
}

func (r *MemoryTaskRepository) Create(_ context.Context, task model.Task) (model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tasks[task.ID] = task
	return task, nil
}

func (r *MemoryTaskRepository) GetAll(_ context.Context) ([]model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tasks := make([]model.Task, 0, len(r.tasks))
	for _, task := range r.tasks {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
	})
	return tasks, nil
}

func (r *MemoryTaskRepository) GetByID(_ context.Context, id string) (model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	task, ok := r.tasks[id]
	if !ok {
		return model.Task{}, ErrTaskNotFound
	}
	return task, nil
}

func (r *MemoryTaskRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.tasks[id]; !ok {
		return ErrTaskNotFound
	}
	delete(r.tasks, id)
	return nil
}
