# Build stage - needs CGO for go-sqlite3 (required by whatsmeow)
FROM golang:1.24 AS builder

# Install build dependencies
RUN apt-get update && apt-get install -y gcc libc6-dev git && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go.mod files first for better caching
COPY bridge/go.mod bridge/go.sum* ./bridge/
COPY pkg/go.mod ./pkg/

# Download dependencies
WORKDIR /app/bridge
RUN go mod download

# Copy source code
WORKDIR /app
COPY bridge/ ./bridge/
COPY pkg/ ./pkg/

# Build with CGO enabled (required for mattn/go-sqlite3)
# i18n files are embedded via go:embed
WORKDIR /app/bridge
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-w -s" -o /bridge ./cmd/bridge

# Runtime stage
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY --from=builder /bridge .

# Create data directory
RUN mkdir -p /data

CMD ["./bridge"]
