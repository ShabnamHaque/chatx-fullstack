# Use official Golang image for building
FROM golang:1.21 AS builder

WORKDIR /app

# Copy source code
COPY . .

# Download dependencies
RUN go mod tidy

# Build the Go binary with static linking
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server .

# Use a lightweight image for deployment
FROM alpine:latest

WORKDIR /root/

# Copy built binary from the builder
COPY --from=builder /app/server .
COPY --from=builder /app/.env .  
# Copy the .env file
# Give execution permission (just in case)
RUN chmod +x server

EXPOSE 8080

CMD ["./server"]
