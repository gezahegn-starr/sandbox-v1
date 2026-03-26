package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save SANDBOX IMAGE_NAME",
	Short: "Save a snapshot of the sandbox as a template",
	Long: `Commit the current state of a sandbox container as a new image.

Example:
  sandbox save copilot-myproject my-agent-template`,
	Args: cobra.ExactArgs(2),
	RunE: runSave,
}

func init() {
	rootCmd.AddCommand(saveCmd)
}

func runSave(_ *cobra.Command, args []string) error {
	sandboxID := args[0]
	imageName := args[1]

	debugLog("committing %s as image %s", sandboxID, imageName)

	c := exec.Command(containerBin(), "commit", sandboxID, imageName)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		return fmt.Errorf("saving sandbox: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Saved snapshot of %s as image: %s\n", sandboxID, imageName)
	return nil
}
