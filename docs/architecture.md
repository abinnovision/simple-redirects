# Architecture Overview

Technical documentation for simple-redirects internals.

## Design Philosophy

- **Single binary** - No external dependencies (Go stdlib only)
- **Configuration as code** - DSL syntax for readability
- **Container-first** - Designed for containerized deployments
- **Minimal footprint** - Alpine-based image, <20MB
- **Zero downtime** - Generate config before nginx start

## System Architecture

```
Configuration (DSL) → config-gen (Go) → nginx.conf.tmpl → nginx.conf → nginx
```

**Flow:**
1. Read DSL from `ROUTES_FILE` or `ROUTES_DSL`
2. Parse DSL into AST
3. Calculate priorities
4. Compile to Config struct
5. Execute Go template
6. Validate with `nginx -t`
7. Start nginx

## Components

### config-gen Binary

**Files:**
- `main.go` (2.9K) - Core logic, env parsing, template execution
- `dsl.go` (13K) - Lexer-based DSL parser
- `nginx.conf.tmpl` (<1K) - Go template for nginx config

**Processing:**
```
Input → Tokenize → Parse → Validate → Sort by Priority → Compile → Template → Output
```

### DSL Parser

**Implementation:** Lexer-based with explicit tokenization

**Pattern types:**
- `PatternExact` - `example.com`
- `PatternWildcardPrefix` - `*.example.com`
- `PatternWildcardSuffix` - `example.*`
- `PatternCatchAll` - `*`
- `PatternWildcard` - `/api/*`
- `PatternNamed` - `/users/:id`

**Operators:**
- `->` Redirect (OperatorRedirect)
- `=>` Proxy (OperatorProxy)
- `!!` Block (OperatorBlock)

## Data Structures

### Core Types (main.go)

```go
type Redirect struct {
    Host   string
    Target string  // Empty for blocks
    Code   int
}

type Proxy struct {
    Host      string
    Target    string
    PathFrom  string  // nginx regex
    PathTo    string  // Rewrite path
    StripPath bool
}

type Config struct {
    Redirects []Redirect
    Proxies   []Proxy
}
```

### DSL Types (dsl.go)

```go
type Route struct {
    Source    SourcePattern   // Host + Path patterns
    Operator  OperatorType    // ->, =>, !!
    Target    TargetPattern   // URL, path, options
    Modifiers []Modifier      // [301], [p:10], [strip_path]
    Priority  int             // Auto-calculated
}
```

## Priority System

**Automatic calculation** based on specificity:

```go
priority := 1000

// Host specificity
switch host.Type {
case PatternExact:          priority -= 100  // 900
case PatternWildcardPrefix: priority += 50   // 1050
case PatternWildcardSuffix: priority += 40   // 1040
case PatternCatchAll:       priority += 100  // 1100
}

// Path specificity
if path != "" {
    switch path.Type {
    case PatternExact:    priority -= 200  // More specific
    case PatternWildcard: priority -= 100  // Less specific
    }
}
```

**Result:** Lower number = higher priority

**Examples:**
- `example.com /health` → 700 (most specific)
- `api.example.com /v1/*` → 780
- `api.example.com` → 900
- `*.example.com` → 1050
- `*` → 1100 (least specific)

**Manual override:** `[p:10]` or `[priority:10]`

## Template System

Uses Go `text/template`:

```go
templatePath := getTemplatePath()  // Checks TEMPLATE_PATH env
tmpl := template.ParseFiles(templatePath)
tmpl.Execute(outputFile, config)
```

**Template structure:**
```nginx
{{range .Redirects}}
server {
    listen 80;
    server_name {{.Host}};
    {{if .Target}}
    return {{.Code}} {{.Target}}$request_uri;
    {{else}}
    return {{.Code}};  # Block - no body
    {{end}}
}
{{end}}

{{range .Proxies}}
server {
    listen 80;
    server_name {{.Host}};
    location ~ {{.PathFrom}} {
        {{if .PathTo}}rewrite {{.PathFrom}} {{.PathTo}} break;{{end}}
        proxy_pass {{.Target}};
        # Standard headers...
    }
}
{{end}}
```

## Build Process

### Multi-stage Docker

