# Start from official Golang image
FROM golang:1.22-alpine

WORKDIR /app

# Copy go.mod and go.sum first (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the Go binary
RUN go build -o backend main.go

# Expose port
EXPOSE 5000

# Run the service
CMD ["./backend"]
