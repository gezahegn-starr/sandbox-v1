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
LOG_FILE="${COPILOT_LOG_FILE:-/home/agent/.copilot/copilot-error.log}"
LOG_LEVEL="${COPILOT_LOG_LEVEL:-error}"
COPILOT_LOG_DIR="/home/agent/.copilot/logs"

# Run copilot, capture exit code on failure.
# On non-zero exit, logs diagnostic info (memory, env, copilot debug logs) to LOG_FILE.
run_copilot() {
    local exit_code=0
    copilot --log-level "$LOG_LEVEL" --log-dir "$COPILOT_LOG_DIR" "$@" || exit_code=$?
    if [ "$exit_code" -ne 0 ]; then
        {
            echo "========================================"
            echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] copilot exited with code $exit_code"
            echo "  args: $*"
            echo "  log_level: $LOG_LEVEL"
            echo "  cwd: $(pwd)"
            echo "  copilot version: $(copilot --version 2>/dev/null || echo 'unknown')"
            echo "  GH_TOKEN set: $([ -n "$GH_TOKEN" ] && echo 'yes' || echo 'no')"
            echo "  GITHUB_TOKEN set: $([ -n "$GITHUB_TOKEN" ] && echo 'yes' || echo 'no')"
            echo "  WORKSPACE_PATH: ${WORKSPACE_PATH:-<unset>}"
            echo "  node version: $(node --version 2>/dev/null || echo 'not found')"
            echo ""
            echo "=== Memory usage ==="
            free -m 2>/dev/null || echo "free command not available"
            echo ""
            echo "=== Exit code analysis ==="
            if [ "$exit_code" -eq 137 ]; then
                echo "  ** Exit 137 = SIGKILL (likely OOM killer) **"
                echo "  Check container memory limit — copilot needs at least 2-4 GB"
                dmesg 2>/dev/null | grep -i "oom\|killed\|out of memory" | tail -5 || true
            elif [ "$exit_code" -eq 1 ]; then
                echo "  ** Exit 1 = general error (check debug logs below) **"
            fi
            echo "========================================"
            # Dump copilot's own debug log files
            if [ -d "$COPILOT_LOG_DIR" ]; then
                echo ""
                echo "=== Copilot debug logs ($COPILOT_LOG_DIR) ==="
                for logfile in "$COPILOT_LOG_DIR"/*.log; do
                    [ -f "$logfile" ] || continue
                    echo "--- $(basename "$logfile") ---"
                    tail -100 "$logfile"
                    echo ""
                done
            else
                echo "No copilot log directory found at $COPILOT_LOG_DIR"
            fi
        } >> "$LOG_FILE"
    fi
    return "$exit_code"
}

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
        run_copilot $YOLO_FLAGS "$@"
        return
    fi

    if [ $# -gt 0 ]; then
        run_copilot $YOLO_FLAGS --prompt "$*"
        return
    fi

    run_copilot $YOLO_FLAGS
}

# Run main with all arguments
main "$@"