**Builder stage:**
```dockerfile
FROM golang:${GO_VERSION}-alpine
COPY cmd/ ./cmd/
RUN CGO_ENABLED=0 go build -ldflags="-w -s" ./cmd/config-gen
```

**Runtime stage:**
```dockerfile
FROM nginx:${NGINX_VERSION}-alpine
COPY --from=builder /build/config-gen /usr/local/bin/
COPY cmd/config-gen/nginx.conf.tmpl /etc/nginx/
COPY entrypoint.sh /entrypoint.sh
```

**Build optimization:**
- `CGO_ENABLED=0` - Static binary
- `-ldflags="-w -s"` - Strip debug info (binary: ~8MB → ~3MB)
- Multi-platform: `linux/amd64`, `linux/arm64`

### Entrypoint Flow

```bash
#!/bin/sh
export TEMPLATE_PATH=/etc/nginx/nginx.conf.tmpl
/usr/local/bin/config-gen  # Generate config
nginx -t                    # Validate
exec nginx -g "daemon off;" # Start
```

## Performance

### Parser

- **Tokenization:** O(n) where n = input length
- **Sorting:** O(m log m) where m = routes
- **Template:** O(m)

**Typical:**
- 100 routes: <10ms
- 1000 routes: <50ms

### Memory

- Binary: ~3MB (stripped)
- Docker image: ~18MB total
- Runtime: <20MB

### Build Time

- Local build: ~2s
- Docker build: ~30s (with cache: ~10s)
- Multi-platform: ~2min (with cache: ~30s)

## Error Handling

### Parse Errors

```
Error on line 5: invalid operator
  api.example.com -=> backend
                  ^^^
Expected: -> (redirect) or => (proxy)
```

### Runtime Errors

All errors write to stderr and exit 1:
- Config file not found
- Invalid DSL syntax
- Template error
- nginx validation failure

## Testing

### Unit Tests

**main_test.go:**
- `TestGetTemplatePath` - Template path resolution

**dsl_test.go (8.9K):**
- `TestParseSimpleRedirect` - Basic redirect
- `TestParseSimpleProxy` - Basic proxy
- `TestParseBlock` - Block operator
- `TestParseWildcardHost` - Wildcard patterns
- `TestParsePathWildcard` - Path wildcards
- `TestParseNamedCaptures` - Named parameters
- `TestParseModifiers` - Modifiers parsing
- `TestParseMultipleRoutes` - Multiple routes
- `TestCompileDSLToConfig` - Config compilation
- `TestPriorityCalculation` - Priority algorithm
- `TestInvalidSyntax` - Error handling
- `TestStripPathIndicator` - strip_path flag

**Coverage:** ~67%

### Integration Tests

Run via `make run`:
```bash
ROUTES_DSL="test.com -> https://example.com" \
OUTPUT_PATH=/tmp/test.conf \
TEMPLATE_PATH=cmd/config-gen/nginx.conf.tmpl \
go run ./cmd/config-gen
```

## Security

### Input Validation

- URL validation in DSL parser
- Pattern syntax checking
- Capture reference validation ($1, $2 must match wildcards)
- No code injection in templates

### Container

- Non-root nginx process
- Read-only root filesystem compatible
- Minimal attack surface (Alpine base)
- No shell in runtime

## Extensibility

### Adding New Operators

1. Add constant to `OperatorType`
2. Add tokenization in `ParseDSL()`
3. Implement `routeToX()` function
4. Update template
5. Add tests

### Custom Templates

```bash
docker run -v /path/to/custom.tmpl:/etc/custom.tmpl \
  -e TEMPLATE_PATH=/etc/custom.tmpl \
  -e ROUTES_FILE=/etc/routes.txt \
  simple-redirects
```

**Available template variables:**
- `.Redirects` - `[]Redirect`
- `.Proxies` - `[]Proxy`

## Monitoring

### Logs

- **stdout:** Generation success/stats
- **stderr:** Errors
- **nginx access:** `/var/log/nginx/access.log`
- **nginx error:** `/var/log/nginx/error.log`

### Health Checks

```bash
# Basic check
curl -f http://localhost/ || exit 1

# Route-specific
curl -I http://localhost/ -H "Host: example.com"
```

## See Also

- [DSL Syntax](dsl.md) - Complete DSL reference
- [Development Guide](development.md) - Build and test
- [README](../README.md) - User-facing overview
