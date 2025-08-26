# Step 1: Build stage
FROM golang:1.22 AS builder

# Set workdir
WORKDIR /app

# Copy go.mod and go.sum first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy rest of the code
COPY . .

# Build the Go binary
RUN go build -o backend main.go

# Step 2: Run stage (lightweight image)
FROM debian:bookworm-slim

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/backend .

# App runs on 5000
EXPOSE 5000

CMD ["./backend"]
