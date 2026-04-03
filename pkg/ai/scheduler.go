package ai

import (
	"sort"
	"time"

	taskdomain "github.com/Ramsi97/flowra-back-end/internal/task/domain"
)

// UserPrefs holds scheduling preferences.
// V1: hardcoded defaults — will be moved to user profile in a future iteration.
type UserPrefs struct {
	WorkStart time.Time // only Hour/Minute are used; date is ignored
	WorkEnd   time.Time
}

// DefaultPrefs returns the default 9 AM – 6 PM working window.
func DefaultPrefs(forDate time.Time) UserPrefs {
	y, m, d := forDate.Date()
	return UserPrefs{
		WorkStart: time.Date(y, m, d, 9, 0, 0, 0, time.UTC),
		WorkEnd:   time.Date(y, m, d, 18, 0, 0, 0, time.UTC),
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
// Tasks that don't fit are skipped gracefully.
//
// Replace body with real AI (Gemini) call in a later iteration.
func BuildSchedule(tasks []taskdomain.Task, prefs UserPrefs) ([]Slot, error) {
	// Filter to "todo" tasks only.
	var todos []taskdomain.Task
	for _, t := range tasks {
		if t.Status == "todo" {
			todos = append(todos, t)
		}
	}

	// Sort: priority ascending (1 = highest), then earliest deadline first.
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
			dur = 60 * time.Minute // default 60 min if unparseable
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

// parseDuration parses strings like "60m", "1h30m", "2h".
func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	return d
}
