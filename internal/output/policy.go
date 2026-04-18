package output

import (
	"fmt"
	"path"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/config/scalar"
)

// Frontmatter keys (dash form) used in safe-outputs:
const (
	fmCreatePullRequest               = "create-pull-request"
	fmAddComment                      = "add-comment"
	fmAddLabels                       = "add-labels"
	fmRemoveLabels                    = "remove-labels"
	fmCreateIssue                     = "create-issue"
	fmUpdatePullRequest               = "update-pull-request"
	fmUpdateIssue                     = "update-issue"
	fmCloseIssue                      = "close-issue"
	fmClosePullRequest                = "close-pull-request"
	fmAddReviewer                     = "add-reviewer"
	fmCreatePullRequestReviewComment  = "create-pull-request-review-comment"
	fmReplyToPullRequestReviewComment = "reply-to-pull-request-review-comment"
	fmResolvePullRequestReviewThread  = "resolve-pull-request-review-thread"
	fmPushToPullRequestBranch         = "push-to-pull-request-branch"
	fmNoop                            = "noop"
)

// kindToFrontmatter maps JSON type (underscore) to frontmatter key (dash).
func kindToFrontmatter(kind OutputKind) string {
	switch kind {
	case KindCreatePullRequest:
		return fmCreatePullRequest
	case KindAddComment:
		return fmAddComment
	case KindAddLabels:
		return fmAddLabels
	case KindRemoveLabels:
		return fmRemoveLabels
	case KindCreateIssue:
		return fmCreateIssue
	case KindUpdatePullRequest:
		return fmUpdatePullRequest
	case KindUpdateIssue:
		return fmUpdateIssue
	case KindCloseIssue:
		return fmCloseIssue
	case KindClosePullRequest:
		return fmClosePullRequest
	case KindAddReviewer:
		return fmAddReviewer
	case KindCreatePullRequestReviewComment:
		return fmCreatePullRequestReviewComment
	case KindReplyToPullRequestReviewComment:
		return fmReplyToPullRequestReviewComment
	case KindResolvePullRequestReviewThread:
		return fmResolvePullRequestReviewThread
	case KindPushToPullRequestBranch:
		return fmPushToPullRequestBranch
	case KindNoop:
		return fmNoop
	default:
		return ""
	}
}

// defaultMaxPerKind when frontmatter omits max:
func defaultMaxPerKind(kind OutputKind) int {
	switch kind {
	case KindCreatePullRequest, KindAddComment, KindCreateIssue,
		KindUpdatePullRequest, KindUpdateIssue, KindCloseIssue, KindPushToPullRequestBranch:
		return 1
	case KindClosePullRequest:
		return 10
	case KindAddReviewer:
		return 3
	case KindCreatePullRequestReviewComment, KindResolvePullRequestReviewThread:
		return 5
	case KindReplyToPullRequestReviewComment:
		return 10
	case KindAddLabels, KindRemoveLabels:
		return 3
	case KindNoop:
		return 10
	default:
		return 1
	}
}

// Policy holds per-run counters and task config lookups.
type Policy struct {
	task *config.Task
	// counts successful executions per kind this run
	counts map[OutputKind]int
}

func newPolicy(task *config.Task) *Policy {
	return &Policy{
		task:   task,
		counts: make(map[OutputKind]int),
	}
}

func (p *Policy) fmBlock(kind OutputKind) map[string]any {
	if p == nil || p.task == nil {
		return nil
	}
	so := p.task.SafeOutputsMap()
	if so == nil {
		return nil
	}
	key := kindToFrontmatter(kind)
	if key == "" {
		return nil
	}
	raw, ok := so[key]
	if !ok {
		return nil
	}
	m, ok := raw.(map[string]any)
	if !ok {
		return nil
	}
	return m
}

// Allowed reports whether the kind is declared in safe-outputs (noop is always allowed).
func (p *Policy) Allowed(kind OutputKind) bool {
	if kind == KindNoop || kind == KindMissingTool || kind == KindMissingData {
		return true
	}
	if p == nil || p.task == nil {
		return false
	}
	return p.task.HasSafeOutputKey(kindToFrontmatter(kind))
}

// CheckMax returns an error if another execution of kind would exceed policy max.
func (p *Policy) CheckMax(kind OutputKind) error {
	if kind == KindNoop || kind == KindMissingTool || kind == KindMissingData {
		return nil
	}
	block := p.fmBlock(kind)
	maxN := scalar.IntFromMap(block, "max")
	if maxN <= 0 {
		maxN = defaultMaxPerKind(kind)
	}
	if p.counts[kind] >= maxN {
		return fmt.Errorf("policy: max %d for %s exceeded", maxN, kind)
	}
	return nil
}

