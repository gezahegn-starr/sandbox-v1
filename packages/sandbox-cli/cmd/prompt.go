package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
)

// listSandboxes returns all copilot-* sandbox containers.
func listSandboxes() ([]Container, error) {
	out, err := exec.Command(containerBin(), "ls", "--all", "--format", "json").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}

	var containers []Container
	if err := json.Unmarshal(out, &containers); err != nil {
		return nil, fmt.Errorf("parsing container list: %w", err)
	}

	var sandboxes []Container
	for _, c := range containers {
		if c.isSandbox() {
			sandboxes = append(sandboxes, c)
		}
	}
	return sandboxes, nil
}

// sandboxSelectOptions builds huh select options labelled with name and status.
func sandboxSelectOptions(sandboxes []Container) []huh.Option[string] {
	opts := make([]huh.Option[string], len(sandboxes))
	for i, c := range sandboxes {
		label := fmt.Sprintf("%-40s %s", c.Config.ID, c.Status)
		opts[i] = huh.NewOption(strings.TrimSpace(label), c.Config.ID)
	}
	return opts
}

// pickSandbox shows an interactive dropdown to select a single sandbox by name.
// Returns an error if the user cancels or no sandboxes exist.
func pickSandbox(title string) (string, error) {
	sandboxes, err := listSandboxes()
	if err != nil {
		return "", err
	}
	if len(sandboxes) == 0 {
		return "", fmt.Errorf("no sandboxes found")
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(sandboxSelectOptions(sandboxes)...).
				Value(&selected),
		),
	)
	if err := form.Run(); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return "", fmt.Errorf("aborted")
		}
		return "", err
	}
	return selected, nil
}

// saveAndRestoreTerminal saves the current terminal state and returns a restore
// function. Call it with defer before running any interactive container process
// so the terminal is always restored even if the process crashes or panics.
//
//	defer saveAndRestoreTerminal()()
func saveAndRestoreTerminal() func() {
	fd := int(os.Stdin.Fd())
	state, err := term.GetState(fd)
	if err != nil {
		// Not a TTY or unsupported — nothing to restore.
		return func() {}
	}
	return func() {
		_ = term.Restore(fd, state)
	}
}
