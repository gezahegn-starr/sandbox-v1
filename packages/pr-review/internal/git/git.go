package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// CurrentBranch returns the name of the currently checked-out git branch.
func CurrentBranch() (string, error) {
	out, err := run("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("detecting current branch: %w", err)
	}
	return strings.TrimSpace(out), nil
}

// HasChanges returns true if there are staged or unstaged changes in the working tree.
func HasChanges() (bool, error) {
	out, err := run("git", "status", "--porcelain")
	if err != nil {
		return false, fmt.Errorf("git status: %w", err)
	}
	return strings.TrimSpace(out) != "", nil
}

// AddAll stages all modified and new files.
func AddAll() error {
	_, err := run("git", "add", "-A")
	if err != nil {
		return fmt.Errorf("git add: %w", err)
	}
	return nil
}

// Commit creates a commit with the given message.
func Commit(message string) error {
	_, err := run("git", "commit", "-m", message)
	if err != nil {
		return fmt.Errorf("git commit: %w", err)
	}
	return nil
}

// Push pushes the current branch to its upstream remote.
func Push() error {
	branch, err := CurrentBranch()
	if err != nil {
		return err
	}
	_, err = run("git", "push", "origin", branch)
	if err != nil {
		return fmt.Errorf("git push: %w", err)
	}
	return nil
}

// Restore reverts all unstaged changes to tracked files, discarding any
// modifications Copilot made that failed validation.
func Restore() error {
	_, err := run("git", "checkout", "--", ".")
	if err != nil {
		return fmt.Errorf("git restore: %w", err)
	}
	// Also remove any newly created untracked files.
	_, err = run("git", "clean", "-fd")
	if err != nil {
		return fmt.Errorf("git clean: %w", err)
	}
	return nil
}

// run executes a git command and returns stdout, returning an error that
// includes stderr on failure.
func run(name string, args ...string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("%w\n%s", err, stderr.String())
	}
	return stdout.String(), nil
}
