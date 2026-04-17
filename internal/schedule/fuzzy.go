// Package schedule provides gh-aw–compatible fuzzy schedule normalization for task cron strings.
package schedule

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

// cronFieldPattern matches valid single cron field tokens (GitHub Actions / POSIX cron).
var cronFieldPattern = regexp.MustCompile(`^[\d\*\-/,]+$`)

// IsCronExpression reports whether input looks like a 5-field GitHub Actions cron string.
func IsCronExpression(input string) bool {
	input = strings.TrimSpace(input)
	fields := strings.Fields(input)
	if len(fields) != 5 {
		return false
	}
	for _, field := range fields {
		if !cronFieldPattern.MatchString(field) {
			return false
		}
	}
	return true
}

// timeSlot is a (hour, minute) pair in the weighted daily pool (github/gh-aw compatible).
type timeSlot struct {
	hour   int
	minute int
}

// bestDailyMinutes are odd minutes preferred during the BEST tier (02:00–05:59 UTC), matching gh-aw.
var bestDailyMinutes = []int{7, 13, 23, 37, 43, 53}

func buildWeightedDailyPool() []timeSlot {
	var pool []timeSlot
	for h := 2; h <= 5; h++ {
		for _, m := range bestDailyMinutes {
			pool = append(pool, timeSlot{h, m}, timeSlot{h, m}, timeSlot{h, m})
		}
	}
	for h := 10; h <= 12; h++ {
		for m := 5; m <= 54; m++ {
			pool = append(pool, timeSlot{h, m}, timeSlot{h, m})
		}
	}
	for h := 19; h <= 23; h++ {
		for m := 5; m <= 54; m++ {
			pool = append(pool, timeSlot{h, m})
		}
	}
	return pool
}

// weightedDailyPool matches github/gh-aw/pkg/parser/schedule_fuzzy_scatter.go (622 slots).
var weightedDailyPool = buildWeightedDailyPool()

func stableHash(s string, modulo int) int {
	if modulo <= 0 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return int(h.Sum32() % uint32(modulo))
}

func weightedDailyTimeSlot(identifier string) (hour, minute int) {
	slot := weightedDailyPool[stableHash(identifier, len(weightedDailyPool))]
	return slot.hour, slot.minute
}

// FuzzyNormalizeSchedule converts gh-aw-style schedule keywords (daily, weekly, hourly) into a
// deterministic cron using the same weighted pool and FNV-1a hashing as github/gh-aw.
// identifier should be stable (e.g. task file path). Raw 5-field cron expressions are returned whitespace-normalized.
func FuzzyNormalizeSchedule(scheduleStr, identifier string) string {
	scheduleStr = strings.TrimSpace(scheduleStr)
	if scheduleStr == "" {
		return ""
	}
	if IsCronExpression(scheduleStr) {
		return strings.Join(strings.Fields(scheduleStr), " ")
	}
	switch strings.ToLower(scheduleStr) {
	case "daily":
		h, m := weightedDailyTimeSlot(identifier)
		return fmt.Sprintf("%d %d * * *", m, h)
	case "weekly":
		dow := stableHash(identifier, 7)
		h, m := weightedDailyTimeSlot(identifier)
		return fmt.Sprintf("%d %d * * %d", m, h, dow)
	case "hourly":
		minute := stableHash(identifier, 50) + 5
		return fmt.Sprintf("%d */1 * * *", minute)
	default:
		return scheduleStr
	}
}
