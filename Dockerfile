#
# MailHog Dockerfile with Health Endpoints
#

FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk --no-cache add --virtual build-dependencies \
    git \
    ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o mailhog .

# Final stage
FROM alpine:3.22

# Add mailhog user/group with uid/gid 1000
# This is a workaround for boot2docker issue #581, see
# https://github.com/boot2docker/boot2docker/issues/581
RUN adduser -D -u 1000 mailhog

# Install ca-certificates for HTTPS support
RUN apk --no-cache add ca-certificates

# Copy the binary from builder stage
COPY --from=builder /app/mailhog /usr/local/bin/mailhog

# Set ownership and permissions
RUN chown mailhog:mailhog /usr/local/bin/mailhog && \
    chmod +x /usr/local/bin/mailhog

USER mailhog

WORKDIR /home/mailhog

# Health check using the new health endpoint
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8025/health || exit 1

ENTRYPOINT ["mailhog"]

# Expose the SMTP and HTTP ports:
EXPOSE 1025 8025
