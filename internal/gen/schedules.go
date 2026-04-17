package gen

import "github.com/an-lee/gh-wm/internal/schedule"

// IsCronExpression reports whether input looks like a 5-field GitHub Actions cron string.
func IsCronExpression(input string) bool {
	return schedule.IsCronExpression(input)
}

// FuzzyNormalizeSchedule converts gh-aw-style schedule keywords into a deterministic cron.
// See [schedule.FuzzyNormalizeSchedule].
func FuzzyNormalizeSchedule(scheduleStr, identifier string) string {
	return schedule.FuzzyNormalizeSchedule(scheduleStr, identifier)
}
