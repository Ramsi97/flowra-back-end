package engine

import (
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
)

// RippleResult contains the updated items and any conflicts encountered.
type RippleResult struct {
	Items     []domain.ScheduleItem
	Conflicts []domain.ScheduleItem // Soft conflicts: items that overlap with Hard tasks
}

const workDayEndHour = 23
const workDayStartHour = 9

// ApplyRipple pushes all items after changedIdx forward to be
// contiguous with the changed item, respecting the day boundary and RestDays.
//
// New features:
// 1. Hard Task respect: If we hit a hard task, we stop rippling and flag a conflict.
// 2. Rest Days: If a task overflows past midnight, it skips user-defined rest days.
func ApplyRipple(items []domain.ScheduleItem, changedIdx int, restDays []int) RippleResult {
	if changedIdx >= len(items)-1 {
		return RippleResult{Items: items}
	}

	res := RippleResult{
		Items:     items,
		Conflicts: []domain.ScheduleItem{},
	}

	ref := res.Items[changedIdx]
	cursor := ref.EndTime

	for i := changedIdx + 1; i < len(res.Items); i++ {
		// If we hit a Hard task (like a meeting), the ripple MUST STOP.
		// We flag a conflict and return early.
		if res.Items[i].IsHard {
			// If our current cursor (where the item should start) is after
			// the hard task's start time, we have a conflict.
			if cursor.After(res.Items[i].StartTime) {
				res.Conflicts = append(res.Conflicts, res.Items[i])
				break
			}
		}

		duration := res.Items[i].EndTime.Sub(res.Items[i].StartTime)

		// Handle multi-day overflow
		// If cursor > 11 PM or day changes, push to next work day
		dayEnd := time.Date(cursor.Year(), cursor.Month(), cursor.Day(), workDayEndHour, 0, 0, 0, cursor.Location())
		if cursor.After(dayEnd) || cursor.Equal(dayEnd) {
			cursor = nextWorkDay(cursor, restDays)
		}

		res.Items[i].StartTime = cursor
		res.Items[i].EndTime = cursor.Add(duration)
		cursor = res.Items[i].EndTime
	}

	return res
}

// nextWorkDay finds the next 9 AM start time, skipping any rest days.
func nextWorkDay(curr time.Time, restDays []int) time.Time {
	next := curr.AddDate(0, 0, 1)
	next = time.Date(next.Year(), next.Month(), next.Day(), workDayStartHour, 0, 0, 0, next.Location())

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
