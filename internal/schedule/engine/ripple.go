package engine

import (
	"fmt"
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
)

type RippleResult struct {
	Items     []domain.ScheduleItem
	Conflicts []domain.ScheduleItem
}

// ApplyRipple pushes all items after changedIdx forward to be
// contiguous with the changed item, respecting dynamic work windows and RestDays.
func ApplyRipple(items []domain.ScheduleItem, changedIdx int, restDays []int, workDayStart, workDayEnd string) RippleResult {
	if changedIdx >= len(items)-1 {
		return RippleResult{Items: items}
	}

	res := RippleResult{
		Items:     items,
		Conflicts: []domain.ScheduleItem{},
	}

	ref := res.Items[changedIdx]
	cursor := ref.EndTime

	// Parse work hours for overflow logic
	startH, startM := parseClock(workDayStart, 9, 0)
	endH, endM := parseClock(workDayEnd, 22, 0) // Defaulting to 10 PM if invalid

	for i := changedIdx + 1; i < len(res.Items); i++ {
		if res.Items[i].IsHard {
			if cursor.After(res.Items[i].StartTime) {
				res.Conflicts = append(res.Conflicts, res.Items[i])
				break
			}
		}

		duration := res.Items[i].EndTime.Sub(res.Items[i].StartTime)

		// Check if we hit the user's specific end of day.
		dayEnd := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), endH, endM, 0, 0, cursor.Location())
		if cursor.After(dayEnd) || cursor.Equal(dayEnd) {
			cursor = nextWorkDay(cursor, restDays, startH, startM)
		}

		res.Items[i].StartTime = cursor
		res.Items[i].EndTime = cursor.Add(duration)
		cursor = res.Items[i].EndTime
	}

	return res
}

func parseClock(clock string, defH, defM int) (int, int) {
	var h, m int
	if _, err := fmt.Sscanf(clock, "%d:%d", &h, &m); err != nil {
		return defH, defM
	}
	return h, m
}

func nextWorkDay(curr time.Time, restDays []int, startH, startM int) time.Time {
	next := curr.AddDate(0, 0, 1)
	next = time.Date(next.Year(), next.Month(), next.Day(), startH, startM, 0, 0, next.Location())

	for isRestDay(next, restDays) {
		next = next.AddDate(0, 0, 1)
	}
	return next
}

func isRestDay(t time.Time, restDays []int) bool {
	wd := int(t.Weekday())
	for _, rd := range restDays {
		if rd == wd {
			return true
		}
	}
	return false
}
