# Build stage
FROM golang:1.24.1-alpine AS builder

# Install only the necessary build dependencies
RUN apk add --no-cache build-base cmake gcc git

WORKDIR /app
COPY go.mod go.sum ./
# Download dependencies first (better layer caching)
RUN go mod download

# Copy source code
COPY . .

# Install goose
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

# Build with optimization flags
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o layerg-crawler
RUN chmod +x layerg-crawler

# Final stage - use scratch for smallest possible image
FROM alpine:3.18.0

# Only install essential certificates
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy only the built binary and config file from builder
COPY --from=builder /app/layerg-crawler /app/
COPY --from=builder /go/bin/goose /app/
COPY .layerg-crawler.yaml /app/.layerg-crawler.yaml

ENTRYPOINT ["./layerg-crawler"]
