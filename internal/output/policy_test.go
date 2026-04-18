package output

import (
	"testing"

	"github.com/an-lee/gh-wm/internal/config"
	"github.com/an-lee/gh-wm/internal/config/scalar"
)

// ---------------------------------------------------------------------------
// kindToFrontmatter
// ---------------------------------------------------------------------------

func TestKindToFrontmatter_AllKinds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		kind    OutputKind
		wantKey string
	}{
		{KindCreatePullRequest, "create-pull-request"},
		{KindAddComment, "add-comment"},
		{KindAddLabels, "add-labels"},
		{KindRemoveLabels, "remove-labels"},
		{KindCreateIssue, "create-issue"},
		{KindUpdatePullRequest, "update-pull-request"},
		{KindUpdateIssue, "update-issue"},
		{KindCloseIssue, "close-issue"},
		{KindClosePullRequest, "close-pull-request"},
		{KindAddReviewer, "add-reviewer"},
		{KindNoop, "noop"},
	}
	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			t.Parallel()
			if got := kindToFrontmatter(tc.kind); got != tc.wantKey {
				t.Fatalf("kindToFrontmatter(%s): got %q, want %q", tc.kind, got, tc.wantKey)
			}
		})
	}
}

func TestKindToFrontmatter_Unknown(t *testing.T) {
	t.Parallel()
	if got := kindToFrontmatter("unknown_kind"); got != "" {
		t.Fatalf("got %q, want empty string", got)
	}
}

// ---------------------------------------------------------------------------
// defaultMaxPerKind
// ---------------------------------------------------------------------------

func TestDefaultMaxPerKind_AllKinds(t *testing.T) {
	t.Parallel()
	cases := []struct {
		kind OutputKind
		want int
	}{
		{KindCreatePullRequest, 1},
		{KindAddComment, 1},
		{KindAddLabels, 3},
		{KindRemoveLabels, 3},
		{KindCreateIssue, 1},
		{KindUpdatePullRequest, 1},
		{KindUpdateIssue, 1},
		{KindCloseIssue, 1},
		{KindClosePullRequest, 10},
		{KindAddReviewer, 3},
		{KindNoop, 10},
	}
	for _, tc := range cases {
		t.Run(string(tc.kind), func(t *testing.T) {
			t.Parallel()
			if got := defaultMaxPerKind(tc.kind); got != tc.want {
				t.Fatalf("defaultMaxPerKind(%s): got %d, want %d", tc.kind, got, tc.want)
			}
		})
	}
}

func TestDefaultMaxPerKind_Unknown(t *testing.T) {
	t.Parallel()
	if got := defaultMaxPerKind("unknown_kind"); got != 1 {
		t.Fatalf("got %d, want 1", got)
	}
}

// ---------------------------------------------------------------------------
// scalar.IntFromMap (max key)
// ---------------------------------------------------------------------------

