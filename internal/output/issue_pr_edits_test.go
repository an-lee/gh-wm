package output

import (
	"context"
	"strings"
	"testing"

	"github.com/an-lee/gh-wm/internal/types"
)

func TestRunUpdateIssue_RejectsInvalidOperation(t *testing.T) {
	t.Parallel()
	err := runUpdateIssue(context.Background(), &types.TaskContext{Repo: "o/r", IssueNumber: 1}, ItemUpdateIssue{
		Title:     "title",
		Operation: "bogus",
	})
	if err == nil || !strings.Contains(err.Error(), `update_issue: invalid operation "bogus"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunUpdateIssue_RejectsNonReplaceOperationWithoutBody(t *testing.T) {
	t.Parallel()
	err := runUpdateIssue(context.Background(), &types.TaskContext{Repo: "o/r", IssueNumber: 1}, ItemUpdateIssue{
		Title:     "title",
		Operation: "append",
	})
	if err == nil || !strings.Contains(err.Error(), `update_issue: operation "append" requires non-empty body`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunUpdatePullRequest_RejectsInvalidOperation(t *testing.T) {
	t.Parallel()
	err := runUpdatePullRequest(context.Background(), &types.TaskContext{Repo: "o/r", PRNumber: 1}, ItemUpdatePullRequest{
		Title:     "title",
		Operation: "bogus",
	})
	if err == nil || !strings.Contains(err.Error(), `update_pull_request: invalid operation "bogus"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunUpdatePullRequest_RejectsNonReplaceOperationWithoutBody(t *testing.T) {
	t.Parallel()
	err := runUpdatePullRequest(context.Background(), &types.TaskContext{Repo: "o/r", PRNumber: 1}, ItemUpdatePullRequest{
		Title:     "title",
		Operation: "prepend",
	})
	if err == nil || !strings.Contains(err.Error(), `update_pull_request: operation "prepend" requires non-empty body`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunCloseIssue_RejectsInvalidStateReason(t *testing.T) {
	t.Parallel()
	err := runCloseIssue(context.Background(), &types.TaskContext{Repo: "o/r", IssueNumber: 1}, ItemCloseIssue{
		StateReason: "bogus",
	})
	if err == nil || !strings.Contains(err.Error(), `close_issue: invalid state_reason "bogus"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
