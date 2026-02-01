# Build stage
FROM golang:1.23-alpine AS builder

# Install git for go mod download
RUN apk add --no-cache git

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod ./
COPY go.sum* ./

# Copy source code
COPY . .

# Download dependencies (will update go.sum if needed)
RUN go mod tidy

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bot ./cmd/bot

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /bot .

# Copy game information file
COPY --from=builder /app/Information.md .

# Copy resources folder (images, etc.)
COPY --from=builder /app/resources ./resources

# Run as non-root user
RUN adduser -D -g '' botuser
USER botuser

# Run the bot
CMD ["./bot"]
