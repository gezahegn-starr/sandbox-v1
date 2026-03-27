package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"

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

	hostMount := ""
	if m := hostSkillsMount(); m != "" {
		hostMount = " -v " + m
	}
	cmdStr := fmt.Sprintf(
		"container create -e GITHUB_TOKEN=${GITHUB_TOKEN} -e WORKSPACE_PATH=%s -v %s:/home/agent/workspace -v %s:%s%s --name %s -i -t agent",
		absPath, absPath, absPath, absPath, hostMount, name,
	)
	debugLog("exec: sh -c %q", cmdStr)

	createCmd := exec.Command("sh", "-c", cmdStr)
	if out, err := createCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("creating sandbox: %w\n%s", err, out)
	}

	// Start and attach
	return attachContainer(name, "")
}

// findContainer looks up a container by name using `container ls --all --format json`.
func findContainer(name string) (*Container, error) {
	out, err := exec.Command(containerBin(), "ls", "--all", "--format", "json").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w\n%s", err, out)
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

// attachContainer stops a running container (if needed) then execs into it,
// replacing the current process entirely so Go is no longer in the signal/IO path.
func attachContainer(name, currentStatus string) error {
	if currentStatus == "running" {
		debugLog("stopping running container before reattach: %s", name)
		stopCmd := exec.Command(containerBin(), "stop", name)
		if out, err := stopCmd.CombinedOutput(); err != nil {
			return fmt.Errorf("stopping sandbox: %w\n%s", err, out)
		}
	}

	binary, err := exec.LookPath(containerBin())
	if err != nil {
		return fmt.Errorf("finding container binary: %w", err)
	}

	args := []string{containerBin(), "start", "-a", "-i", name}
	debugLog("execve: %v", args)
	// Replace this process with the container binary. From here Go is gone —
	// signals, terminal control, and exit codes are all owned by the container.
	if err := syscall.Exec(binary, args, os.Environ()); err != nil {
		return fmt.Errorf("attaching to sandbox: %w", err)
	}
	return nil // unreachable
}
