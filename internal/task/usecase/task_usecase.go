package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/task/domain"
	"github.com/Ramsi97/flowra-back-end/internal/task/repository/interfaces"
)

type taskUseCase struct {
	repo interfaces.TaskRepository
}

func NewTaskUseCase(repo interfaces.TaskRepository) domain.TaskUseCase {
	return &taskUseCase{repo: repo}
}

func (u *taskUseCase) Create(userID string, task *domain.Task) error {
	if task.Title == "" {
		return errors.New("title is required")
	}
	if task.Status == "" {
		task.Status = "todo"
	}
	now := time.Now()
	task.UserID = userID
	task.CreatedAt = now
	task.UpdatedAt = now

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return u.repo.Create(ctx, task)
}

func (u *taskUseCase) GetByID(userID, id string) (*domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	task, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, errors.New("task not found")
	}
	if task.UserID != userID {
		return nil, errors.New("task not found")
	}
	return task, nil
}

func (u *taskUseCase) ListByUser(userID string) ([]domain.Task, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return u.repo.FindByUserID(ctx, userID)
}

func (u *taskUseCase) Update(userID, id string, input domain.UpdateTaskInput) (*domain.Task, error) {
	// Ensure the task belongs to this user first.
	if _, err := u.GetByID(userID, id); err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	updated, err := u.repo.Update(ctx, id, input)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, errors.New("task not found")
	}
	return updated, nil
}

func (u *taskUseCase) Delete(userID, id string) error {
	// Ensure the task belongs to this user first.
	if _, err := u.GetByID(userID, id); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return u.repo.Delete(ctx, id)
}
