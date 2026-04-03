package engine

import (
	"time"

	"github.com/Ramsi97/flowra-back-end/internal/schedule/domain"
)

// workDayEnd is the latest allowed end time (11 PM) to prevent items spilling past midnight.
const workDayEndHour = 23

// ApplyRipple pushes all items after changedIdx forward (or backward) to be
// contiguous with the changed item, respecting the day boundary.
//
// Contract:
//   - items must be sorted by StartTime (ascending).
//   - items[changedIdx] already has its new StartTime / EndTime set.
//   - All items with index > changedIdx on the same day are pushed forward.
func ApplyRipple(items []domain.ScheduleItem, changedIdx int) []domain.ScheduleItem {
	if changedIdx >= len(items)-1 {
		return items // nothing to ripple
	}

	ref := items[changedIdx]
	dayEndBase := time.Date(ref.EndTime.Year(), ref.EndTime.Month(), ref.EndTime.Day(),
		workDayEndHour, 0, 0, 0, ref.EndTime.Location())

	cursor := ref.EndTime

	for i := changedIdx + 1; i < len(items); i++ {
		duration := items[i].EndTime.Sub(items[i].StartTime)

		// Stop rippling if item is on a different day.
		if items[i].StartTime.YearDay() != ref.StartTime.YearDay() ||
			items[i].StartTime.Year() != ref.StartTime.Year() {
			break
		}

		newStart := cursor
		newEnd := newStart.Add(duration)

		// Clamp to work-day boundary — truncate duration rather than overflow.
		if newEnd.After(dayEndBase) {
			newEnd = dayEndBase
		}
		if newStart.After(dayEndBase) {
			// No more room; leave remaining items untouched.
			break
		}

		items[i].StartTime = newStart
		items[i].EndTime = newEnd
		cursor = newEnd
	}

	return items
}
