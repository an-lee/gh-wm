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
