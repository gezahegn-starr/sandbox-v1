package report

import (
	"fmt"
	"os"
	"strings"
	"time"

	"pr-review/internal/gh"
	"pr-review/internal/review"
)

// Item is a single entry in the human-review report.
type Item struct {
	Comment        gh.ReviewComment
	Reason         string         // why human review is needed
	ValidationFail *review.ValidationResult // non-nil if an auto-fix attempt failed validation
}

// Report holds all data needed to render the human-review markdown document.
type Report struct {
	PRMeta         *gh.PRMeta
	Items          []Item
	AutoFixed      int // number of comments successfully auto-fixed
	GeneratedAt    time.Time
}

// New creates a new Report.
func New(meta *gh.PRMeta) *Report {
	return &Report{
		PRMeta:      meta,
		GeneratedAt: time.Now(),
	}
}

// AddHumanItem appends a comment that requires human attention.
func (r *Report) AddHumanItem(comment gh.ReviewComment, reason string, valResult *review.ValidationResult) {
	r.Items = append(r.Items, Item{
		Comment:        comment,
		Reason:         reason,
		ValidationFail: valResult,
	})
}

// Write renders the report to a markdown file at the given path.
func (r *Report) Write(outputPath string) error {
	content := r.render()
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing report to %s: %w", outputPath, err)
	}
	return nil
}

func (r *Report) render() string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# PR Review Report — PR #%d\n\n", r.PRMeta.Number)
	fmt.Fprintf(&sb, "> **%s**  \n", r.PRMeta.Title)
	fmt.Fprintf(&sb, "> %s  \n", r.PRMeta.URL)
	fmt.Fprintf(&sb, "> Generated: %s\n\n", r.GeneratedAt.Format(time.RFC1123))

	fmt.Fprintf(&sb, "## Summary\n\n")
	fmt.Fprintf(&sb, "| | Count |\n|---|---|\n")
	fmt.Fprintf(&sb, "| ✅ Auto-fixed | %d |\n", r.AutoFixed)
	fmt.Fprintf(&sb, "| 🙋 Needs human review | %d |\n\n", len(r.Items))

	if len(r.Items) == 0 {
		sb.WriteString("All review comments were addressed automatically. 🎉\n")
		return sb.String()
	}

	sb.WriteString("## Comments Requiring Human Review\n\n")

	for i, item := range r.Items {
		c := item.Comment
		commentURL := c.URL
		if commentURL == "" {
			// Construct a best-effort URL from the PR URL and file path.
			commentURL = fmt.Sprintf("%s/files#%s", r.PRMeta.URL, c.Path)
		}

		fmt.Fprintf(&sb, "### %d. `%s` (line %d)\n\n", i+1, c.Path, c.Line)
		fmt.Fprintf(&sb, "🔗 [View comment on GitHub](%s)\n\n", commentURL)
		fmt.Fprintf(&sb, "**Review comment:**\n\n> %s\n\n",
			strings.ReplaceAll(strings.TrimSpace(c.Body), "\n", "\n> "))
		fmt.Fprintf(&sb, "**Why human review is needed:** %s\n\n", item.Reason)

		if item.ValidationFail != nil {
			sb.WriteString("**Auto-fix was attempted but failed validation:**\n\n")
			for _, cr := range item.ValidationFail.Commands {
				if !cr.Passed {
					fmt.Fprintf(&sb, "<details>\n<summary><code>%s</code> — failed</summary>\n\n```\n%s\n```\n\n</details>\n\n",
						cr.Command, cr.Output)
				}
			}
		}

		if c.DiffHunk != "" {
			fmt.Fprintf(&sb, "<details>\n<summary>Diff context</summary>\n\n```diff\n%s\n```\n\n</details>\n\n", c.DiffHunk)
		}

		sb.WriteString("---\n\n")
	}

	return sb.String()
}
