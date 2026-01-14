# Build stage
FROM golang:1.24-alpine AS builder

# Install necessary build tools
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /build

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o server \
    main.go

# Runtime stage
FROM alpine:3.19

# Install ca-certificates and timezone data
RUN apk add --no-cache ca-certificates tzdata

# Set timezone to Asia/Shanghai (adjust as needed)
ENV TZ=Asia/Shanghai

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/server .

# Copy config file
COPY config.yaml .

# Create directories for logs and static files
RUN mkdir -p log static

# Expose port
EXPOSE 8888

# Run the application
CMD ["./server"]
