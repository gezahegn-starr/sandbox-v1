package review

import (
	"strings"
	"testing"

	"pr-review/internal/gh"
)

// --- classifier ---

func TestParseClassifyResponse_Fixable(t *testing.T) {
	cases := []struct {
		name   string
		output string
	}{
		{"plain", "FIXABLE: rename the variable to follow conventions"},
		{"with leading whitespace", "  FIXABLE: add nil check"},
		{"mixed case prefix is not matched", "fixable: lower"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			verdict, reason := parseClassifyResponse(tc.output)
			if strings.HasPrefix(strings.TrimSpace(tc.output), "FIXABLE:") {
				if verdict != VerdictFixable {
					t.Errorf("expected FIXABLE, got %s", verdict)
				}
				if reason == "" {
					t.Error("expected non-empty reason")
				}
			} else {
				// lowercase prefix should fall through to HUMAN_REQUIRED
				if verdict != VerdictHumanRequired {
					t.Errorf("expected HUMAN_REQUIRED for unparseable input, got %s", verdict)
				}
			}
		})
	}
}

func TestParseClassifyResponse_HumanRequired(t *testing.T) {
	output := "HUMAN_REQUIRED: this is a design decision about the API contract"
	verdict, reason := parseClassifyResponse(output)
	if verdict != VerdictHumanRequired {
		t.Errorf("expected HUMAN_REQUIRED, got %s", verdict)
	}
	if !strings.Contains(reason, "design decision") {
		t.Errorf("unexpected reason: %s", reason)
	}
}

func TestParseClassifyResponse_Unparseable(t *testing.T) {
	verdict, _ := parseClassifyResponse("something completely unexpected")
	if verdict != VerdictHumanRequired {
		t.Errorf("unparseable response should default to HUMAN_REQUIRED, got %s", verdict)
	}
}

func TestBuildClassifyPrompt_ContainsKeyFields(t *testing.T) {
	meta := &gh.PRMeta{Number: 99, Title: "Fix login bug"}
	comment := gh.ReviewComment{
		Path:     "auth/login.go",
		Line:     42,
		Body:     "This function should return an error instead of panicking.",
		DiffHunk: "@@ -10,6 +10,7 @@",
	}

	prompt := buildClassifyPrompt(meta, comment, "package auth\nfunc Login() {}")

	for _, want := range []string{"PR #99", "Fix login bug", "auth/login.go", "line 42", "panicking", "FIXABLE:", "HUMAN_REQUIRED:"} {
		if !strings.Contains(prompt, want) {
			t.Errorf("prompt missing %q", want)
		}
	}
}
