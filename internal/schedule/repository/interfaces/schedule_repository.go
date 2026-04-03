package interfaces

import (
	"context"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
)

// ScheduleRepository defines all data-access operations for schedule items.
type ScheduleRepository interface {
	InsertMany(ctx context.Context, items []domain.ScheduleItem) error
	FindByID(ctx context.Context, id string) (*domain.ScheduleItem, error)
	FindByUserAndDate(ctx context.Context, userID string, date time.Time) ([]domain.ScheduleItem, error)
	FindByUserAndDateRange(ctx context.Context, userID string, start, end time.Time) ([]domain.ScheduleItem, error)
	// UpdateItemTimes updates start_time and end_time of a single item.
	UpdateItemTimes(ctx context.Context, id string, start, end time.Time) error
	Update(ctx context.Context, id string, input domain.UpdateItemInput) (*domain.ScheduleItem, error)
	DeleteByID(ctx context.Context, id string) error
	DeleteByUserAndDate(ctx context.Context, userID string, date time.Time) error
	DeleteByUserAndDateRange(ctx context.Context, userID string, start, end time.Time) error
	DeleteByUserAndMonth(ctx context.Context, userID string, year, month int) error
}
