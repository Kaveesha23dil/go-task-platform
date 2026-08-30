package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Kaveesha23dil/go-task-platform/internal/model"
	"github.com/Kaveesha23dil/go-task-platform/internal/repository"
)

var (
	ErrInvalidTask     = errors.New("task type and payload are required")
	ErrInvalidPriority = errors.New("priority must be between 1 and 5")
)

type CreateTaskInput struct {
	Type     string
	Payload  string
	Priority int
}

type TaskService struct {
	repository repository.TaskRepository
	now        func() time.Time
	newID      func() string
}

func NewTaskService(taskRepository repository.TaskRepository) *TaskService {
	return &TaskService{
		repository: taskRepository,
		now:        time.Now,
		newID:      uuid.NewString,
	}
}

func (s *TaskService) Create(ctx context.Context, input CreateTaskInput) (model.Task, error) {
	taskType := strings.TrimSpace(input.Type)
	payload := strings.TrimSpace(input.Payload)
	if taskType == "" || payload == "" {
		return model.Task{}, ErrInvalidTask
	}
	if input.Priority < 1 || input.Priority > 5 {
		return model.Task{}, ErrInvalidPriority
	}

	now := s.now().UTC()
	task := model.Task{
		ID:        s.newID(),
		Type:      taskType,
		Payload:   payload,
		Status:    model.TaskStatusQueued,
		Priority:  input.Priority,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return s.repository.Create(ctx, task)
}

func (s *TaskService) GetAll(ctx context.Context) ([]model.Task, error) {
	return s.repository.GetAll(ctx)
}

func (s *TaskService) GetByID(ctx context.Context, id string) (model.Task, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *TaskService) Delete(ctx context.Context, id string) error {
	return s.repository.Delete(ctx, id)
}
