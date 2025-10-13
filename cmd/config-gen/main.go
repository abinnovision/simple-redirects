package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"
)

type Redirect struct {
	Host   string `json:"host"`
	Target string `json:"target"`
	Code   int    `json:"code"`
}

type Proxy struct {
	Host      string `json:"host"`
	Target    string `json:"target"`
	PathFrom  string `json:"path_from,omitempty"`
	PathTo    string `json:"path_to,omitempty"`
	StripPath bool   `json:"strip_path,omitempty"`
}

type Config struct {
	Redirects []Redirect `json:"redirects"`
	Proxies   []Proxy    `json:"proxies"`
}

func getTemplatePath() string {
	// Try environment variable first
	if path := os.Getenv("TEMPLATE_PATH"); path != "" {
		return path
	}

	// Try current directory
	if _, err := os.Stat("nginx.conf.tmpl"); err == nil {
		return "nginx.conf.tmpl"
	}

	// Try executable directory
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		tmplPath := filepath.Join(exeDir, "nginx.conf.tmpl")
		if _, err := os.Stat(tmplPath); err == nil {
			return tmplPath
		}
	}

	// Default fallback
	return "nginx.conf.tmpl"
}

func main() {
	config := Config{}

	// Priority order: DSL file > DSL inline

	// 1. Check for DSL file
	if dslFile := os.Getenv("ROUTES_FILE"); dslFile != "" {
		content, err := os.ReadFile(dslFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading ROUTES_FILE %s: %v\n", dslFile, err)
			os.Exit(1)
		}

		routes, err := ParseDSL(string(content))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing DSL routes: %v\n", err)
			os.Exit(1)
		}

		config, err = CompileDSLToConfig(routes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error compiling routes: %v\n", err)
			os.Exit(1)
		}
	} else if routesDSL := os.Getenv("ROUTES_DSL"); routesDSL != "" {
		// 2. Check for inline DSL
		routes, err := ParseDSL(routesDSL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing ROUTES_DSL: %v\n", err)
			os.Exit(1)
		}

		config, err = CompileDSLToConfig(routes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error compiling routes: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Fprintf(os.Stderr, "Error: No configuration provided\n")
		fmt.Fprintf(os.Stderr, "Set one of: ROUTES_FILE or ROUTES_DSL\n")
		os.Exit(1)
	}

	// Generate nginx config
	templatePath := getTemplatePath()
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing template: %v\n", err)
		os.Exit(1)
	}

	outputPath := os.Getenv("OUTPUT_PATH")
	if outputPath == "" {
		outputPath = "/etc/nginx/nginx.conf"
	}

	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating nginx config at %s: %v\n", outputPath, err)
		os.Exit(1)
	}
	defer f.Close()

	if err := tmpl.Execute(f, config); err != nil {
		fmt.Fprintf(os.Stderr, "Error executing template: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generated nginx.conf successfully:")
	fmt.Printf("- %d redirects\n", len(config.Redirects))
	fmt.Printf("- %d proxies\n", len(config.Proxies))
}
