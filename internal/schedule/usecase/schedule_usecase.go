package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
	"github.com/Ramsi97/flowra-back-end/internal/schedule/engine"
	schedinterfaces "github.com/Ramsi97/flowra-back-end/internal/schedule/repository/interfaces"
	taskdomain "github.com/Ramsi97/flowra-back-end/internal/task/domain"
	taskinterfaces "github.com/Ramsi97/flowra-back-end/internal/task/repository/interfaces"
	"github.com/Ramsi97/flowra-back-end/pkg/ai"
)

type scheduleUseCase struct {
	schedRepo schedinterfaces.ScheduleRepository
	taskRepo  taskinterfaces.TaskRepository
}

func NewScheduleUseCase(schedRepo schedinterfaces.ScheduleRepository, taskRepo taskinterfaces.TaskRepository) domain.ScheduleUseCase {
	return &scheduleUseCase{
		schedRepo: schedRepo,
		taskRepo:  taskRepo,
	}
}

// ---------- helpers ----------

func parseDate(date string) (time.Time, error) {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date format (expected YYYY-MM-DD): %w", err)
	}
	return t.UTC(), nil
}

func ctx5s() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (u *scheduleUseCase) buildItemsForDate(userID string, date time.Time) ([]domain.ScheduleItem, error) {
	ctx, cancel := ctx5s()
	defer cancel()

	tasks, err := u.taskRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	prefs := ai.DefaultPrefs(date)
	slots, err := ai.BuildSchedule(tasks, prefs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]domain.ScheduleItem, 0, len(slots))
	for _, s := range slots {
		items = append(items, domain.ScheduleItem{
			UserID:    userID,
			TaskID:    s.TaskID,
			Title:     s.Title,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Status:    "pending",
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	return items, nil
}

// ---------- ScheduleUseCase implementation ----------

// Generate clears the day then inserts a fresh schedule.
func (u *scheduleUseCase) Generate(userID, date string) ([]domain.ScheduleItem, error) {
	t, err := parseDate(date)
	if err != nil {
		return nil, err
	}

	// Always replace existing schedule for that day.
	if err := u.ClearDay(userID, date); err != nil {
		return nil, err
	}

	items, err := u.buildItemsForDate(userID, t)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return []domain.ScheduleItem{}, nil
	}

	ctx, cancel := ctx5s()
	defer cancel()
	if err := u.schedRepo.InsertMany(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

// Regenerate is identical to Generate — it always overwrites.
func (u *scheduleUseCase) Regenerate(userID, date string) ([]domain.ScheduleItem, error) {
	return u.Generate(userID, date)
}

// UpdateItem updates a single schedule item and ripples forward.
func (u *scheduleUseCase) UpdateItem(userID, itemID string, input domain.UpdateItemInput) ([]domain.ScheduleItem, error) {
	ctx, cancel := ctx5s()
	defer cancel()

	// Fetch item to verify ownership and get current state.
	item, err := u.schedRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.UserID != userID {
		return nil, errors.New("schedule item not found")
	}

	// Compute new start / end.
	newStart := item.StartTime
	if input.StartTime != nil {
		newStart = *input.StartTime
	}
	duration := item.EndTime.Sub(item.StartTime)
	if input.DurationMinutes != nil {
		duration = time.Duration(*input.DurationMinutes) * time.Minute
	}
	newEnd := newStart.Add(duration)

	// Persist this item's change (start + end).
	if err := u.schedRepo.UpdateItemTimes(ctx, itemID, newStart, newEnd); err != nil {
		return nil, err
	}

	// Fetch full day for ripple.
	y, m, d := newStart.Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	allItems, err := u.schedRepo.FindByUserAndDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].StartTime.Before(allItems[j].StartTime)
	})

	// Update the changed item in memory so ripple uses the new times.
	changedIdx := -1
	for i, it := range allItems {
		if it.ID == itemID {
			allItems[i].StartTime = newStart
			allItems[i].EndTime = newEnd
			changedIdx = i
			break
		}
	}

	if changedIdx == -1 {
		return allItems, nil
	}

	allItems = engine.ApplyRipple(allItems, changedIdx)

	// Persist rippled updates.
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	for i := changedIdx + 1; i < len(allItems); i++ {
		it := allItems[i]
		_ = u.schedRepo.UpdateItemTimes(ctx2, it.ID, it.StartTime, it.EndTime)
	}

	return allItems, nil
}

