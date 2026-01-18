# Multi-stage Dockerfile for Job Scheduler
# Supports all three components: API Server, Scheduler, Consumer

# ============================================================================
# Build Stage
# ============================================================================
FROM golang:1.24-alpine AS builder

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

# Build binary
# CGO_ENABLED=0 creates a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o job-scheduler .

# ============================================================================
# Runtime Stage
# ============================================================================
FROM alpine:latest

# Install CA certificates for TLS connections (PostgreSQL, Kafka)
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 jobscheduler && \
    adduser -D -u 1000 -G jobscheduler jobscheduler

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/job-scheduler .

# Change ownership to non-root user
RUN chown -R jobscheduler:jobscheduler /app

# Switch to non-root user
USER jobscheduler

# Expose port for API server
EXPOSE 8000

# Default command (can be overridden)
# Supports: api-server, scheduler first-time, scheduler retry, consumer
CMD ["./job-scheduler", "api-server"]
