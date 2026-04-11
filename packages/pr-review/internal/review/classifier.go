package review

import (
	"context"
	"fmt"
	"strings"

	"pr-review/internal/copilot"
	"pr-review/internal/gh"
)

// Verdict is the classification result for a PR comment.
type Verdict string

const (
	VerdictFixable       Verdict = "FIXABLE"
	VerdictHumanRequired Verdict = "HUMAN_REQUIRED"
)

// Classification holds the result of classifying a single comment.
type Classification struct {
	Comment gh.ReviewComment
	Verdict Verdict
	Reason  string
}

// Classify asks Copilot whether the given PR comment can be addressed
// automatically (code change) or requires human judgment.
func Classify(ctx context.Context, meta *gh.PRMeta, comment gh.ReviewComment, fileContent string) (*Classification, error) {
	prompt := buildClassifyPrompt(meta, comment, fileContent)

	resp, err := copilot.Ask(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("classifying comment %d: %w", comment.ID, err)
	}

	verdict, reason := parseClassifyResponse(resp.Output)

	return &Classification{
		Comment: comment,
		Verdict: verdict,
		Reason:  reason,
	}, nil
}

func buildClassifyPrompt(meta *gh.PRMeta, comment gh.ReviewComment, fileContent string) string {
	var sb strings.Builder

	sb.WriteString("You are a code review assistant. Classify the following pull request review comment.\n\n")
	fmt.Fprintf(&sb, "PR: %s (PR #%d)\n", meta.Title, meta.Number)
	fmt.Fprintf(&sb, "File: %s (line %d)\n", comment.Path, comment.Line)
	fmt.Fprintf(&sb, "Diff context:\n%s\n\n", comment.DiffHunk)

	if fileContent != "" {
		fmt.Fprintf(&sb, "Current file content:\n```\n%s\n```\n\n", fileContent)
	}

	fmt.Fprintf(&sb, "Review comment:\n%s\n\n", comment.Body)

	sb.WriteString(`Respond with EXACTLY one of these two lines (nothing else):
FIXABLE: <one-sentence reason why this can be auto-fixed with a code change>
HUMAN_REQUIRED: <one-sentence reason why this needs human judgment>

A comment is FIXABLE if it requests a concrete code change (rename, refactor, add/remove code, fix a bug, update a comment/doc).
A comment is HUMAN_REQUIRED if it raises a design question, requests a decision, asks for clarification, discusses architecture, or requires understanding of business context.`)

	return sb.String()
}

func parseClassifyResponse(output string) (Verdict, string) {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "FIXABLE:") {
			return VerdictFixable, strings.TrimSpace(strings.TrimPrefix(line, "FIXABLE:"))
		}
		if strings.HasPrefix(line, "HUMAN_REQUIRED:") {
			return VerdictHumanRequired, strings.TrimSpace(strings.TrimPrefix(line, "HUMAN_REQUIRED:"))
		}
	}
	// Default to human required if we can't parse the response.
	return VerdictHumanRequired, fmt.Sprintf("Could not parse Copilot classification response: %s", output)
}
