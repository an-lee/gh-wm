package cmd

import (
	"bytes"
	"testing"
)

func TestRootHelp(t *testing.T) {
	t.Parallel()
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if buf.Len() < 50 {
		t.Fatal("expected help output")
	}
}
