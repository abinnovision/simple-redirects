# Simple Redirects

Lightweight nginx configuration generator for HTTP redirects, reverse proxies, and URL blocking. Configured via DSL syntax, packaged as an Alpine-based Docker container.

## Features

- **HTTP Redirects** - 301/302 redirects with configurable status codes
- **Reverse Proxy** - Backend proxying with path transformation support
- **URL Blocking** - Return error codes for specific hosts
- **DSL Syntax** - Intuitive one-line routing rules
- **Flexible Configuration** - File-based or inline DSL
- **Minimal Footprint** - <20MB Alpine-based Docker image

## Quick Start

```bash
# Create routes file
cat > routes.txt <<'EOF'
old.example.com -> https://new.example.com
api.example.com => http://backend:8080
bad-actor.com !! 403
EOF

# Run container
docker run -d -p 80:80 \
  -v $(pwd)/routes.txt:/etc/routes.txt:ro \
  -e ROUTES_FILE=/etc/routes.txt \
  ghcr.io/abinnovision/simple-redirects:latest
```

## Configuration

### DSL Syntax (Recommended)

Simple, readable routing rules:

```
# Redirects
old.example.com -> https://new.example.com [301]

# Proxies with path transformation
api.example.com /v1/* => http://backend:8080/api/$1

# Wildcards
*.saas.example.com => http://tenant-backend:3000

# Blocking
spam.example.com !! 403
```

**[→ Complete DSL Documentation](docs/dsl.md)**

### Environment Variables

**Priority order:**
1. `ROUTES_FILE` - Path to DSL file (recommended for production)
2. `ROUTES_DSL` - Inline DSL string (useful for testing)

**Examples:**

```bash
# File-based DSL (recommended)
docker run -v ./routes.txt:/etc/routes.txt:ro -e ROUTES_FILE=/etc/routes.txt ...

# Inline DSL
docker run -e ROUTES_DSL="old.com -> https://new.com" ...
```

**See [docs/dsl.md](docs/dsl.md)** for complete configuration reference.

## Building

### Local Development

```bash
# Using Makefile (recommended)
make build          # Build binary to dist/
make test           # Run tests
make ci             # Run all CI checks
make help           # Show all available targets

# Or use Go directly
go build -o dist/simple-redirects ./cmd/config-gen
go test ./...

# Build Docker image
make docker
# Or: ./scripts/build.sh
```

### Docker Tags

- `latest` - Latest stable release
- `edge` - Latest main branch build
- `sha-<hash>` - Specific commit
- `v1.2.3` - Semantic version

## Contributing

This project follows [Conventional Commits](https://www.conventionalcommits.org/) for automated versioning:

```bash
# Feature (minor bump)
git commit -m "feat: add WebSocket support"

# Fix (patch bump)
git commit -m "fix: handle empty paths correctly"

# Breaking change (major bump)
git commit -m "feat!: change config format"
```

**[→ Development Guide](docs/development.md)**

## Documentation

- **[DSL Syntax Guide](docs/dsl.md)** - Complete DSL and configuration reference
- **[Development Guide](docs/development.md)** - Build, test, and contribute
- **[Architecture](docs/architecture.md)** - Technical deep dive

## License

See [LICENSE](LICENSE) file for details.
