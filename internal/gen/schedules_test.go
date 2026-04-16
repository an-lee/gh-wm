package gen

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollectSchedulesFromTasksDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte(`---
on:
  schedule: daily
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte(`---
on:
  schedule: weekly
---

y
`), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := CollectSchedulesFromTasksDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("%v", out)
	}
}

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
