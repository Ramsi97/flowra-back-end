package ai

import (
	"fmt"
	"sort"
	"time"

	taskdomain "github.com/Ramsi97/flowra-back-end/internal/task/domain"
)

// UserPrefs holds scheduling preferences for a specific day.
type UserPrefs struct {
	WorkStart    time.Time // Specific date/time when work starts
	WorkEnd      time.Time // Specific date/time when work ends
	RestDays     []int     // 0=Sun, 6=Sat
	WorkDayStart string    // raw "HH:MM"
	WorkDayEnd   string    // raw "HH:MM"
}

// DefaultPrefs returns a 9 AM – 6 PM window for the given date.
func DefaultPrefs(forDate time.Time) UserPrefs {
	y, m, d := forDate.Date()
	return UserPrefs{
		WorkStart:    time.Date(y, m, d, 9, 0, 0, 0, time.UTC),
		WorkEnd:      time.Date(y, m, d, 18, 0, 0, 0, time.UTC),
		RestDays:     []int{0, 6},
		WorkDayStart: "09:00",
		WorkDayEnd:   "18:00",
	}
}

// ParseWorkTime converts a "HH:MM" string and a date into a UTC time.Time.
func ParseWorkTime(date time.Time, clock string) (time.Time, error) {
	var h, m int
	if _, err := fmt.Sscanf(clock, "%d:%d", &h, &m); err != nil {
		return time.Time{}, err
	}
	y, mon, d := date.Date()
	return time.Date(y, mon, d, h, m, 0, 0, time.UTC), nil
}

// Slot is the output of the AI scheduler: a proposed time block for one task.
type Slot struct {
	TaskID    string
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

// BuildSchedule packs tasks into the user's work window.
func BuildSchedule(tasks []taskdomain.Task, exceptions map[string]taskdomain.TaskException, prefs UserPrefs) ([]Slot, error) {
	// 1. Check if today is a rest day.
	today := int(prefs.WorkStart.Weekday())
	for _, rd := range prefs.RestDays {
		if rd == today {
			return []Slot{}, nil
		}
	}

	// 2. Filter to "todo" tasks and apply exceptions.
	var todos []taskdomain.Task
	for _, t := range tasks {
		if t.Status != "todo" {
			continue
		}

		if ex, ok := exceptions[t.ID]; ok {
			if ex.IsSkipped {
				continue
			}
			if ex.NewDuration != "" {
				t.Duration = ex.NewDuration
			}
		}
		todos = append(todos, t)
	}

	// 3. Sort: priority ascending, then earliest deadline.
	sort.Slice(todos, func(i, j int) bool {
		if todos[i].Priority != todos[j].Priority {
			return todos[i].Priority < todos[j].Priority
		}
		// Handle nil deadlines: tasks WITH deadlines come before tasks WITHOUT.
		if todos[i].Deadline == nil && todos[j].Deadline != nil {
			return false
		}
		if todos[i].Deadline != nil && todos[j].Deadline == nil {
			return true
		}
		if todos[i].Deadline != nil && todos[j].Deadline != nil {
			return todos[i].Deadline.Before(*todos[j].Deadline)
		}
		return false
	})

	var slots []Slot
	cursor := prefs.WorkStart

	for _, task := range todos {
		dur, _ := time.ParseDuration(task.Duration)
		if dur <= 0 {
			dur = 30 * time.Minute // Defaulting to 30m as requested
		}

		// Handle NewStartTime exception.
		if ex, ok := exceptions[task.ID]; ok && ex.NewStartTime != nil {
			if ex.NewStartTime.After(cursor) {
				cursor = *ex.NewStartTime
			}
		}

		if cursor.Add(dur).After(prefs.WorkEnd) {
			break
		}

		slots = append(slots, Slot{
			TaskID:    task.ID,
			Title:     task.Title,
			StartTime: cursor,
			EndTime:   cursor.Add(dur),
		})
		cursor = cursor.Add(dur)
	}

	return slots, nil
}
