package review

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"pr-review/internal/copilot"
	"pr-review/internal/gh"
)

// FixResult holds the outcome of applying a Copilot fix for a single comment.
type FixResult struct {
	Comment      gh.ReviewComment
	FilesChanged []string
	Skipped      bool   // true if Copilot indicated no change was needed
	SkipReason   string
}

// Fix asks Copilot to apply a code fix for the given comment and writes the
// changes to disk. It returns the list of files it modified.
func Fix(ctx context.Context, meta *gh.PRMeta, comment gh.ReviewComment, fileContent string) (*FixResult, error) {
	prompt := buildFixPrompt(meta, comment, fileContent)

	resp, err := copilot.Ask(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("fixing comment %d: %w", comment.ID, err)
	}

	result := &FixResult{Comment: comment}

	// Check if Copilot indicated no change was needed.
	lower := strings.ToLower(resp.Output)
	if strings.Contains(lower, "no change needed") || strings.Contains(lower, "already correct") {
		result.Skipped = true
		result.SkipReason = resp.Output
		return result, nil
	}

	// Extract file blocks from the response and write them to disk.
	changed, err := applyFileBlocks(resp.Output)
	if err != nil {
		return nil, fmt.Errorf("applying Copilot output for comment %d: %w", comment.ID, err)
	}

	result.FilesChanged = changed
	return result, nil
}

func buildFixPrompt(meta *gh.PRMeta, comment gh.ReviewComment, fileContent string) string {
	var sb strings.Builder

	sb.WriteString("You are a code assistant applying pull request review feedback. ")
	sb.WriteString("Make the minimal code change to address the review comment below.\n\n")
	fmt.Fprintf(&sb, "PR: %s (PR #%d)\n", meta.Title, meta.Number)
	fmt.Fprintf(&sb, "File: %s (line %d)\n\n", comment.Path, comment.Line)

	if fileContent != "" {
		fmt.Fprintf(&sb, "Current file content:\n```\n%s\n```\n\n", fileContent)
	}

	fmt.Fprintf(&sb, "Diff context:\n%s\n\n", comment.DiffHunk)
	fmt.Fprintf(&sb, "Review comment to address:\n%s\n\n", comment.Body)

	sb.WriteString("Respond with the COMPLETE updated file content using this exact format for EACH file you change:\n\n")
	sb.WriteString("FILE: <relative/path/to/file>\n")
	sb.WriteString("```\n")
	sb.WriteString("<complete file content>\n")
	sb.WriteString("```\n\n")
	sb.WriteString("If no code change is needed, respond with: NO CHANGE NEEDED: <reason>")

	return sb.String()
}

// fileBlockRe matches FILE: <path> followed by a fenced code block.
var fileBlockRe = regexp.MustCompile("(?m)^FILE: (.+)\n```[^\n]*\n((?s:.+?))\n```")

// applyFileBlocks parses Copilot's response and writes each file block to disk.
func applyFileBlocks(output string) ([]string, error) {
	matches := fileBlockRe.FindAllStringSubmatch(output, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no FILE: blocks found in Copilot response; raw output:\n%s", output)
	}

	var changed []string
	for _, m := range matches {
		relPath := strings.TrimSpace(m[1])
		content := m[2]

		// Ensure the parent directory exists.
		if err := os.MkdirAll(filepath.Dir(relPath), 0o755); err != nil {
			return changed, fmt.Errorf("creating dirs for %s: %w", relPath, err)
		}

		if err := os.WriteFile(relPath, []byte(content), 0o644); err != nil {
			return changed, fmt.Errorf("writing %s: %w", relPath, err)
		}
		changed = append(changed, relPath)
	}

	return changed, nil
}
