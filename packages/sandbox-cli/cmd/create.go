package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// hostSkillsMount returns a -v flag to mount the host's ~/.copilot/skills directory
// into the container as read-only staging, or an empty string if it doesn't exist.
func hostSkillsMount() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	hostDir := filepath.Join(home, ".copilot", "skills")
	if _, err := os.Stat(hostDir); err != nil {
		return ""
	}
	return fmt.Sprintf("%s:/home/agent/.copilot-host-skills:ro", hostDir)
}

var createCmd = &cobra.Command{
	Use:   "create [OPTIONS] [WORKSPACE_PATH]",
	Short: "Create a sandbox for an agent",
	Long: `Create a new Apple container sandbox for an AI agent.

If WORKSPACE_PATH is provided, the directory is mounted into the container and
a Copilot config is written. The sandbox is named copilot-<project> by default.`,
	RunE: runCreate,
}

var (
	createName   string
	createImage  string
	createCPUs   int
	createMemory int
)

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringVar(&createName, "name", "", "Sandbox name (default: copilot-<workspace dir>)")
	createCmd.Flags().StringVar(&createImage, "image", "agent", "Container image to use")
	createCmd.Flags().IntVar(&createCPUs, "cpus", 0, "Number of CPUs (0 = use container default)")
	createCmd.Flags().IntVar(&createMemory, "memory", 0, "Memory in MB (0 = use container default)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	var absPath string

	if len(args) > 0 {
		p, err := filepath.Abs(args[0])
		if err != nil {
			return fmt.Errorf("resolving workspace path: %w", err)
		}
		absPath = p
	}

	name := createName
	if name == "" {
		if absPath != "" {
			name = sandboxName(absPath)
		} else {
			return fmt.Errorf("provide a --name or a WORKSPACE_PATH")
		}
	}

	cmdArgs := []string{"create"}
	cmdArgs = append(cmdArgs, "-e", "GITHUB_TOKEN=${GITHUB_TOKEN}")
	if absPath != "" {
		cmdArgs = append(cmdArgs, "-e", fmt.Sprintf("WORKSPACE_PATH=%s", absPath))
		cmdArgs = append(cmdArgs, "-v", fmt.Sprintf("%s:/home/agent/workspace", absPath))
		cmdArgs = append(cmdArgs, "-v", fmt.Sprintf("%s:%s", absPath, absPath))
	}
	if mount := hostSkillsMount(); mount != "" {
		cmdArgs = append(cmdArgs, "-v", mount)
	}
	if createCPUs > 0 {
		cmdArgs = append(cmdArgs, "--cpus", fmt.Sprintf("%d", createCPUs))
	}
	if createMemory > 0 {
		cmdArgs = append(cmdArgs, "--memory", fmt.Sprintf("%dm", createMemory))
	}
	cmdArgs = append(cmdArgs, "--name", name, "-i", "-t", createImage)

	cmdStr := containerBin() + " " + strings.Join(cmdArgs, " ")
	debugLog("exec: sh -c %q", cmdStr)

	c := exec.Command("sh", "-c", cmdStr)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err := c.Run(); err != nil {
		return fmt.Errorf("creating sandbox: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Created sandbox: %s\n", name)
	return nil
}

// sandboxName derives the canonical copilot-<project> name from a path.
func sandboxName(absPath string) string {
	parts := strings.Split(absPath, "/")
	project := parts[len(parts)-1]
	return fmt.Sprintf("copilot-%s", project)
}

