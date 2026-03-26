package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset all VM sandboxes and clean up state",
	Long: `Stop and remove all sandboxes (containers named copilot-*).

Prompts for confirmation unless --force is passed.`,
	Args: cobra.NoArgs,
	RunE: runReset,
}

var resetForce bool

func init() {
	rootCmd.AddCommand(resetCmd)
	resetCmd.Flags().BoolVarP(&resetForce, "force", "f", false, "Skip confirmation prompt")
}

func runReset(_ *cobra.Command, _ []string) error {
	out, err := exec.Command(containerBin(), "ls", "--all", "--format", "json").CombinedOutput()
	if err != nil {
		return fmt.Errorf("listing containers: %w\n%s", err, out)
	}

	var containers []Container
	if err := json.Unmarshal(out, &containers); err != nil {
		return fmt.Errorf("parsing container list: %w", err)
	}

	var sandboxes []Container
	for _, c := range containers {
		if c.isSandbox() {
			sandboxes = append(sandboxes, c)
		}
	}

	if len(sandboxes) == 0 {
		fmt.Println("No sandboxes found.")
		return nil
	}

	names := make([]string, len(sandboxes))
	for i, s := range sandboxes {
		names[i] = s.Config.ID
	}

	if !resetForce {
		fmt.Printf("This will remove the following sandboxes:\n  %s\n\nContinue? [y/N] ", strings.Join(names, "\n  "))
		var answer string
		fmt.Scanln(&answer)
		answer = strings.ToLower(strings.TrimSpace(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	var failed []string
	for _, s := range sandboxes {
		id := s.Config.ID
		if s.Status == "running" {
			debugLog("stopping %s", id)
			if out, err := exec.Command(containerBin(), "stop", id).CombinedOutput(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: stopping %s: %v: %s\n", id, err, out)
			}
		}
		debugLog("removing %s", id)
		if out, err := exec.Command(containerBin(), "rm", id).CombinedOutput(); err != nil {
			fmt.Fprintf(os.Stderr, "error removing %s: %v: %s\n", id, err, out)
			failed = append(failed, id)
		} else {
			fmt.Printf("Removed: %s\n", id)
		}
	}

	if len(failed) > 0 {
		return fmt.Errorf("failed to remove: %s", strings.Join(failed, ", "))
	}
	return nil
}
