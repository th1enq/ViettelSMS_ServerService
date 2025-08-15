# Build stage
FROM golang:1.24.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Make entrypoint script executable
RUN chmod +x scripts/entrypoint.sh

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main cmd/server/main.go

# Build migrate binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o migrate cmd/migrate/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

# Copy config file
COPY --from=builder /app/config.dev.yaml .

# Copy migration files (if needed for migrate command)
COPY --from=builder /app/migrations/ ./migrations/

# Copy entrypoint script
COPY --from=builder /app/scripts/entrypoint.sh .
RUN chmod +x entrypoint.sh

# Create logs directory and set permissions
RUN mkdir -p logs && chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port (from config.yaml, user_service port is 8082)
EXPOSE 8081

# Health check - using a simple port check since grpc_health_probe might not be available
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD nc -z localhost 8081 || exit 1

# Run the entrypoint script which handles migrations and starts the app
CMD ["./entrypoint.sh"] 