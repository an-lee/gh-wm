package ghclient

import (
	ghauth "github.com/cli/go-gh/v2/pkg/auth"
)

// Token returns the GitHub token for github.com (same as `gh auth token`).
func Token() (string, string) {
	return ghauth.TokenForHost("github.com")
}
