package review

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pr-review/internal/gh"
)

// --- fixer ---

func TestApplyFileBlocks_WritesFiles(t *testing.T) {
	dir := t.TempDir()
	orig, _ := os.Getwd()
	_ = os.Chdir(dir)
	defer os.Chdir(orig)

	output := `Here are the changes:

FILE: subpkg/hello.go
` + "```" + `
package subpkg

func Hello() string { return "hello" }
` + "```" + `
`

	changed, err := applyFileBlocks(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changed) != 1 {
		t.Fatalf("expected 1 changed file, got %d", len(changed))
	}

	content, err := os.ReadFile(filepath.Join(dir, "subpkg/hello.go"))
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if !strings.Contains(string(content), `func Hello()`) {
		t.Errorf("unexpected file content: %s", content)
	}
}

func TestApplyFileBlocks_NoBlocks(t *testing.T) {
	_, err := applyFileBlocks("NO CHANGE NEEDED: already correct")
	if err == nil {
		t.Error("expected error when no FILE: blocks found")
	}
}

func TestBuildFixPrompt_ContainsKeyFields(t *testing.T) {
	meta := &gh.PRMeta{Number: 7, Title: "Refactor handler"}
	comment := gh.ReviewComment{
		Path:     "server/handler.go",
		Line:     15,
		Body:     "Use errors.As instead of type assertion here.",
		DiffHunk: "@@ -14,3 +14,4 @@",
	}

	prompt := buildFixPrompt(meta, comment, "package server\n")

	for _, want := range []string{"PR #7", "server/handler.go", "errors.As", "FILE:", "NO CHANGE NEEDED"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("fix prompt missing %q", want)
		}
	}
}
