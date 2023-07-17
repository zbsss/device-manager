# Start from a base image with Go installed
FROM golang:1.17

RUN apt-get update && apt-get install -y \
    gcc \
    pocl-opencl-icd \
    opencl-headers \
    ocl-icd-opencl-dev

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o main ./cmd/device

# Run the binary program produced by `go build`
CMD ["./main"]