func TestMaxIntFromMap(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		m    map[string]any
		want int
	}{
		{"nil", nil, 0},
		{"missing", map[string]any{}, 0},
		{"wrong_type_string", map[string]any{"max": "x"}, 0},
		{"wrong_type_bool", map[string]any{"max": true}, 0},
		{"float64", map[string]any{"max": float64(42)}, 42},
		{"int", map[string]any{"max": int(7)}, 7},
		{"int64", map[string]any{"max": int64(99)}, 99},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := scalar.IntFromMap(tc.m, "max"); got != tc.want {
				t.Fatalf("IntFromMap max: got %d, want %d", got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// scalar.StringFromMap
// ---------------------------------------------------------------------------

func TestStringFromMap(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		m    map[string]any
		key  string
		want string
	}{
		{"nil_map", nil, "k", ""},
		{"missing_key", map[string]any{}, "k", ""},
		{"wrong_type", map[string]any{"k": 123}, "k", ""},
		{"valid", map[string]any{"k": "  hello  "}, "k", "hello"},
		{"empty_string", map[string]any{"k": "  "}, "k", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := scalar.StringFromMap(tc.m, tc.key); got != tc.want {
				t.Fatalf("StringFromMap(%s): got %q, want %q", tc.key, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// stringSliceFromMap
// ---------------------------------------------------------------------------

func TestStringSliceFromMap(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		m    map[string]any
		key  string
		want []string
	}{
		{"nil_map", nil, "k", nil},
		{"missing_key", map[string]any{}, "k", nil},
		{"wrong_type", map[string]any{"k": "not-a-slice"}, "k", nil},
		{"wrong_type_int", map[string]any{"k": []any{1, 2}}, "k", nil},
		{"valid", map[string]any{"k": []any{"a", "b", "c"}}, "k", []string{"a", "b", "c"}},
		{"filters_empty", map[string]any{"k": []any{"a", "", "b", "  ", "c"}}, "k", []string{"a", "b", "c"}},
		{"filters_nonstrings", map[string]any{"k": []any{"a", 1, "b"}}, "k", []string{"a", "b"}},
		{"empty_slice", map[string]any{"k": []any{}}, "k", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := scalar.StringSliceFromMap(tc.m, tc.key)
			if len(got) != len(tc.want) {
				t.Fatalf("StringSliceFromMap(%s): got %v, want %v", tc.key, got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Fatalf("StringSliceFromMap(%s): got %v, want %v", tc.key, got, tc.want)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// newPolicy
// ---------------------------------------------------------------------------

func TestNewPolicy(t *testing.T) {
	t.Parallel()
	task := &config.Task{Name: "test-task"}
	p := newPolicy(task)
	if p == nil {
		t.Fatal("newPolicy returned nil")
	}
	if p.task != task {
		t.Fatalf("task: got %v, want %v", p.task, task)
	}
	if p.counts == nil {
		t.Fatal("counts map should be initialized")
	}
}

func TestNewPolicy_NilTask(t *testing.T) {
	t.Parallel()
	p := newPolicy(nil)
	if p == nil {
		t.Fatal("newPolicy(nil) returned nil")
	}
}

// ---------------------------------------------------------------------------
// Policy.Allowed
// ---------------------------------------------------------------------------

func TestAllowed_KindNoop(t *testing.T) {
	t.Parallel()
	p := newPolicy(nil)
	if !p.Allowed(KindNoop) {
		t.Fatal("KindNoop should always be allowed")
	}
}

func TestAllowed_NilPolicy(t *testing.T) {
	t.Parallel()
	var p *Policy
	if p.Allowed(KindCreatePullRequest) {
		t.Fatal("nil policy should not allow non-noop kinds")
	}
}

func TestAllowed_NilTask(t *testing.T) {
	t.Parallel()
	p := newPolicy(nil)
	if p.Allowed(KindAddComment) {
		t.Fatal("nil task should not allow kinds")
	}
}

func TestAllowed_TaskWithoutSafeOutputs(t *testing.T) {
	t.Parallel()
	task := &config.Task{Name: "no-safe-outputs", Frontmatter: map[string]any{}}
	p := newPolicy(task)
	for _, kind := range []OutputKind{
		KindCreatePullRequest, KindAddComment, KindAddLabels, KindRemoveLabels, KindCreateIssue,
		KindUpdatePullRequest, KindUpdateIssue, KindCloseIssue, KindClosePullRequest, KindAddReviewer,
	} {
		if p.Allowed(kind) {
			t.Fatalf("task without safe-outputs should not allow %s", kind)
		}
	}
}

func TestAllowed_TaskWithSafeOutputs(t *testing.T) {
	t.Parallel()
	task := &config.Task{
		Frontmatter: map[string]any{
			"safe-outputs": map[string]any{
				"add-comment":         map[string]any{},
				"create-pull-request": map[string]any{},
			},
		},
	}
	p := newPolicy(task)
	if !p.Allowed(KindAddComment) {
		t.Fatal("add-comment should be allowed")
	}
	if !p.Allowed(KindCreatePullRequest) {
		t.Fatal("create-pull-request should be allowed")
	}
	if p.Allowed(KindCreateIssue) {
		t.Fatal("create-issue should not be allowed")
	}
}

// ---------------------------------------------------------------------------
// Policy.CheckMax
// ---------------------------------------------------------------------------

func TestCheckMax_KindNoop(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	if err := p.CheckMax(KindNoop); err != nil {
		t.Fatalf("CheckMax(KindNoop) should never error: %v", err)
	}
}

func TestCheckMax_UsesDefault(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{"noop": map[string]any{}},
	}}
	p := newPolicy(task)
	// default for noop is 10; within limit
	for i := 0; i < 10; i++ {
		if err := p.CheckMax(KindNoop); err != nil {
			t.Fatalf("CheckMax(KindNoop) iteration %d: %v", i, err)
		}
	}
}

func TestCheckMax_ExceedsDefault(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{"add-comment": map[string]any{}},
	}}
	p := newPolicy(task)
	// default for add-comment is 1
	if err := p.CheckMax(KindAddComment); err != nil {
		t.Fatalf("CheckMax at count=0 should not error: %v", err)
	}
	p.RecordSuccess(KindAddComment)   // count=1
	err := p.CheckMax(KindAddComment) // count=1 >= max=1 → error
	if err == nil {
		t.Fatal("expected error when count=1 >= default max=1")
	}
}

func TestCheckMax_ExplicitMax(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-comment": map[string]any{"max": 2},
		},
	}}
	p := newPolicy(task)
	// max 2 for add-comment
	if err := p.CheckMax(KindAddComment); err != nil {
		t.Fatalf("CheckMax(KindAddComment) at 0: %v", err)
	}
	p.RecordSuccess(KindAddComment)
	if err := p.CheckMax(KindAddComment); err != nil {
		t.Fatalf("CheckMax(KindAddComment) at 1: %v", err)
	}
	p.RecordSuccess(KindAddComment)
	err := p.CheckMax(KindAddComment)
	if err == nil {
		t.Fatal("expected error when exceeding explicit max=2")
	}
}

func TestCheckMax_ZeroMaxTreatedAsDefault(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"max": 0},
		},
	}}
	p := newPolicy(task)
	// max 0 means use default (3 for add-labels)
	// RecordSuccess increments counts without checking
	p.RecordSuccess(KindAddLabels)
	if got, want := p.counts[KindAddLabels], 1; got != want {
		t.Fatalf("count: got %d, want %d", got, want)
	}
}

