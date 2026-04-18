package output

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/config/scalar"
	"github.com/an-lee/gh-wm/internal/types"
	"github.com/gofrs/flock"
)

// ValidateAndAppend validates one safe-output item against task policy (including prior NDJSON lines
// in outPath), mutates item fields where policy applies (title-prefix, merged labels), then appends
// one JSON line. Used by `gh-wm emit` and future MCP tooling.
func ValidateAndAppend(ctx context.Context, glob *config.GlobalConfig, task *config.Task, tc *types.TaskContext, kind OutputKind, item map[string]any, outPath string) error {
	if task == nil {
		return fmt.Errorf("emit: task is nil")
	}
	if item == nil {
		return fmt.Errorf("emit: item is nil")
	}
	if strings.TrimSpace(string(kind)) == "" {
		return fmt.Errorf("emit: empty kind")
	}
	outPath = strings.TrimSpace(outPath)
	if outPath == "" {
		return fmt.Errorf("emit: output path is empty (set WM_SAFE_OUTPUT_FILE)")
	}

	item["type"] = string(kind)

	lock := flock.New(outPath)
	if err := lock.Lock(); err != nil {
		return fmt.Errorf("emit: lock %s: %w", outPath, err)
	}
	defer func() { _ = lock.Unlock() }()

	existing, err := readNDJSONLinesUnlocked(outPath)
	if err != nil {
		return err
	}

	p := newPolicy(task)
	p.SeedCountsFromExistingOutputs(existing)

	if !p.Allowed(kind) {
		return fmt.Errorf("emit: kind %q not permitted by safe-outputs:", kind)
	}
	if err := p.CheckMax(kind); err != nil {
		return fmt.Errorf("emit: %w", err)
	}

	_ = glob
	applyPolicyMutations(task, p, kind, item)

	if err := validateEmitPayload(kind, task, tc, item); err != nil {
		return err
	}

	line, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("emit: marshal item: %w", err)
	}

	f, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("emit: open %s: %w", outPath, err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(append(line, '\n')); err != nil {
		return fmt.Errorf("emit: write: %w", err)
	}
	_ = ctx // reserved for future cancellation / timeouts
	return nil
}

// SeedCountsFromExistingOutputs replays prior successful emits so max: limits apply across append calls.
func (p *Policy) SeedCountsFromExistingOutputs(items []map[string]any) {
	if p == nil {
		return
	}
	for _, raw := range items {
		if raw == nil {
			continue
		}
		kind := ParseOutputKind(ItemType(raw))
		if kind == "" || kind == KindNoop {
			continue
		}
		if kind == KindMissingTool || kind == KindMissingData {
			continue
		}
		p.RecordSuccess(kind)
	}
}

func readNDJSONLinesUnlocked(path string) ([]map[string]any, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("emit: read %s: %w", path, err)
	}
	if len(strings.TrimSpace(string(b))) == 0 {
		return nil, nil
	}
	var out []map[string]any
	sc := bufio.NewScanner(strings.NewReader(string(b)))
	// Lines can be long (comment bodies); allow large tokens.
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNum := 0
	for sc.Scan() {
		lineNum++
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			slog.Info("wm: emit: skip malformed jsonl line", "path", path, "line", lineNum, "err", err)
			continue
		}
		out = append(out, m)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("emit: scan %s: %w", path, err)
	}
	return out, nil
}

