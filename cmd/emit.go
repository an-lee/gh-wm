package cmd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/output"
	"github.com/an-lee/gh-wm/internal/types"
)

var emitCmd = &cobra.Command{
	Use:   "emit",
	Short: "Append one validated safe-output line (NDJSON) for the current WM_TASK run",
	Long: `Reads WM_REPO_ROOT, WM_TASK, WM_SAFE_OUTPUT_FILE, and optional GITHUB_REPOSITORY /
WM_ISSUE_NUMBER / WM_PR_NUMBER from the environment (set by gh wm run). Each subcommand
appends one JSON object to the per-run output.jsonl file.`,
	SilenceUsage: true,
}

func init() {
	emitCmd.AddCommand(
		emitNoopCmd,
		emitAddCommentCmd,
		emitAddLabelsCmd,
		emitRemoveLabelsCmd,
		emitCreateIssueCmd,
		emitCreatePRCmd,
		emitMissingToolCmd,
		emitMissingDataCmd,
	)
	rootCmd.AddCommand(emitCmd)

	emitNoopCmd.Flags().String("message", "", "completion message")

	emitAddCommentCmd.Flags().String("body", "", "comment body (required)")
	_ = emitAddCommentCmd.MarkFlagRequired("body")
	emitAddCommentCmd.Flags().Int("target", 0, "issue/PR number (default: WM_ISSUE_NUMBER / WM_PR_NUMBER)")

	emitAddLabelsCmd.Flags().StringSlice("labels", nil, "labels to add (repeat flag or comma-separated)")
	_ = emitAddLabelsCmd.MarkFlagRequired("labels")
	emitAddLabelsCmd.Flags().Int("target", 0, "issue/PR number")

	emitRemoveLabelsCmd.Flags().StringSlice("labels", nil, "labels to remove")
	_ = emitRemoveLabelsCmd.MarkFlagRequired("labels")
	emitRemoveLabelsCmd.Flags().Int("target", 0, "issue/PR number")

	emitCreateIssueCmd.Flags().String("title", "", "issue title")
	_ = emitCreateIssueCmd.MarkFlagRequired("title")
	emitCreateIssueCmd.Flags().String("body", "", "issue body")
	emitCreateIssueCmd.Flags().StringSlice("labels", nil, "labels")
	emitCreateIssueCmd.Flags().StringSlice("assignees", nil, "assignees")

	emitCreatePRCmd.Flags().String("title", "", "PR title")
	emitCreatePRCmd.Flags().String("body", "", "PR body")
	emitCreatePRCmd.Flags().Bool("draft", false, "open as draft")
	emitCreatePRCmd.Flags().StringSlice("labels", nil, "labels")

	emitMissingToolCmd.Flags().String("tool", "", "tool or capability name")
	emitMissingToolCmd.Flags().String("reason", "", "why it is unavailable")
	emitMissingDataCmd.Flags().String("what", "", "data that was needed")
	emitMissingDataCmd.Flags().String("reason", "", "why it is unavailable")
}

func loadEmitContext() (*config.GlobalConfig, *config.Task, *types.TaskContext, string, error) {
	repoRoot := strings.TrimSpace(os.Getenv("WM_REPO_ROOT"))
	if repoRoot == "" {
		return nil, nil, nil, "", fmt.Errorf("WM_REPO_ROOT is not set")
	}
	taskName := strings.TrimSpace(os.Getenv("WM_TASK"))
	if taskName == "" {
		return nil, nil, nil, "", fmt.Errorf("WM_TASK is not set")
	}
	outPath := strings.TrimSpace(os.Getenv("WM_SAFE_OUTPUT_FILE"))
	if outPath == "" {
		return nil, nil, nil, "", fmt.Errorf("WM_SAFE_OUTPUT_FILE is not set")
	}
	glob, tasks, err := config.Load(repoRoot)
	if err != nil {
		return nil, nil, nil, "", err
	}
	glob = config.DefaultGlobal(glob)
	var task *config.Task
	for _, t := range tasks {
		if t.Name == taskName {
			task = t
			break
		}
	}
	if task == nil {
		return nil, nil, nil, "", fmt.Errorf("task not found: %s", taskName)
	}
	tc := taskContextFromEmitEnv()
	return glob, task, tc, outPath, nil
}

func taskContextFromEmitEnv() *types.TaskContext {
	return &types.TaskContext{
		Repo:        strings.TrimSpace(os.Getenv("GITHUB_REPOSITORY")),
		RepoPath:    strings.TrimSpace(os.Getenv("WM_REPO_ROOT")),
		TaskName:    strings.TrimSpace(os.Getenv("WM_TASK")),
		IssueNumber: intFromEnv("WM_ISSUE_NUMBER"),
		PRNumber:    intFromEnv("WM_PR_NUMBER"),
	}
}

