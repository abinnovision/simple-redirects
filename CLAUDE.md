# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## Project Overview

Go application that generates nginx configurations for HTTP redirects, reverse proxies, and URL blocking. Configured via DSL syntax, packaged as an Alpine-based Docker container.

**Key Features:**
- DSL-based routing rules (`->` redirect, `=>` proxy, `!!` block)
- Priority-based route matching
- Multi-platform Docker images
- No external dependencies (Go stdlib only)

## Project Structure

```
.
├── cmd/config-gen/       # Main application
│   ├── main.go          # Core logic
│   ├── dsl.go           # DSL parser
│   ├── *_test.go        # Tests
│   └── nginx.conf.tmpl  # nginx template
├── docs/                # Complete documentation
│   ├── dsl.md          # DSL syntax + configuration reference
│   ├── development.md  # Build, test, contribute
│   └── architecture.md # Technical deep dive
├── examples/           # Example configs
├── scripts/            # Build scripts
├── dist/               # Build output (gitignored)
├── Makefile            # Build automation
├── Dockerfile          # Multi-stage build
└── .tool-versions      # Go version (source of truth)
```

## Quick Reference

### Common Commands

```bash
make help       # Show all targets
make build      # Build to dist/
make test       # Run tests
make ci         # Run all checks (fmt, vet, lint, test)
make docker     # Build Docker image
make clean      # Clean dist/
```

### Running Locally

```bash
# Build and test with example
ROUTES_DSL="old.com -> https://new.com" \
OUTPUT_PATH=/tmp/nginx.conf \
TEMPLATE_PATH=cmd/config-gen/nginx.conf.tmpl \
go run ./cmd/config-gen

# Or use Makefile
make run
```

### File Locations

- **Source:** `cmd/config-gen/main.go`, `dsl.go`
- **Tests:** `cmd/config-gen/*_test.go`
- **Template:** `cmd/config-gen/nginx.conf.tmpl`
- **Docs:** `docs/dsl.md`, `docs/development.md`, `docs/architecture.md`
- **Examples:** `examples/routes.txt`

## Configuration

**Environment variables:**
- `ROUTES_FILE` - Path to DSL file (recommended)
- `ROUTES_DSL` - Inline DSL string
- `TEMPLATE_PATH` - Custom template path
- `OUTPUT_PATH` - Custom output path (default: `/etc/nginx/nginx.conf`)

**See [docs/dsl.md](docs/dsl.md)** for complete DSL syntax and configuration reference.

## DSL Quick Reference

```bash
# Redirects
old.com -> https://new.com [301]

# Proxies
api.com => http://backend:8080
api.com /v1/* => http://backend:8080/$1

# Blocks
spam.com !! 403

# Wildcards
*.example.com => http://backend:8080
example.com /api/* => http://backend:8080/$1

# Priority (lower = higher precedence)
api.com /health => http://backend:8080/healthz [p:1]
api.com => http://backend:8080 [p:100]
```

**See [docs/dsl.md](docs/dsl.md)** for complete syntax reference.

## Development Workflow

1. **Make changes** to `cmd/config-gen/*.go`
2. **Format:** `make fmt`
3. **Test:** `make test`
4. **Check:** `make ci` (runs fmt-check, vet, lint, test)
5. **Build:** `make build`
6. **Commit** following [Conventional Commits](https://www.conventionalcommits.org/)

**Pre-commit hooks** (optional):
```bash
pip install pre-commit
make install-hooks
```

**See [docs/development.md](docs/development.md)** for complete development guide.

## Architecture Notes

**Core types** (in `main.go`):
- `Redirect` - HTTP redirects and blocks (target empty = block)
- `Proxy` - Reverse proxy with path transformation
- `Config` - Container for redirects and proxies

**DSL Parser** (in `dsl.go`):
- Lexer-based tokenization
- AST construction
- Automatic priority calculation
- Pattern matching (exact, wildcard, catch-all)

**Priority system:**
- Exact host + exact path: ~700 (highest)
- Exact host + wildcard path: ~800
- Exact host: 900
- Wildcard host: 1050
- Catch-all: 1100 (lowest)

**See [docs/architecture.md](docs/architecture.md)** for technical deep dive.

## Testing

```bash
make test              # Run all tests
make test-coverage     # With coverage report
make test-race         # With race detection
```

**Test files:**
- `cmd/config-gen/main_test.go` - Core logic tests
- `cmd/config-gen/dsl_test.go` - DSL parser tests (comprehensive)

## Docker

```bash
# Build locally
make docker

# Run with DSL file
docker run -d -p 80:80 \
  -v $(pwd)/examples/routes.txt:/etc/routes.txt:ro \
  -e ROUTES_FILE=/etc/routes.txt \
  ghcr.io/abinnovision/simple-redirects:latest

# Run with inline DSL
docker run -d -p 80:80 \
  -e ROUTES_DSL="old.com -> https://new.com" \
  ghcr.io/abinnovision/simple-redirects:latest
```

**Multi-stage build:**
1. Builder: `golang:1.25.2-alpine` (from `.tool-versions`)
2. Runtime: `nginx:1.27.3-alpine`

**Build args:** `GO_VERSION`, `NGINX_VERSION`

## CI/CD

**Workflow:** `.github/workflows/build.yaml`

**On every push:**
1. Check & Build - fmt, vet, staticcheck, build, test
2. (main only) Build Docker - SHA tags: `sha-abc1234`, `edge`

**On release:**
1. Semantic versioning via release-please
2. Docker images: `latest`, `v1.2.3`, `v1.2`, `v1`

**Verify locally before pushing:**
```bash
make ci
```

## Go Version

**Source of truth:** `.tool-versions`
```
golang 1.25.2
```

All build processes (local, Docker, CI) read from this file.

**Update:**
```bash
echo "golang 1.26.0" > .tool-versions
```

## Contributing

1. Follow [Conventional Commits](https://www.conventionalcommits.org/)
   - `feat:` - New feature (minor bump)
   - `fix:` - Bug fix (patch bump)
   - `feat!:` - Breaking change (major bump)

2. Run `make ci` before committing

3. Create PR to `main` branch

**See [docs/development.md](docs/development.md)** for complete guide.

## Documentation

- **[docs/dsl.md](docs/dsl.md)** - Complete DSL syntax and configuration
- **[docs/development.md](docs/development.md)** - Build, test, contribute
- **[docs/architecture.md](docs/architecture.md)** - Technical deep dive
- **[README.md](README.md)** - User-facing overview
- **[examples/routes.txt](examples/routes.txt)** - Example DSL file
