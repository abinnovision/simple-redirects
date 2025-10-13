package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// TokenType represents the type of token
type TokenType int

const (
	TokenHost TokenType = iota
	TokenPath
	TokenOperatorRedirect // ->
	TokenOperatorProxy    // =>
	TokenOperatorBlock    // !!
	TokenURL
	TokenModifier
	TokenWildcard
	TokenCapture
)

// Token represents a lexical token
type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

// OperatorType represents the routing operator
type OperatorType int

const (
	OperatorRedirect OperatorType = iota
	OperatorProxy
	OperatorBlock
)

// PatternType represents the type of pattern matching
type PatternType int

const (
	PatternExact          PatternType = iota
	PatternWildcardPrefix             // *.example.com
	PatternWildcardSuffix             // example.*
	PatternWildcard                   // /api/*
	PatternCatchAll                   // *
	PatternNamed                      // /:id
)

// HostPattern represents a host matching pattern
type HostPattern struct {
	Type  PatternType
	Value string
	Regex *regexp.Regexp
}

// PathPattern represents a path matching pattern
type PathPattern struct {
	Type     PatternType
	Value    string
	Captures []string
	Regex    *regexp.Regexp
}

// SourcePattern represents the source of a route
type SourcePattern struct {
	Host HostPattern
	Path PathPattern
}

// TargetPattern represents the target of a route
type TargetPattern struct {
	URL       string
	Path      string
	StripPath bool
}

// Modifier represents a route modifier
type Modifier struct {
	Name  string
	Value string
}

// Route represents a parsed DSL route
type Route struct {
	Source    SourcePattern
	Operator  OperatorType
	Target    TargetPattern
	Modifiers []Modifier
	Priority  int
	Line      int
}

// ParseDSL parses a DSL configuration string into routes
func ParseDSL(input string) ([]*Route, error) {
	lines := strings.Split(input, "\n")
	routes := make([]*Route, 0)

	for lineNum, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		route, err := parseLine(line, lineNum+1)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum+1, err)
		}

		routes = append(routes, route)
	}

	// Calculate priorities and sort
	for _, route := range routes {
		if route.Priority == 0 {
			route.Priority = calculatePriority(route)
		}
	}

	sortRoutesByPriority(routes)

	return routes, nil
}

// parseLine parses a single line of DSL
func parseLine(line string, lineNum int) (*Route, error) {
	route := &Route{Line: lineNum}

	// Extract modifiers first (anything in brackets at the end)
	modifiers, lineWithoutMods := extractModifiers(line)
	route.Modifiers = modifiers

	// Determine operator type
	var operator string
	if strings.Contains(lineWithoutMods, " -> ") {
		operator = " -> "
		route.Operator = OperatorRedirect
	} else if strings.Contains(lineWithoutMods, " => ") {
		operator = " => "
		route.Operator = OperatorProxy
	} else if strings.Contains(lineWithoutMods, " !! ") {
		operator = " !! "
		route.Operator = OperatorBlock
	} else {
		return nil, fmt.Errorf("no valid operator found (expected ->, =>, or !!)")
	}

	// Split by operator
	parts := strings.Split(lineWithoutMods, operator)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid syntax: expected SOURCE OPERATOR TARGET")
	}

	sourcePart := strings.TrimSpace(parts[0])
	targetPart := strings.TrimSpace(parts[1])

	// Parse source pattern
	source, err := parseSourcePattern(sourcePart)
	if err != nil {
		return nil, fmt.Errorf("invalid source pattern: %w", err)
	}
	route.Source = source

	// Parse target pattern
	target, err := parseTargetPattern(targetPart, route.Operator)
	if err != nil {
		return nil, fmt.Errorf("invalid target pattern: %w", err)
	}
	route.Target = target

	// Apply modifiers
	err = applyModifiers(route)
	if err != nil {
		return nil, err
	}

	return route, nil
}

// extractModifiers extracts modifiers from the end of a line
func extractModifiers(line string) ([]Modifier, string) {
	modifiers := make([]Modifier, 0)

	// Find all bracketed sections at the end
	re := regexp.MustCompile(`\s*\[([^\]]+)\]\s*$`)
	matches := re.FindStringSubmatch(line)

	if len(matches) == 0 {
		return modifiers, line
	}

	modContent := matches[1]
	lineWithout := strings.TrimSpace(re.ReplaceAllString(line, ""))

	// Parse comma-separated modifiers
	parts := strings.Split(modContent, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		// Check for key:value format
		if strings.Contains(part, ":") {
			kv := strings.SplitN(part, ":", 2)
			modifiers = append(modifiers, Modifier{
				Name:  strings.TrimSpace(kv[0]),
				Value: strings.TrimSpace(kv[1]),
			})
		} else {
			// Just a flag or number
			modifiers = append(modifiers, Modifier{
				Name:  part,
				Value: "",
			})
		}
	}

	return modifiers, lineWithout
}