func intFromEnv(key string) int {
	s := strings.TrimSpace(os.Getenv(key))
	if s == "" {
		return 0
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func runEmit(ctx context.Context, kind output.OutputKind, item map[string]any) error {
	glob, task, tc, outPath, err := loadEmitContext()
	if err != nil {
		return err
	}
	return output.ValidateAndAppend(ctx, glob, task, tc, kind, item, outPath)
}

var emitNoopCmd = &cobra.Command{
	Use:   "noop",
	Short: "Record completion with no GitHub writes",
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, _ := cmd.Flags().GetString("message")
		return runEmit(cmd.Context(), output.KindNoop, map[string]any{"message": msg})
	},
}

var emitAddCommentCmd = &cobra.Command{
	Use:   "add-comment",
	Short: "Request adding an issue/PR comment",
	RunE: func(cmd *cobra.Command, args []string) error {
		body, _ := cmd.Flags().GetString("body")
		target, _ := cmd.Flags().GetInt("target")
		return runEmit(cmd.Context(), output.KindAddComment, map[string]any{
			"body":   body,
			"target": target,
		})
	},
}

var emitAddLabelsCmd = &cobra.Command{
	Use:   "add-labels",
	Short: "Request adding labels to an issue/PR",
	RunE: func(cmd *cobra.Command, args []string) error {
		labels, _ := cmd.Flags().GetStringSlice("labels")
		target, _ := cmd.Flags().GetInt("target")
		item := map[string]any{"target": target}
		if len(labels) > 0 {
			arr := make([]any, len(labels))
			for i, l := range labels {
				arr[i] = l
			}
			item["labels"] = arr
		}
		return runEmit(cmd.Context(), output.KindAddLabels, item)
	},
}

var emitRemoveLabelsCmd = &cobra.Command{
	Use:   "remove-labels",
	Short: "Request removing labels from an issue/PR",
	RunE: func(cmd *cobra.Command, args []string) error {
		labels, _ := cmd.Flags().GetStringSlice("labels")
		target, _ := cmd.Flags().GetInt("target")
		item := map[string]any{"target": target}
		if len(labels) > 0 {
			arr := make([]any, len(labels))
			for i, l := range labels {
				arr[i] = l
			}
			item["labels"] = arr
		}
		return runEmit(cmd.Context(), output.KindRemoveLabels, item)
	},
}

var emitCreateIssueCmd = &cobra.Command{
	Use:   "create-issue",
	Short: "Request creating an issue",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		labels, _ := cmd.Flags().GetStringSlice("labels")
		assignees, _ := cmd.Flags().GetStringSlice("assignees")
		item := map[string]any{"title": title, "body": body}
		if len(labels) > 0 {
			arr := make([]any, len(labels))
			for i, l := range labels {
				arr[i] = l
			}
			item["labels"] = arr
		}
		if len(assignees) > 0 {
			arr := make([]any, len(assignees))
			for i, a := range assignees {
				arr[i] = a
			}
			item["assignees"] = arr
		}
		return runEmit(cmd.Context(), output.KindCreateIssue, item)
	},
}

var emitCreatePRCmd = &cobra.Command{
	Use:   "create-pull-request",
	Short: "Request opening a pull request from the current branch",
	RunE: func(cmd *cobra.Command, args []string) error {
		title, _ := cmd.Flags().GetString("title")
		body, _ := cmd.Flags().GetString("body")
		draft, _ := cmd.Flags().GetBool("draft")
		labels, _ := cmd.Flags().GetStringSlice("labels")
		item := map[string]any{"title": title, "body": body}
		if cmd.Flags().Changed("draft") {
			item["draft"] = draft
		}
		if len(labels) > 0 {
			arr := make([]any, len(labels))
			for i, l := range labels {
				arr[i] = l
			}
			item["labels"] = arr
		}
		return runEmit(cmd.Context(), output.KindCreatePullRequest, item)
	},
}

var emitMissingToolCmd = &cobra.Command{
	Use:   "missing-tool",
	Short: "Report unavailable functionality",
	RunE: func(cmd *cobra.Command, args []string) error {
		tool, _ := cmd.Flags().GetString("tool")
		reason, _ := cmd.Flags().GetString("reason")
		return runEmit(cmd.Context(), output.KindMissingTool, map[string]any{"tool": tool, "reason": reason})
	},
}

var emitMissingDataCmd = &cobra.Command{
	Use:   "missing-data",
	Short: "Report unavailable information",
	RunE: func(cmd *cobra.Command, args []string) error {
		what, _ := cmd.Flags().GetString("what")
		reason, _ := cmd.Flags().GetString("reason")
		return runEmit(cmd.Context(), output.KindMissingData, map[string]any{"what": what, "reason": reason})
	},
}
