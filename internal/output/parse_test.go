package output

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseAgentOutputFile_Missing(t *testing.T) {
	t.Parallel()
	ao, err := ParseAgentOutputFile(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatal(err)
	}
	if ao != nil {
		t.Fatal("expected nil")
	}
}

func TestParseAgentOutputFile_Valid(t *testing.T) {
	t.Parallel()
	p := filepath.Join(t.TempDir(), "output.json")
	if err := os.WriteFile(p, []byte(`{"items":[{"type":"noop","message":"ok"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	ao, err := ParseAgentOutputFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if ao == nil || len(ao.Items) != 1 {
		t.Fatalf("got %#v", ao)
	}
	if ParseOutputKind(ItemType(ao.Items[0])) != KindNoop {
		t.Fatal()
	}
}

func TestParseOutputKind_Dashes(t *testing.T) {
	t.Parallel()
	if ParseOutputKind("create-pull-request") != KindCreatePullRequest {
		t.Fatal()
	}
	if ParseOutputKind("add_comment") != KindAddComment {
		t.Fatal()
	}
}
