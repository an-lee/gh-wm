package config

// GlobalConfig is .wm/config.yml
type GlobalConfig struct {
	Version  int    `yaml:"version"`
	Engine   string `yaml:"engine"`
	Model    string `yaml:"model"`
	MaxTurns int    `yaml:"max_turns"`
	Context  struct {
		Files []string `yaml:"files"`
	} `yaml:"context"`
	PR struct {
		Draft     bool     `yaml:"draft"`
		Reviewers []string `yaml:"reviewers"`
	} `yaml:"pr"`
}

// WMExtension is wm: in task frontmatter
type WMExtension struct {
	StateLabels map[string]string `yaml:"state_labels"`
}

// Task holds one .wm/tasks/*.md file
type Task struct {
	Name        string         // filename without .md
	Path        string         // absolute path
	Frontmatter map[string]any // raw YAML
	Body        string         // markdown prompt
}

// OnMap returns the on: block
func (t *Task) OnMap() map[string]any {
	if t == nil || t.Frontmatter == nil {
		return nil
	}
	on, _ := t.Frontmatter["on"].(map[string]any)
	return on
}

// Engine returns engine: from frontmatter or empty
func (t *Task) Engine() string {
	if t == nil {
		return ""
	}
	if e, ok := t.Frontmatter["engine"].(string); ok {
		return e
	}
	return ""
}

// ScheduleString extracts schedule from on: block for union in wm-agent.yml
func (t *Task) ScheduleString() string {
	on := t.OnMap()
	if on == nil {
		return ""
	}
	s, ok := on["schedule"]
	if !ok {
		return ""
	}
	switch v := s.(type) {
	case string:
		return v
	default:
		return ""
	}
}
