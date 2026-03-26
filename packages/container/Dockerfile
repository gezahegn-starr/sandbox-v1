FROM ubuntu:24.04

# Copilot CLI version (passed via --build-arg or defaults to latest)
ARG COPILOT_VERSION=latest

ARG GITHUB_TOKEN
ARG WORKSPACE_PATH

ENV GITHUB_TOKEN=${GITHUB_TOKEN}
ENV WORKSPACE_PATH=${WORKSPACE_PATH}

LABEL org.opencontainers.image.version="${COPILOT_VERSION}"
LABEL com.docker.sandboxes="templates"
LABEL com.docker.sandboxes.base="ubuntu:24.04"
LABEL com.docker.sandboxes.flavor="copilot-cli"

# Prevent interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=UTC

# Install system dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    # Essential tools
    ca-certificates \
    curl \
    wget \
    # Development tools
    git \
    build-essential \
    # Shell utilities
    bash \
    ripgrep \
    tree \
    vim-tiny \
    # Process management
    sudo \
    procps \
    # Python (for various tools and MCP servers)
    python3 \
    python3-pip \
    python3-venv \
    && rm -rf /var/lib/apt/lists/*

# Install Node.js 24 - Required for Copilot CLI (v24+ is mandatory)
RUN curl -fsSL https://deb.nodesource.com/setup_24.x | bash - \
    && apt-get install -y --no-install-recommends nodejs \
    && rm -rf /var/lib/apt/lists/*

# Create non-root user 'agent' with sudo access (matching Docker sandbox conventions)
RUN useradd -m -s /bin/bash -G sudo agent \
    && echo "agent ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers.d/agent \
    && chmod 0440 /etc/sudoers.d/agent

# Install Go
RUN curl -fsSL "https://go.dev/dl/go1.26.0.linux-arm64.tar.gz" -o /tmp/go.tar.gz \
    && tar -C /usr/local -xzf /tmp/go.tar.gz \
    && rm /tmp/go.tar.gz
ENV PATH="/usr/local/go/bin:$PATH"

# Install GitHub Copilot CLI globally
RUN npm install -g @github/copilot@${COPILOT_VERSION}

# Set up user environment
USER agent

# RUN mkdir -p /home/agent/workspace

# Create directories for Copilot CLI configuration and state
RUN mkdir -p /home/agent/.config/copilot \
    && mkdir -p /home/agent/.copilot \
    && mkdir -p /home/agent/.local/state/copilot \
    && mkdir -p /home/agent/.local/share/copilot

# Set XDG directories for Copilot CLI
ENV XDG_CONFIG_HOME=/home/agent/.config
ENV XDG_STATE_HOME=/home/agent/.local/state
ENV XDG_DATA_HOME=/home/agent/.local/share

# Add npm global bin to PATH for user-level installs
ENV NPM_CONFIG_PREFIX=/home/agent/.npm-global
ENV PATH="/home/agent/.npm-global/bin:$PATH"

# Configure shell
COPY --chown=agent:agent .bashrc /home/agent/.bashrc

# Set up entrypoint
COPY --chown=agent:agent entrypoint.sh /home/agent/entrypoint.sh
RUN chmod +x /home/agent/entrypoint.sh

# Working directory for mounted workspaces
WORKDIR /home/agent/workspace

ENTRYPOINT ["/home/agent/entrypoint.sh"]
CMD []
