FROM golang:1.23.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

COPY . .

# Set environment variables for static binary compilation
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the Go application
RUN go build -o app .

FROM alpine:latest

# Install necessary certificates for HTTPS
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/app .

# Expose port 9090 for the application
EXPOSE 9090

# Run the application
CMD ["./app"]
