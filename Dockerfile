# Use the official Go image as a base image
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy the Go source code into the container
COPY . .

# Set GIN_MODE to "release"
ENV GIN_MODE=release

ENV DB_HOST=mysql-container
ENV DB_PORT=3306
ENV DB_USERNAME=test
ENV DB_PASSWORD=test
ENV DB_NAME=testdb
ENV APPLICATION_PORT=3535

# Build the Go application
RUN go build -o main .

# Use a smaller base image for the final image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Expose the port the application runs on
EXPOSE 3535

# Command to run the executable
CMD ["./main"]
