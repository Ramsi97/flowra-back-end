package domain

import (
	"context"
	"time"
)

// Task represents a user-defined piece of work.
type Task struct {
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Duration    string     `json:"duration"` // e.g. "60m", "1h30m"
	Priority    int        `json:"priority"` // 1 = highest
	IsHard      bool       `json:"is_hard"`  // hard deadline?
	Status      string     `json:"status"`   // "todo" | "done" | "skipped"
	Deadline    *time.Time `json:"deadline"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RecurringTask links a recurring pattern to a Task.
type RecurringTask struct {
	ID         string    `json:"id"`
	TaskID     string    `json:"task_id"`
	Type       string    `json:"type"`         // "daily" | "weekly"
	DaysOfWeek []int     `json:"days_of_week"` // 0=Sunday … 6=Saturday
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// UpdateTaskInput contains the fields a caller may change on an existing task.
type UpdateTaskInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Duration    *string    `json:"duration"`
	Priority    *int       `json:"priority"`
	IsHard      *bool      `json:"is_hard"`
	Status      *string    `json:"status"`
	Deadline    *time.Time `json:"deadline"`
}

// TaskUseCase defines all business operations for tasks.
type TaskUseCase interface {
	Create(userID string, task *Task) error
	GetByID(userID, id string) (*Task, error)
	ListByUser(userID string) ([]Task, error)
	Update(userID, id string, input UpdateTaskInput) (*Task, error)
	Delete(userID, id string) error

	// AI Assistant methods
	SuggestDraftTasks(ctx context.Context, description string) ([]Task, error)
	RefineDraftTasks(ctx context.Context, drafts []Task, instruction string) ([]Task, error)
}
