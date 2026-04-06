package domain

import (
	"context"
	"time"
)

// TaskException represents an override for a recurring task on a specific date.
type TaskException struct {
	ID           string     `json:"id" bson:"_id,omitempty"`
	TaskID       string     `json:"task_id" bson:"task_id"`
	Date         string     `json:"date" bson:"date"` // YYYY-MM-DD
	NewDuration  string     `json:"duration" bson:"new_duration"`
	NewStartTime *time.Time `json:"start_time" bson:"new_start_time"`
	IsSkipped    bool       `json:"is_skipped" bson:"is_skipped"`
	CreatedAt    time.Time  `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" bson:"updated_at"`
}

// ExceptionRepository defines data-access operations for task exceptions.
type ExceptionRepository interface {
	Create(ctx context.Context, ex *TaskException) error
	FindByTaskAndDate(ctx context.Context, taskID, date string) (*TaskException, error)
	ListByTask(ctx context.Context, taskID string) ([]TaskException, error)
	Delete(ctx context.Context, id string) error
}
