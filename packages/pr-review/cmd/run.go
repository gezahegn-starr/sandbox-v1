package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"pr-review/internal/gh"
	"pr-review/internal/git"
	"pr-review/internal/report"
	"pr-review/internal/review"
)

var (
	flagRepo     string
	flagPR       int
	flagDryRun   bool
	flagOutput   string
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Fetch PR comments, apply Copilot fixes, and push changes",
	Long: `Fetches all review comments on a pull request, classifies each one with
Copilot, auto-fixes the actionable ones, validates the changes, pushes to the
branch, and writes a markdown report of items that need human review.`,
	RunE: runReview,
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.Flags().StringVar(&flagRepo, "repo", "", "Repository in owner/repo format (auto-detected if omitted)")
	runCmd.Flags().IntVar(&flagPR, "pr", 0, "PR number (auto-detected from current branch if omitted)")
	runCmd.Flags().BoolVar(&flagDryRun, "dry-run", false, "Classify and plan without applying changes or pushing")
	runCmd.Flags().StringVar(&flagOutput, "output", "pr-review-report.md", "Path for the human-review markdown report")
}

// Styles for terminal output.
var (
	styleHeader  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	styleWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	styleDim     = lipgloss.NewStyle().Faint(true)
)

func runReview(_ *cobra.Command, _ []string) error {
	ctx := context.Background()

	// Warn if no repo-scoped token is available for GitHub API calls.
	if os.Getenv("PR_REVIEW_TOKEN") == "" {
		fmt.Fprintln(os.Stderr, styleWarn.Render(
			"⚠  PR_REVIEW_TOKEN is not set. Falling back to GITHUB_TOKEN.\n"+
				"   GITHUB_TOKEN is Copilot-scoped and may fail on private repositories.\n"+
				"   Set PR_REVIEW_TOKEN to a token with 'repo' scope for private repo support."))
		fmt.Fprintln(os.Stderr)
	}

	printHeader("🔍 Fetching PR data")
	meta, comments, err := gh.FetchPR(flagRepo, flagPR)
	if err != nil {
		return fmt.Errorf("fetching PR: %w", err)
	}

	fmt.Printf("  PR #%d: %s\n", meta.Number, meta.Title)
	fmt.Printf("  Branch: %s → %s\n", meta.HeadRefName, meta.BaseRefName)
	fmt.Printf("  %d review comment(s) found\n\n", len(comments))

	if len(comments) == 0 {
		fmt.Println(styleSuccess.Render("No review comments to process. ✅"))
		return nil
	}

	rpt := report.New(meta)

	// --- Phase 1: Classify all comments ---
	printHeader("🤖 Classifying comments with Copilot")

	type classifiedComment struct {
		comment gh.ReviewComment
		cls     *review.Classification
		content string
	}

	var classified []classifiedComment
	for i, c := range comments {
		fmt.Printf("  [%d/%d] %s:%d ... ", i+1, len(comments), c.Path, c.Line)

		content, _ := gh.FetchFileContent(c.Path)

		cls, err := review.Classify(ctx, meta, c, content)
		if err != nil {
			fmt.Println(styleWarn.Render("⚠ classify failed, marking as human"))
			debugLog("classify error: %v", err)
			rpt.AddHumanItem(c, fmt.Sprintf("Classification error: %v", err), nil)
			continue
		}

		if cls.Verdict == review.VerdictHumanRequired {
			fmt.Println(styleWarn.Render("🙋 human required"))
			rpt.AddHumanItem(c, cls.Reason, nil)
		} else {
			fmt.Println(styleSuccess.Render("✅ fixable"))
		}

		classified = append(classified, classifiedComment{comment: c, cls: cls, content: content})
	}

	// --- Phase 2: Apply fixes for FIXABLE comments ---
	printHeader("🔧 Applying Copilot fixes")

	var staged int
	for _, cc := range classified {
		if cc.cls.Verdict != review.VerdictFixable {
			continue
		}

		c := cc.comment
		fmt.Printf("  Fixing %s:%d ... ", c.Path, c.Line)

		if flagDryRun {
			fmt.Println(styleDim.Render("[dry-run, skipped]"))
			continue
		}

		fixResult, err := review.Fix(ctx, meta, c, cc.content)
		if err != nil {
			fmt.Println(styleError.Render("✗ fix failed"))
			debugLog("fix error: %v", err)
			rpt.AddHumanItem(c, fmt.Sprintf("Copilot fix error: %v", err), nil)
			continue
		}

		if fixResult.Skipped {
			fmt.Println(styleDim.Render("no change needed"))
			continue
		}

		debugLog("files changed: %v", fixResult.FilesChanged)

		// Validate the fix.
		fmt.Print("validating ... ")
		valResult, err := review.Validate()
		if err != nil {
			fmt.Println(styleError.Render("✗ validation error"))
			debugLog("validation error: %v", err)
			_ = git.Restore()
			rpt.AddHumanItem(c, fmt.Sprintf("Validation error: %v", err), nil)
			continue
		}

		if !valResult.Passed {
			fmt.Println(styleError.Render("✗ tests/lint failed"))
			_ = git.Restore()
			rpt.AddHumanItem(c, "Auto-fix failed tests/lint — see report for details", valResult)
			continue
		}

		fmt.Println(styleSuccess.Render("✅ passed"))
		staged++
	}

	// --- Phase 3: Commit & push ---
	if !flagDryRun && staged > 0 {
		printHeader("📤 Committing and pushing changes")

		hasChanges, err := git.HasChanges()
		if err != nil {
			return fmt.Errorf("checking git status: %w", err)
		}

		if hasChanges {
			commitMsg := fmt.Sprintf("chore: apply Copilot fixes for PR #%d review comments", meta.Number)
			if err := git.AddAll(); err != nil {
				return fmt.Errorf("staging changes: %w", err)
			}
			if err := git.Commit(commitMsg); err != nil {
				return fmt.Errorf("committing: %w", err)
			}
			if err := git.Push(); err != nil {
				return fmt.Errorf("pushing: %w", err)
			}
			fmt.Println(styleSuccess.Render("  Changes pushed ✅"))
		} else {
			fmt.Println(styleDim.Render("  No changes to commit."))
		}
	}

	// --- Phase 4: Write human-review report ---
	rpt.AutoFixed = staged
	if len(rpt.Items) > 0 || flagDryRun {
		printHeader("📝 Writing human-review report")
		if err := rpt.Write(flagOutput); err != nil {
			return fmt.Errorf("writing report: %w", err)
		}
		fmt.Printf("  Report written to: %s\n", flagOutput)
	}

	// --- Summary ---
	fmt.Println()
	printHeader("Summary")
	fmt.Printf("  ✅ Auto-fixed and pushed : %d comment(s)\n", staged)
	fmt.Printf("  🙋 Needs human review    : %d comment(s)\n", len(rpt.Items))
	if len(rpt.Items) > 0 {
		fmt.Printf("  👉 See %s for details\n", flagOutput)
	}

	if !flagDryRun && len(rpt.Items) > 0 {
		// Exit with code 1 to signal that manual review is still needed.
		os.Exit(1)
	}

	return nil
}

func printHeader(title string) {
	fmt.Println(styleHeader.Render(title))
}
