package schedule

import (
	"strings"
	"testing"
)

func TestFuzzyNormalizeSchedule(t *testing.T) {
	t.Parallel()
	id := "/repo/.wm/tasks/doc.md"
	d1 := FuzzyNormalizeSchedule("daily", id)
	d2 := FuzzyNormalizeSchedule("daily", id)
	if d1 != d2 {
		t.Fatalf("daily not stable: %q vs %q", d1, d2)
	}
	if d1 == FuzzyNormalizeSchedule("daily", "/other/path.md") {
		t.Fatal("daily should differ by identifier")
	}
	if FuzzyNormalizeSchedule("custom", id) != "custom" {
		t.Fatal()
	}
	raw := "0  0   * * *"
	if FuzzyNormalizeSchedule(raw, id) != "0 0 * * *" {
		t.Fatal("raw cron should normalize whitespace")
	}
}

func TestFuzzyNormalizeSchedule_EveryNHours(t *testing.T) {
	t.Parallel()
	id := "/repo/.wm/tasks/doc.md"
	e3a := FuzzyNormalizeSchedule("every 3 hours", id)
	e3b := FuzzyNormalizeSchedule("every 3 hours", id)
	if e3a != e3b {
		t.Fatalf("every 3 hours not stable: %q vs %q", e3a, e3b)
	}
	if !strings.Contains(e3a, "*/3") {
		t.Fatalf("expected */3 in cron, got %q", e3a)
	}
	if FuzzyNormalizeSchedule("every 1 hour", id) != FuzzyNormalizeSchedule("hourly", id) {
		t.Fatal("every 1 hour should match hourly")
	}
	if FuzzyNormalizeSchedule("every 1 hours", id) != FuzzyNormalizeSchedule("hourly", id) {
		t.Fatal("every 1 hours should match hourly")
	}
	if FuzzyNormalizeSchedule("every 24 hours", id) != FuzzyNormalizeSchedule("daily", id) {
		t.Fatal("every 24 hours should match daily")
	}
	if FuzzyNormalizeSchedule("EVERY  3  HOURS", id) != e3a {
		t.Fatalf("case and whitespace: got %q want %q", FuzzyNormalizeSchedule("EVERY  3  HOURS", id), e3a)
	}
	if FuzzyNormalizeSchedule("every 0 hours", id) != "every 0 hours" {
		t.Fatal("every 0 hours should pass through")
	}
	if FuzzyNormalizeSchedule("every 25 hours", id) != "every 25 hours" {
		t.Fatal("every 25 hours should pass through")
	}
}

func TestIsCronExpression(t *testing.T) {
	t.Parallel()
	if !IsCronExpression("0 0 * * *") {
		t.Fatal()
	}
	if IsCronExpression("daily") {
		t.Fatal()
	}
	if !strings.Contains(FuzzyNormalizeSchedule("daily", "x"), " ") {
		t.Fatal("daily should produce 5-field cron")
	}
}
