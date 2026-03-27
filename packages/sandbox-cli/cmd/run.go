package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run WORKSPACE_PATH",
	Short: "Run an agent in a sandbox",
	Long: `Run an agent in a sandbox for the given workspace directory.

If a sandbox for the project already exists it is reused (restarted if stopped).
Otherwise a new sandbox is created, configured with a Copilot config, and started.
An interactive session is attached so the agent runs in the foreground.`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(_ *cobra.Command, args []string) error {
	// No argument: pick from existing sandboxes interactively.
	if len(args) == 0 {
		name, err := pickSandbox("Select a sandbox to run")
		if err != nil {
			return err
		}
		existing, err := findContainer(name)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "Reusing existing sandbox: %s\n", name)
		status := ""
		if existing != nil {
			status = existing.Status
		}
		return attachContainer(name, status)
	}

	absPath, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("resolving workspace path: %w", err)
	}

	name := sandboxName(absPath)
	existing, err := findContainer(name)
	if err != nil {
		return err
	}

	if existing != nil {
		fmt.Fprintf(os.Stderr, "Reusing existing sandbox: %s\n", name)
		return attachContainer(name, existing.Status)
	}

	// Create new sandbox
	fmt.Fprintf(os.Stderr, "Creating sandbox: %s\n", name)

	cmdStr := fmt.Sprintf(
		"container create -e GITHUB_TOKEN=${GITHUB_TOKEN} -e WORKSPACE_PATH=%s -v %s:/home/agent/workspace -v %s:%s --name %s -i -t agent",
		absPath, absPath, absPath, absPath, name,
	)
	debugLog("exec: sh -c %q", cmdStr)

	createCmd := exec.Command("sh", "-c", cmdStr)
	if out, err := createCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating sandbox: %w\n%s", err, out)
	}

	if err := writeCopilotConfig(name, absPath); err != nil {
		return err
	}

	// Start and attach
	return attachContainer(name, "")
}

// findContainer looks up a container by name using `container ls --all --format json`.
func findContainer(name string) (*Container, error) {
	out, err := exec.Command(containerBin(), "ls", "--all", "--format", "json").CombinedOutput()
	if err != nil {
		debugLog("container ls failed: %v", err)
		return nil, nil
	}

	var containers []Container
	if err := json.Unmarshal(out, &containers); err != nil {
		return nil, fmt.Errorf("parsing container list: %w", err)
	}

	for i := range containers {
		if containers[i].Config.ID == name {
			return &containers[i], nil
		}
	}
	return nil, nil
}

// attachContainer stops a running container (if needed) then starts it with an attached session.
func attachContainer(name, currentStatus string) error {
	if currentStatus == "running" {
		debugLog("stopping running container before reattach: %s", name)
		stopCmd := exec.Command(containerBin(), "stop", name)
		if out, err := stopCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("stopping sandbox: %w\n%s", err, out)
		}
	}

	debugLog("attaching to container: %s", name)
	defer saveAndRestoreTerminal()()
	c := exec.Command(containerBin(), "start", "-a", "-i", name)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	defer forwardSignals(c)()

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("attaching to sandbox: %w", err)
	}
	return nil
}
