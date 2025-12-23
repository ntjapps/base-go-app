# Multi-stage Dockerfile for the Go worker

FROM golang:alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /src

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o /worker ./cmd/worker

# Final image
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=builder /worker /worker

# Drop privileges by creating a non-root user (optional)
RUN addgroup -S worker && adduser -S worker -G worker
USER worker

ENTRYPOINT ["/worker"]
