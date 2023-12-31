# Build stage
FROM golang:1.17 AS build

# Install dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    pocl-opencl-icd \
    opencl-headers \
    ocl-icd-opencl-dev

# Set the Current Working Directory inside the container
WORKDIR /src

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o /out/main ./cmd/device

# Run stage
FROM debian:buster-slim

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    pocl-opencl-icd \
    ocl-icd-opencl-dev \
    && rm -rf /var/lib/apt/lists/*

COPY --from=build /out/main /app/main

# Run the binary program produced by `go build`
CMD ["/app/main"]
