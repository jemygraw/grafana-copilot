# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git make build-base

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN GOOS=linux go build -o grafana-copilot main.go

# Runtime stage
FROM alpine:3.22

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata

# Create data directory
RUN mkdir -p /data

# Copy binary from builder
COPY --from=builder /app/grafana-copilot .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./grafana-copilot"]