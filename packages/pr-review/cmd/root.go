package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var debugMode bool

var rootCmd = &cobra.Command{
	Use:   "pr-review",
	Short: "Automatically respond to PR review comments using GitHub Copilot.",
	Long: `pr-review fetches pull request review comments, uses GitHub Copilot to
classify each one (auto-fixable vs. needs human judgment), applies code fixes,
validates them with the repo's test/lint suite, pushes the changes, and
produces a report of items that need human attention.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "D", false, "Enable debug logging")
}

// debugLog prints only when --debug is set.
func debugLog(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}
