package output

import "testing"

func TestNormalizeUpdateOperation(t *testing.T) {
	t.Parallel()
	if got := normalizeUpdateOperation(""); got != "replace" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeUpdateOperation("Replace"); got != "replace" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeUpdateOperation("replace-island"); got != "replace_island" {
		t.Fatalf("got %q", got)
	}
}

func TestIsValidUpdateOperation(t *testing.T) {
	t.Parallel()
	if !isValidUpdateOperation("") || !isValidUpdateOperation("append") || !isValidUpdateOperation("prepend") {
		t.Fatal("expected valid")
	}
	if isValidUpdateOperation("nope") {
		t.Fatal("expected invalid")
	}
}

func TestReplaceGhWMIsland(t *testing.T) {
	t.Parallel()
	cur := "intro\n<!-- gh-wm:island -->old\nline<!-- /gh-wm:island -->\ntrailer"
	got, err := replaceGhWMIsland(cur, "new")
	if err != nil {
		t.Fatal(err)
	}
	if want := "intro\n<!-- gh-wm:island -->\nnew\n<!-- /gh-wm:island -->\ntrailer"; got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestReplaceGhWMIsland_Errors(t *testing.T) {
	t.Parallel()
	if _, err := replaceGhWMIsland("no markers", "x"); err == nil {
		t.Fatal("expected error")
	}
}
