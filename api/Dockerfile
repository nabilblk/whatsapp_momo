# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go.mod files first for better caching
COPY api/go.mod api/go.sum* ./api/
COPY pkg/go.mod ./pkg/

# Download dependencies
WORKDIR /app/api
RUN go mod download

# Copy source code
WORKDIR /app
COPY api/ ./api/
COPY pkg/ ./pkg/

# Build (i18n files are embedded via go:embed)
WORKDIR /app/api
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /api ./cmd/api

# Runtime stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata wget

WORKDIR /app

COPY --from=builder /api .

# Create data directory
RUN mkdir -p /data

EXPOSE 8080

CMD ["./api"]