// ---------------------------------------------------------------------------
// Policy.RecordSuccess
// ---------------------------------------------------------------------------

func TestRecordSuccess_NilPolicy(t *testing.T) {
	t.Parallel()
	var p *Policy
	p.RecordSuccess(KindCreatePullRequest) // must not panic
}

func TestRecordSuccess_KindNoopNoOp(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	p.RecordSuccess(KindNoop)
	if got := p.counts[KindNoop]; got != 0 {
		t.Fatalf("KindNoop should not increment counts, got %d", got)
	}
}

func TestRecordSuccess_Increments(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	p.RecordSuccess(KindAddComment)
	p.RecordSuccess(KindAddComment)
	p.RecordSuccess(KindCreatePullRequest)
	if got := p.counts[KindAddComment]; got != 2 {
		t.Fatalf("KindAddComment: got %d, want 2", got)
	}
	if got := p.counts[KindCreatePullRequest]; got != 1 {
		t.Fatalf("KindCreatePullRequest: got %d, want 1", got)
	}
}

// ---------------------------------------------------------------------------
// Policy.ApplyTitlePrefix
// ---------------------------------------------------------------------------

func TestApplyTitlePrefix_NoBlock(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	got := p.ApplyTitlePrefix(KindCreatePullRequest, "My Title")
	if got != "My Title" {
		t.Fatalf("got %q, want %q", got, "My Title")
	}
}

