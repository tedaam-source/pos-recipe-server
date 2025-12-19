# Build stage
FROM golang:1.24 as builder


WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum* ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
# CGO_ENABLED=0 is important for the resulting binary to be static and run in scratch/distroless
RUN CGO_ENABLED=0 GOOS=linux go build -v -o server ./cmd/api

# Run stage
FROM gcr.io/distroless/static-debian12

WORKDIR /

COPY --from=builder /app/server /server

# Expose port 8080 (Cloud Run default, but configurable via PORT env)
ENV PORT 8080

# Run the web service on container startup.
CMD ["/server"]