func applyPolicyMutations(task *config.Task, p *Policy, kind OutputKind, item map[string]any) {
	switch kind {
	case KindCreateIssue:
		t := strings.TrimSpace(scalar.StringField(item, "title"))
		t = p.ApplyTitlePrefix(KindCreateIssue, t)
		item["title"] = t
		item["labels"] = toStringSliceAny(p.MergeLabels(KindCreateIssue, scalar.StringSliceField(item, "labels")))
	case KindCreatePullRequest:
		t := strings.TrimSpace(scalar.StringField(item, "title"))
		if t == "" {
			t = fmt.Sprintf("[%s] wm task", task.Name)
		}
		t = p.ApplyTitlePrefix(KindCreatePullRequest, t)
		item["title"] = t
		item["labels"] = toStringSliceAny(p.MergeLabels(KindCreatePullRequest, scalar.StringSliceField(item, "labels")))
	case KindUpdateIssue:
		t := strings.TrimSpace(scalar.StringField(item, "title"))
		if t != "" {
			t = p.ApplyTitlePrefix(KindUpdateIssue, t)
			item["title"] = t
		}
	case KindUpdatePullRequest:
		t := strings.TrimSpace(scalar.StringField(item, "title"))
		if t != "" {
			t = p.ApplyTitlePrefix(KindUpdatePullRequest, t)
			item["title"] = t
		}
	default:
		// no in-map mutations
	}
}

