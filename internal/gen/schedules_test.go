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

func TestNormalizeSchedule(t *testing.T) {
	t.Parallel()
	if normalizeSchedule("daily") != "0 0 * * *" {
		t.Fatal()
	}
	if normalizeSchedule("custom") != "custom" {
		t.Fatal()
	}
}
