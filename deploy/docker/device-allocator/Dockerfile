# Build stage
FROM golang:1.17 AS build

# Set the Current Working Directory inside the container
WORKDIR /src

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o /out/main ./cmd/device-allocator

# Run stage
FROM debian:buster-slim

COPY --from=build /out/main /app/main

# Run the binary program produced by `go build`
CMD ["/app/main"]