// parseSourcePattern parses a source pattern (host + optional path)
func parseSourcePattern(source string) (SourcePattern, error) {
	pattern := SourcePattern{}

	// Check for catch-all
	if source == "*" {
		pattern.Host = HostPattern{
			Type:  PatternCatchAll,
			Value: "*",
		}
		return pattern, nil
	}

	// Split into host and path
	parts := strings.Fields(source)
	if len(parts) == 0 {
		return pattern, fmt.Errorf("empty source pattern")
	}

	// Parse host
	hostStr := parts[0]
	hostPattern, err := parseHostPattern(hostStr)
	if err != nil {
		return pattern, err
	}
	pattern.Host = hostPattern

	// Parse path if present
	if len(parts) > 1 {
		pathStr := parts[1]
		pathPattern, err := parsePathPattern(pathStr)
		if err != nil {
			return pattern, err
		}
		pattern.Path = pathPattern
	}

	return pattern, nil
}

// parseHostPattern parses a host pattern
func parseHostPattern(host string) (HostPattern, error) {
	pattern := HostPattern{Value: host}

	// Check for wildcard patterns
	if strings.HasPrefix(host, "*.") {
		pattern.Type = PatternWildcardPrefix
		pattern.Value = strings.TrimPrefix(host, "*.")
	} else if strings.HasSuffix(host, ".*") {
		pattern.Type = PatternWildcardSuffix
		pattern.Value = strings.TrimSuffix(host, ".*")
	} else {
		pattern.Type = PatternExact
	}

	return pattern, nil
}

// parsePathPattern parses a path pattern
func parsePathPattern(path string) (PathPattern, error) {
	pattern := PathPattern{Value: path}

	// Check for named captures (:name)
	if strings.Contains(path, ":") {
		pattern.Type = PatternNamed
		// Extract capture names
		re := regexp.MustCompile(`:(\w+)`)
		matches := re.FindAllStringSubmatch(path, -1)
		for _, match := range matches {
			pattern.Captures = append(pattern.Captures, match[1])
		}
		// Build regex pattern
		regexPattern := re.ReplaceAllString(path, `([^/]+)`)
		regexPattern = "^" + regexPattern + "$"
		compiled, err := regexp.Compile(regexPattern)
		if err != nil {
			return pattern, fmt.Errorf("invalid path pattern: %w", err)
		}
		pattern.Regex = compiled
	} else if strings.Contains(path, "*") {
		pattern.Type = PatternWildcard
		// Count wildcards for captures
		pattern.Captures = make([]string, strings.Count(path, "*"))
		// Build regex pattern
		regexPattern := regexp.QuoteMeta(path)
		regexPattern = strings.ReplaceAll(regexPattern, `\*`, `(.*)`)
		regexPattern = "^" + regexPattern + "$"
		compiled, err := regexp.Compile(regexPattern)
		if err != nil {
			return pattern, fmt.Errorf("invalid path pattern: %w", err)
		}
		pattern.Regex = compiled
	} else {
		pattern.Type = PatternExact
	}

	return pattern, nil
}

// parseTargetPattern parses a target pattern
func parseTargetPattern(target string, operator OperatorType) (TargetPattern, error) {
	pattern := TargetPattern{URL: target}

	if operator == OperatorBlock {
		// For block operator, target is the status code
		return pattern, nil
	}

	// Check for path stripping indicator
	if strings.HasSuffix(target, "/_") {
		pattern.StripPath = true
		pattern.URL = strings.TrimSuffix(target, "/_")
	}

	return pattern, nil
}

// applyModifiers applies parsed modifiers to the route
func applyModifiers(route *Route) error {
	for _, mod := range route.Modifiers {
		// Check for HTTP status codes
		if _, err := strconv.Atoi(mod.Name); err == nil {
			// It's a numeric status code - will be used in config generation
			continue
		}

		// Check for priority
		if mod.Name == "priority" || mod.Name == "p" {
			priority, err := strconv.Atoi(mod.Value)
			if err != nil {
				return fmt.Errorf("invalid priority value: %s", mod.Value)
			}
			route.Priority = priority
		}

		// Check for strip_path
		if mod.Name == "strip_path" {
			route.Target.StripPath = true
		}
	}

	return nil
}

// calculatePriority calculates implicit priority based on specificity
// Lower number = higher priority (more specific)
func calculatePriority(route *Route) int {
	priority := 1000

	// Host specificity - more specific = lower number
	switch route.Source.Host.Type {
	case PatternExact:
		priority -= 100
	case PatternWildcardPrefix:
		priority += 50 // Less specific, higher number
	case PatternWildcardSuffix:
		priority += 40
	case PatternCatchAll:
		priority += 100 // Least specific, highest number
	}

	// Path specificity - having a path makes it more specific
	if route.Source.Path.Value != "" {
		switch route.Source.Path.Type {
		case PatternExact:
			priority -= 200
		case PatternWildcard:
			segments := strings.Count(route.Source.Path.Value, "/")
			priority -= 100 + (segments * 10)
		case PatternNamed:
			segments := strings.Count(route.Source.Path.Value, "/")
			priority -= 80 + (segments * 10)
		}
	}

	return priority
}

