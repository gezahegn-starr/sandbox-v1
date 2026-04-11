#!/bin/bash
# Runs `mise install` if the workspace has a .mise.toml or .tool-versions,
# then hands off to the base entrypoint.

set -e

WORKSPACE="${WORKSPACE_PATH:-/home/agent/workspace}"

if [ -f "$WORKSPACE/.mise.toml" ] || [ -f "$WORKSPACE/.tool-versions" ]; then
    echo "mise: installing tools globally from $WORKSPACE..."
    cd "$WORKSPACE"
    mise trust "$WORKSPACE" 2>/dev/null || true
    mise install
    mise ls --current 2>/dev/null | awk '{print $1"@"$2}' | xargs -r mise use -g
fi

exec /home/agent/entrypoint.sh "$@"
