# Use the official Go image as the base image
FROM golang:1.23-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests to OpenAI API
RUN apk --no-cache add ca-certificates

# Set the working directory
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy the embedded PDF file (if it exists separately)
# Note: Since we're using go:embed, the PDF is already embedded in the binary

# Expose port (optional, since this is a CLI app)
# EXPOSE 8080

# Set environment variables (can be overridden at runtime)
ENV OPENAI_API_KEY=""

# Run the application
CMD ["./main"]
