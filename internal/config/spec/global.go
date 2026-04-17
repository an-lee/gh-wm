package spec

// GlobalSpec is a typed view of .wm/config.yml fields gh-wm interprets (for docs/schema tooling).
type GlobalSpec struct {
	Version                      int      `json:"version"`
	Engine                       string   `json:"engine,omitempty"`
	Model                        string   `json:"model,omitempty"`
	MaxTurns                     int      `json:"max_turns,omitempty"`
	ClaudeOutputFormat           string   `json:"claude_output_format,omitempty"`
	WorkflowRunsOn               []string `json:"workflow_runs_on,omitempty"`
	WorkflowGhWMExtensionVersion string   `json:"workflow_gh_wm_extension_version,omitempty"`
	ContextFiles                 []string `json:"context_files,omitempty"`
	PRDraft                      bool     `json:"pr_draft,omitempty"`
	PRReviewers                  []string `json:"pr_reviewers,omitempty"`
}
