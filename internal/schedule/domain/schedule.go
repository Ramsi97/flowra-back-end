package domain

import "time"

// ScheduleItem is a single scheduled block of time linked to a task.
type ScheduleItem struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	TaskID    string    `json:"task_id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"` // "pending" | "done" | "skipped"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateItemInput contains the mutable fields for a schedule item edit.
type UpdateItemInput struct {
	StartTime       *time.Time `json:"start_time"`
	DurationMinutes *int       `json:"duration_minutes"`
}

// GenerateRequest is the payload for POST /schedule/generate.
type GenerateRequest struct {
	Date string `json:"date"` // "YYYY-MM-DD"
}

// FixRequest is the payload for POST /schedule/fix.
type FixRequest struct {
	CurrentTime time.Time `json:"current_time"`
}

// ScheduleUseCase defines all business operations for schedules.
type ScheduleUseCase interface {
	Generate(userID, date string) ([]ScheduleItem, error)
	Regenerate(userID, date string) ([]ScheduleItem, error)
	UpdateItem(userID, itemID string, input UpdateItemInput) ([]ScheduleItem, error)
	DeleteItem(userID, itemID string) error
	ClearDay(userID, date string) error
	ClearWeek(userID, startDate string) error
	ClearMonth(userID, month string) error
	Fix(userID string, currentTime time.Time) ([]ScheduleItem, error)
	RemoveTask(userID, itemID string) error
}
