FROM debian:bookworm-slim

# Arguments for version control
ARG JOSS_VERSION="latest"
ARG REPO_OWNER="josprox"
ARG REPO_NAME="JosSecurity-language"

# Install dependencies
RUN apt-get update && apt-get install -y \
    curl \
    unzip \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Install JosSecurity
RUN if [ "$JOSS_VERSION" = "latest" ]; then \
        DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/latest/download/jossecurity-linux.zip"; \
    else \
        DOWNLOAD_URL="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${JOSS_VERSION}/jossecurity-linux.zip"; \
    fi \
    && curl -fsSL "$DOWNLOAD_URL" -o /tmp/jossecurity.zip \
    && unzip /tmp/jossecurity.zip -d /tmp/ \
    && ARCH=$(uname -m) \
    && if [ "$ARCH" = "x86_64" ]; then BINARY="joss-linux-amd64"; \
       elif [ "$ARCH" = "aarch64" ]; then BINARY="joss-linux-arm64"; \
       elif [ "$ARCH" = "armv7l" ]; then BINARY="joss-linux-armv7"; \
       else echo "Unsupported architecture: $ARCH"; exit 1; fi \
    && find /tmp -name "$BINARY" -exec mv {} /usr/local/bin/joss \; \
    && chmod +x /usr/local/bin/joss \
    && rm -rf /tmp/*

# Copy entrypoint script
COPY docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Setup working directory
WORKDIR /app

# Expose default port
EXPOSE 8000

# Entrypoint
ENTRYPOINT ["docker-entrypoint.sh"]

# Default command
CMD ["joss", "server", "start"]
