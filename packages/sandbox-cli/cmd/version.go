package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show sandbox version information",
	Args:  cobra.NoArgs,
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(_ *cobra.Command, _ []string) error {
	fmt.Printf("sandbox version %s\n\n", version)

	// Also show underlying container runtime version
	c := exec.Command(containerBin(), "system", "version")
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	// Non-fatal if container CLI is unavailable
	if err := c.Run(); err != nil {
		debugLog("container system version: %v", err)
	}
	return nil
}
