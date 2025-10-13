# Development Guide

Build, test, and contribute to simple-redirects.

## Prerequisites

- **Go 1.25.2+** - Managed via `.tool-versions` (use [asdf](https://asdf-vm.com/))
- **Docker** - For container builds
- **Make** - Build automation

## Quick Start

```bash
git clone https://github.com/abinnovision/simple-redirects.git
cd simple-redirects

# Install Go (if using asdf)
asdf install

# Show available commands
make help

# Build and test
make build
make test
make ci        # Run all checks

# Build Docker image
make docker
```

## Common Tasks

### Building

```bash
make build              # Build to dist/
make build-linux        # Linux amd64
make build-darwin       # macOS amd64/arm64
make build-all          # All platforms
```

### Testing

```bash
make test               # Run tests
make test-coverage      # With coverage report
make test-race          # With race detection
```

### Code Quality

```bash
make fmt                # Format code
make fmt-check          # Check formatting
make vet                # Static analysis
make lint               # Run staticcheck
make ci                 # All checks (fmt, vet, lint, test)
```

### Docker

```bash
make docker             # Build with defaults
make docker GO_VERSION=1.26.0 NGINX_VERSION=1.28.0  # Custom versions
```

## Pre-commit Hooks

Automatically format and lint staged files:

```bash
# Install (one-time)
pip install pre-commit
make install-hooks

# Now hooks run automatically on commit
# Skip with: git commit --no-verify
```

**Hooks:**
- `go fmt` - Auto-format
- `go vet` - Static analysis
- `staticcheck` - Advanced linting
- `go test` - Run tests (on push)
- Trailing whitespace, EOF fixes, YAML validation

## Project Structure

```
cmd/config-gen/
├── main.go           # Core application logic
├── main_test.go      # Core tests
├── dsl.go            # DSL parser (lexer, AST, priority)
├── dsl_test.go       # Parser tests (comprehensive)
└── nginx.conf.tmpl   # nginx template
```

## Development Workflow

1. **Make changes** to `cmd/config-gen/*.go`
2. **Format:** `make fmt`
3. **Test:** `make test`
4. **Check:** `make ci`
5. **Commit:** Follow [Conventional Commits](https://www.conventionalcommits.org/)

```bash
# Examples
git commit -m "feat: add WebSocket support"
git commit -m "fix: handle empty paths correctly"
git commit -m "feat!: change config format"  # Breaking change
```

## Testing

### Unit Tests

```bash
# All tests
go test ./...

# Specific test
go test -run TestParseDSL ./cmd/config-gen

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Testing

```bash
# Create test config
cat > /tmp/routes.txt <<'EOF'
test.local -> https://example.com
EOF

# Test locally
ROUTES_FILE=/tmp/routes.txt \
OUTPUT_PATH=/tmp/nginx.conf \
TEMPLATE_PATH=cmd/config-gen/nginx.conf.tmpl \
go run ./cmd/config-gen

# Or use Makefile
make run
```

## Go Version Management

**.tool-versions is the single source of truth.**

All builds (local, Docker, CI/CD) read from this file:
```
golang 1.25.2
```

**Update Go version:**
```bash
echo "golang 1.26.0" > .tool-versions
# No other changes needed - all builds will use new version
```

## CI/CD Pipeline

**File:** `.github/workflows/build.yaml`

**On every push:**
1. Check & Build job (fmt, vet, staticcheck, build, test)

**On push to main:**
2. Build Docker images with SHA tags (`sha-abc1234`, `edge`)
3. Create release PR (if conventional commits warrant it)

**On release:**
4. Build Docker images with semantic version tags (`latest`, `v1.2.3`, etc.)

**Before pushing:**
```bash
make ci  # Verify all checks pass
```

## Docker Details

### Multi-stage Build

**Stage 1 (Builder):** `golang:${GO_VERSION}-alpine`
- Copies source code
- Builds static binary with `CGO_ENABLED=0`
- Strips debug symbols: `-ldflags="-w -s"`

**Stage 2 (Runtime):** `nginx:${NGINX_VERSION}-alpine`
- Copies binary from builder
- Copies template
- Runs entrypoint script

### Build Args

- `GO_VERSION` - Go compiler version (default from `.tool-versions`)
- `NGINX_VERSION` - nginx version (default: `1.27.3`)

### Testing Docker Build

```bash
# Build
make docker

# Run with file
docker run -d -p 8080:80 \
  -v $(pwd)/examples/routes.txt:/etc/routes.txt:ro \
  -e ROUTES_FILE=/etc/routes.txt \
  simple-redirects:latest

# Run with inline DSL
docker run -d -p 8080:80 \
  -e ROUTES_DSL="test.local -> https://example.com" \
  simple-redirects:latest

# Check logs
docker logs <container-id>

# Test endpoint
curl -I http://localhost:8080 -H "Host: test.local"
```

## Debugging

### Parser Errors

Test DSL parsing directly:

```bash
go run ./cmd/config-gen <<< 'ROUTES_DSL=invalid syntax' 2>&1
```

### Template Errors

```bash
# Test template generation
ROUTES_DSL="test.com -> https://example.com" \
OUTPUT_PATH=/tmp/test.conf \
TEMPLATE_PATH=cmd/config-gen/nginx.conf.tmpl \
go run ./cmd/config-gen

# Validate nginx config
nginx -t -c /tmp/test.conf
```

### Test Failures

```bash
# Verbose output
go test -v ./...

# Specific test with verbose
go test -v -run TestParseDSL ./cmd/config-gen

# Show test coverage gaps
go test -cover ./...
```

## Release Process

Releases are automated via [release-please](https://github.com/googleapis/release-please):

1. **Commit** using conventional commits
2. **Push** to `main`
3. **Release PR** is created/updated automatically
4. **Merge** release PR to trigger release and Docker publish

**Commit types:**
- `feat:` → Minor version bump (1.0.0 → 1.1.0)
- `fix:` → Patch version bump (1.0.0 → 1.0.1)
- `feat!:` → Major version bump (1.0.0 → 2.0.0)

**Other types** (`chore:`, `docs:`, `refactor:`, `test:`, `ci:`) don't trigger releases.

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feat/my-feature`
3. Make changes following the development workflow
4. Run `make ci` to verify
5. Commit using conventional commits
6. Push and create a pull request

## Troubleshooting

**staticcheck not found:**
```bash
make install-tools
# Or: go install honnef.co/go/tools/cmd/staticcheck@latest
```

**Tests fail on CI but pass locally:**
```bash
make test-race  # Check for race conditions
make ci         # Run exact same checks as CI
```

**Docker build fails:**
```bash
# Check build args
docker build \
  --build-arg GO_VERSION=$(grep "^golang" .tool-versions | awk '{print $2}') \
  --build-arg NGINX_VERSION=1.27.3 \
  -t simple-redirects:test .
```

**Template not found:**
```bash
# Ensure template is in correct location
ls -la cmd/config-gen/nginx.conf.tmpl

# Set TEMPLATE_PATH explicitly
export TEMPLATE_PATH=cmd/config-gen/nginx.conf.tmpl
```

## See Also

- [DSL Syntax](dsl.md) - Complete DSL reference
- [Architecture](architecture.md) - Technical deep dive
- [README](../README.md) - User-facing overview
