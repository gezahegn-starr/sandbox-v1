package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec SANDBOX COMMAND [ARGS...]",
	Short: "Execute a command inside a sandbox",
	Long: `Execute a command inside a running sandbox.

Example:
  sandbox exec copilot-myproject -- ls /home/agent/workspace`,
	Args:               cobra.MinimumNArgs(2),
	RunE:               runExec,
	DisableFlagParsing: true,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func runExec(_ *cobra.Command, args []string) error {
	// args[0] = sandbox name/id, rest = command (strip leading "--" if present)
	sandboxID := args[0]
	cmdArgs := args[1:]
	if len(cmdArgs) > 0 && cmdArgs[0] == "--" {
		cmdArgs = cmdArgs[1:]
	}
	if len(cmdArgs) == 0 {
		return fmt.Errorf("no command specified")
	}

	execArgs := append([]string{"exec", sandboxID}, cmdArgs...)
	debugLog("exec: container %v", execArgs)

	c := exec.Command(containerBin(), execArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("exec failed: %w", err)
	}
	return nil
}
