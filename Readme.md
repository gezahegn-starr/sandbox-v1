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

```shell
# Build the mise feature image
docker build --tag agent:mise --file packages/container/Dockerfile.mise packages/container
```

Adding a new feature is as simple as creating a new `Dockerfile.<feature>` in `packages/container/` that starts with `FROM agent`.

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