func TestApplyTitlePrefix_EmptyPrefix(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{"create-pull-request": map[string]any{}},
	}}
	p := newPolicy(task)
	got := p.ApplyTitlePrefix(KindCreatePullRequest, "My Title")
	if got != "My Title" {
		t.Fatalf("got %q, want %q", got, "My Title")
	}
}

func TestApplyTitlePrefix_AddsPrefix(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"create-pull-request": map[string]any{"title-prefix": "[WIP] "},
		},
	}}
	p := newPolicy(task)
	got := p.ApplyTitlePrefix(KindCreatePullRequest, "My Title")
	if got != "[WIP]My Title" {
		t.Fatalf("got %q, want %q", got, "[WIP]My Title")
	}
}

func TestApplyTitlePrefix_AlreadyHasPrefix(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"create-pull-request": map[string]any{"title-prefix": "[WIP]"},
		},
	}}
	p := newPolicy(task)
	got := p.ApplyTitlePrefix(KindCreatePullRequest, "[WIP] My Title")
	if got != "[WIP] My Title" {
		t.Fatalf("got %q, want %q (should not double-prefix)", got, "[WIP] My Title")
	}
}

func TestApplyTitlePrefix_EmptyTitle(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"create-pull-request": map[string]any{"title-prefix": "[WIP]"},
		},
	}}
	p := newPolicy(task)
	got := p.ApplyTitlePrefix(KindCreatePullRequest, "")
	if got != "[WIP]" {
		t.Fatalf("got %q, want %q", got, "[WIP]")
	}
}

// ---------------------------------------------------------------------------
// Policy.MergeLabels
// ---------------------------------------------------------------------------

func TestMergeLabels_NoBlock(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	got := p.MergeLabels(KindAddLabels, []string{"bug", "help-wanted"})
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestMergeLabels_PolicyOnly(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{"policy-label"}},
		},
	}}
	p := newPolicy(task)
	got := p.MergeLabels(KindAddLabels, nil)
	if len(got) != 1 || got[0] != "policy-label" {
		t.Fatalf("got %v", got)
	}
}

func TestMergeLabels_Deduplicates(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{"shared", "policy-only"}},
		},
	}}
	p := newPolicy(task)
	got := p.MergeLabels(KindAddLabels, []string{"shared", "agent-only"})
	if len(got) != 3 {
		t.Fatalf("got %v", got)
	}
	// policy labels come first, then agent labels (deduped)
	if got[0] != "shared" || got[1] != "policy-only" || got[2] != "agent-only" {
		t.Fatalf("unexpected order: %v", got)
	}
}

func TestMergeLabels_PolicyLabelsFirst(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{"p1", "p2"}},
		},
	}}
	p := newPolicy(task)
	got := p.MergeLabels(KindAddLabels, []string{"a1", "a2"})
	if len(got) != 4 {
		t.Fatalf("got %v", got)
	}
	if got[0] != "p1" || got[1] != "p2" {
		t.Fatalf("policy labels should come first: %v", got)
	}
}

func TestMergeLabels_FiltersEmpty(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{"", "valid"}},
		},
	}}
	p := newPolicy(task)
	got := p.MergeLabels(KindAddLabels, []string{"", "agent"})
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

// ---------------------------------------------------------------------------
// Policy.DefaultDraft
// ---------------------------------------------------------------------------

func TestDefaultDraft_NilGlobalNilBlock(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	if got := p.DefaultDraft(nil, KindCreatePullRequest); got != false {
		t.Fatalf("DefaultDraft with nil global: got %v, want false", got)
	}
}

func TestDefaultDraft_GlobalTrue(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	glob := &config.GlobalConfig{}
	glob.PR.Draft = true
	if got := p.DefaultDraft(glob, KindCreatePullRequest); !got {
		t.Fatal("DefaultDraft should be true from global")
	}
}