// DeleteItem removes one item and leaves a gap (no ripple).
func (u *scheduleUseCase) DeleteItem(userID, itemID string) error {
	ctx, cancel := ctx5s()
	defer cancel()
	item, err := u.schedRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item == nil || item.UserID != userID {
		return errors.New("schedule item not found")
	}
	return u.schedRepo.DeleteByID(ctx, itemID)
}

// ClearDay deletes all items for the given date string.
func (u *scheduleUseCase) ClearDay(userID, date string) error {
	t, err := parseDate(date)
	if err != nil {
		return err
	}
	ctx, cancel := ctx5s()
	defer cancel()
	return u.schedRepo.DeleteByUserAndDate(ctx, userID, t)
}

// ClearWeek deletes 7 days starting from startDate.
func (u *scheduleUseCase) ClearWeek(userID, startDate string) error {
	start, err := parseDate(startDate)
	if err != nil {
		return err
	}
	end := start.AddDate(0, 0, 7)
	ctx, cancel := ctx5s()
	defer cancel()
	return u.schedRepo.DeleteByUserAndDateRange(ctx, userID, start, end)
}

// ClearMonth deletes all items for the given month string (YYYY-MM).
func (u *scheduleUseCase) ClearMonth(userID, month string) error {
	var year, mon int
	if _, err := fmt.Sscanf(month, "%d-%d", &year, &mon); err != nil {
		return errors.New("invalid month format (expected YYYY-MM)")
	}
	ctx, cancel := ctx5s()
	defer cancel()
	return u.schedRepo.DeleteByUserAndMonth(ctx, userID, year, mon)
}

// Fix keeps past items unchanged and reschedules remaining future tasks from currentTime.
func (u *scheduleUseCase) Fix(userID string, currentTime time.Time) ([]domain.ScheduleItem, error) {
	ctx, cancel := ctx5s()
	defer cancel()

	y, m, d := currentTime.Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	allItems, err := u.schedRepo.FindByUserAndDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	// Collect future items to delete.
	var futureIDs []string
	for _, it := range allItems {
		if !it.StartTime.Before(currentTime) {
			futureIDs = append(futureIDs, it.ID)
		}
	}

	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	for _, id := range futureIDs {
		_ = u.schedRepo.DeleteByID(ctx2, id)
	}

	// Build fresh schedule from currentTime.
	tasks, err := u.taskRepo.FindByUserID(ctx2, userID)
	if err != nil {
		return nil, err
	}

	workEnd := time.Date(y, m, d, 18, 0, 0, 0, time.UTC)
	if workEnd.Before(currentTime) {
		return []domain.ScheduleItem{}, nil
	}

	prefs := ai.UserPrefs{WorkStart: currentTime, WorkEnd: workEnd}
	slots, err := ai.BuildSchedule(tasks, prefs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newItems := make([]domain.ScheduleItem, 0, len(slots))
	for _, s := range slots {
		newItems = append(newItems, domain.ScheduleItem{
			UserID:    userID,
			TaskID:    s.TaskID,
			Title:     s.Title,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			Status:    "pending",
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if len(newItems) > 0 {
		ctx3, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel3()
		if err := u.schedRepo.InsertMany(ctx3, newItems); err != nil {
			return nil, err
		}
	}
	return newItems, nil
}

// RemoveTask removes an item from the schedule and resets the task status to "todo".
func (u *scheduleUseCase) RemoveTask(userID, itemID string) error {
	ctx, cancel := ctx5s()
	defer cancel()

	item, err := u.schedRepo.FindByID(ctx, itemID)
	if err != nil {
		return err
	}
	if item == nil || item.UserID != userID {
		return errors.New("schedule item not found")
	}

	if err := u.schedRepo.DeleteByID(ctx, itemID); err != nil {
		return err
	}

	// Reset task back to "todo".
	if item.TaskID != "" {
		todo := "todo"
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		_, _ = u.taskRepo.Update(ctx2, item.TaskID, taskdomain.UpdateTaskInput{Status: &todo})
	}
	return nil
}
