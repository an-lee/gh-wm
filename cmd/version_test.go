package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestVersionCommand(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, "gh-wm") {
		t.Fatalf("expected gh-wm in output: %q", out)
	}
	if !strings.Contains(out, Version) {
		t.Fatalf("expected %q in output: %q", Version, out)
	}
}

func TestRootVersionFlag(t *testing.T) {
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if !strings.Contains(out, Version) {
		t.Fatalf("expected %q in output: %q", Version, out)
	}
}
