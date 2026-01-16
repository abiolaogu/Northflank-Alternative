# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=${VERSION:-dev} -X main.commit=${COMMIT:-unknown} -X main.buildDate=${BUILD_DATE:-unknown}" \
    -o /app/orchestrator \
    ./cmd/orchestrator

# Final stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/orchestrator /app/orchestrator

# Create non-root user
RUN addgroup -g 1000 openpaas && \
    adduser -u 1000 -G openpaas -s /bin/sh -D openpaas && \
    chown -R openpaas:openpaas /app

USER openpaas

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health/live || exit 1

ENTRYPOINT ["/app/orchestrator"]
