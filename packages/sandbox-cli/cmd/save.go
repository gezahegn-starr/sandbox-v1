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
	Args: cobra.RangeArgs(0, 2),
	RunE: runSave,
}

func init() {
	rootCmd.AddCommand(saveCmd)
}

func runSave(_ *cobra.Command, args []string) error {
	var sandboxID, imageName string

	if len(args) == 0 {
		var err error
		sandboxID, err = pickSandbox("Select a sandbox to save")
		if err != nil {
			return err
		}
	} else {
		sandboxID = args[0]
	}

	if len(args) < 2 {
		return fmt.Errorf("image name required: sandbox save %s IMAGE_NAME", sandboxID)
	}
	imageName = args[1]

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
