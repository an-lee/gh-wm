package ghclient

import "testing"

func TestAddIssueLabel_InvalidRepo(t *testing.T) {
	t.Parallel()
	if err := AddIssueLabel("bad", 1, "x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCurrentRepo(t *testing.T) {
	WithFakeGH(t)
	repo, err := CurrentRepo()
	if err != nil {
		t.Fatal(err)
	}
	if repo != "test-owner/test-repo" {
		t.Fatalf("got %q", repo)
	}
}

func TestToken(t *testing.T) {
	t.Parallel()
	_, src := Token()
	_ = src // host token source
}
