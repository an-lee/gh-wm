package cmd

import (
	"testing"
)

func TestIsGitHubShorthand(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"githubnext/agentics/daily-doc-updater", true},
		{"https://example.com/a", false},
		{"http://example.com/a", false},
		{"./local/task.md", false},
		{"/abs/task.md", false},
		{"owner/repo", false},
		{"", false},
		{".hidden/foo", false},
	}
	for _, tt := range tests {
		if got := isGitHubShorthand(tt.in); got != tt.want {
			t.Errorf("isGitHubShorthand(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestResolveSourceToURL(t *testing.T) {
	u, err := resolveSourceToURL("  https://example.com/x  ")
	if err != nil {
		t.Fatal(err)
	}
	if u != "https://example.com/x" {
		t.Fatalf("url passthrough: got %q", u)
	}

	u, err = resolveSourceToURL("o/r/workflows/foo.md")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://raw.githubusercontent.com/o/r/main/workflows/foo.md"
	if u != want {
		t.Fatalf("resolve shorthand: got %q want %q", u, want)
	}

	u, err = resolveSourceToURL("o/r/.wm/tasks/foo.md")
	if err != nil {
		t.Fatal(err)
	}
	want2 := "https://raw.githubusercontent.com/o/r/main/.wm/tasks/foo.md"
	if u != want2 {
		t.Fatalf("resolve wm path: got %q want %q", u, want2)
	}

	if _, err := resolveSourceToURL("bad"); err == nil {
		t.Fatal("expected error for invalid source")
	}
}

func TestNormalizeTaskFileName(t *testing.T) {
	if got := normalizeTaskFileName("x"); got != "x.md" {
		t.Fatalf("got %q", got)
	}
	if got := normalizeTaskFileName("x.MD"); got != "x.MD" {
		t.Fatalf("got %q", got)
	}
}

func TestParseGitHubShorthand(t *testing.T) {
	o, r, task := parseGitHubShorthand("a/b/c-d")
	if o != "a" || r != "b" || task != "c-d" {
		t.Fatalf("got %q %q %q", o, r, task)
	}
}
