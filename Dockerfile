ARG GO_VERSION=1.25.2
ARG NGINX_VERSION=1.27.3

# Build stage
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /build

# Copy go module files first for better caching
COPY go.mod ./

# Copy source code
COPY cmd/ ./cmd/

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o config-gen ./cmd/config-gen

# Runtime stage
FROM nginx:${NGINX_VERSION}-alpine

# Copy the built binary from builder stage
COPY --from=builder /build/config-gen /usr/local/bin/config-gen

# Copy the nginx template
COPY cmd/config-gen/nginx.conf.tmpl /etc/nginx/nginx.conf.tmpl

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh

# Make scripts executable
RUN chmod +x /entrypoint.sh /usr/local/bin/config-gen

# Expose HTTP port
EXPOSE 80

ENTRYPOINT ["/entrypoint.sh"]
CMD ["nginx", "-g", "daemon off;"]
