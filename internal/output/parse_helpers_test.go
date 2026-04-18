package output

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/config/scalar"
)

// --- scalar.StringField ---

func TestStringField_NilMap(t *testing.T) {
	t.Parallel()
	if got := scalar.StringField(nil, "key"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestStringField_MissingKey(t *testing.T) {
	t.Parallel()
	if got := scalar.StringField(map[string]any{}, "key"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestStringField_WrongType(t *testing.T) {
	t.Parallel()
	if got := scalar.StringField(map[string]any{"key": 42}, "key"); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestStringField_Valid(t *testing.T) {
	t.Parallel()
	if got := scalar.StringField(map[string]any{"key": "  hello  "}, "key"); got != "hello" {
		t.Fatalf("got %q", got)
	}
}

func TestStringField_PreservesUntrimmed(t *testing.T) {
	t.Parallel()
	if got := scalar.StringField(map[string]any{"key": "  hello  "}, "key"); got == "  hello  " {
		t.Fatal("should trim")
	}
}

// --- scalar.StringSliceField ---

func TestStringSliceField_NilMap(t *testing.T) {
	t.Parallel()
	if got := scalar.StringSliceField(nil, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceField_MissingKey(t *testing.T) {
	t.Parallel()
	if got := scalar.StringSliceField(map[string]any{}, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceField_WrongType(t *testing.T) {
	t.Parallel()
	if got := scalar.StringSliceField(map[string]any{"key": "not-an-array"}, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceField_Valid(t *testing.T) {
	t.Parallel()
	got := scalar.StringSliceField(map[string]any{"key": []any{"a", "b", "  c  "}}, "key")
	want := []string{"a", "b", "c"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestStringSliceField_IgnoresEmptyStrings(t *testing.T) {
	t.Parallel()
	got := scalar.StringSliceField(map[string]any{"key": []any{"a", "", "  "}}, "key")
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("got %#v", got)
	}
}

func TestStringSliceField_IgnoresNonStrings(t *testing.T) {
	t.Parallel()
	got := scalar.StringSliceField(map[string]any{"key": []any{"a", 42, true}}, "key")
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("got %#v", got)
	}
}

// --- scalar.IntField ---

func TestIntField_NilMap(t *testing.T) {
	t.Parallel()
	if got := scalar.IntField(nil, "key"); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestIntField_MissingKey(t *testing.T) {
	t.Parallel()
	if got := scalar.IntField(map[string]any{}, "key"); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestIntField_WrongType(t *testing.T) {
	t.Parallel()
	if got := scalar.IntField(map[string]any{"key": "not-an-int"}, "key"); got != 0 {
		t.Fatalf("got %d", got)
	}
}

func TestIntField_Float64(t *testing.T) {
	t.Parallel()
	if got := scalar.IntField(map[string]any{"key": float64(42)}, "key"); got != 42 {
		t.Fatalf("got %d", got)
	}
}

func TestIntField_Int(t *testing.T) {
	t.Parallel()
	if got := scalar.IntField(map[string]any{"key": 42}, "key"); got != 42 {
		t.Fatalf("got %d", got)
	}
}

func TestIntField_Int64(t *testing.T) {
	t.Parallel()
	if got := scalar.IntField(map[string]any{"key": int64(42)}, "key"); got != 42 {
		t.Fatalf("got %d", got)
	}
}

// --- scalar.BoolPtrField ---

func TestBoolPtrField_NilMap(t *testing.T) {
	t.Parallel()
	if got := scalar.BoolPtrField(nil, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestBoolPtrField_MissingKey(t *testing.T) {
	t.Parallel()
	if got := scalar.BoolPtrField(map[string]any{}, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestBoolPtrField_WrongType(t *testing.T) {
	t.Parallel()
	if got := scalar.BoolPtrField(map[string]any{"key": "not-a-bool"}, "key"); got != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestBoolPtrField_True(t *testing.T) {
	t.Parallel()
	got := scalar.BoolPtrField(map[string]any{"key": true}, "key")
	if got == nil || !*got {
		t.Fatalf("got %#v", got)
	}
}

func TestBoolPtrField_False(t *testing.T) {
	t.Parallel()
	got := scalar.BoolPtrField(map[string]any{"key": false}, "key")
	if got == nil || *got {
		t.Fatalf("got %#v", got)
	}
}

// --- mapToCreatePR ---

func TestMapToCreatePR_Basic(t *testing.T) {
	t.Parallel()
	item := map[string]any{
		"title":  "  My PR  ",
		"body":   "  Description  ",
		"draft":  true,
		"labels": []any{"bug", "  urgent  "},
	}
	got := mapToCreatePR(item)
	if got.Title != "My PR" {
		t.Fatalf("title: got %q", got.Title)
	}
	if got.Body != "Description" {
		t.Fatalf("body: got %q", got.Body)
	}
	if got.Draft == nil || !*got.Draft {
		t.Fatalf("draft: got %#v", got.Draft)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "bug" || got.Labels[1] != "urgent" {
		t.Fatalf("labels: got %#v", got.Labels)
	}
}

func TestMapToCreatePR_Empty(t *testing.T) {
	t.Parallel()
	got := mapToCreatePR(map[string]any{})
	if got.Title != "" || got.Body != "" || got.Draft != nil || got.Labels != nil {
		t.Fatalf("got %#v", got)
	}
}

func TestMapToCreatePR_DraftFalse(t *testing.T) {
	t.Parallel()
	item := map[string]any{"draft": false}
	got := mapToCreatePR(item)
	if got.Draft == nil || *got.Draft {
		t.Fatalf("got %#v", got.Draft)
	}
}

// --- mapToAddComment ---

func TestMapToAddComment_Basic(t *testing.T) {
	t.Parallel()
	item := map[string]any{
		"body":   "  Great work!  ",
		"target": float64(42),
	}
	got := mapToAddComment(item)
	if got.Body != "Great work!" {
		t.Fatalf("body: got %q", got.Body)
	}
	if got.Target != 42 {
		t.Fatalf("target: got %d", got.Target)
	}
}

func TestMapToAddComment_Empty(t *testing.T) {
	t.Parallel()
	got := mapToAddComment(map[string]any{})
	if got.Body != "" || got.Target != 0 {
		t.Fatalf("got %#v", got)
	}
}

// --- mapToLabels ---

func TestMapToLabels_Basic(t *testing.T) {
	t.Parallel()
	item := map[string]any{
		"labels": []any{"  bug  ", "  enhancement  "},
		"target": float64(5),
	}
	got := mapToLabels(item)
	if len(got.Labels) != 2 || got.Labels[0] != "bug" || got.Labels[1] != "enhancement" {
		t.Fatalf("labels: got %#v", got.Labels)
	}
	if got.Target != 5 {
		t.Fatalf("target: got %d", got.Target)
	}
}

func TestMapToLabels_Empty(t *testing.T) {
	t.Parallel()
	got := mapToLabels(map[string]any{})
	if got.Labels != nil || got.Target != 0 {
		t.Fatalf("got %#v", got)
	}
}

// --- mapToCreateIssue ---

func TestMapToCreateIssue_Basic(t *testing.T) {
	t.Parallel()
	item := map[string]any{
		"title":     "  Bug Report  ",
		"body":      "  Something is broken  ",
		"labels":    []any{"bug", "  high-priority  "},
		"assignees": []any{"alice", "  bob  "},
	}
	got := mapToCreateIssue(item)
	if got.Title != "Bug Report" {
		t.Fatalf("title: got %q", got.Title)
	}
	if got.Body != "Something is broken" {
		t.Fatalf("body: got %q", got.Body)
	}
	if len(got.Labels) != 2 || got.Labels[0] != "bug" || got.Labels[1] != "high-priority" {
		t.Fatalf("labels: got %#v", got.Labels)
	}
	if len(got.Assignees) != 2 || got.Assignees[0] != "alice" || got.Assignees[1] != "bob" {
		t.Fatalf("assignees: got %#v", got.Assignees)
	}
}

func TestMapToCreateIssue_Empty(t *testing.T) {
	t.Parallel()
	got := mapToCreateIssue(map[string]any{})
	if got.Title != "" || got.Body != "" || got.Labels != nil || got.Assignees != nil {
		t.Fatalf("got %#v", got)
	}
}

// --- mapToNoop ---

func TestMapToNoop_Basic(t *testing.T) {
	t.Parallel()
	item := map[string]any{"message": "  nothing to do  "}
	got := mapToNoop(item)
	if got.Message != "nothing to do" {
		t.Fatalf("got %q", got.Message)
	}
}

func TestMapToNoop_Empty(t *testing.T) {
	t.Parallel()
	got := mapToNoop(map[string]any{})
	if got.Message != "" {
		t.Fatalf("got %q", got.Message)
	}
}

// --- ParseOutputKind edge cases ---

func TestParseOutputKind_AllKinds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  OutputKind
	}{
		{"create_pull_request", KindCreatePullRequest},
		{"create-pull-request", KindCreatePullRequest},
		{"add_comment", KindAddComment},
		{"add-comment", KindAddComment},
		{"add_labels", KindAddLabels},
		{"add-labels", KindAddLabels},
		{"remove_labels", KindRemoveLabels},
		{"remove-labels", KindRemoveLabels},
		{"create_issue", KindCreateIssue},
		{"create-issue", KindCreateIssue},
		{"update_pull_request", KindUpdatePullRequest},
		{"update-pull-request", KindUpdatePullRequest},
		{"update_issue", KindUpdateIssue},
		{"update-issue", KindUpdateIssue},
		{"close_issue", KindCloseIssue},
		{"close-issue", KindCloseIssue},
		{"close_pull_request", KindClosePullRequest},
		{"close-pull-request", KindClosePullRequest},
		{"add_reviewer", KindAddReviewer},
		{"add-reviewer", KindAddReviewer},
		{"create_pull_request_review_comment", KindCreatePullRequestReviewComment},
		{"create-pull-request-review-comment", KindCreatePullRequestReviewComment},
		{"reply_to_pull_request_review_comment", KindReplyToPullRequestReviewComment},
		{"reply-to-pull-request-review-comment", KindReplyToPullRequestReviewComment},
		{"resolve_pull_request_review_thread", KindResolvePullRequestReviewThread},
		{"resolve-pull-request-review-thread", KindResolvePullRequestReviewThread},
		{"noop", KindNoop},
		{"unknown_type", ""},
		{"", ""},
	}
	for _, tc := range tests {
		got := ParseOutputKind(tc.input)
		if got != tc.want {
			t.Errorf("ParseOutputKind(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

// --- ItemType edge cases ---

func TestItemType_AllForms(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input map[string]any
		want  string
	}{
		{map[string]any{"type": "noop"}, "noop"},
		{map[string]any{"type": "NOOP"}, "noop"},
		{map[string]any{"type": "  noop  "}, "noop"},
		{map[string]any{"type": "create-pr"}, "create_pr"},
		{map[string]any{"type": "CREATE-PR"}, "create_pr"},
		{map[string]any{"type": "create_pr"}, "create_pr"},
		{map[string]any{}, ""},
		{map[string]any{"type": 42}, ""},
		{nil, ""},
	}
	for _, tc := range tests {
		got := ItemType(tc.input)
		if got != tc.want {
			t.Errorf("ItemType(%#v) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