func TestDefaultDraft_BlockOverridesGlobal(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"create-pull-request": map[string]any{"draft": false},
		},
	}}
	p := newPolicy(task)
	glob := &config.GlobalConfig{}
	glob.PR.Draft = true
	if got := p.DefaultDraft(glob, KindCreatePullRequest); got {
		t.Fatal("block draft=false should override global draft=true")
	}
}

func TestDefaultDraft_NilBlockUsesGlobal(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"create-pull-request": map[string]any{},
		},
	}}
	p := newPolicy(task)
	glob := &config.GlobalConfig{}
	glob.PR.Draft = true
	if got := p.DefaultDraft(glob, KindCreatePullRequest); !got {
		t.Fatal("nil block should use global")
	}
}

func TestDefaultDraft_WrongTypeInBlock(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"create-pull-request": map[string]any{"draft": "not-a-bool"},
		},
	}}
	p := newPolicy(task)
	// wrong type → fallback to global (false)
	if got := p.DefaultDraft(nil, KindCreatePullRequest); got {
		t.Fatal("wrong type in block should not override nil global")
	}
}

// ---------------------------------------------------------------------------
// Policy.ResolveDraft
// ---------------------------------------------------------------------------

func TestResolveDraft_AgentTrue(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	b := true
	if got := p.ResolveDraft(nil, KindCreatePullRequest, &b); !got {
		t.Fatal("agent value true should be returned")
	}
}

func TestResolveDraft_AgentFalse(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	b := false
	if got := p.ResolveDraft(nil, KindCreatePullRequest, &b); got {
		t.Fatal("agent value false should be returned")
	}
}

func TestResolveDraft_AgentNil(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"create-pull-request": map[string]any{"draft": true},
		},
	}}
	p := newPolicy(task)
	glob := &config.GlobalConfig{}
	glob.PR.Draft = false
	if got := p.ResolveDraft(glob, KindCreatePullRequest, nil); !got {
		t.Fatal("nil agent should use policy default (true from block)")
	}
}

// ---------------------------------------------------------------------------
// Policy.LabelAllowed
// ---------------------------------------------------------------------------

func TestLabelAllowed_NoBlock(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	if p.LabelAllowed(KindAddLabels, "bug") {
		t.Fatal("no block should not allow any label")
	}
}

func TestLabelAllowed_BlockedByPattern(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"blocked": []any{"bot-*"}},
		},
	}}
	p := newPolicy(task)
	if p.LabelAllowed(KindAddLabels, "bot-security") {
		t.Fatal("bot-security should be blocked by bot-*")
	}
}

func TestLabelAllowed_NotBlocked(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"blocked": []any{"bot-*"}},
		},
	}}
	p := newPolicy(task)
	if !p.LabelAllowed(KindAddLabels, "bug") {
		t.Fatal("bug should not be blocked")
	}
}

func TestLabelAllowed_AllowedListExact(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"allowed": []any{"bug", "enhancement"}},
		},
	}}
	p := newPolicy(task)
	if !p.LabelAllowed(KindAddLabels, "bug") {
		t.Fatal("bug should be allowed")
	}
	if p.LabelAllowed(KindAddLabels, "help-wanted") {
		t.Fatal("help-wanted not in allowed list")
	}
}

func TestLabelAllowed_AllowedListWithBlocked(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{
				"allowed": []any{"bug", "enhancement"},
				"blocked": []any{"*-wip"},
			},
		},
	}}
	p := newPolicy(task)
	if p.LabelAllowed(KindAddLabels, "bug-wip") {
		t.Fatal("bug-wip is in both allowed and blocked; blocked should win")
	}
}

func TestLabelAllowed_EmptyAllowedMeansAllExceptBlocked(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"blocked": []any{"wip-*"}},
		},
	}}
	p := newPolicy(task)
	if !p.LabelAllowed(KindAddLabels, "bug") {
		t.Fatal("bug should be allowed (empty allowed list means all except blocked)")
	}
}

