# Build stage
FROM golang:1.24.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make gcc musl-dev sqlite-dev

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
RUN make build

# Runtime stage
FROM alpine:3.22.0

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

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

# Run the application
CMD ["./threadmirror", "server"] 
# CMD ["./dlv", "exec", "--listen=:40000", "--headless=true", "--api-version=2", "--accept-multiclient", "./threadmirror", "server"]
