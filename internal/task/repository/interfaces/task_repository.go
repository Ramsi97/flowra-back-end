package interfaces

import (
	"context"

	"github.com/Ramsi97/flowra-back-end/internal/task/domain"
)

// TaskRepository defines data-access operations for tasks.
type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	FindByID(ctx context.Context, id string) (*domain.Task, error)
	FindByUserID(ctx context.Context, userID string) ([]domain.Task, error)
	Update(ctx context.Context, id string, input domain.UpdateTaskInput) (*domain.Task, error)
	Delete(ctx context.Context, id string) error
}
