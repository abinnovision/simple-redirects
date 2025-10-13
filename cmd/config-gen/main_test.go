package main

import (
	"os"
	"testing"
)

func TestGetTemplatePath(t *testing.T) {
	// Test with environment variable
	testPath := "/custom/path/template.tmpl"
	os.Setenv("TEMPLATE_PATH", testPath)
	defer os.Unsetenv("TEMPLATE_PATH")

	result := getTemplatePath()
	if result != testPath {
		t.Errorf("Expected template path '%s', got '%s'", testPath, result)
	}

	// Test without environment variable (should return default)
	os.Unsetenv("TEMPLATE_PATH")
	result = getTemplatePath()
	// The result will vary based on file existence, but should not be empty
	if result == "" {
		t.Error("Expected non-empty template path")
	}
}
