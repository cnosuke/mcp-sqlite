FROM golang:1.24 AS builder

WORKDIR /app

# Install git and SQLite development libraries for CGO
RUN apt-get update && apt-get install -y \
    git \
    make \
    gcc \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files first for better layer caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN make build-for-linux-amd64

RUN mkdir -p /app/logs

# Final stage - using Distroless base image with SQLite support
FROM gcr.io/distroless/base-debian12

# Set metadata
LABEL org.opencontainers.image.source="https://github.com/cnosuke/mcp-sqlite"
LABEL org.opencontainers.image.description="MCP server for SQLite functionality"

WORKDIR /app

# Create logs directory and make it writable by nonroot user
COPY --from=builder /app/config.yml /app/config.yml

# Create a logs directory that's writable by nonroot
COPY --from=builder --chown=nonroot:nonroot /app/bin/mcp-sqlite-linux-amd64 /app/mcp-sqlite

# Copy logs directory from builder stage
# Because the distroless image doesn't have a shell, we need to ensure the directory exists
COPY --from=builder --chown=nonroot:nonroot /app/logs /app/logs
ENV LOG_PATH=/app/logs/mcp-sqlite.log

# Distroless image uses nonroot user by default (uid=65532)
USER nonroot:nonroot

# Set the entrypoint
ENTRYPOINT ["/app/mcp-sqlite"]

# Default command
CMD ["server", "--config", "config.yml"]
