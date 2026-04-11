FROM agent

# Install the GitHub CLI (gh) — not included in the base agent image.
# gh is needed for: gh pr view, gh api (PR comments), gh repo view, gh copilot.
USER root
RUN curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg \
    && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
    && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list \
    && apt-get update \
    && apt-get install -y --no-install-recommends gh \
    && rm -rf /var/lib/apt/lists/*

USER agent

# Configure git to use GITHUB_TOKEN as a credential helper so that
# `git push` works for both public and private repositories at runtime.
# The helper reads GITHUB_TOKEN from the environment each time git needs auth.
RUN git config --global credential.helper \
      '!f() { echo username=x-access-token; echo password=${GITHUB_TOKEN}; }; f' \
    && git config --global credential.useHttpPath true

# Copy source and build the pr-review binary.
COPY --chown=agent:agent . /build/pr-review
WORKDIR /build/pr-review

RUN go mod download && \
    go build -o /usr/local/bin/pr-review . && \
    rm -rf /build

WORKDIR /home/agent/workspace

# The entrypoint and default user come from the base agent image.
