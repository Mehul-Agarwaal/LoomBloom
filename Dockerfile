# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install make
RUN apk add --no-cache make

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Generate templ files and build
RUN go run github.com/a-h/templ/cmd/templ generate
RUN go build -o /app/bin/loombloom ./cmd/server

# Run stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/bin/loombloom .

# Set default port
ENV PORT=8080
EXPOSE 8080

# Run the binary
CMD ["./loombloom"]
