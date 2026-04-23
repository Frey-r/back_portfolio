# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies for go-sqlite3 (requires CGO)
RUN apk add --no-cache gcc musl-dev

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled (required for sqlite)
RUN CGO_ENABLED=1 GOOS=linux go build -o /app/portfolio-binary ./cmd/server

# Run stage
FROM alpine:3.19

WORKDIR /app

# Install Certificates
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/portfolio-binary /app/server
COPY --from=builder /app/.env.example /app/.env

# Create data directory for SQLite
RUN mkdir -p /app/data

EXPOSE 8082

CMD ["/app/server"]