// sortRoutesByPriority sorts routes by priority (lower = higher priority)
// Stable sort - maintains original order for equal priorities
func sortRoutesByPriority(routes []*Route) {
	// Simple insertion sort - stable and fine for small route sets
	for i := 1; i < len(routes); i++ {
		key := routes[i]
		j := i - 1
		// Move elements with higher priority number (lower specificity) to the right
		for j >= 0 && routes[j].Priority > key.Priority {
			routes[j+1] = routes[j]
			j--
		}
		routes[j+1] = key
	}
}

// CompileDSLToConfig converts DSL routes to the existing Config format
func CompileDSLToConfig(routes []*Route) (Config, error) {
	config := Config{
		Redirects: []Redirect{},
		Proxies:   []Proxy{},
	}

	for _, route := range routes {
		switch route.Operator {
		case OperatorRedirect:
			redirect, err := routeToRedirect(route)
			if err != nil {
				return config, fmt.Errorf("line %d: %w", route.Line, err)
			}
			config.Redirects = append(config.Redirects, redirect)

		case OperatorProxy:
			proxy, err := routeToProxy(route)
			if err != nil {
				return config, fmt.Errorf("line %d: %w", route.Line, err)
			}
			config.Proxies = append(config.Proxies, proxy)

		case OperatorBlock:
			// Blocks are implemented as redirects with error codes
			redirect, err := routeToBlock(route)
			if err != nil {
				return config, fmt.Errorf("line %d: %w", route.Line, err)
			}
			config.Redirects = append(config.Redirects, redirect)
		}
	}

	return config, nil
}

// routeToRedirect converts a Route to a Redirect
func routeToRedirect(route *Route) (Redirect, error) {
	redirect := Redirect{
		Host:   hostPatternToString(route.Source.Host),
		Target: route.Target.URL,
		Code:   301, // Default
	}

	// Check for status code in modifiers
	for _, mod := range route.Modifiers {
		if code, err := strconv.Atoi(mod.Name); err == nil {
			redirect.Code = code
			break
		}
	}

	return redirect, nil
}

// routeToProxy converts a Route to a Proxy
func routeToProxy(route *Route) (Proxy, error) {
	proxy := Proxy{
		Host:      hostPatternToString(route.Source.Host),
		Target:    route.Target.URL,
		StripPath: route.Target.StripPath,
	}

	// Handle path patterns
	if route.Source.Path.Value != "" {
		switch route.Source.Path.Type {
		case PatternExact:
			proxy.PathFrom = route.Source.Path.Value
		case PatternWildcard:
			// Convert wildcard to nginx regex
			proxy.PathFrom = pathPatternToNginxRegex(route.Source.Path)
		case PatternNamed:
			// Convert named captures to nginx regex
			proxy.PathFrom = pathPatternToNginxRegex(route.Source.Path)
		}

		// Handle target path transformation
		if strings.Contains(route.Target.URL, "$") {
			// Split URL and path
			urlParts := strings.SplitN(route.Target.URL, "/", 4)
			if len(urlParts) >= 4 {
				proxy.Target = strings.Join(urlParts[:3], "/")
				proxy.PathTo = "/" + urlParts[3]
			}
		}
	}

	return proxy, nil
}

// routeToBlock converts a Route to a blocking Redirect
func routeToBlock(route *Route) (Redirect, error) {
	statusCode := 403 // Default block code

	// Check for status code in modifiers or target
	for _, mod := range route.Modifiers {
		if c, err := strconv.Atoi(mod.Name); err == nil {
			statusCode = c
			break
		}
	}

	// Check if target is a status code
	if c, err := strconv.Atoi(route.Target.URL); err == nil {
		statusCode = c
	}

	redirect := Redirect{
		Host:   hostPatternToString(route.Source.Host),
		Target: "", // Empty target means return status
		Code:   statusCode,
	}

	return redirect, nil
}

// hostPatternToString converts a HostPattern to string for nginx
func hostPatternToString(pattern HostPattern) string {
	switch pattern.Type {
	case PatternCatchAll:
		return "_" // nginx default server
	case PatternWildcardPrefix:
		return "*." + pattern.Value
	case PatternWildcardSuffix:
		return pattern.Value + ".*"
	default:
		return pattern.Value
	}
}

// pathPatternToNginxRegex converts a PathPattern to nginx regex
func pathPatternToNginxRegex(pattern PathPattern) string {
	if pattern.Regex != nil {
		// Extract the regex pattern without ^ and $
		regexStr := pattern.Regex.String()
		regexStr = strings.TrimPrefix(regexStr, "^")
		regexStr = strings.TrimSuffix(regexStr, "$")
		return regexStr
	}
	return pattern.Value
}
