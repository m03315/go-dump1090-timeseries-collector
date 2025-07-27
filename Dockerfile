# Stage 1: Build the Go application
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker cache
# This means Go modules won't be re-downloaded unless go.mod/go.sum change
COPY go.mod .
COPY go.sum .

# Download dependencies - this step can be cached
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the Go application
# CGO_ENABLED=0 creates a statically linked binary, making it more portable
# -o specifies the output binary name
# ./cmd/collector is an example if your main package is in cmd/collector, otherwise use ./
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o collector .

# Stage 2: Create the final, minimal image
# Use a minimal base image like scratch or alpine for the smallest possible image
FROM alpine:latest

# Set the working directory for the application
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/collector .

# Expose the port if your application serves an API or connects to a specific port
# (e.g., if you were to add a health check endpoint or web UI)
# Note: This is *not* the port your collector connects *out* to (30003)
# EXPOSE 8080 # Uncomment and change if your app serves a port

# Define default environment variables (can be overridden during docker run)
ENV DUMP1090_HOST="dump1090-server"
ENV DUMP1090_PORT="30003"
ENV INFLUX_URL="http://influxdb:8086"
ENV INFLUXDB_TOKEN="your-influxdb-token"
ENV INFLUXDB_DATABASE="your-bucket"

# Command to run the executable
ENTRYPOINT ["./collector"]

# CMD can provide default arguments to ENTRYPOINT, or act as the main command if no ENTRYPOINT is set
# CMD ["--config", "/etc/collector/config.yaml"] # Example if using a config file