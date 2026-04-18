package gh

import (
	"context"
	"fmt"

	"github.com/cli/go-gh/v2/pkg/api"
)

// ResolveReviewThread marks a pull request review thread resolved via GraphQL.
func ResolveReviewThread(ctx context.Context, threadID string) error {
	c, err := api.DefaultGraphQLClient()
	if err != nil {
		return err
	}
	q := `mutation ResolveReviewThread($id: ID!) {
  resolveReviewThread(input: {threadId: $id}) {
    thread { isResolved }
  }
}`
	var data struct {
		ResolveReviewThread *struct {
			Thread struct {
				IsResolved bool `json:"isResolved"`
			} `json:"thread"`
		} `json:"resolveReviewThread"`
	}
	vars := map[string]interface{}{"id": threadID}
	if err := c.DoWithContext(ctx, q, vars, &data); err != nil {
		return fmt.Errorf("graphql resolveReviewThread: %w", err)
	}
	if data.ResolveReviewThread == nil {
		return fmt.Errorf("graphql resolveReviewThread: empty response")
	}
	return nil
}
