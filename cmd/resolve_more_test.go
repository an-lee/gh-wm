package cmd

import (
	"bytes"
	"os"
	"testing"
)

func TestResolveCommand_MissingPayload(t *testing.T) {
	t.Setenv("GITHUB_EVENT_PATH", "")
	t.Cleanup(func() { _ = os.Unsetenv("GITHUB_EVENT_PATH") })
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"resolve", "--repo-root", ".", "--event-name", "issues"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error")
	}
}