// RecordSuccess increments the per-kind counter after a successful GitHub operation.
func (p *Policy) RecordSuccess(kind OutputKind) {
	if p == nil || kind == KindNoop || kind == KindMissingTool || kind == KindMissingData {
		return
	}
	p.counts[kind]++
}

// ApplyTitlePrefix prepends title-prefix from policy when set.
func (p *Policy) ApplyTitlePrefix(kind OutputKind, title string) string {
	block := p.fmBlock(kind)
	prefix := scalar.StringFromMap(block, "title-prefix")
	if prefix == "" {
		return title
	}
	t := strings.TrimSpace(title)
	if t == "" {
		return prefix
	}
	if strings.HasPrefix(t, prefix) {
		return t
	}
	return prefix + t
}

// MergeLabels merges policy default labels with agent labels (dedupe, preserve order).
func (p *Policy) MergeLabels(kind OutputKind, agent []string) []string {
	block := p.fmBlock(kind)
	def := scalar.StringSliceFromMap(block, "labels")
	seen := make(map[string]struct{})
	var out []string
	for _, x := range def {
		if x == "" {
			continue
		}
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	for _, x := range agent {
		if x == "" {
			continue
		}
		if _, ok := seen[x]; ok {
			continue
		}
		seen[x] = struct{}{}
		out = append(out, x)
	}
	return out
}

// DefaultDraft returns draft from policy + global.
func (p *Policy) DefaultDraft(glob *config.GlobalConfig, kind OutputKind) bool {
	d := false
	if glob != nil {
		d = glob.PR.Draft
	}
	block := p.fmBlock(kind)
	if block != nil {
		if b, ok := block["draft"].(bool); ok {
			d = b
		}
	}
	return d
}

// ResolveDraft uses agent value when set, else policy default.
func (p *Policy) ResolveDraft(glob *config.GlobalConfig, kind OutputKind, agent *bool) bool {
	def := p.DefaultDraft(glob, kind)
	if agent != nil {
		return *agent
	}
	return def
}

// LabelAllowed checks allowed/blocked lists for add_labels / remove_labels.
func (p *Policy) LabelAllowed(kind OutputKind, label string) bool {
	block := p.fmBlock(kind)
	if block == nil {
		return false
	}
	for _, pat := range scalar.StringSliceFromMap(block, "blocked") {
		if pat == "" {
			continue
		}
		if match, _ := path.Match(pat, label); match {
			return false
		}
	}
	allowed := scalar.StringSliceFromMap(block, "allowed")
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if a == label {
			return true
		}
	}
	return false
}

// RemoveLabelAllowedForRemove: when allowed is empty, any label not blocked can be removed; else must be in allowed.
func (p *Policy) RemoveLabelAllowed(label string) bool {
	block := p.fmBlock(KindRemoveLabels)
	if block == nil {
		return false
	}
	for _, pat := range scalar.StringSliceFromMap(block, "blocked") {
		if pat == "" {
			continue
		}
		if match, _ := path.Match(pat, label); match {
			return false
		}
	}
	allowed := scalar.StringSliceFromMap(block, "allowed")
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if a == label {
			return true
		}
	}
	return false
}

// PushPRTitleMatchesPolicy returns an error if title-prefix is set in safe-outputs and the PR title does not start with it.
func (p *Policy) PushPRTitleMatchesPolicy(title string) error {
	if p == nil {
		return nil
	}
	block := p.fmBlock(KindPushToPullRequestBranch)
	if block == nil {
		return nil
	}
	prefix := strings.TrimSpace(scalar.StringFromMap(block, "title-prefix"))
	if prefix == "" {
		return nil
	}
	t := strings.TrimSpace(title)
	if strings.HasPrefix(t, prefix) {
		return nil
	}
	return fmt.Errorf("push_to_pull_request_branch: PR title %q does not start with required prefix %q", t, prefix)
}

// PushPRHasRequiredLabels returns an error if safe-outputs lists labels and the PR is missing any of them.
func (p *Policy) PushPRHasRequiredLabels(prLabelNames []string) error {
	if p == nil {
		return nil
	}
	block := p.fmBlock(KindPushToPullRequestBranch)
	if block == nil {
		return nil
	}
	required := scalar.StringSliceFromMap(block, "labels")
	if len(required) == 0 {
		return nil
	}
	have := make(map[string]struct{})
	for _, l := range prLabelNames {
		if l != "" {
			have[l] = struct{}{}
		}
	}
	for _, r := range required {
		if _, ok := have[r]; !ok {
			return fmt.Errorf("push_to_pull_request_branch: PR missing required label %q", r)
		}
	}
	return nil
}
