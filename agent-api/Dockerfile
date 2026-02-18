FROM golang:1.21 AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum* ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o taskapp main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Copy the binary
COPY --from=builder /app/taskapp .

# Copy source code and OpenAPI spec (for AI Assistant to analyze and use)
COPY --from=builder /app /app/source

# Create log directory
RUN mkdir -p /var/log

# Expose ports
# 8080: Task Management API
# 8888: AI Agent (Willknow)
EXPOSE 8080 8888

# Run the application
CMD ["./taskapp"]