func TestLabelAllowed_FiltersEmptyBlockedPatterns(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"blocked": []any{"", "allowed-label"}},
		},
	}}
	p := newPolicy(task)
	// empty pattern is filtered; allowed-label is not a valid label name
	if !p.LabelAllowed(KindAddLabels, "bug") {
		t.Fatal("bug should be allowed with only empty/blocked patterns")
	}
}

// ---------------------------------------------------------------------------
// Policy.RemoveLabelAllowed
// ---------------------------------------------------------------------------

func TestRemoveLabelAllowed_NoBlock(t *testing.T) {
	t.Parallel()
	task := &config.Task{}
	p := newPolicy(task)
	if p.RemoveLabelAllowed("bug") {
		t.Fatal("no block should not allow any removal")
	}
}

func TestRemoveLabelAllowed_BlockedByPattern(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"remove-labels": map[string]any{"blocked": []any{"pinned"}},
		},
	}}
	p := newPolicy(task)
	if p.RemoveLabelAllowed("pinned") {
		t.Fatal("pinned should be blocked")
	}
}

func TestRemoveLabelAllowed_AllowedListExact(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"remove-labels": map[string]any{"allowed": []any{"wip", "needs-review"}},
		},
	}}
	p := newPolicy(task)
	if !p.RemoveLabelAllowed("wip") {
		t.Fatal("wip should be allowed for removal")
	}
	if p.RemoveLabelAllowed("bug") {
		t.Fatal("bug not in allowed list for removal")
	}
}

func TestRemoveLabelAllowed_EmptyAllowedMeansAllExceptBlocked(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"remove-labels": map[string]any{"blocked": []any{"pinned"}},
		},
	}}
	p := newPolicy(task)
	if !p.RemoveLabelAllowed("bug") {
		t.Fatal("bug should be removable (empty allowed means all except blocked)")
	}
	if p.RemoveLabelAllowed("pinned") {
		t.Fatal("pinned should be blocked")
	}
}

func TestRemoveLabelAllowed_BlockedWinsOverAllowed(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"remove-labels": map[string]any{
				"allowed": []any{"pinned"},
				"blocked": []any{"pinned"},
			},
		},
	}}
	p := newPolicy(task)
	if p.RemoveLabelAllowed("pinned") {
		t.Fatal("blocked should win over allowed")
	}
}

// ---------------------------------------------------------------------------
// Policy.fmBlock (internal helper) — tested indirectly via Allowed/CheckMax
// ---------------------------------------------------------------------------

// fmBlock helper: constructs a policy and calls fmBlock directly.
func makePolicyWithSafeOutputs(so map[string]any) *Policy {
	return newPolicy(&config.Task{
		Frontmatter: map[string]any{
			"safe-outputs": so,
		},
	})
}

func TestFmBlock_SafeOutputsExistsButKeyMissing(t *testing.T) {
	t.Parallel()
	// safe-outputs has add-comment, but we ask for add-labels
	p := makePolicyWithSafeOutputs(map[string]any{
		"add-comment": map[string]any{"max": 1},
	})
	block := p.fmBlock(KindAddLabels)
	if block != nil {
		t.Fatalf("expected nil for missing key, got %v", block)
	}
}

func TestFmBlock_KeyExistsButWrongType(t *testing.T) {
	t.Parallel()
	// safe-outputs has add-comment as a string, not a map
	p := makePolicyWithSafeOutputs(map[string]any{
		"add-comment": "not-a-map",
	})
	block := p.fmBlock(KindAddComment)
	if block != nil {
		t.Fatalf("expected nil for wrong type, got %v", block)
	}
}

func TestFmBlock_KeyExistsWithValidMap(t *testing.T) {
	t.Parallel()
	p := makePolicyWithSafeOutputs(map[string]any{
		"add-comment": map[string]any{"max": 5, "title-prefix": "PR: "},
	})
	block := p.fmBlock(KindAddComment)
	if block == nil {
		t.Fatal("expected non-nil block for valid key")
	}
	if got := scalar.IntFromMap(block, "max"); got != 5 {
		t.Fatalf("max: got %d, want 5", got)
	}
	// scalar.StringFromMap trims whitespace, so "PR: " becomes "PR:"
	if got := scalar.StringFromMap(block, "title-prefix"); got != "PR:" {
		t.Fatalf("title-prefix: got %q, want %q", got, "PR:")
	}
}

