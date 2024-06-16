# Stage 1: Build the Go application
FROM golang:1.18 AS build

WORKDIR /app

# Copy the Go application code
COPY . .

# Download dependencies
RUN go mod tidy

# Build the application
RUN go build -o nexasecure-server

# Stage 2: Run the Go application
FROM alpine:latest

WORKDIR /root/

# Copy the built application from the builder stage
COPY --from=build /app/nexasecure-server .

# Expose the application port
EXPOSE 3000

# Start the Go application
CMD ["./nexasecure-server"]
