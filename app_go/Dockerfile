# ============================================
# STAGE 1: Build the Go application
# ============================================

FROM golang:1.25-alpine AS builder

#Installing git
RUN apk add --no-cache git

# Set wroking dir
WORKDIR /app

# Copying go.mod first (for better caching)
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copying the source code
COPY main.go ./

# Build the application
# CGO_ENABLED=0: Creates a static binary (no C dependencies)
# -ldflags="-w -s": Strips debug info to reduce binary size
# -o devops-info-service: Output binary name
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o devops-info-service main.go

# ============================================
# STAGE 2: Create minimal runtime image
# ============================================
FROM alpine:3.19

# Add CA certificates for HTTPS requests 
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -S appuser && adduser -S appuser -G appuser

# Setting working dir
WORKDIR /app

# Copying only the binary from the builder stage
COPY --from=builder /app/devops-info-service .

# Change ownership to non-root user
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose the port
EXPOSE 8080

# Run the application
CMD [ "./devops-info-service" ]