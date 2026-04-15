package config

import "testing"

func TestSplitFrontmatter(t *testing.T) {
	const doc = `---
foo: bar
---

# Hello
`
	y, body, err := SplitFrontmatter(doc)
	if err != nil {
		t.Fatal(err)
	}
	if y != "foo: bar" {
		t.Fatalf("yaml: %q", y)
	}
	if body != "# Hello" {
		t.Fatalf("body: %q", body)
	}
}
