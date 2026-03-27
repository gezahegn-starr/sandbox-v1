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
	Args:               cobra.ArbitraryArgs,
	RunE:               runExec,
	DisableFlagParsing: true,
}

func init() {
	rootCmd.AddCommand(execCmd)
}

func runExec(_ *cobra.Command, args []string) error {
	var sandboxID string
	var cmdArgs []string

	if len(args) == 0 {
		var err error
		sandboxID, err = pickSandbox("Select a sandbox to exec into")
		if err != nil {
			return err
		}

		// Default to bash
		cmdArgs = []string{"bash"}
	} else {
		sandboxID = args[0]
		cmdArgs = args[1:]

		// strip leading "--" separator
		if len(cmdArgs) > 0 && cmdArgs[0] == "--" {
			cmdArgs = cmdArgs[1:]
		}
		if len(cmdArgs) == 0 {
			return fmt.Errorf("no command specified")
		}
	}

	execArgs := append([]string{"exec", "-it", sandboxID}, cmdArgs...)
	debugLog("exec: container %v", execArgs)

	defer saveAndRestoreTerminal()()
	c := exec.Command(containerBin(), execArgs...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	defer forwardSignals(c)()

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("exec failed: %w", err)
	}
	return nil
}
