# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files from workers directory
COPY go.mod go.sum ./
RUN go mod download

# Copy common and indexer code
COPY common/ ./common/
COPY indexer/ ./indexer/

# Build the indexer worker
WORKDIR /app/indexer
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o worker-indexer .

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/indexer/worker-indexer .

CMD ["./worker-indexer"]
