// Package gh wraps the GitHub REST API via github.com/cli/go-gh/v2/pkg/api (gh auth).
package gh

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/cli/go-gh/v2/pkg/api"
)

var (
	restOnce sync.Once
	restCli  *api.RESTClient
	restErr  error
)

// REST returns a process-wide REST client (lazy, uses gh credentials).
func REST() (*api.RESTClient, error) {
	restOnce.Do(func() {
		restCli, restErr = api.DefaultRESTClient()
	})
	return restCli, restErr
}

func splitRepo(repo string) (owner, name string, err error) {
	parts := strings.Split(repo, "/")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid repo: %s", repo)
	}
	return parts[0], parts[1], nil
}

// PostJSON POSTs a JSON body to a GitHub REST path (e.g. "repos/o/r/issues/1/labels").
func PostJSON(ctx context.Context, path string, payload any) error {
	c, err := REST()
	if err != nil {
		return err
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	resp, err := c.RequestWithContext(ctx, "POST", path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return nil
}

// DeletePath issues DELETE to path (caller must encode path segments).
func DeletePath(ctx context.Context, path string) error {
	c, err := REST()
	if err != nil {
		return err
	}
	resp, err := c.RequestWithContext(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	return nil
}

// GetJSON GETs and decodes JSON into dest.
func GetJSON(ctx context.Context, path string, dest any) error {
	c, err := REST()
	if err != nil {
		return err
	}
	resp, err := c.RequestWithContext(ctx, "GET", path, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}
