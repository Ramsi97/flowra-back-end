package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	authinterfaces "github.com/Ramsi97/flowra-back-end/internal/auth/repository/interfaces"
	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
	"github.com/Ramsi97/flowra-back-end/internal/schedule/engine"
	schedinterfaces "github.com/Ramsi97/flowra-back-end/internal/schedule/repository/interfaces"
	taskdomain "github.com/Ramsi97/flowra-back-end/internal/task/domain"
	taskinterfaces "github.com/Ramsi97/flowra-back-end/internal/task/repository/interfaces"
	"github.com/Ramsi97/flowra-back-end/pkg/ai"
)

type scheduleUseCase struct {
	schedRepo     schedinterfaces.ScheduleRepository
	taskRepo      taskinterfaces.TaskRepository
	exceptionRepo taskdomain.ExceptionRepository
	authRepo      authinterfaces.AuthRepository
}

func NewScheduleUseCase(
	schedRepo schedinterfaces.ScheduleRepository,
	taskRepo taskinterfaces.TaskRepository,
	exceptionRepo taskdomain.ExceptionRepository,
	authRepo authinterfaces.AuthRepository,
) domain.ScheduleUseCase {
	return &scheduleUseCase{
		schedRepo:     schedRepo,
		taskRepo:      taskRepo,
		exceptionRepo: exceptionRepo,
		authRepo:      authRepo,
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

func (u *scheduleUseCase) getPrefs(userID string, date time.Time) ai.UserPrefs {
	prefs := ai.DefaultPrefs(date)
	ctx, cancel := ctx5s()
	defer cancel()

	user, err := u.authRepo.FindByID(ctx, userID)
	if err == nil && user != nil {
		if len(user.RestDays) > 0 {
			prefs.RestDays = user.RestDays
		}
	}
	return prefs
}

func (u *scheduleUseCase) buildItemsForDate(userID string, date time.Time) ([]domain.ScheduleItem, error) {
	ctx, cancel := ctx5s()
	defer cancel()

	tasks, err := u.taskRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Fetch exceptions for these tasks on this date.
	exceptions := make(map[string]taskdomain.TaskException)
	dateStr := date.Format("2006-01-02")
	for _, t := range tasks {
		ex, err := u.exceptionRepo.FindByTaskAndDate(ctx, t.ID, dateStr)
		if err == nil && ex != nil {
			exceptions[t.ID] = *ex
		}
	}

	prefs := u.getPrefs(userID, date)
	slots, err := ai.BuildSchedule(tasks, exceptions, prefs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	items := make([]domain.ScheduleItem, 0, len(slots))
	for _, s := range slots {
		// Find original task to get IsHard flag.
		var isHard bool
		for _, t := range tasks {
			if t.ID == s.TaskID {
				isHard = t.IsHard
				break
			}
		}

		items = append(items, domain.ScheduleItem{
			UserID:    userID,
			TaskID:    s.TaskID,
			Title:     s.Title,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			IsHard:    isHard,
			Status:    "pending",
			CreatedAt: now,
			UpdatedAt: now,
		})
	}
	return items, nil
}

// ---------- ScheduleUseCase implementation ----------

func (u *scheduleUseCase) Generate(userID, date string) ([]domain.ScheduleItem, error) {
	t, err := parseDate(date)
	if err != nil {
		return nil, err
	}

	if err := u.ClearDay(userID, date); err != nil {
		return nil, err
	}

	items, err := u.buildItemsForDate(userID, t)
	if err != nil {
		return nil, err
	}

	ctx, cancel := ctx5s()
	defer cancel()
	if err := u.schedRepo.InsertMany(ctx, items); err != nil {
		return nil, err
	}
	return items, nil
}

func (u *scheduleUseCase) Regenerate(userID, date string) ([]domain.ScheduleItem, error) {
	return u.Generate(userID, date)
}

func (u *scheduleUseCase) UpdateItem(userID, itemID string, input domain.UpdateItemInput) ([]domain.ScheduleItem, error) {
	ctx, cancel := ctx5s()
	defer cancel()

	item, err := u.schedRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.UserID != userID {
		return nil, errors.New("schedule item not found")
	}

	newStart := item.StartTime
	if input.StartTime != nil {
		newStart = *input.StartTime
	}
	duration := item.EndTime.Sub(item.StartTime)
	if input.DurationMinutes != nil {
		duration = time.Duration(*input.DurationMinutes) * time.Minute
	}
	newEnd := newStart.Add(duration)

	if err := u.schedRepo.UpdateItemTimes(ctx, itemID, newStart, newEnd); err != nil {
		return nil, err
	}

	y, m, d := newStart.Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	allItems, err := u.schedRepo.FindByUserAndDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	sort.Slice(allItems, func(i, j int) bool {
		return allItems[i].StartTime.Before(allItems[j].StartTime)
	})

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

	prefs := u.getPrefs(userID, date)
	rippleRes := engine.ApplyRipple(allItems, changedIdx, prefs.RestDays)

	// Persist rippled updates.
	for i := changedIdx + 1; i < len(rippleRes.Items); i++ {
		it := rippleRes.Items[i]
		_ = u.schedRepo.UpdateItemTimes(ctx, it.ID, it.StartTime, it.EndTime)
	}

	// NOTE: In a real app, we would return rippleRes.Conflicts to the user.
	// For now we return the items list.
	return rippleRes.Items, nil
}

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

func (u *scheduleUseCase) ClearDay(userID, date string) error {
	t, err := parseDate(date)
	if err != nil {
		return err
	}
	ctx, cancel := ctx5s()
	defer cancel()
	return u.schedRepo.DeleteByUserAndDate(ctx, userID, t)
}

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

func (u *scheduleUseCase) ClearMonth(userID, month string) error {
	var year, mon int
	if _, err := fmt.Sscanf(month, "%d-%d", &year, &mon); err != nil {
		return errors.New("invalid month format (expected YYYY-MM)")
	}
	ctx, cancel := ctx5s()
	defer cancel()
	return u.schedRepo.DeleteByUserAndMonth(ctx, userID, year, mon)
}

func (u *scheduleUseCase) Fix(userID string, currentTime time.Time) ([]domain.ScheduleItem, error) {
	ctx, cancel := ctx5s()
	defer cancel()

	y, m, d := currentTime.Date()
	date := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	allItems, err := u.schedRepo.FindByUserAndDate(ctx, userID, date)
	if err != nil {
		return nil, err
	}

	var futureIDs []string
	for _, it := range allItems {
		if !it.StartTime.Before(currentTime) {
			futureIDs = append(futureIDs, it.ID)
		}
	}

	for _, id := range futureIDs {
		_ = u.schedRepo.DeleteByID(ctx, id)
	}

	tasks, err := u.taskRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Fetch exceptions for Fix.
	exceptions := make(map[string]taskdomain.TaskException)
	dateStr := date.Format("2006-01-02")
	for _, t := range tasks {
		ex, err := u.exceptionRepo.FindByTaskAndDate(ctx, t.ID, dateStr)
		if err == nil && ex != nil {
			exceptions[t.ID] = *ex
		}
	}

	prefs := u.getPrefs(userID, currentTime)
	prefs.WorkStart = currentTime // schedule from now

	if prefs.WorkEnd.Before(currentTime) {
		return []domain.ScheduleItem{}, nil
	}

	slots, err := ai.BuildSchedule(tasks, exceptions, prefs)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	newItems := make([]domain.ScheduleItem, 0, len(slots))
	for _, s := range slots {
		var isHard bool
		for _, t := range tasks {
			if t.ID == s.TaskID {
				isHard = t.IsHard
				break
			}
		}

		newItems = append(newItems, domain.ScheduleItem{
			UserID:    userID,
			TaskID:    s.TaskID,
			Title:     s.Title,
			StartTime: s.StartTime,
			EndTime:   s.EndTime,
			IsHard:    isHard,
			Status:    "pending",
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	if len(newItems) > 0 {
		if err := u.schedRepo.InsertMany(ctx, newItems); err != nil {
			return nil, err
		}
	}
	return newItems, nil
}

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

	if item.TaskID != "" {
		todo := "todo"
		ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel2()
		_, _ = u.taskRepo.Update(ctx2, item.TaskID, taskdomain.UpdateTaskInput{Status: &todo})
	}
	return nil
}
