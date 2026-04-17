package spec

// validGitHubReaction duplicates GitHub reactions API accepted values (keep in sync with config.ValidGitHubReaction).
func validGitHubReaction(s string) bool {
	_, ok := map[string]struct{}{
		"+1": {}, "-1": {}, "laugh": {}, "confused": {}, "heart": {},
		"hooray": {}, "rocket": {}, "eyes": {},
	}[s]
	return ok
}
