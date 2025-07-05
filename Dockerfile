# Build stage
FROM golang:1.24.2-bookworm AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends gcc build-essential musl-dev libsqlite3-dev && rm -rf /var/lib/apt/lists/*

# Set working directory
WORKDIR /app

# Build Delve
RUN go install github.com/go-delve/delve/cmd/dlv@v1.25.0

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
ENV CGO_ENABLED=1
RUN go build -o bin/threadmirror ./cmd/*.go

# Runtime stage
FROM chromedp/headless-shell:137.0.7106.2

# Install runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*

# Create non-root user
RUN groupadd -g 1001 appgroup && \
    useradd -u 1001 -g appgroup -s /usr/sbin/nologin -m appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /go/bin/dlv .
COPY --from=builder /app/bin/threadmirror .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080