func toStringSliceAny(in []string) []any {
	if len(in) == 0 {
		return nil
	}
	out := make([]any, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

func validateEmitPayload(kind OutputKind, task *config.Task, tc *types.TaskContext, item map[string]any) error {
	switch kind {
	case KindNoop:
		return nil
	case KindMissingTool, KindMissingData:
		return nil
	case KindAddComment:
		body := strings.TrimSpace(scalar.StringField(item, "body"))
		if body == "" {
			return fmt.Errorf("emit: add_comment: empty body")
		}
		target := scalar.IntField(item, "target")
		n := resolveCommentTarget(tc, target)
		if n <= 0 {
			return fmt.Errorf("emit: add_comment: no issue or PR number (set --target or run with issue/PR context)")
		}
		return nil
	case KindAddLabels:
		labels := scalar.StringSliceField(item, "labels")
		if len(labels) == 0 {
			return fmt.Errorf("emit: add_labels: empty labels")
		}
		p := newPolicy(task)
		for _, label := range labels {
			if label == "" {
				continue
			}
			if !p.LabelAllowed(KindAddLabels, label) {
				return fmt.Errorf("emit: add_labels: label %q not allowed by policy", label)
			}
		}
		target := scalar.IntField(item, "target")
		n := resolveLabelTarget(tc, target)
		if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: add_labels: no issue/PR number or repository")
		}
		return nil
	case KindRemoveLabels:
		labels := scalar.StringSliceField(item, "labels")
		if len(labels) == 0 {
			return fmt.Errorf("emit: remove_labels: empty labels")
		}
		p := newPolicy(task)
		for _, label := range labels {
			if label == "" {
				continue
			}
			if !p.RemoveLabelAllowed(label) {
				return fmt.Errorf("emit: remove_labels: label %q not allowed by policy", label)
			}
		}
		target := scalar.IntField(item, "target")
		n := resolveLabelTarget(tc, target)
		if n <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: remove_labels: no issue/PR number or repository")
		}
		return nil
	case KindCreateIssue:
		t := strings.TrimSpace(scalar.StringField(item, "title"))
		if t == "" {
			return fmt.Errorf("emit: create_issue: empty title")
		}
		if tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: create_issue: GITHUB_REPOSITORY not set")
		}
		return nil
	case KindCreatePullRequest:
		if tc == nil || strings.TrimSpace(tc.RepoPath) == "" {
			return fmt.Errorf("emit: create_pull_request: WM_REPO_ROOT / repo path not set")
		}
		return nil
	case KindUpdateIssue:
		title := strings.TrimSpace(scalar.StringField(item, "title"))
		body := strings.TrimSpace(scalar.StringField(item, "body"))
		if title == "" && body == "" {
			return fmt.Errorf("emit: update_issue: need non-empty title and/or body")
		}
		target := scalar.IntField(item, "target")
		if resolveIssueTarget(tc, target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: update_issue: no issue number or GITHUB_REPOSITORY")
		}
		return nil
	case KindUpdatePullRequest:
		title := strings.TrimSpace(scalar.StringField(item, "title"))
		body := strings.TrimSpace(scalar.StringField(item, "body"))
		if title == "" && body == "" {
			return fmt.Errorf("emit: update_pull_request: need non-empty title and/or body")
		}
		target := scalar.IntField(item, "target")
		if resolvePRTarget(tc, target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: update_pull_request: no PR/issue number or GITHUB_REPOSITORY")
		}
		return nil
	case KindCloseIssue:
		target := scalar.IntField(item, "target")
		if resolveIssueTarget(tc, target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: close_issue: no issue number or GITHUB_REPOSITORY")
		}
		if sr := strings.TrimSpace(scalar.StringField(item, "state_reason")); sr != "" && !validIssueCloseReason(sr) {
			return fmt.Errorf("emit: close_issue: state_reason must be completed, not_planned, or duplicate")
		}
		return nil
	case KindClosePullRequest:
		target := scalar.IntField(item, "target")
		if resolvePRTarget(tc, target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: close_pull_request: no PR number or GITHUB_REPOSITORY")
		}
		return nil
	case KindAddReviewer:
		reviewers := scalar.StringSliceField(item, "reviewers")
		if len(reviewers) == 0 {
			return fmt.Errorf("emit: add_reviewer: empty reviewers")
		}
		target := scalar.IntField(item, "target")
		if resolvePRTarget(tc, target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: add_reviewer: no PR number or GITHUB_REPOSITORY")
		}
		return nil
	case KindCreatePullRequestReviewComment:
		if strings.TrimSpace(scalar.StringField(item, "body")) == "" {
			return fmt.Errorf("emit: create_pull_request_review_comment: empty body")
		}
		if strings.TrimSpace(scalar.StringField(item, "commit_id")) == "" {
			return fmt.Errorf("emit: create_pull_request_review_comment: empty commit_id")
		}
		if strings.TrimSpace(scalar.StringField(item, "path")) == "" {
			return fmt.Errorf("emit: create_pull_request_review_comment: empty path")
		}
		if scalar.IntField(item, "line") <= 0 {
			return fmt.Errorf("emit: create_pull_request_review_comment: line must be positive")
		}
		if _, err := normalizeReviewCommentSide(scalar.StringField(item, "side")); err != nil {
			return fmt.Errorf("emit: create_pull_request_review_comment: %w", err)
		}
		start := scalar.IntField(item, "start_line")
		line := scalar.IntField(item, "line")
		if start > 0 && start > line {
			return fmt.Errorf("emit: create_pull_request_review_comment: start_line must be <= line")
		}
		target := scalar.IntField(item, "target")
		if resolvePRTarget(tc, target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: create_pull_request_review_comment: no PR number or GITHUB_REPOSITORY")
		}
		return nil
	case KindReplyToPullRequestReviewComment:
		if strings.TrimSpace(scalar.StringField(item, "body")) == "" {
			return fmt.Errorf("emit: reply_to_pull_request_review_comment: empty body")
		}
		if scalar.IntField(item, "comment_id") <= 0 {
			return fmt.Errorf("emit: reply_to_pull_request_review_comment: invalid comment_id")
		}
		target := scalar.IntField(item, "target")
		if resolvePRTarget(tc, target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: reply_to_pull_request_review_comment: no PR number or GITHUB_REPOSITORY")
		}
		return nil
	case KindResolvePullRequestReviewThread:
		if strings.TrimSpace(scalar.StringField(item, "thread_id")) == "" {
			return fmt.Errorf("emit: resolve_pull_request_review_thread: empty thread_id")
		}
		target := scalar.IntField(item, "target")
		if resolvePRTarget(tc, target) <= 0 || tc == nil || strings.TrimSpace(tc.Repo) == "" {
			return fmt.Errorf("emit: resolve_pull_request_review_thread: no PR number or GITHUB_REPOSITORY")
		}
		return nil
	default:
		return fmt.Errorf("emit: unknown kind %q", kind)
	}
}

func validIssueCloseReason(s string) bool {
	x := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(s, "-", "_")))
	x = strings.ReplaceAll(x, " ", "_")
	switch x {
	case "completed", "not_planned", "duplicate":
		return true
	default:
		return false
	}
}
