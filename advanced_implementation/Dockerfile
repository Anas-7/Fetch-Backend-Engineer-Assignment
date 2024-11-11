FROM golang:1.23.3 AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# Copy the entire directory
COPY . .

# Set the CGO_ENABLED=0 to build a static binary otherwise it doesn't work
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64

# Build the message_queue binary
WORKDIR /app/message_queue
RUN go build -o message_queue

# Build the server binary
WORKDIR /app/server
RUN go build -o server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/server
COPY --from=builder /app/server/server .

WORKDIR /root/message_queue
COPY --from=builder /app/message_queue/message_queue .

# 9090 is for the server that we will ping using Postman and 15672 is for the RabbitMQ management console
EXPOSE 9090
EXPOSE 15672

CMD ["sh"]
