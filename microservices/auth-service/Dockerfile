# Build stage
FROM golang:1.24.3-alpine AS builder
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o auth-service ./cmd

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

# Copy binary and config
COPY --from=builder /app/auth-service .
COPY --from=builder /app/config ./config/
COPY --from=builder /app/migrations ./migrations/

EXPOSE 50051
CMD ["./auth-service"]