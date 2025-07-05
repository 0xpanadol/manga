# Stage 1: Build the application
FROM golang:1.24.4 AS builder

# This ARG will be passed from docker-compose.yml
# It defaults to 'api' if not provided.
ARG APP_NAME=api

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Use the ARG to specify which application to build.
# The output binary will always be named 'app' for consistency.
RUN CGO_ENABLED=0 GOOS=linux go build -v -o app ./cmd/${APP_NAME}

# Stage 2: Create the final, small image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/app .

# Copy the .env
COPY .env ./

# We don't need to copy migrations or .env for the worker in this setup,
# but it's fine to leave them for the api.

# The command to run the executable.
# The binary is now just called 'app'.
CMD ["./app"]