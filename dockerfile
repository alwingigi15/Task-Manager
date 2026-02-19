# Stage 1: Build the Go app
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files and download dependencies (for caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the binary (disable CGO for static binary, target Linux)
RUN CGO_ENABLED=0 GOOS=linux go build -o task-manager ./main.go  # Adjust if main is in cmd/

# Stage 2: Create a lightweight runtime image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/task-manager .

# Copy your .env file (for runtime configs like DB_HOST, PORT=8019)
COPY .env .

# Expose the port from your .env (8019)
EXPOSE 8019

# Run the app
CMD ["./task-manager"]