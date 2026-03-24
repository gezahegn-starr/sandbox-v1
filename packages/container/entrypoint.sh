#!/bin/bash
# Entrypoint script for GitHub Copilot CLI Sandbox

set -e

setup_github_token() {
    if [ -n "$GITHUB_TOKEN" ]; then
        export GH_TOKEN="$GITHUB_TOKEN"
    fi
}

setup_workspace() {
    if [ -n "$WORKSPACE_PATH" ] && [ -d "$WORKSPACE_PATH" ]; then
        cd "$WORKSPACE_PATH"
    else
        cd /home/agent/workspace
    fi
}

YOLO_FLAGS="--allow-all-tools"

main() {
    setup_github_token
    setup_workspace

    local sandbox_name
    sandbox_name=$(basename "$(pwd)")
    echo "Starting copilot agent in sandbox '${sandbox_name}'..."
    echo "Workspace: $(pwd)"

    if [ "$1" = "copilot" ]; then
        shift
        exec copilot $YOLO_FLAGS "$@"
    fi

    if [ $# -gt 0 ]; then
        exec copilot $YOLO_FLAGS --prompt "$*"
    fi

    exec copilot $YOLO_FLAGS
}

# Run main with all arguments
main "$@"
