package awexpr

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestScanBody_SkipsFences(t *testing.T) {
	body := "Hello `${{ wm.task_name }}`\n```\ncode ${{ secrets.FOO }}\n```\nAfter ${{ github.repository }}"
	sp := ScanBody(body)
	if len(sp) != 2 {
		t.Fatalf("want 2 spans, got %d: %#v", len(sp), sp)
	}
	if sp[0].Expr != "wm.task_name" {
		t.Errorf("first expr: %q", sp[0].Expr)
	}
	if sp[1].Expr != "github.repository" {
		t.Errorf("second expr: %q", sp[1].Expr)
	}
}

func TestValidateBody_RejectsSecrets(t *testing.T) {
	_, err := ValidateBody(`x ${{ secrets.FOO }}`)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateBody_AllowsWM(t *testing.T) {
	w, err := ValidateBody(`# T\n${{ wm.sanitized.text }}`)
	if err != nil {
		t.Fatal(err)
	}
	_ = w
}

func TestValidateBody_RejectsHandlebars(t *testing.T) {
	_, err := ValidateBody("{{#if github.event}}x{{/if}}")
	if err == nil {
		t.Fatal("expected error for {{#")
	}
}

func TestExpand_Issue(t *testing.T) {
	payload := map[string]any{
		"issue": map[string]any{
			"number": float64(42),
			"title":  "Hi",
			"body":   "There",
		},
	}
	ctx := BuildContext(&types.TaskContext{
		TaskName: "t1",
		Event:    &types.GitHubEvent{Name: "issues", Payload: payload},
	})
	out, err := Expand(`Issue #${{ github.event.issue.number }}: ${{ wm.sanitized.text }}`, ctx)
	if err != nil {
		t.Fatal(err)
	}
	want := "Issue #42: Hi\n\nThere"
	if out != want {
		t.Fatalf("got %q want %q", out, want)
	}
}

func TestExpand_OR(t *testing.T) {
	payload := map[string]any{
		"pull_request": map[string]any{"number": float64(7)},
	}
	ctx := BuildContext(&types.TaskContext{
		Event: &types.GitHubEvent{Name: "pull_request", Payload: payload},
	})
	out, err := Expand(`PR ${{ github.event.issue.number || github.event.pull_request.number }}`, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if out != "PR 7" {
		t.Fatalf("got %q", out)
	}
}

func TestValidateBody_OR(t *testing.T) {
	_, err := ValidateBody(`${{ github.event.issue.number || github.event.pull_request.number }}`)
	if err != nil {
		t.Fatal(err)
	}
}
