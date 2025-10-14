package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// BuildTarget represents a single binary to build.
type BuildTarget struct {
	Name            string   `json:"name"`              // Binary name (e.g., "cli", "server")
	Path            string   `json:"path"`              // Build path (e.g., "./cmd/cli", "./cmd/server")
	OutputName      string   `json:"output_name"`       // Custom output name (optional)
	LdFlags         string   `json:"ldflags"`           // Custom ldflags (optional)
	BuildFlags      string   `json:"build_flags"`       // Custom build flags (optional)
	AdditionalFiles []string `json:"additional_files"`  // Target-specific additional files
}

// ProjectConfig represents the configuration for a multi-binary project.
type ProjectConfig struct {
	ProjectName           string        `json:"project_name"`
	Version               string        `json:"version"`
	VersionFile           string        `json:"version_file"`
	PlatformsFile         string        `json:"platforms_file"`
	DefaultLdFlags        string        `json:"default_ldflags"`
	DefaultBuildFlags     string        `json:"default_build_flags"`
	GlobalAdditionalFiles []string      `json:"global_additional_files"`
	Targets               []BuildTarget `json:"targets"`
}

// Config holds all configuration for the build process.
type Config struct {
	ProjectName     string
	BinDir          string
	VersionFile     string
	PlatformsFile   string
	ChecksumsFile   string
	LdFlags         string
	BuildFlags      string
	AdditionalFiles []string
	ProjectConfig   *ProjectConfig // New: multi-target config
	ExtraBuildArgs  []string
}

// LoadProjectConfig loads a project configuration from a JSON file.
func LoadProjectConfig(configPath, baseDir string) (*ProjectConfig, error) {
	// Make path absolute if it's relative
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(baseDir, configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config ProjectConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	// Validate configuration
	if len(config.Targets) == 0 {
		return nil, fmt.Errorf("no build targets specified in config")
	}

	for i, target := range config.Targets {
		if target.Name == "" {
			return nil, fmt.Errorf("target %d is missing a name", i)
		}
		if target.Path == "" {
			return nil, fmt.Errorf("target %s is missing a path", target.Name)
		}
	}

	return &config, nil
}