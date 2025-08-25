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
RUN go build -o main cmd/server/main.go

# Build migrate binary
RUN go build -o migrate cmd/migrate/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .
COPY --from=builder /app/migrate .

# Copy migration files (if needed for migrate command)
COPY --from=builder /app/migrations/ ./migrations/

# Copy entrypoint script
COPY --from=builder /app/scripts/entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expose port (from config.yaml, user_service port is 8082)
EXPOSE 8080
# Run the entrypoint script which handles migrations and starts the app
CMD ["./entrypoint.sh"] 