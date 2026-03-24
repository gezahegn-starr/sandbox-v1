# GitHub Copilot CLI Container Sandbox

## Build the container:
```shell
container build --tag agent --file Dockerfile .
```

## Run the container:
```shell
container run -e GITHUB_TOKEN=${GITHUB_TOKEN} -v ~/path/to/project:/home/agent/workspace -it agent
```

## Go CLI (`packages/go`)

A CLI tool that manages the lifecycle of a Copilot agent container for a given project directory. It will:

1. **Reuse** an existing container (named `copilot-<project>`) if one already exists, restarting it as needed.
2. **Create** a new container if none exists — mounting the project directory and injecting the `GITHUB_TOKEN`.
3. **Configure** the container by writing a Copilot config file (`config.json`) with the project's trusted folder paths.
4. **Attach** an interactive session to the running container.

### Build
```shell
cd packages/go
go build -o sandbox .
```

### Usage
```shell
./sandbox /path/to/your/project
```

### Install globally (run from anywhere)

1. Build the binary to `~/go/bin`:
```shell
cd packages/go
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
sandbox /path/to/your/project
```

The tool expects the `GITHUB_TOKEN` environment variable to be set and the `agent` container image to be built (see above).
