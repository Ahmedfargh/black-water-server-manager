# Stage 1: Build
FROM golang:1.25 AS builder

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o server-manager main.go

# Stage 2: Final Image
FROM ubuntu:22.04

# Avoid interactive prompts during build
ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies for server management
RUN apt-get update && apt-get install -y --no-install-recommends \
    polkitd \
    iproute2 \
    procps \
    sudo \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server-manager .
COPY --from=builder /app/.env.example .env

# Expose the application port (default :8080)
EXPOSE 8080

# Run the application
CMD ["./server-manager"]
