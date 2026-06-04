# Stage 1: Build the Go app
FROM golang:1.24-alpine AS build_base

# Install necessary packages
# RUN apk add --no-cache git build-base

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod .
COPY go.sum .

# Download dependencies
RUN go mod download

# Copy the entire project to the workspace
COPY . .

# Build the Go app
RUN go build -o keelBackend .

# Stage 2: Create a minimal runtime image
FROM alpine:latest

# Install necessary packages
#RUN apk add --no-cache ca-certificates

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the built binary and necessary files from the build stage
COPY --from=build_base /app/keelBackend /app/keelBackend
COPY --from=build_base /app/config /app/config
COPY --from=build_base /app/com.code.github.com/glodb/keel/services /app/com.code.github.com/glodb/keel/services

# Run the binary program produced by `go build`
ENTRYPOINT ["./keelBackend"]