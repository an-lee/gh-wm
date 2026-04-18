package gen

import (
	"fmt"
	"sort"
	"strings"

	"github.com/an-lee/gh-wm/internal/config"
)

// WorkflowTriggers is the union of GitHub Actions workflow `on:` keys derived from all tasks.
type WorkflowTriggers struct {
	IssuesWildcard       bool
	IssuesTypes          []string
	IssueCommentWildcard bool
	IssueCommentTypes    []string
	PullRequestWildcard  bool
	PullRequestTypes     []string
	// PullRequestReviewComment triggers inline review comment webhooks (e.g. /command on a diff thread).
	PullRequestReviewCommentWildcard bool
	PullRequestReviewCommentTypes    []string
	Schedules                        []string
	// WorkflowDispatch is always true when rendering (manual runs); set from tasks for API completeness.
	WorkflowDispatch bool
}

// CollectTriggersFromTasksDir reads all tasks under tasksDir and unions triggers for wm-agent.yml.
// Unknown on: keys (e.g. reaction) are ignored. If no task defines schedule, Schedules defaults to ["0 22 * * 1-5"].
// workflow_dispatch inputs are always emitted by renderOnBlock; WorkflowDispatch is set true if any task lists workflow_dispatch.
func CollectTriggersFromTasksDir(tasksDir string) (WorkflowTriggers, error) {
	var wt WorkflowTriggers
	tasks, err := config.LoadTasksDir(tasksDir)
	if err != nil {
		return wt, err
	}

	issuesSet := make(map[string]struct{})
	icSet := make(map[string]struct{})
	prSet := make(map[string]struct{})
	prcSet := make(map[string]struct{})
	var schedules []string

	for _, t := range tasks {
		on := t.OnMap()
		if on == nil {
			continue
		}

		if _, ok := on["workflow_dispatch"]; ok {
			wt.WorkflowDispatch = true
		}

		if sc, ok := on["slash_command"].(map[string]any); ok && sc != nil {
			events := stringSliceFromYAML(sc["events"])
			if len(events) == 0 {
				icSet["created"] = struct{}{}
			} else {
				for _, e := range events {
					e = strings.TrimSpace(strings.ToLower(e))
					if e == "issue_comment" || e == "pull_request_comment" {
						icSet["created"] = struct{}{}
					}
					if e == "pull_request_review_comment" {
						prcSet["created"] = struct{}{}
					}
				}
			}
		}

		if m, ok := on["issues"].(map[string]any); ok && m != nil {
			typesVal, hasTypes := m["types"]
			if !hasTypes {
				wt.IssuesWildcard = true
			} else {
				sl := stringSliceFromYAML(typesVal)
				if len(sl) == 0 {
					wt.IssuesWildcard = true
				} else {
					for _, s := range sl {
						issuesSet[s] = struct{}{}
					}
				}
			}
		}

		if m, ok := on["issue_comment"].(map[string]any); ok && m != nil {
			typesVal, hasTypes := m["types"]
			if !hasTypes {
				wt.IssueCommentWildcard = true
			} else {
				sl := stringSliceFromYAML(typesVal)
				if len(sl) == 0 {
					wt.IssueCommentWildcard = true
				} else {
					for _, s := range sl {
						icSet[s] = struct{}{}
					}
				}
			}
		}

		if m, ok := on["pull_request"].(map[string]any); ok && m != nil {
			typesVal, hasTypes := m["types"]
			if !hasTypes {
				wt.PullRequestWildcard = true
			} else {
				sl := stringSliceFromYAML(typesVal)
				if len(sl) == 0 {
					wt.PullRequestWildcard = true
				} else {
					for _, s := range sl {
						prSet[s] = struct{}{}
					}
				}
			}
		}

		if m, ok := on["pull_request_review_comment"].(map[string]any); ok && m != nil {
			typesVal, hasTypes := m["types"]
			if !hasTypes {
				wt.PullRequestReviewCommentWildcard = true
			} else {
				sl := stringSliceFromYAML(typesVal)
				if len(sl) == 0 {
					wt.PullRequestReviewCommentWildcard = true
				} else {
					for _, s := range sl {
						prcSet[s] = struct{}{}
					}
				}
			}
		}

		s := t.ScheduleString()
		if s != "" {
			schedules = append(schedules, FuzzyNormalizeSchedule(s, t.Path))
		}
	}

	wt.Schedules = dedupe(schedules)
	if len(wt.Schedules) == 0 {
		wt.Schedules = []string{"0 22 * * 1-5"}
	}

	if !wt.IssuesWildcard {
		wt.IssuesTypes = sortedKeys(issuesSet)
	}
	if !wt.IssueCommentWildcard {
		wt.IssueCommentTypes = sortedKeys(icSet)
	}
	if !wt.PullRequestWildcard {
		wt.PullRequestTypes = sortedKeys(prSet)
	}
	if !wt.PullRequestReviewCommentWildcard {
		wt.PullRequestReviewCommentTypes = sortedKeys(prcSet)
	}

	return wt, nil
}

func stringSliceFromYAML(v any) []string {
	arr, ok := v.([]any)
	if !ok || len(arr) == 0 {
		return nil
	}
	var out []string
	for _, x := range arr {
		s, ok := x.(string)
		if ok && strings.TrimSpace(s) != "" {
			out = append(out, strings.TrimSpace(s))
		}
	}
	return out
}

func sortedKeys(m map[string]struct{}) []string {
	if len(m) == 0 {
		return nil
	}
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// renderOnBlock returns the full `on:` YAML block (including the `on:` key) with two-space indentation
// for keys under `on:` (matching wm-agent.yml style).
func renderOnBlock(t WorkflowTriggers) string {
	var b strings.Builder
	b.WriteString("on:\n")

	if t.IssuesWildcard {
		b.WriteString("  issues:\n")
	} else if len(t.IssuesTypes) > 0 {
		b.WriteString("  issues:\n")
		b.WriteString(fmt.Sprintf("    types: [%s]\n", strings.Join(t.IssuesTypes, ", ")))
	}

	if t.IssueCommentWildcard {
		b.WriteString("  issue_comment:\n")
	} else if len(t.IssueCommentTypes) > 0 {
		b.WriteString("  issue_comment:\n")
		b.WriteString(fmt.Sprintf("    types: [%s]\n", strings.Join(t.IssueCommentTypes, ", ")))
	}

	if t.PullRequestWildcard {
		b.WriteString("  pull_request:\n")
	} else if len(t.PullRequestTypes) > 0 {
		b.WriteString("  pull_request:\n")
		b.WriteString(fmt.Sprintf("    types: [%s]\n", strings.Join(t.PullRequestTypes, ", ")))
	}

	if t.PullRequestReviewCommentWildcard {
		b.WriteString("  pull_request_review_comment:\n")
	} else if len(t.PullRequestReviewCommentTypes) > 0 {
		b.WriteString("  pull_request_review_comment:\n")
		b.WriteString(fmt.Sprintf("    types: [%s]\n", strings.Join(t.PullRequestReviewCommentTypes, ", ")))
	}

	b.WriteString("  schedule:\n")
	for _, c := range t.Schedules {
		b.WriteString(fmt.Sprintf("    - cron: \"%s\"\n", c))
	}

	b.WriteString(`  workflow_dispatch:
    inputs:
      issue_number:
        description: Issue or PR number (optional)
        required: false
      task_name:
        description: Run only this task (optional; skips event matching when set)
        required: false
`)

	return b.String()
}
