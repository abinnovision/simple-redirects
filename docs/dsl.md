# DSL Syntax Reference

Complete reference for the routing DSL (Domain-Specific Language).

## Overview

The DSL provides a concise, readable syntax for defining routing rules. Each rule is a single line with a clear structure:

```
SOURCE_PATTERN OPERATOR TARGET_PATTERN [MODIFIERS]
```

## Basic Syntax

### Redirect

```
old.example.com -> https://new.example.com
```

Performs HTTP redirect (301 by default).

### Proxy

```
api.example.com => http://backend:8080
```

Reverse proxies traffic to backend.

### Block

```
spam.example.com !! 403
```

Returns HTTP error code, blocking access.

## Source Patterns

### Exact Host

```
example.com -> https://new.com
```

Matches exact hostname only.

### Wildcard Prefix

```
*.example.com => http://backend:8080
```

Matches any subdomain (e.g., `api.example.com`, `www.example.com`).

### Wildcard Suffix

```
example.* -> https://example.com
```

Matches any TLD (e.g., `example.net`, `example.org`).

### Catch-All

```
* => http://default-backend:8080
```

Matches any hostname not matched by more specific rules.

### Path Patterns

**Exact path:**
```
example.com /health => http://backend:8080/healthz
```

**Wildcard path:**
```
example.com /api/* => http://backend:8080/api/$1
```

Captures everything after `/api/` and passes to backend.

**Named parameters:**
```
example.com /users/:id => http://backend:8080/user/$1
```

Captures `:id` as `$1` for use in target.

## Multiple Routes Per Host

**You can define the same host multiple times with different paths.** The priority system ensures the most specific routes are matched first:

```
# Most specific: exact path match
api.example.com /health => http://backend:8080/healthz [p:1]

# Moderately specific: wildcard paths
api.example.com /v2/* => http://backend-v2:8080/$1 [p:10]
api.example.com /v1/* => http://backend-v1:8080/$1 [p:20]

# Least specific: catch-all for this host
api.example.com => http://backend-v1:8080 [p:100]
```

Routes are evaluated in priority order:
- Lower priority numbers are checked first
- More specific patterns automatically get higher priority (lower numbers)
- Manual priority modifiers allow fine-tuning

## Operators

| Operator | Type | Description |
|----------|------|-------------|
| `->` | Redirect | HTTP redirect (301/302) |
| `=>` | Proxy | Reverse proxy |
| `!!` | Block | Return error code |

## Target Patterns

### Redirect Targets

**Full URL:**
```
old.com -> https://new.com
```

**Preserve path:**
```
old.com -> https://new.com$request_uri
```

nginx variable `$request_uri` is automatically appended for redirects.

### Proxy Targets

**Simple proxy:**
```
example.com => http://backend:8080
```

**Path transformation:**
```
example.com /api/* => http://backend:8080/v1/$1
```

**Path stripping:**
```
example.com /api/* => http://backend:8080 [strip_path]
```

### Block Targets

```
spam.com !! 403
spam.com !! 404
spam.com !!     # Returns 403 by default
```

## Modifiers

Modifiers are optional and specified in square brackets.

### Status Code

**Redirect:**
```
example.com -> https://new.com [301]
example.com -> https://new.com [302]
```

**Block:**
```
example.com !! 403
example.com !! 404
```

### Priority

Manual priority override:

```
example.com /api => http://backend:8080 [p:10]
example.com => http://fallback:8080 [priority:100]
```

Lower numbers = higher priority.

### Options

**Strip path:**
```
example.com /api/* => http://backend:8080 [strip_path]
```

Removes matched path before proxying.

### Combined Modifiers

```
example.com /api/* => http://backend:8080 [p:10,strip_path]
```

## Priority System

Routes are automatically prioritized by specificity:

**Host specificity** (base priority = 1000):
- Exact host: -100 → priority 900
- Wildcard prefix (`*.example.com`): +50 → priority 1050
- Wildcard suffix (`example.*`): +50 → priority 1050
- Catch-all (`*`): +100 → priority 1100

