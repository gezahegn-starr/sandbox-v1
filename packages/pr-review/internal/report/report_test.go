package report_test

import (
	"os"
	"strings"
	"testing"
	"time"

	"pr-review/internal/gh"
	"pr-review/internal/report"
	"pr-review/internal/review"
)

func TestReport_AllAutoFixed(t *testing.T) {
	meta := &gh.PRMeta{Number: 10, Title: "Clean up handlers", URL: "https://github.com/o/r/pull/10"}
	rpt := report.New(meta)
	rpt.AutoFixed = 3

	tmp, _ := os.CreateTemp(t.TempDir(), "report-*.md")
	tmp.Close()

	if err := rpt.Write(tmp.Name()); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, _ := os.ReadFile(tmp.Name())
	md := string(content)

	if !strings.Contains(md, "PR #10") {
		t.Error("missing PR number")
	}
	if !strings.Contains(md, "Auto-fixed | 3") {
		t.Error("missing auto-fixed count")
	}
	if !strings.Contains(md, "All review comments were addressed automatically") {
		t.Error("missing all-fixed message")
	}
}

func TestReport_WithHumanItems(t *testing.T) {
	meta := &gh.PRMeta{
		Number: 20,
		Title:  "Add caching layer",
		URL:    "https://github.com/o/r/pull/20",
	}
	rpt := report.New(meta)
	rpt.GeneratedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rpt.AutoFixed = 1

	comment := gh.ReviewComment{
		Path: "cache/store.go",
		Line: 88,
		Body: "Should we use Redis or Memcached here? Depends on the deployment.",
		URL:  "https://github.com/o/r/pull/20#discussion_r1",
	}
	rpt.AddHumanItem(comment, "Requires a deployment decision", nil)

	// Item with failed validation
	failedComment := gh.ReviewComment{
		Path:     "cache/evict.go",
		Line:     12,
		Body:     "Extract this into a helper.",
		DiffHunk: "@@ -10,5 +10,6 @@",
	}
	valResult := &review.ValidationResult{
		Passed: false,
		Commands: []review.CommandResult{
			{Command: "go test ./...", Output: "FAIL cache/evict_test.go:30", Passed: false},
		},
	}
	rpt.AddHumanItem(failedComment, "Auto-fix failed tests/lint", valResult)

	tmp, _ := os.CreateTemp(t.TempDir(), "report-*.md")
	tmp.Close()

	if err := rpt.Write(tmp.Name()); err != nil {
		t.Fatalf("Write failed: %v", err)
	}

	content, _ := os.ReadFile(tmp.Name())
	md := string(content)

	checks := []string{
		"PR #20",
		"cache/store.go",
		"Requires a deployment decision",
		"cache/evict.go",
		"go test ./...",
		"FAIL cache/evict_test.go",
		"View comment on GitHub",
		"Needs human review | 2",
		"Auto-fixed | 1",
	}
	for _, want := range checks {
		if !strings.Contains(md, want) {
			t.Errorf("report missing %q", want)
		}
	}
}
