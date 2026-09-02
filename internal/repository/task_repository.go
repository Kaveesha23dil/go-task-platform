package repository

import (
	"context"
	"errors"

	"github.com/Kaveesha23dil/go-task-platform/internal/model"
)

var ErrTaskNotFound = errors.New("task not found")

// TaskRepository keeps persistence details behind an interface so another
// implementation, such as PostgreSQL, can replace the in-memory store later.
type TaskRepository interface {
	Create(ctx context.Context, task model.Task) (model.Task, error)
	Update(ctx context.Context, task model.Task) (model.Task, error)
	GetAll(ctx context.Context) ([]model.Task, error)
	GetByID(ctx context.Context, id string) (model.Task, error)
	Delete(ctx context.Context, id string) error
}
