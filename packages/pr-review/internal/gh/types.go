package gh

// PRMeta holds high-level metadata about a pull request.
type PRMeta struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	HeadRefName string `json:"headRefName"`
	BaseRefName string `json:"baseRefName"`
	URL         string `json:"url"`
	RepoOwner   string // populated separately
	RepoName    string // populated separately
}

// ReviewComment is a single inline review comment on a PR, as returned by the
// GitHub REST API via `gh api /repos/{owner}/{repo}/pulls/{number}/comments`.
// Field names match the REST API's snake_case JSON.
type ReviewComment struct {
	ID          int64  `json:"id"`
	Body        string `json:"body"`
	Path        string `json:"path"`
	Line        int    `json:"line"`       // end line (may be 0 for deleted lines)
	StartLine   int    `json:"start_line"` // first line of a multi-line comment
	DiffHunk    string `json:"diff_hunk"`
	URL         string `json:"html_url"`
	User        Author `json:"user"`
	InReplyToID int64  `json:"in_reply_to_id"`
}

// Review is a top-level PR review (APPROVED, CHANGES_REQUESTED, COMMENTED),
// as returned by `gh pr view --json reviews`.
type Review struct {
	ID    string `json:"id"`
	Body  string `json:"body"`
	State string `json:"state"`
	URL   string `json:"url"`
}

// Author is a GitHub user reference.
type Author struct {
	Login string `json:"login"`
}

// prViewPayload is the raw JSON shape returned by `gh pr view --json …`.
// Note: inline review comments are fetched separately via the REST API.
type prViewPayload struct {
	Number      int    `json:"number"`
	Title       string `json:"title"`
	URL         string `json:"url"`
	HeadRefName string `json:"headRefName"`
	BaseRefName string `json:"baseRefName"`
}