func TestFmBlock_NilPolicy(t *testing.T) {
	t.Parallel()
	var p *Policy
	block := p.fmBlock(KindAddComment)
	if block != nil {
		t.Fatal("expected nil for nil policy")
	}
}

func TestFmBlock_NilTask(t *testing.T) {
	t.Parallel()
	p := newPolicy(nil)
	block := p.fmBlock(KindAddComment)
	if block != nil {
		t.Fatal("expected nil for nil task")
	}
}

func TestFmBlock_UnknownKind(t *testing.T) {
	t.Parallel()
	p := makePolicyWithSafeOutputs(map[string]any{
		"add-comment": map[string]any{},
	})
	block := p.fmBlock("unknown_kind")
	if block != nil {
		t.Fatal("expected nil for unknown kind")
	}
}

// ---------------------------------------------------------------------------
// Policy.LabelAllowed — additional edge cases
// ---------------------------------------------------------------------------

func TestLabelAllowed_NotInAllowedList(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"allowed": []any{"bug", "enhancement"}},
		},
	}}
	p := newPolicy(task)
	if p.LabelAllowed(KindAddLabels, "wontfix") {
		t.Fatal("wontfix should not be allowed (not in allowed list)")
	}
}

func TestLabelAllowed_BlockedWinsOverAllowed(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{
				"allowed": []any{"bug-wip"},
				"blocked": []any{"*-wip"},
			},
		},
	}}
	p := newPolicy(task)
	// bug-wip matches allowed but also matches blocked; blocked should win
	if p.LabelAllowed(KindAddLabels, "bug-wip") {
		t.Fatal("blocked pattern *-wip should win over allowed bug-wip")
	}
}

func TestLabelAllowed_ExplicitEmptyAllowed(t *testing.T) {
	t.Parallel()
	// Explicitly set allowed: [] (empty list) — scalar.StringSliceFromMap returns
	// nil for empty []any{}, so len(nil)==0 → "allow all non-blocked" (same as missing key).
	// To restrict to nothing, omit the allowed key rather than setting it to [].
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"allowed": []any{}},
		},
	}}
	p := newPolicy(task)
	if !p.LabelAllowed(KindAddLabels, "any-label") {
		t.Fatal("empty []any{} maps to nil slice → len==0 → allow all non-blocked")
	}
}

func TestLabelAllowed_BlockedWithGlob(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"blocked": []any{"team-*"}},
		},
	}}
	p := newPolicy(task)
	// "team-security" matches "team-*" → blocked
	if p.LabelAllowed(KindAddLabels, "team-security") {
		t.Fatal("team-security should be blocked by team-*")
	}
	// "needs-review" does NOT match "team-*" and no allowed list → allowed
	if !p.LabelAllowed(KindAddLabels, "needs-review") {
		t.Fatal("needs-review should be allowed (not blocked)")
	}
}

func TestLabelAllowed_EmptyAllowedPatternFiltered(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"allowed": []any{"", "valid-label"}},
		},
	}}
	p := newPolicy(task)
	// valid-label is in the non-empty allowed list → allowed
	if !p.LabelAllowed(KindAddLabels, "valid-label") {
		t.Fatal("valid-label should be allowed")
	}
	// "other" is not in allowed list → not allowed (empty string filtered but list not empty)
	if p.LabelAllowed(KindAddLabels, "other") {
		t.Fatal("other should not be allowed (not in allowed list)")
	}
}

// ---------------------------------------------------------------------------
// Policy.RemoveLabelAllowed — additional edge cases
// ---------------------------------------------------------------------------

