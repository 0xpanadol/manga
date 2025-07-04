# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go module files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
# CGO_ENABLED=0 is important for a static binary
# -o builds the output into a specific file
RUN CGO_ENABLED=0 GOOS=linux go build -v -o manga-api ./cmd/api

# Stage 2: Create the final, small image
FROM alpine:latest

# We need ca-certificates for making HTTPS requests (e.g., to other APIs)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/manga-api .

# This application does not have any static assets to copy yet,
# but if it did, they would be copied here.

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./manga-api"]