// Package schedule provides gh-aw–compatible fuzzy schedule normalization for task cron strings.
package schedule

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
)

// cronFieldPattern matches valid single cron field tokens (GitHub Actions / POSIX cron).
var cronFieldPattern = regexp.MustCompile(`^[\d\*\-/,]+$`)

// everyNHoursPattern matches "every N hour(s)" for fuzzy schedule expansion (case-insensitive).
var everyNHoursPattern = regexp.MustCompile(`(?i)^every\s+(\d+)\s+hours?$`)

// weeklyOnPattern matches "weekly on <weekday>" for fuzzy schedule expansion (case-insensitive).
var weeklyOnPattern = regexp.MustCompile(`(?i)^weekly\s+on\s+(\S+)$`)

// weekdayNameToDOW maps English weekday names/abbreviations to GitHub cron DOW (0=Sunday … 6=Saturday).
func weekdayNameToDOW(s string) (dow int, ok bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "sun", "sunday":
		return 0, true
	case "mon", "monday":
		return 1, true
	case "tue", "tues", "tuesday":
		return 2, true
	case "wed", "weds", "wednesday":
		return 3, true
	case "thu", "thur", "thurs", "thursday":
		return 4, true
	case "fri", "friday":
		return 5, true
	case "sat", "saturday":
		return 6, true
	default:
		return 0, false
	}
}

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

// scatteredHourlyMinute returns a deterministic minute in 5–54 for hourly / every-N-hours crons.
func scatteredHourlyMinute(identifier string) int {
	return stableHash(identifier, 50) + 5
}

// parseEveryNHours returns (n, true) if s matches "every N hour(s)" with a decimal N; otherwise (0, false).
func parseEveryNHours(s string) (n int, matched bool) {
	m := everyNHoursPattern.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, false
	}
	v, err := strconv.Atoi(m[1])
	if err != nil {
		return 0, true
	}
	return v, true
}

// FuzzyNormalizeSchedule converts gh-aw-style schedule keywords (daily, weekly, hourly, weekly on <weekday>) into a
// deterministic cron using the same weighted pool and FNV-1a hashing as github/gh-aw.
// identifier should be stable (e.g. task file path). Raw 5-field cron expressions are returned whitespace-normalized.
// "every N hours" (1≤N≤23) expands to M */N * * * with scattered M; N=1 matches hourly; N=24 matches daily.
func FuzzyNormalizeSchedule(scheduleStr, identifier string) string {
	scheduleStr = strings.TrimSpace(scheduleStr)
	if scheduleStr == "" {
		return ""
	}
	if IsCronExpression(scheduleStr) {
		return strings.Join(strings.Fields(scheduleStr), " ")
	}
	if n, matched := parseEveryNHours(scheduleStr); matched {
		if n < 1 || n > 24 {
			return scheduleStr
		}
		if n == 24 {
			h, m := weightedDailyTimeSlot(identifier)
			return fmt.Sprintf("%d %d * * *", m, h)
		}
		minute := scatteredHourlyMinute(identifier)
		return fmt.Sprintf("%d */%d * * *", minute, n)
	}
	if m := weeklyOnPattern.FindStringSubmatch(scheduleStr); m != nil {
		dow, ok := weekdayNameToDOW(m[1])
		if !ok {
			return scheduleStr
		}
		h, min := weightedDailyTimeSlot(identifier)
		return fmt.Sprintf("%d %d * * %d", min, h, dow)
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
		return fmt.Sprintf("%d */1 * * *", scatteredHourlyMinute(identifier))
	default:
		return scheduleStr
	}
}
