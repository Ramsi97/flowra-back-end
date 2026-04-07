package domain

import (
	"context"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
)

// FocusStatus represents the current state of a user's focus session.
type FocusStatus struct {
	IsActive    bool                 `json:"is_active"`
	CurrentItem *domain.ScheduleItem `json:"current_item,omitempty"`
	BlockedApps []string             `json:"blocked_apps"`
}

// UpdateConfigInput contains fields to update focus preferences.
type UpdateConfigInput struct {
	BlockedApps      *[]string `json:"blocked_apps"`
	FocusModeEnabled *bool     `json:"focus_mode_enabled"`
}

// FocusUseCase defines operations for mobile focus mode.
type FocusUseCase interface {
	GetStatus(ctx context.Context, userID string, now time.Time) (*FocusStatus, error)
	UpdateConfig(ctx context.Context, userID string, input UpdateConfigInput) error
}
