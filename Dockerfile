# Build stage
FROM golang:1.23 AS builder

WORKDIR /app
COPY . .

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o homework-object-storage

# Final stage
FROM alpine:latest

# Install necessary tools
RUN apk add --no-cache bash curl

# Copy the statically built Go binary
COPY --from=builder /app/homework-object-storage /usr/local/bin/homework-object-storage

# Command to run your application
ENTRYPOINT ["/usr/local/bin/homework-object-storage"]
