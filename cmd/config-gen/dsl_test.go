package main

import (
	"testing"
)

func TestParseSimpleRedirect(t *testing.T) {
	input := "old.example.com -> https://new.example.com"
	routes, err := ParseDSL(input)

	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Operator != OperatorRedirect {
		t.Error("Expected OperatorRedirect")
	}

	if route.Source.Host.Value != "old.example.com" {
		t.Errorf("Expected host 'old.example.com', got '%s'", route.Source.Host.Value)
	}

	if route.Target.URL != "https://new.example.com" {
		t.Errorf("Expected target 'https://new.example.com', got '%s'", route.Target.URL)
	}
}

func TestParseSimpleProxy(t *testing.T) {
	input := "api.example.com => http://backend:8080"
	routes, err := ParseDSL(input)

	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Operator != OperatorProxy {
		t.Error("Expected OperatorProxy")
	}

	if route.Source.Host.Value != "api.example.com" {
		t.Errorf("Expected host 'api.example.com', got '%s'", route.Source.Host.Value)
	}

	if route.Target.URL != "http://backend:8080" {
		t.Errorf("Expected target 'http://backend:8080', got '%s'", route.Target.URL)
	}
}

func TestParseBlock(t *testing.T) {
	input := "bad.example.com !! 403"
	routes, err := ParseDSL(input)

	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Operator != OperatorBlock {
		t.Error("Expected OperatorBlock")
	}
}

func TestParseWildcardHost(t *testing.T) {
	tests := []struct {
		input        string
		expectedType PatternType
		expectedVal  string
	}{
		{"*.example.com -> https://example.com", PatternWildcardPrefix, "example.com"},
		{"example.* -> https://example.com", PatternWildcardSuffix, "example"},
		{"* => http://backend:8080", PatternCatchAll, "*"},
	}

	for _, tt := range tests {
		routes, err := ParseDSL(tt.input)
		if err != nil {
			t.Fatalf("Failed to parse '%s': %v", tt.input, err)
		}

		if len(routes) != 1 {
			t.Fatalf("Expected 1 route for '%s', got %d", tt.input, len(routes))
		}

		route := routes[0]
		if route.Source.Host.Type != tt.expectedType {
			t.Errorf("Expected host type %v, got %v for '%s'", tt.expectedType, route.Source.Host.Type, tt.input)
		}

		if route.Source.Host.Value != tt.expectedVal {
			t.Errorf("Expected host value '%s', got '%s' for '%s'", tt.expectedVal, route.Source.Host.Value, tt.input)
		}
	}
}

func TestParsePathWildcard(t *testing.T) {
	input := "example.com /api/* => http://backend:8080/v1/$1"
	routes, err := ParseDSL(input)

	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Source.Path.Type != PatternWildcard {
		t.Errorf("Expected PatternWildcard, got %v", route.Source.Path.Type)
	}

	if route.Source.Path.Value != "/api/*" {
		t.Errorf("Expected path '/api/*', got '%s'", route.Source.Path.Value)
	}

	if len(route.Source.Path.Captures) != 1 {
		t.Errorf("Expected 1 capture, got %d", len(route.Source.Path.Captures))
	}
}

