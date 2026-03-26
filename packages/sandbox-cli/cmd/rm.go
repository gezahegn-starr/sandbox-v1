package cmd

import (
	"fmt"
	"os/exec"
	"os"

	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [OPTIONS] SANDBOX [SANDBOX...]",
	Short: "Remove one or more sandboxes",
	Long: `Remove one or more sandboxes.

Use --force to stop and remove running sandboxes.`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRm,
}

var rmForce bool

func init() {
	rootCmd.AddCommand(rmCmd)
	rmCmd.Flags().BoolVarP(&rmForce, "force", "f", false, "Force remove running sandboxes (stops them first)")
}

func runRm(_ *cobra.Command, args []string) error {
	var firstErr error
	for _, id := range args {
		if rmForce {
			// Best-effort stop before remove
			debugLog("stopping %s before rm", id)
			exec.Command(containerBin(), "stop", id).Run() //nolint:errcheck
		}
		debugLog("removing %s", id)
		if out, err := exec.Command(containerBin(), "rm", id).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "error removing %s: %v: %s\n", id, err, out)
			if firstErr == nil {
				firstErr = err
			}
		} else {
			fmt.Printf("Removed: %s\n", id)
		}
	}
	return firstErr
}
