package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/Kaveesha23dil/go-task-platform/internal/model"
	"github.com/Kaveesha23dil/go-task-platform/internal/queue"
	"github.com/Kaveesha23dil/go-task-platform/internal/repository"
)

var (
	ErrInvalidTask      = errors.New("task type and payload are required")
	ErrInvalidPriority  = errors.New("priority must be between 1 and 5")
	ErrQueueUnavailable = errors.New("task queue is unavailable")
)

type CreateTaskInput struct {
	Type     string
	Payload  string
	Priority int
}

type TaskService struct {
	repository repository.TaskRepository
	queue      queue.Submitter
	now        func() time.Time
	newID      func() string
}

func NewTaskService(taskRepository repository.TaskRepository, taskQueue ...queue.Submitter) *TaskService {
	var submitter queue.Submitter
	if len(taskQueue) > 0 {
		submitter = taskQueue[0]
	}
	return &TaskService{
		repository: taskRepository,
		queue:      submitter,
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
	created, err := s.repository.Create(ctx, task)
	if err != nil {
		return model.Task{}, err
	}
	if s.queue == nil {
		return created, nil
	}
	if err := s.queue.Submit(ctx, queue.Job{TaskID: created.ID}); err != nil {
		// Roll back so a task is never left queued when no worker can receive it.
		_ = s.repository.Delete(context.Background(), created.ID)
		return model.Task{}, errors.Join(ErrQueueUnavailable, err)
	}
	return created, nil
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

func (s *TaskService) UpdateStatus(ctx context.Context, id string, status model.TaskStatus) error {
	task, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return err
	}
	task.Status = status
	task.UpdatedAt = s.now().UTC()
	_, err = s.repository.Update(ctx, task)
	return err
}