func TestParseNamedCaptures(t *testing.T) {
	input := "example.com /users/:id/posts/:postId => http://backend:8080/api"
	routes, err := ParseDSL(input)

	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(routes) != 1 {
		t.Fatalf("Expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Source.Path.Type != PatternNamed {
		t.Errorf("Expected PatternNamed, got %v", route.Source.Path.Type)
	}

	if len(route.Source.Path.Captures) != 2 {
		t.Errorf("Expected 2 captures, got %d", len(route.Source.Path.Captures))
	}

	if route.Source.Path.Captures[0] != "id" {
		t.Errorf("Expected capture 'id', got '%s'", route.Source.Path.Captures[0])
	}

	if route.Source.Path.Captures[1] != "postId" {
		t.Errorf("Expected capture 'postId', got '%s'", route.Source.Path.Captures[1])
	}
}

func TestParseModifiers(t *testing.T) {
	tests := []struct {
		input         string
		expectedCode  int
		expectedPrio  int
		expectedStrip bool
	}{
		{"old.com -> https://new.com [302]", 302, 0, false},
		{"old.com -> https://new.com [301]", 301, 0, false},
		{"api.com => http://backend [p:10]", 0, 10, false},
		{"api.com => http://backend [priority:5]", 0, 5, false},
		{"api.com /api/* => http://backend [strip_path]", 0, 0, true},
		{"api.com => http://backend [p:10, strip_path]", 0, 10, true},
	}

	for _, tt := range tests {
		routes, err := ParseDSL(tt.input)
		if err != nil {
			t.Fatalf("Failed to parse '%s': %v", tt.input, err)
		}

		route := routes[0]

		// Check status code
		if tt.expectedCode > 0 {
			found := false
			for _, mod := range route.Modifiers {
				if mod.Name == "302" || mod.Name == "301" {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Expected status code modifier for '%s'", tt.input)
			}
		}

		// Check priority
		if tt.expectedPrio > 0 && route.Priority != tt.expectedPrio {
			t.Errorf("Expected priority %d, got %d for '%s'", tt.expectedPrio, route.Priority, tt.input)
		}

		// Check strip_path
		if route.Target.StripPath != tt.expectedStrip {
			t.Errorf("Expected strip_path %v, got %v for '%s'", tt.expectedStrip, route.Target.StripPath, tt.input)
		}
	}
}

func TestParseMultipleRoutes(t *testing.T) {
	input := `
# Redirects
old.example.com -> https://new.example.com

# API Gateway
api.example.com /v1/* => http://backend:8080/api/$1

# Block bad actors
bad.example.com !! 403
`

	routes, err := ParseDSL(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	if len(routes) != 3 {
		t.Fatalf("Expected 3 routes, got %d", len(routes))
	}

	// Check that we have all three operators (order may vary due to priority sorting)
	operators := make(map[OperatorType]int)
	for _, route := range routes {
		operators[route.Operator]++
	}

	if operators[OperatorRedirect] != 1 {
		t.Errorf("Expected 1 redirect, got %d", operators[OperatorRedirect])
	}
	if operators[OperatorProxy] != 1 {
		t.Errorf("Expected 1 proxy, got %d", operators[OperatorProxy])
	}
	if operators[OperatorBlock] != 1 {
		t.Errorf("Expected 1 block, got %d", operators[OperatorBlock])
	}
}

func TestCompileDSLToConfig(t *testing.T) {
	input := `
old.example.com -> https://new.example.com
api.example.com => http://backend:8080
bad.example.com !! 403
`

	routes, err := ParseDSL(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	config, err := CompileDSLToConfig(routes)
	if err != nil {
		t.Fatalf("Failed to compile: %v", err)
	}

	if len(config.Redirects) != 2 { // redirect + block
		t.Errorf("Expected 2 redirects, got %d", len(config.Redirects))
	}

	if len(config.Proxies) != 1 {
		t.Errorf("Expected 1 proxy, got %d", len(config.Proxies))
	}
}

func TestPriorityCalculation(t *testing.T) {
	input := `
*.example.com => http://catchall:8080
api.example.com /v1/* => http://backend:8080
api.example.com => http://api:8080
example.com /health => http://health:8080
`

	routes, err := ParseDSL(input)
	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	// Debug: print routes in order
	for i, route := range routes {
		t.Logf("Route %d: %s%s (priority: %d)", i, hostPatternToString(route.Source.Host), route.Source.Path.Value, route.Priority)
	}

	// Routes should be sorted by priority (lower number = higher priority)
	// Most specific should come first
	if routes[0].Source.Host.Value != "example.com" || routes[0].Source.Path.Value != "/health" {
		t.Errorf("Expected most specific route (example.com /health) to be first, got %s%s",
			hostPatternToString(routes[0].Source.Host), routes[0].Source.Path.Value)
	}

	// Least specific should come last
	lastRoute := routes[len(routes)-1]
	if lastRoute.Source.Host.Type != PatternWildcardPrefix {
		t.Errorf("Expected wildcard route to be last, got %s (type: %v)",
			hostPatternToString(lastRoute.Source.Host), lastRoute.Source.Host.Type)
	}
}

func TestInvalidSyntax(t *testing.T) {
	tests := []string{
		"invalid syntax without operator",
		"example.com",                  // No operator
		"example.com > http://backend", // Wrong operator
	}

	for _, input := range tests {
		_, err := ParseDSL(input)
		if err == nil {
			t.Errorf("Expected error for invalid input: '%s'", input)
		}
	}

	// Empty input should return empty routes, not error
	routes, err := ParseDSL("")
	if err != nil {
		t.Errorf("Empty input should not error: %v", err)
	}
	if len(routes) != 0 {
		t.Errorf("Empty input should return 0 routes, got %d", len(routes))
	}
}

func TestStripPathIndicator(t *testing.T) {
	input := "example.com /api/* => http://backend:8080/_"
	routes, err := ParseDSL(input)

	if err != nil {
		t.Fatalf("Failed to parse: %v", err)
	}

	route := routes[0]
	if !route.Target.StripPath {
		t.Error("Expected strip_path to be true when target ends with /_")
	}

	if route.Target.URL != "http://backend:8080" {
		t.Errorf("Expected URL without /_, got '%s'", route.Target.URL)
	}
}
