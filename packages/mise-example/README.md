# mise example

A minimal project for testing the `agent:mise` feature image.

## What it does

`.mise.toml` pins two runtimes:

| Tool   | Version |
|--------|---------|
| Node   | 22      |
| Python | 3.12    |

When you open a shell in this directory, mise automatically activates those versions.

## How to test

### 1. Build the images

```shell
# From the repo root
docker build --tag agent --file packages/container/Dockerfile packages/container
docker build --tag agent:mise --file packages/container/Dockerfile.mise packages/container
```

### 2. Run the sandbox pointing at this directory

```shell
sandbox run --image agent:mise packages/mise-example
```

Or with the raw container command:

```shell
docker run --rm -it \
  -v $(pwd)/packages/mise-example:/home/agent/workspace \
  agent:mise bash
```

### 3. Inside the container, verify tools are ready

```shell
bash check.sh
```

`mise install` runs automatically on container startup, so the tools are already installed.

Expected output:

```
=== mise version ===
mise 2025.x.x

=== active tools ===
node    22.x.x
python  3.12.x

=== node ===
v22.x.x

=== python ===
Python 3.12.x

All checks passed!
```
