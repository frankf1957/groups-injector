# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files (if you create them)
COPY go.mod go.sum* ./

# Download dependencies (if go.mod exists)
RUN if [ -f go.mod ]; then go mod download; fi

# Copy source code
COPY main.go .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o groups-injector .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/groups-injector .

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./groups-injector"]
