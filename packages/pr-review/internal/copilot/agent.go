package copilot

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

const defaultTimeout = 3 * time.Minute

// Response holds the raw text output from a Copilot invocation.
type Response struct {
	Output string
}

// Ask invokes the Copilot CLI in headless/piped mode with the given prompt and
// returns the response.
//
// The container environment is expected to have GITHUB_TOKEN / GH_TOKEN set.
// The standalone `copilot` binary (installed by the base agent image) is tried
// first; `gh copilot` is used as a fallback when running outside the container.
func Ask(ctx context.Context, prompt string) (*Response, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	output, err := runCopilot(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("copilot invocation failed: %w", err)
	}

	return &Response{Output: strings.TrimSpace(output)}, nil
}

// runCopilot tries the standalone `copilot --prompt` binary first (always
// present in the agent container), then falls back to `gh copilot`.
func runCopilot(ctx context.Context, prompt string) (string, error) {
	// Primary: standalone copilot binary — matches entrypoint.sh usage:
	//   copilot --allow-all-tools --prompt "<prompt>"
	if _, err := exec.LookPath("copilot"); err == nil {
		out, err := runStandaloneCopilot(ctx, prompt)
		if err == nil {
			return out, nil
		}
		// Fall through to gh copilot on failure.
	}

	// Fallback: gh copilot (useful when running outside the agent container).
	return runGhCopilot(ctx, prompt)
}

// runStandaloneCopilot uses the `copilot` binary directly, matching how the
// base agent image's entrypoint.sh invokes it.
func runStandaloneCopilot(ctx context.Context, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, "copilot", "--allow-all-tools", "--prompt", prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("copilot --prompt: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

// runGhCopilot uses `gh copilot -- -p "<prompt>"` as a fallback.
// The `--` ensures the -p flag is forwarded to the Copilot CLI, not consumed by gh.
func runGhCopilot(ctx context.Context, prompt string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "copilot", "--", "-p", prompt)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("gh copilot: %w\nstderr: %s", err, stderr.String())
	}

	return stdout.String(), nil
}
