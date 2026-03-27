package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop SANDBOX [SANDBOX...]",
	Short: "Stop one or more sandboxes without removing them",
	Args:  cobra.ArbitraryArgs,
	RunE:  runStop,
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func runStop(_ *cobra.Command, args []string) error {
	if len(args) == 0 {
		name, err := pickSandbox("Select a sandbox to stop")
		if err != nil {
			return err
		}
		args = []string{name}
	}

	var firstErr error
	for _, id := range args {
		debugLog("stopping %s", id)
		c := exec.Command(containerBin(), "stop", id)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "error stopping %s: %v\n", id, err)
			if firstErr == nil {
				firstErr = err
			}
		} else {
			fmt.Printf("Stopped: %s\n", id)
		}
	}
	return firstErr
}
