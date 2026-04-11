package review

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ValidationResult holds the outcome of running the repo's test/lint suite.
type ValidationResult struct {
	Passed   bool
	Commands []CommandResult
}

// CommandResult holds the output of a single validation command.
type CommandResult struct {
	Command string
	Output  string
	Passed  bool
}

// Validate detects and runs the repository's existing test/lint commands
// from the current working directory.
func Validate() (*ValidationResult, error) {
	cmds, err := detectCommands()
	if err != nil {
		return nil, err
	}
	if len(cmds) == 0 {
		return &ValidationResult{Passed: true}, nil
	}

	result := &ValidationResult{Passed: true}
	for _, c := range cmds {
		cr := runCommand(c)
		result.Commands = append(result.Commands, cr)
		if !cr.Passed {
			result.Passed = false
		}
	}
	return result, nil
}

// detectCommands inspects the workspace and returns commands to run for
// validation, in priority order.
func detectCommands() ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	var cmds []string

	// Go module
	if fileExists(filepath.Join(cwd, "go.mod")) {
		cmds = append(cmds, "go vet ./...", "go test ./...")
	}

	// Node / npm
	if fileExists(filepath.Join(cwd, "package.json")) {
		if hasScript(filepath.Join(cwd, "package.json"), "lint") {
			cmds = append(cmds, "npm run lint --if-present")
		}
		if hasScript(filepath.Join(cwd, "package.json"), "test") {
			cmds = append(cmds, "npm test --if-present")
		}
	}

	// Makefile targets
	if fileExists(filepath.Join(cwd, "Makefile")) {
		if makeTargetExists("lint") {
			cmds = append(cmds, "make lint")
		}
		if makeTargetExists("test") {
			cmds = append(cmds, "make test")
		}
	}

	return cmds, nil
}

func runCommand(command string) CommandResult {
	parts := strings.Fields(command)
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	output := stdout.String()
	if stderr.Len() > 0 {
		output += "\n" + stderr.String()
	}

	return CommandResult{
		Command: command,
		Output:  strings.TrimSpace(output),
		Passed:  err == nil,
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// hasScript checks whether a package.json file contains the named script key.
func hasScript(packageJSON, script string) bool {
	data, err := os.ReadFile(packageJSON)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), fmt.Sprintf("%q", script))
}

// makeTargetExists checks whether a Makefile defines the named target.
func makeTargetExists(target string) bool {
	out, err := exec.Command("make", "-n", target).CombinedOutput()
	if err != nil {
		return false
	}
	return len(out) > 0
}
