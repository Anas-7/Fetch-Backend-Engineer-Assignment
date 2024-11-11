# Stage 1: Build the Go application
FROM golang:1.23.3 AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Set environment variables for static binary compilation
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the Go application
RUN go build -o app .

# Stage 2: Create a lightweight runtime container
FROM alpine:latest

# Install necessary certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set the working directory inside the container
WORKDIR /root/

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/app .

# Expose port 9090 for the application
EXPOSE 9090

# Run the application
CMD ["./app"]
