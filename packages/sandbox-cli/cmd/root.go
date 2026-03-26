package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

const version = "0.1.0"

var debugMode bool

var rootCmd = &cobra.Command{
	Use:   "sandbox",
	Short: "Local sandbox environments for AI agents, using Apple containers.",
	Long:  "Local sandbox environments for AI agents, using Apple containers.",
	// Print usage like docker sandbox (no automatic "help" subcommand confusion)
	SilenceErrors: true,
	SilenceUsage:  true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "ERROR:", err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "D", false, "Enable debug logging")

	// Customize usage template to mirror docker sandbox style
	rootCmd.SetUsageTemplate(`Usage:  sandbox [OPTIONS] COMMAND

Local sandbox environments for AI agents, using Apple containers.

Options:
{{.InheritedFlags.FlagUsages | trimRightSpace}}

Management Commands:
  create      Create a sandbox for an agent
  network     Manage sandbox networking

Commands:
  exec        Execute a command inside a sandbox
  ls          List sandboxes
  reset       Reset all sandboxes and clean up state
  rm          Remove one or more sandboxes
  run         Run an agent in a sandbox
  save        Save a snapshot of the sandbox as a template
  stop        Stop one or more sandboxes without removing them
  version     Show sandbox version information

Run 'sandbox COMMAND --help' for more information on a command.
`)
}

// debugLog prints a message only when debug mode is enabled.
func debugLog(format string, args ...interface{}) {
	if debugMode {
		fmt.Fprintf(os.Stderr, "[debug] "+format+"\n", args...)
	}
}

// containerBin returns the Apple container binary path.
func containerBin() string {
	return "container"
}
