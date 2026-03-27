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

# Merge host ~/.copilot/skills into the container's ~/.copilot/skills directory.
# Host skills are mounted read-only at /home/agent/.copilot-host-skills.
# Existing container skills are preserved; host skills fill in any gaps.
setup_copilot_config() {
    local config="/home/agent/.copilot/config.json"
    [ -f "$config" ] && return 0
    [ -z "$WORKSPACE_PATH" ] && return 0
    mkdir -p /home/agent/.copilot
    cat > "$config" << EOF
{
  "banner": "never",
  "trusted_folders": ["/home/agent/workspace", "$WORKSPACE_PATH"]
}
EOF
}

merge_host_skills() {
    local host_dir="/home/agent/.copilot-host-skills"
    local dest_dir="/home/agent/.copilot/skills"

    [ -d "$host_dir" ] || return 0

    mkdir -p "$dest_dir"

    # Copy host skill directories without overwriting existing container skills
    find -L "$host_dir" -mindepth 1 -maxdepth 1 -type d | while read -r skill_dir; do
        local skill_name
        skill_name=$(basename "$skill_dir")
        if [ ! -d "$dest_dir/$skill_name" ]; then
            cp -rL "$skill_dir" "$dest_dir/$skill_name"
        fi
    done
}

YOLO_FLAGS="--allow-all-tools"

main() {
    setup_github_token
    setup_workspace
    setup_copilot_config
    merge_host_skills

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
