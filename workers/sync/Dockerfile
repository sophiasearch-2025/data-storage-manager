# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files from workers directory
COPY go.mod go.sum ./
RUN go mod download

# Copy common and sync code
COPY common/ ./common/
COPY sync/ ./sync/

# Build the sync worker
WORKDIR /app/sync
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker-sync .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/sync/worker-sync .

CMD ["./worker-sync"]