func TestRemoveLabelAllowed_NotInAllowedList(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"remove-labels": map[string]any{"allowed": []any{"wip", "stale"}},
		},
	}}
	p := newPolicy(task)
	if p.RemoveLabelAllowed("bug") {
		t.Fatal("bug not in allowed list, should not be removable")
	}
}

func TestRemoveLabelAllowed_BlockedWithGlob(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"remove-labels": map[string]any{"blocked": []any{"team-*"}},
		},
	}}
	p := newPolicy(task)
	if p.RemoveLabelAllowed("team-ops") {
		t.Fatal("team-ops should be blocked by team-*")
	}
	if !p.RemoveLabelAllowed("bug") {
		t.Fatal("bug should be removable (not blocked)")
	}
}

func TestRemoveLabelAllowed_ExplicitEmptyAllowed(t *testing.T) {
	t.Parallel()
	// allowed: []any{} → StringSliceFromMap returns nil → len==0 → allow all non-blocked.
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"remove-labels": map[string]any{"allowed": []any{}},
		},
	}}
	p := newPolicy(task)
	if !p.RemoveLabelAllowed("any-label") {
		t.Fatal("empty []any{} maps to nil → len==0 → allow all non-blocked")
	}
}

func TestRemoveLabelAllowed_EmptyBlockedPatternFiltered(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"remove-labels": map[string]any{
				"blocked": []any{"", "pinned"},
			},
		},
	}}
	p := newPolicy(task)
	if !p.RemoveLabelAllowed("bug") {
		t.Fatal("bug should be removable (empty blocked pattern filtered)")
	}
}

// ---------------------------------------------------------------------------
// Policy.MergeLabels — additional edge cases
// ---------------------------------------------------------------------------

func TestMergeLabels_EmptyStringsInDef(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{"", "policy", ""}},
		},
	}}
	p := newPolicy(task)
	got := p.MergeLabels(KindAddLabels, nil)
	if len(got) != 1 || got[0] != "policy" {
		t.Fatalf("expected [policy], got %v", got)
	}
}

func TestMergeLabels_EmptyStringsInAgent(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{}},
		},
	}}
	p := newPolicy(task)
	got := p.MergeLabels(KindAddLabels, []string{"", "agent", ""})
	if len(got) != 1 || got[0] != "agent" {
		t.Fatalf("expected [agent], got %v", got)
	}
}

func TestMergeLabels_DedupAcrossDefAndAgent(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{"shared", "def-only"}},
		},
	}}
	p := newPolicy(task)
	got := p.MergeLabels(KindAddLabels, []string{"shared", "agent-only"})
	// shared appears in both; def comes first, agent's shared is skipped
	if len(got) != 3 {
		t.Fatalf("expected 3 labels, got %v", got)
	}
	if got[0] != "shared" || got[1] != "def-only" || got[2] != "agent-only" {
		t.Fatalf("unexpected order: %v", got)
	}
}

func TestMergeLabels_AgentDedupAfterDef(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{"a"}},
		},
	}}
	p := newPolicy(task)
	// agent has duplicate "b" entries; second "b" should be skipped
	got := p.MergeLabels(KindAddLabels, []string{"b", "b"})
	if len(got) != 2 {
		t.Fatalf("expected 2 labels, got %v", got)
	}
	if got[0] != "a" || got[1] != "b" {
		t.Fatalf("unexpected order: %v", got)
	}
}

func TestMergeLabels_SecondDefDedup(t *testing.T) {
	t.Parallel()
	task := &config.Task{Frontmatter: map[string]any{
		"safe-outputs": map[string]any{
			"add-labels": map[string]any{"labels": []any{"a", "a"}},
		},
	}}
	p := newPolicy(task)
	// def has duplicate "a"; second "a" should be skipped
	got := p.MergeLabels(KindAddLabels, nil)
	if len(got) != 1 || got[0] != "a" {
		t.Fatalf("expected [a], got %v", got)
	}
}
