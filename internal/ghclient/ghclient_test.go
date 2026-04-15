package ghclient

import "testing"

func TestAddIssueLabel_InvalidRepo(t *testing.T) {
	t.Parallel()
	if err := AddIssueLabel("bad", 1, "x"); err == nil {
		t.Fatal("expected error")
	}
}

func TestCurrentRepo(t *testing.T) {
	t.Parallel()
	_, err := CurrentRepo()
	if err == nil {
		// gh may work in CI
		return
	}
	// expected when gh not authenticated
	if err.Error() == "" {
		t.Fatal("empty error")
	}
}

func TestToken(t *testing.T) {
	t.Parallel()
	_, src := Token()
	_ = src // host token source
}
