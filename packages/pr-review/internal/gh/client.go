package gh

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// FetchPR retrieves PR metadata and all inline review comments for the given PR.
// If prNumber is 0 it is auto-detected from the current branch.
// repo should be "owner/repo"; if empty it is auto-detected from the git remote.
func FetchPR(repo string, prNumber int) (*PRMeta, []ReviewComment, error) {
	owner, name, err := resolveRepo(repo)
	if err != nil {
		return nil, nil, err
	}

	// Fetch PR metadata via `gh pr view --json`.
	viewArgs := []string{"pr", "view",
		"--json", "number,title,url,headRefName,baseRefName",
		"--repo", fmt.Sprintf("%s/%s", owner, name),
	}
	if prNumber != 0 {
		viewArgs = append(viewArgs, fmt.Sprintf("%d", prNumber))
	}

	viewOut, err := ghRun(viewArgs...)
	if err != nil {
		return nil, nil, err
	}

	var payload prViewPayload
	if err := json.Unmarshal(viewOut, &payload); err != nil {
		return nil, nil, fmt.Errorf("parsing gh pr view output: %w", err)
	}

	meta := &PRMeta{
		Number:      payload.Number,
		Title:       payload.Title,
		URL:         payload.URL,
		HeadRefName: payload.HeadRefName,
		BaseRefName: payload.BaseRefName,
		RepoOwner:   owner,
		RepoName:    name,
	}

	// Fetch inline review comments via the REST API.
	// `gh pr view --json` does not expose reviewComments; we must use `gh api`.
	comments, err := fetchReviewComments(owner, name, payload.Number)
	if err != nil {
		return nil, nil, err
	}

	return meta, comments, nil
}

// fetchReviewComments uses `gh api` to retrieve all inline PR review comments,
// filtering out reply threads (only top-level comments are returned).
func fetchReviewComments(owner, repo string, prNumber int) ([]ReviewComment, error) {
	endpoint := fmt.Sprintf("repos/%s/%s/pulls/%d/comments", owner, repo, prNumber)
	out, err := ghRun("api", "--paginate", endpoint)
	if err != nil {
		return nil, err
	}

	var comments []ReviewComment
	if err := json.Unmarshal(out, &comments); err != nil {
		return nil, fmt.Errorf("parsing review comments: %w", err)
	}

	// Only surface top-level comments (not replies to existing threads).
	var toplevel []ReviewComment
	for _, c := range comments {
		if c.InReplyToID == 0 {
			toplevel = append(toplevel, c)
		}
	}
	return toplevel, nil
}

// FetchFileContent returns the current on-disk content of the given file path.
func FetchFileContent(path string) (string, error) {
	out, err := exec.Command("cat", path).Output()
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", path, err)
	}
	return string(out), nil
}

// resolveRepo returns owner and repo name. If repo string is provided as
// "owner/repo" it is split directly. Otherwise `gh repo view` is called to
// auto-detect from the current directory.
func resolveRepo(repo string) (owner, name string, err error) {
	if repo != "" {
		parts := strings.SplitN(repo, "/", 2)
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
		return "", "", fmt.Errorf("invalid --repo format, expected owner/repo, got: %s", repo)
	}

	out, err := ghRun("repo", "view", "--json", "nameWithOwner")
	if err != nil {
		return "", "", err
	}

	var v struct {
		NameWithOwner string `json:"nameWithOwner"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		return "", "", fmt.Errorf("parsing gh repo view: %w", err)
	}

	parts := strings.SplitN(v.NameWithOwner, "/", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unexpected nameWithOwner format: %s", v.NameWithOwner)
	}
	return parts[0], parts[1], nil
}

// ghRun runs a gh CLI subcommand, ensuring GH_TOKEN is always set from
// GITHUB_TOKEN so that private repositories are accessible. Returns a
// human-friendly error with a token-scope hint when GitHub returns a
// "Could not resolve" error (the classic symptom of insufficient token scope).
func ghRun(args ...string) ([]byte, error) {
	cmd := exec.Command("gh", args...)
	cmd.Env = ghEnv()

	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		return nil, formatGhError(args, err, stderr)
	}
	return out, nil
}

// ghEnv returns os.Environ() with GH_TOKEN set to PR_REVIEW_TOKEN (preferred)
// or GITHUB_TOKEN (fallback). The gh CLI uses GH_TOKEN for authentication.
// PR_REVIEW_TOKEN should be a token with full 'repo' scope for private repo access.
// GITHUB_TOKEN is typically scoped to Copilot only and may not access private repos.
func ghEnv() []string {
	token := os.Getenv("PR_REVIEW_TOKEN")
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}

	env := os.Environ()
	if token == "" {
		return env
	}

	// Overwrite any existing GH_TOKEN with our chosen token.
	out := make([]string, 0, len(env)+1)
	for _, e := range env {
		if !strings.HasPrefix(e, "GH_TOKEN=") {
			out = append(out, e)
		}
	}
	return append(out, "GH_TOKEN="+token)
}

// formatGhError produces a human-friendly error message, adding a token-scope
// hint when the error looks like a private repository access problem.
func formatGhError(args []string, err error, stderr string) error {
	base := fmt.Errorf("gh %s failed: %w\n%s", strings.Join(args, " "), err, stderr)

	if strings.Contains(stderr, "Could not resolve to a Repository") ||
		strings.Contains(stderr, "Could not resolve to a Repository with the name") {
		return fmt.Errorf("%w\n\nHint: this usually means your token cannot access the repository.\n"+
			"For private repos, GITHUB_TOKEN must have the 'repo' scope (classic token)\n"+
			"or 'Contents: read' + 'Pull requests: read' permissions (fine-grained token).\n"+
			"Check your token scopes with: gh auth status", base)
	}

	return base
}


