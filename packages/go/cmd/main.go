package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run() {
	containerCmd := exec.Command("container", strings.Fields("ls --all --format json")...)
	out, err := containerCmd.CombinedOutput()
	containers := []Container{}

	if err == nil {
		err := json.Unmarshal(out, &containers)
		if err != nil {
			log.Fatal("Error listing containers",err)
		}
	}
	inputPath := os.Args[1]

	absPath, err := filepath.Abs(inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error resolving path: %v\n", err)
		os.Exit(1)
	}

	paths := strings.Split(absPath, "/")
	projectName := ""
	if len(paths) > 0 {
		projectName = paths[len(paths) -1]
	}

	var container Container

	containerName := fmt.Sprintf("copilot-%s", projectName)

	for _, c := range containers {
		if c.Config.ID == containerName {
			container = c
			break
		}
	}

	containerExists := container.Config.ID == containerName
	if containerExists {
		fmt.Fprintf(os.Stderr, "Reusing existing container: %s\n", containerName)

		// If running, stop it first so we can restart with the entrypoint
		if container.Status == "running" {
			stopCmd := exec.Command("sh", "-c", fmt.Sprintf("container stop %s", containerName))
			if err := stopCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "error stopping container: %v\n", err)
				os.Exit(1)
			}
		}

		// Start and attach — entrypoint will run (starting copilot)
		cmd := exec.Command("sh", "-c", fmt.Sprintf("container start -a -i %s", containerName))
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				os.Exit(exitErr.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Container doesn't exist — create a new one
	cmdStr := fmt.Sprintf("container create -e GITHUB_TOKEN=${GITHUB_TOKEN} -e WORKSPACE_PATH=%s -v %s:/home/agent/workspace -v %s:%s --name %s -i -t agent", absPath, absPath, absPath,absPath, containerName)

	cmd := exec.Command("sh", "-c", cmdStr)

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "error creating container: %v\n", err)
		os.Exit(1)
	}

	// Start the container so we can exec into it to write config
	startCmd := exec.Command("sh", "-c", fmt.Sprintf("container start %s", containerName))
	if err := startCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error starting container: %v\n", err)
		os.Exit(1)
	}

	// Write copilot config into the container
	content := fmt.Sprintf(`{
  "banner": "never",
	"trusted_folders": ["/home/agent/workspace","%s"]
}`, absPath)
	configStr := fmt.Sprintf("container exec %s sh -c 'cat << EOF > /home/agent/.copilot/config.json\n%s\nEOF'", containerName, content)

	configCmd := exec.Command("sh", "-c", configStr)
	if err := configCmd.Run(); err != nil {
		log.Fatal("Config error: ", err)
	}

	// Stop the container so we can restart it cleanly with the entrypoint
	stopCmd := exec.Command("sh", "-c", fmt.Sprintf("container stop %s", containerName))
	if err := stopCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error stopping container: %v\n", err)
		os.Exit(1)
	}

	// Restart and attach — the entrypoint will run again (starting copilot)
	cmd = exec.Command("sh", "-c", fmt.Sprintf("container start -a -i %s", containerName))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
