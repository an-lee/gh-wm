package ghclient

import "testing"

func TestSplitRepo_Errors(t *testing.T) {
	t.Parallel()
	if err := RemoveIssueLabel("nope", 1, "l"); err == nil {
		t.Fatal("invalid repo")
	}
	if _, err := ListIssueCommentBodies("bad", 1); err == nil {
		t.Fatal("invalid repo")
	}
	if err := PostIssueComment("bad", 1, "x"); err == nil {
		t.Fatal("invalid repo")
	}
}
