# Jenkins MCP Server Dockerfile
# Multi-stage build for optimized image size

# Stage 1: Build the application
FROM golang:1.23.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
# CGO_ENABLED=0 for static binary
# -ldflags="-w -s" to reduce binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o jenkins-mcp-server \
    ./

# Stage 2: Create minimal runtime image
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 mcp && \
    adduser -D -u 1000 -G mcp mcp

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/jenkins-mcp-server /app/jenkins-mcp-server

# Change ownership to non-root user
RUN chown -R mcp:mcp /app

# Switch to non-root user
USER mcp

# Set environment variables with defaults
ENV JENKINS_URL="" \
    JENKINS_API_TOKEN="" \
    JENKINS_USERNAME="" \
    JENKINS_PASSWORD="" \
    JENKINS_TIMEOUT="30s" \
    JENKINS_TLS_SKIP_VERIFY="false" \
    JENKINS_CA_CERT="" \
    JENKINS_MAX_RETRIES="3" \
    JENKINS_RETRY_BACKOFF="1s"

# Health check (optional - checks if binary exists and is executable)
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD [ -x /app/jenkins-mcp-server ] || exit 1

# The server communicates via stdio, so no ports to expose
# EXPOSE is not needed for MCP servers

# Run the application
ENTRYPOINT ["/app/jenkins-mcp-server"]

# Optional: Allow passing config file path as argument
# Usage: docker run ... jenkins-mcp-server --config /app/config.yaml
CMD []
