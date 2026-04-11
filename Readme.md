# GitHub Copilot CLI Container Sandbox

## Build the container

### Base image
```shell
docker build --tag agent --file packages/container/Dockerfile packages/container
```

### Feature images

Feature Dockerfiles extend the base image with additional tooling. Build the base first, then any feature image you need.

| Dockerfile | Tag | What it adds |
|---|---|---|
| `Dockerfile.mise` | `agent:mise` | [mise](https://mise.jdx.dev) polyglot runtime manager |
| `packages/pr-review/Dockerfile` | `agent:pr-review` | Automated PR review responder (`pr-review` CLI) |

```shell
# Build the mise feature image
docker build --tag agent:mise --file packages/container/Dockerfile.mise packages/container

# Build the pr-review feature image
docker build --tag agent:pr-review --file packages/pr-review/Dockerfile packages/pr-review
```

Adding a new feature is as simple as creating a new `Dockerfile.<feature>` in `packages/container/` that starts with `FROM agent`.

## `pr-review` CLI (`packages/pr-review`)

A CLI tool that automatically responds to pull request review comments using GitHub Copilot. It will:

1. **Fetch** all open review comments on a PR via the `gh` CLI
2. **Classify** each comment with Copilot — auto-fixable code change vs. needs human judgment
3. **Fix** actionable comments by invoking Copilot in headless mode and applying the changes to disk
4. **Validate** each fix by running the repo's existing test/lint commands
5. **Push** successful fixes to the PR branch
6. **Report** any comments requiring human review in a markdown file with GitHub links

### Build

```shell
cd packages/pr-review
go build -o pr-review .
```

### Usage

Run inside a container (recommended, using the `agent:pr-review` image):

```shell
sandbox run --image agent:pr-review /path/to/your/project
# then inside the container:
pr-review run
```

Or run directly (requires `gh` CLI authenticated and Copilot CLI available):

```shell
pr-review run --repo owner/repo --pr 42
```

### Flags

| Flag | Default | Description |
|---|---|---|
| `--repo` | auto-detected | Repository in `owner/repo` format |
| `--pr` | auto-detected | PR number (uses current branch if omitted) |
| `--dry-run` | false | Classify and plan without applying changes or pushing |
| `--output` | `pr-review-report.md` | Path for the human-review markdown report |
| `--debug` | false | Verbose debug logging |

### Output

- Changes that pass validation are committed and pushed automatically.
- A `pr-review-report.md` file is written listing every comment that needs human attention, with direct GitHub links and context.
- Exit code `0` = all comments resolved; exit code `1` = some comments need human review.

## Run the container

Use the `sandbox` CLI (see below) to create and run sandboxes. To use a feature image, pass `--image`:

```shell
# Default (agent image)
sandbox run /path/to/your/project

# With a feature image
sandbox run --image agent:mise /path/to/your/project
```

## `sandbox` CLI (`packages/sandbox-cli`)

A CLI tool that manages the lifecycle of a Copilot agent container for a given project directory. It will:

1. **Reuse** an existing container (named `copilot-<project>`) if one already exists, restarting it as needed.
2. **Create** a new container if none exists — mounting the project directory and injecting the `GITHUB_TOKEN`.
3. **Configure** the container by writing a Copilot config file (`config.json`) with the project's trusted folder paths.
4. **Attach** an interactive session to the running container.

### Build
```shell
cd packages/sandbox-cli
go build -o sandbox .
```

### Usage
```shell
./sandbox run /path/to/your/project
```

### Install globally (run from anywhere)

1. Build the binary to `~/go/bin`:
```shell
cd packages/sandbox-cli
go build -o ~/go/bin/sandbox .
```

2. Add `~/go/bin` to your `$PATH` (one-time setup). Add this line to `~/.zshrc`:
```shell
export PATH="$PATH:$HOME/go/bin"
```

3. Reload your shell:
```shell
source ~/.zshrc
```

4. Now you can run `sandbox` from anywhere:
```shell
sandbox run /path/to/your/project
```

The tool expects the `GITHUB_TOKEN` environment variable to be set and the `agent` container image to be built (see above).
