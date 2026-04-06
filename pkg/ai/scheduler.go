package ai

import (
	"sort"
	"time"

	taskdomain "github.com/Ramsi97/flowra-back-end/internal/task/domain"
)

// UserPrefs holds scheduling preferences.
type UserPrefs struct {
	WorkStart time.Time // only Hour/Minute are used; date is ignored
	WorkEnd   time.Time
	RestDays  []int // 0=Sun, 6=Sat
}

// DefaultPrefs returns the default 9 AM – 6 PM working window with Sat/Sun rest.
func DefaultPrefs(forDate time.Time) UserPrefs {
	y, m, d := forDate.Date()
	return UserPrefs{
		WorkStart: time.Date(y, m, d, 9, 0, 0, 0, time.UTC),
		WorkEnd:   time.Date(y, m, d, 18, 0, 0, 0, time.UTC),
		RestDays:  []int{0, 6},
	}
}

// Slot is the output of the AI scheduler: a proposed time block for one task.
type Slot struct {
	TaskID    string
	Title     string
	StartTime time.Time
	EndTime   time.Time
}

// BuildSchedule is the V1 stub scheduler.
//
// It sorts todo tasks by (Priority ASC, Deadline ASC) and packs them
// sequentially inside the user's work window.
//
// Now supports:
// 1. Rest Days: Returns empty if today is a rest day.
// 2. Exceptions: Applies date-specific overrides from the 'exceptions' map.
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

		// Apply Exception if exists for this task.
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

	// 3. Sort: priority ascending (1 = highest), then earliest deadline first.
	sort.Slice(todos, func(i, j int) bool {
		if todos[i].Priority != todos[j].Priority {
			return todos[i].Priority < todos[j].Priority
		}
		return todos[i].Deadline.Before(todos[j].Deadline)
	})

	var slots []Slot
	cursor := prefs.WorkStart

	for _, task := range todos {
		dur := parseDuration(task.Duration)
		if dur <= 0 {
			dur = 60 * time.Minute
		}

		// Handle NewStartTime exception if it exists.
		if ex, ok := exceptions[task.ID]; ok && ex.NewStartTime != nil {
			// If an exception forces a start time, we jump the cursor.
			// This might create a gap (intended).
			if ex.NewStartTime.After(cursor) {
				cursor = *ex.NewStartTime
			}
		}

		if cursor.Add(dur).After(prefs.WorkEnd) {
			break // no more room today
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

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}