**Path specificity:**
- Exact path: -200
- Wildcard path: -100
- No path: 0

**Examples:**

| Route | Priority | Explanation |
|-------|----------|-------------|
| `example.com /health` | 700 | Exact host (900) + exact path (-200) |
| `example.com /api/*` | 800 | Exact host (900) + wildcard path (-100) |
| `example.com` | 900 | Exact host only |
| `*.example.com /api/*` | 950 | Wildcard host (1050) + wildcard path (-100) |
| `*.example.com` | 1050 | Wildcard host only |
| `*` | 1100 | Catch-all |

**Manual override:**
```
# Force specific priority
example.com => http://backend:8080 [p:5]
```

## Comments and Blank Lines

```
# This is a comment
example.com -> https://new.com

# Blank lines are ignored

# Multiple comments in a row
```

Lines starting with `#` are ignored. Empty lines are ignored.

## Complete Examples

### Simple Redirect

```
www.example.com -> https://example.com
```

### Multi-tenant SaaS

```
# Per-tenant backends
customer1.saas.com => http://customer1:3000
customer2.saas.com => http://customer2:3000

# Wildcard for remaining tenants
*.saas.com => http://default:3000
```

### API Gateway with Versioning

```
# Health check endpoint
api.example.com /health => http://backend:8080/healthz [p:1]

# Version-specific routing
api.example.com /v2/* => http://backend-v2:8080/$1 [p:10]
api.example.com /v1/* => http://backend-v1:8080/$1 [p:20]

# Default to v1 for unversioned requests
api.example.com => http://backend-v1:8080 [p:100]
```

### Mixed Redirects and Proxies

```
# Redirect old domains
old.example.com -> https://new.example.com
www.example.com -> https://example.com [302]

# Proxy API traffic
api.example.com => http://backend:8080

# Block unwanted traffic
spam.example.com !! 403
```

## Syntax Errors

Common errors and their messages:

**Invalid operator:**
```
example.com <- http://backend:8080
```
Error: Invalid operator, expected `->`, `=>`, or `!!`

**Missing target:**
```
example.com ->
```
Error: Missing target after operator

**Invalid wildcard:**
```
*.*.example.com => http://backend:8080
```
Error: Only one wildcard allowed per hostname

**Invalid capture reference:**
```
example.com /api/* => http://backend:8080/$2
```
Error: Only 1 wildcard in source, but $2 referenced

## Configuration

### Environment Variables

**Priority order** (first match wins):

1. `ROUTES_FILE` - Path to DSL file (recommended for production)
2. `ROUTES_DSL` - Inline DSL string (useful for testing)

**Additional variables:**
- `TEMPLATE_PATH` - Custom nginx template path
- `OUTPUT_PATH` - Custom output path (default: `/etc/nginx/nginx.conf`)

### Docker Usage

**File-based (recommended):**
```bash
docker run -d -p 80:80 \
  -v /path/to/routes.txt:/etc/routes.txt:ro \
  -e ROUTES_FILE=/etc/routes.txt \
  ghcr.io/abinnovision/simple-redirects:latest
```

**Inline DSL:**
```bash
docker run -d -p 80:80 \
  -e ROUTES_DSL="old.com -> https://new.com
api.com => http://backend:8080" \
  ghcr.io/abinnovision/simple-redirects:latest
```

### Template Customization

Mount custom nginx template:

```bash
docker run -d -p 80:80 \
  -v /path/to/custom.tmpl:/etc/nginx/custom.tmpl:ro \
  -e TEMPLATE_PATH=/etc/nginx/custom.tmpl \
  -e ROUTES_FILE=/etc/routes.txt \
  ghcr.io/abinnovision/simple-redirects:latest
```

**Template variables available:**
- `.Redirects` - Array of redirect/block configurations
- `.Proxies` - Array of proxy configurations

## See Also

- [Architecture](architecture.md) - Technical deep dive
- [Development Guide](development.md) - Build and test
