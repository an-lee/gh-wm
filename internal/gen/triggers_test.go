package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectTriggersFromTasksDir_Schedules(t *testing.T) {
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
	wt, err := CollectTriggersFromTasksDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(wt.Schedules) != 2 {
		t.Fatalf("schedules: got %v", wt.Schedules)
	}
}

func TestCollectTriggersFromTasksDir_SlashCommandIssueComment(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte(`---
on:
  slash_command:
    name: agent
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	wt, err := CollectTriggersFromTasksDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if wt.IssueCommentWildcard {
		t.Fatal("expected non-wildcard issue_comment")
	}
	if len(wt.IssueCommentTypes) != 1 || wt.IssueCommentTypes[0] != "created" {
		t.Fatalf("issue_comment types: %v", wt.IssueCommentTypes)
	}
}

func TestCollectTriggersFromTasksDir_IssuesUnion(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte(`---
on:
  issues:
    types: [opened]
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte(`---
on:
  issues:
    types: [reopened]
---

y
`), 0o644); err != nil {
		t.Fatal(err)
	}
	wt, err := CollectTriggersFromTasksDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if wt.IssuesWildcard {
		t.Fatal("unexpected wildcard")
	}
	if len(wt.IssuesTypes) != 2 {
		t.Fatalf("issues types: %v", wt.IssuesTypes)
	}
}

func TestCollectTriggersFromTasksDir_IssuesWildcard(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte(`---
on:
  issues:
    types: [opened]
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.md"), []byte(`---
on:
  issues: {}
---

y
`), 0o644); err != nil {
		t.Fatal(err)
	}
	wt, err := CollectTriggersFromTasksDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if !wt.IssuesWildcard {
		t.Fatal("expected issues wildcard when one task omits types")
	}
}

func TestCollectTriggersFromTasksDir_UnknownKeysIgnored(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte(`---
on:
  reaction: eyes
  schedule: daily
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	wt, err := CollectTriggersFromTasksDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(wt.Schedules) != 1 {
		t.Fatalf("schedules: %v", wt.Schedules)
	}
}

func TestRenderOnBlock_Minimal(t *testing.T) {
	t.Parallel()
	s := renderOnBlock(WorkflowTriggers{Schedules: []string{"0 0 * * *"}})
	if !strings.Contains(s, "schedule:") || !strings.Contains(s, `cron: "0 0 * * *"`) {
		t.Fatalf("unexpected: %s", s)
	}
	if !strings.Contains(s, "workflow_dispatch:") {
		t.Fatal("expected workflow_dispatch")
	}
	if strings.Contains(s, "issues:") || strings.Contains(s, "issue_comment:") || strings.Contains(s, "pull_request:") {
		t.Fatal("did not expect event keys")
	}
}

func TestRenderOnBlock_WildcardIssues(t *testing.T) {
	t.Parallel()
	s := renderOnBlock(WorkflowTriggers{IssuesWildcard: true, Schedules: []string{"0 0 * * *"}})
	if !strings.Contains(s, "  issues:\n") {
		t.Fatalf("expected issues wildcard block: %q", s)
	}
	if strings.Contains(s, "    types:") {
		t.Fatal("wildcard issues should not list types")
	}
}

func TestCollectTriggersFromTasksDir_SlashCommandReviewCommentOnly(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "a.md"), []byte(`---
on:
  slash_command:
    name: grumpy
    events: [pull_request_review_comment]
---

x
`), 0o644); err != nil {
		t.Fatal(err)
	}
	wt, err := CollectTriggersFromTasksDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if wt.IssueCommentWildcard || len(wt.IssueCommentTypes) > 0 {
		t.Fatalf("expected no issue_comment trigger, got wildcard=%v types=%v", wt.IssueCommentWildcard, wt.IssueCommentTypes)
	}
	if len(wt.PullRequestReviewCommentTypes) != 1 || wt.PullRequestReviewCommentTypes[0] != "created" {
		t.Fatalf("pull_request_review_comment types: %v", wt.PullRequestReviewCommentTypes)
	}
	s := renderOnBlock(wt)
	if !strings.Contains(s, "pull_request_review_comment:") {
		t.Fatalf("expected pr review comment in workflow: %s", s)
	}
}
