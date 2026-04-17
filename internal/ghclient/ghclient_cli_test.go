package ghclient

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// withFakeGHForCmd runs fn with a fake gh binary on PATH.
func withFakeGHForCmd(t *testing.T, cmdResponse string, exitCode int, fn func()) {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("fake gh script is unix-only")
	}
	dir := t.TempDir()
	gh := filepath.Join(dir, "gh")
	script := "#!/bin/sh\n"
	if cmdResponse != "" {
		script += "echo '" + cmdResponse + "'\n"
	}
	if exitCode != 0 {
		script += "exit 1\n"
	}
	if err := os.WriteFile(gh, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	fn()
}

func TestAddIssueLabel_Success(t *testing.T) {
	withFakeGHForCmd(t, "", 0, func() {
		err := AddIssueLabel("owner/repo", 123, "bug")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAddIssueLabel_GHError(t *testing.T) {
	withFakeGHForCmd(t, "HTTP 422", 1, func() {
		err := AddIssueLabel("owner/repo", 123, "bug")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestRemoveIssueLabel_Success(t *testing.T) {
	withFakeGHForCmd(t, "", 0, func() {
		err := RemoveIssueLabel("owner/repo", 123, "bug")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestRemoveIssueLabel_GHError(t *testing.T) {
	withFakeGHForCmd(t, "HTTP 404", 1, func() {
		err := RemoveIssueLabel("owner/repo", 123, "bug")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestListIssueCommentBodies_Success(t *testing.T) {
	withFakeGHForCmd(t, `[{"body":"hello"},{"body":"world"}]`, 0, func() {
		bodies, err := ListIssueCommentBodies("owner/repo", 123)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(bodies) != 2 || bodies[0] != "hello" || bodies[1] != "world" {
			t.Fatalf("got %v", bodies)
		}
	})
}

func TestListIssueCommentBodies_GHError(t *testing.T) {
	withFakeGHForCmd(t, "Server error", 1, func() {
		_, err := ListIssueCommentBodies("owner/repo", 123)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestListIssueCommentBodies_InvalidJSON(t *testing.T) {
	withFakeGHForCmd(t, `not json`, 0, func() {
		_, err := ListIssueCommentBodies("owner/repo", 123)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestCreateIssue_Success(t *testing.T) {
	withFakeGHForCmd(t, "", 0, func() {
		err := CreateIssue(nil, "owner/repo", "title", "body", []string{"bug"}, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateIssue_WithAssignees(t *testing.T) {
	withFakeGHForCmd(t, "", 0, func() {
		err := CreateIssue(nil, "owner/repo", "title", "body", []string{"bug"}, []string{"alice", "bob"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestCreateIssue_GHError(t *testing.T) {
	withFakeGHForCmd(t, "rate limit", 1, func() {
		err := CreateIssue(nil, "owner/repo", "title", "body", nil, nil)
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestPostIssueComment_Success(t *testing.T) {
	withFakeGHForCmd(t, "", 0, func() {
		err := PostIssueComment("owner/repo", 123, "hello world")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestPostIssueComment_GHError(t *testing.T) {
	withFakeGHForCmd(t, "unauthorized", 1, func() {
		err := PostIssueComment("owner/repo", 123, "hello")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAddIssueReaction_Success(t *testing.T) {
	withFakeGHForCmd(t, "", 0, func() {
		err := AddIssueReaction("owner/repo", 123, "rocket")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAddIssueReaction_AlreadyExists(t *testing.T) {
	// already_exists response is treated as success (no-op per API design).
	withFakeGHForCmd(t, `{"message":"Resource already exists"}`, 1, func() {
		err := AddIssueReaction("owner/repo", 123, "rocket")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAddIssueReaction_GHError(t *testing.T) {
	withFakeGHForCmd(t, "server error", 1, func() {
		err := AddIssueReaction("owner/repo", 123, "rocket")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAddIssueCommentReaction_Success(t *testing.T) {
	withFakeGHForCmd(t, "", 0, func() {
		err := AddIssueCommentReaction("owner/repo", 456, "+1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAddIssueCommentReaction_AlreadyExists(t *testing.T) {
	withFakeGHForCmd(t, `"code":"already_exists"`, 1, func() {
		err := AddIssueCommentReaction("owner/repo", 456, "+1")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestAddIssueCommentReaction_GHError(t *testing.T) {
	withFakeGHForCmd(t, "internal error", 1, func() {
		err := AddIssueCommentReaction("owner/repo", 456, "+1")
		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestReactionAlreadyExists(t *testing.T) {
	t.Parallel()
	tests := []struct {
		msg      string
		expected bool
	}{
		{"already_exists", true},
		{`"code":"already_exists"`, true},
		{`{"message":"Resource already exists"}`, true},
		{"already_exists: true", true},
		{"other error", false},
		{"", false},
	}
	for _, tc := range tests {
		got := reactionAlreadyExists(tc.msg)
		if got != tc.expected {
			t.Errorf("reactionAlreadyExists(%q) = %v, want %v", tc.msg, got, tc.expected)
		}
	}
}

func TestSplitRepo(t *testing.T) {
	t.Parallel()
	tests := []struct {
		repo     string
		wantOwn  string
		wantName string
		wantErr  bool
	}{
		{"owner/repo", "owner", "repo", false},
		{"my-org/my-project", "my-org", "my-project", false},
		{"justone", "", "", true},
		{"a/b/c", "", "", true},
		{"", "", "", true},
	}
	for _, tc := range tests {
		owner, name, err := splitRepo(tc.repo)
		if tc.wantErr {
			if err == nil {
				t.Errorf("splitRepo(%q) = (_, %q, nil), want error", tc.repo, name)
			}
		} else {
			if err != nil {
				t.Errorf("splitRepo(%q) error: %v", tc.repo, err)
				continue
			}
			if owner != tc.wantOwn || name != tc.wantName {
				t.Errorf("splitRepo(%q) = (%q, %q), want (%q, %q)", tc.repo, owner, name, tc.wantOwn, tc.wantName)
			}
		}
	}
}
