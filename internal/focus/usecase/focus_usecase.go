package usecase

import (
	"context"
	"time"

	authinterfaces "github.com/Ramsi97/flowra-back-end/internal/auth/repository/interfaces"
	focusdomain "github.com/Ramsi97/flowra-back-end/internal/focus/domain"
	scheduldomain "github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
	schedinterfaces "github.com/Ramsi97/flowra-back-end/internal/schedule/repository/interfaces"
)

type focusUseCase struct {
	authRepo  authinterfaces.AuthRepository
	schedRepo schedinterfaces.ScheduleRepository
}

func NewFocusUseCase(ar authinterfaces.AuthRepository, sr schedinterfaces.ScheduleRepository) focusdomain.FocusUseCase {
	return &focusUseCase{
		authRepo:  ar,
		schedRepo: sr,
	}
}

func (u *focusUseCase) GetStatus(ctx context.Context, userID string, now time.Time) (*focusdomain.FocusStatus, error) {
	// 1. Fetch user to get blocked apps and global toggle
	user, err := u.authRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return &focusdomain.FocusStatus{IsActive: false, BlockedApps: []string{}}, nil
	}

	// 2. If focus mode is globally disabled, return inactive
	if !user.FocusModeEnabled {
		return &focusdomain.FocusStatus{IsActive: false, BlockedApps: user.BlockedApps}, nil
	}

	// 3. Fetch schedule for today to see if we are in a work session
	y, m, d := now.Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	items, err := u.schedRepo.FindByUserAndDate(ctx, userID, today)
	if err != nil {
		return nil, err
	}

	var activeItem *scheduldomain.ScheduleItem
	isActive := false

	for _, item := range items {
		// If current time is between start and end of a pending/todo item
		if (now.After(item.StartTime) || now.Equal(item.StartTime)) && now.Before(item.EndTime) {
			if item.Status == "pending" || item.Status == "todo" {
				isActive = true
				activeItem = &item
				break
			}
		}
	}

	return &focusdomain.FocusStatus{
		IsActive:    isActive,
		CurrentItem: activeItem,
		BlockedApps: user.BlockedApps,
	}, nil
}

func (u *focusUseCase) UpdateConfig(ctx context.Context, userID string, input focusdomain.UpdateConfigInput) error {
	user, err := u.authRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return nil
	}

	if input.BlockedApps != nil {
		user.BlockedApps = *input.BlockedApps
	}
	if input.FocusModeEnabled != nil {
		user.FocusModeEnabled = *input.FocusModeEnabled
	}

	return u.authRepo.UpdateUser(ctx, user)
}
