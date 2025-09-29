package main

/////////////////////////////////////////////////////////////////////
// Cross compile go programs for other platforms
// Enhanced version with support for multiple main packages
// Developed with Claude AI 3.7 Sonnet, working under my guidance and
// instructions.
// muquit@muquit.com Mar-26-2025
/// v1.1.0 - Enhanced for modular Go projects
/////////////////////////////////////////////////////////////////////

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/muquit/go-xbuild-go/pkg/version"
)

const (
	me      = "go-xbuild-go"
	url     = "https://github.com/muquit/go-xbuild-go"
)

var buildForPi = true

// BuildTarget represents a single binary to build
type BuildTarget struct {
	Name            string   `json:"name"`            // Binary name (e.g., "cli", "server")
	Path            string   `json:"path"`            // Build path (e.g., "./cmd/cli", "./cmd/server")
	OutputName      string   `json:"output_name"`     // Custom output name (optional)
	LdFlags         string   `json:"ldflags"`         // Custom ldflags (optional)
	BuildFlags      string   `json:"build_flags"`     // Custom build flags (optional)
	AdditionalFiles []string `json:"additional_files"` // Target-specific additional files
}

// ProjectConfig represents the configuration for a multi-binary project
type ProjectConfig struct {
	ProjectName     string        `json:"project_name"`
	Version         string        `json:"version"`
	VersionFile     string        `json:"version_file"`
	PlatformsFile   string        `json:"platforms_file"`
	DefaultLdFlags  string        `json:"default_ldflags"`
	DefaultBuildFlags string      `json:"default_build_flags"`
	GlobalAdditionalFiles []string `json:"global_additional_files"`
	Targets         []BuildTarget `json:"targets"`
}

// Configuration constants (legacy support)
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

func main() {
	var showVersion bool
	var showHelp bool
	var makeRelease bool
	var releaseNote string
	var releaseNoteFile string
	var additionalFiles string
	var configFile string
	var listTargets bool
	var buildArgs string
	var platformsFile string

	flag.StringVar(&buildArgs, "build-args", "", "Additional go build arguments (e.g., '-tags systray -race')")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&showHelp, "help", false, "Show help information and exit")
	flag.BoolVar(&buildForPi, "pi", true, "Build Raspberry Pi")
	flag.BoolVar(&makeRelease, "release", false, "Create a GitHub release")
	flag.StringVar(&releaseNote, "release-note", "", "Release note text (required if -release-note-file not specified and release_notes.md doesn't exist)")
	flag.StringVar(&releaseNoteFile, "release-note-file", "", "File containing release notes (required if -release-note not specified and release_notes.md doesn't exist)")
	flag.StringVar(&additionalFiles, "additional-files", "", "Comma-separated list of additional files to include in archives")
	flag.StringVar(&configFile, "config", "", "Path to build configuration file (JSON)")

	flag.StringVar(&platformsFile,"platforms-file","platforms.txt","Path of platforms.txt")

	flag.BoolVar(&listTargets, "list-targets", false, "List available build targets and exit")

flag.Usage = func() {
	// Determine output destination - stdout if help explicitly requested, stderr otherwise
	out := os.Stderr
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "-help" || arg == "--help" {
			out = os.Stdout
			break
		}
	}
	
	fmt.Fprintf(out, "%s %s\n", me, version.Get())
	fmt.Fprintf(out, "A program to cross compile go programs and release any software to github\n\n")
	
	fmt.Fprintf(out, "Usage:\n")
	fmt.Fprintf(out, "  %s [options]                    # Build using defaults or config file\n", me)
	fmt.Fprintf(out, "  %s -config build-config.json   # Build using custom config\n", me)
	fmt.Fprintf(out, "  %s -release                     # Create GitHub release from ./bin\n\n", me)
	
	fmt.Fprintf(out, "Quick Start:\n")
	fmt.Fprintf(out, "  1. Create/edit platforms.txt (uncomment desired platforms)\n")
	fmt.Fprintf(out, "  2. Create VERSION file (e.g., v1.0.1)\n")
	fmt.Fprintf(out, "  3. Run %s\n\n", me)
	
	fmt.Fprintf(out, "Options:\n")
	
	// Set flag output to match our choice
	flag.CommandLine.SetOutput(out)
	flag.PrintDefaults()
	
	fmt.Fprintf(out, "\nEnvironment Variables (for GitHub release):\n")
	fmt.Fprintf(out, "  GITHUB_TOKEN     GitHub API token (required for -release)\n")
	fmt.Fprintf(out, "  GH_CLI_PATH      Custom path to GitHub CLI executable (optional)\n")
	
	fmt.Fprintf(out, "\nAutomatically Included Files:\n")
	fmt.Fprintf(out, "  README.md, LICENSE.txt, LICENSE, platforms.txt, <project>.1\n")
	fmt.Fprintf(out, "  (Don't specify these in -additional-files)\n")
	
	fmt.Fprintf(out, "\nConfig File:\n")
	fmt.Fprintf(out, "  Optional JSON file for advanced configuration (any project type, single or multi main).\n")
	fmt.Fprintf(out, "  Useful for multi-target builds, custom flags, or organized projects.\n\n")
	
	fmt.Fprintf(out, "A Minimal example config file (build-config.json):\n")
	fmt.Fprintf(out, `{
  "project_name": "myproject",
  "version_file": "VERSION",
  "platforms_file": "platforms.txt",
  "default_ldflags": "-s -w",
  "default_build_flags": "-trimpath",
  "targets": [
    {
      "name": "cli",
      "path": "./cmd/cli",
      "output_name": "mycli"
    },
    {
      "name": "server",
      "path": "./cmd/server",
      "output_name": "myserver"
    }
  ]
}
`)

	fmt.Fprintf(out, "Please consult documentaiton for details\n")
}


	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s\n", me, version.Get())
		os.Exit(0)
	}

	if showHelp {
		flag.Usage()
		os.Exit(0)
	}

	// Get current directory
	myDir, err := os.Getwd()
	if err != nil {
		fail("Could not get current directory: " + err.Error())
	}

	// Set up configuration
	config := Config{
		ProjectName:   filepath.Base(myDir),
		BinDir:        filepath.Join(myDir, "bin"),
		VersionFile:   filepath.Join(myDir, "VERSION"),
		PlatformsFile: filepath.Join(myDir, "platforms.txt"),
		ChecksumsFile: "checksums.txt",
		LdFlags:       "-s -w",
		BuildFlags:    "-trimpath",
	}

	// specify an alternate one
	config.PlatformsFile = platformsFile

	// Load project configuration if specified
	if configFile != "" {
		projectConfig, err := loadProjectConfig(configFile, myDir)
		if err != nil {
			fail("Failed to load config file: " + err.Error())
		}
		config.ProjectConfig = projectConfig
		
		// Override some values from project config
		if projectConfig.ProjectName != "" {
			config.ProjectName = projectConfig.ProjectName
		}
		if projectConfig.VersionFile != "" {
			if filepath.IsAbs(projectConfig.VersionFile) {
				config.VersionFile = projectConfig.VersionFile
			} else {
				config.VersionFile = filepath.Join(myDir, projectConfig.VersionFile)
			}
		}
		if projectConfig.PlatformsFile != "" {
			if filepath.IsAbs(projectConfig.PlatformsFile) {
				config.PlatformsFile = projectConfig.PlatformsFile
			} else {
				config.PlatformsFile = filepath.Join(myDir, projectConfig.PlatformsFile)
			}
		}
	}
	// parse additional build args
	if buildArgs != "" {
	parsedArgs, err := parseArguments(buildArgs)
	if err != nil {
		fail("Failed to parse build arguments: " + err.Error())
	}
	config.ExtraBuildArgs = parsedArgs
}

	// Handle additional files from command line
	if additionalFiles != "" {
		config.AdditionalFiles = strings.Split(additionalFiles, ",")
		// Trim spaces from each file path
		for i, file := range config.AdditionalFiles {
			config.AdditionalFiles[i] = strings.TrimSpace(file)
		}
	}

	// List targets if requested
	if listTargets {
		if config.ProjectConfig != nil {
			fmt.Printf("Available build targets for %s:\n", config.ProjectName)
			for _, target := range config.ProjectConfig.Targets {
				fmt.Printf("  - %s (path: %s)\n", target.Name, target.Path)
			}
		} else {
			fmt.Printf("No multi-target configuration found. Running in legacy single-binary mode.\n")
			fmt.Printf("Target: %s (current directory)\n", config.ProjectName)
		}
		os.Exit(0)
	}

	// ./bin must have the archives for release
	if makeRelease {
		err = createRelease(&config, releaseNote, releaseNoteFile)
		if err != nil {
			fail(err.Error())
		}
		os.Exit(0)
	}

	// Otherwise, run the main process
	if config.ProjectConfig != nil {
		fmt.Printf("Building multi-target project: %s\n", config.ProjectName)
		err = processMultiTarget(&config)
	} else {
		fmt.Printf("Building single-target project: %s\n", config.ProjectName)
		err = process(&config)
	}
	
	if err != nil {
		fail(err.Error())
	}
}

// Load project configuration from JSON file
func loadProjectConfig(configPath, baseDir string) (*ProjectConfig, error) {
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

// Process multi-target builds
func processMultiTarget(config *Config) error {
	// Initialize
	if err := initialize(config); err != nil {
		return err
	}

	// Get version
	version, err := getVersion(config)
	if err != nil {
		return err
	}

	projectConfig := config.ProjectConfig
	fmt.Printf("Building %s version %s with %d targets\n", config.ProjectName, version, len(projectConfig.Targets))
	fmt.Printf("The binaries are cross compiled with %s\n", url)

	// Build each target
	for _, target := range projectConfig.Targets {
		fmt.Printf("\n=== Building target: %s ===\n", target.Name)
		
		// Create target-specific config
		targetConfig := *config
		targetConfig.ProjectName = target.Name
		if target.OutputName != "" {
			targetConfig.ProjectName = target.OutputName
		}

		// Set target-specific build parameters
		targetConfig.LdFlags = projectConfig.DefaultLdFlags
		if target.LdFlags != "" {
			targetConfig.LdFlags = target.LdFlags
		}
		
		targetConfig.BuildFlags = projectConfig.DefaultBuildFlags
		if target.BuildFlags != "" {
			targetConfig.BuildFlags = target.BuildFlags
		}

		// Combine global and target-specific additional files
		targetConfig.AdditionalFiles = append(projectConfig.GlobalAdditionalFiles, target.AdditionalFiles...)
		targetConfig.AdditionalFiles = append(targetConfig.AdditionalFiles, config.AdditionalFiles...) // Add CLI files

		// Clean existing checksums for this target
		checksumFile := filepath.Join(config.BinDir, fmt.Sprintf("%s-%s-%s", targetConfig.ProjectName, version, config.ChecksumsFile))
		if err := os.RemoveAll(checksumFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old checksums file: %v", err)
		}

		// Build for platforms in platforms.txt
		if err := buildForPlatformsWithPath(&targetConfig, version, target.Path); err != nil {
			return fmt.Errorf("failed to build target %s: %v", target.Name, err)
		}

		if buildForPi {
			// Build for Raspberry Pi variants
			if err := buildPiWithPath(&targetConfig, version, target.Path, "", "7"); err != nil {
				return fmt.Errorf("failed to build target %s for Pi: %v", target.Name, err)
			}
			if err := buildPiWithPath(&targetConfig, version, target.Path, "-jessie", "6"); err != nil {
				return fmt.Errorf("failed to build target %s for Pi Jessie: %v", target.Name, err)
			}
		}

		fmt.Printf("Target %s build complete\n", target.Name)
	}

	fmt.Printf("\nAll targets build complete. Artifacts are in %s\n", config.BinDir)
	return nil
}

// Build for platforms with custom path
func buildForPlatformsWithPath(config *Config, version, buildPath string) error {
	file, err := os.Open(config.PlatformsFile)
	if err != nil {
		return fmt.Errorf("failed to open platforms file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "/")
		if len(parts) < 2 {
			continue
		}

		goos := parts[0]
		goarch := parts[1]

		fmt.Printf("\n> Building for %s/%s\n", goos, goarch)

		distDir := fmt.Sprintf("%s-%s-%s-%s.d", config.ProjectName, version, goos, goarch)
		binaryName := fmt.Sprintf("%s-%s-%s-%s", config.ProjectName, version, goos, goarch)
		if goos == "windows" {
			binaryName += ".exe"
		}

		// Set environment variables
		env := []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
		}

		// Build binary with custom path
		if err := gobuildWithPath(config, binaryName, buildPath, env); err != nil {
			return fmt.Errorf("failed to build for %s/%s: %v", goos, goarch, err)
		}

		// Copy files
		if err := copyFiles(config, binaryName, distDir); err != nil {
			return err
		}

		// Create archive
		if err := createArchive(config, version, distDir, goos); err != nil {
			return err
		}

		// Remove binary
		if err := os.Remove(binaryName); err != nil {
			return fmt.Errorf("failed to remove binary: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read platforms file: %v", err)
	}

	return nil
}

// Build for Raspberry Pi with custom path
func buildPiWithPath(config *Config, version, buildPath, variant, armVersion string) error {
	distDir := fmt.Sprintf("%s-%s-raspberry-pi%s.d", config.ProjectName, version, variant)
	binaryName := fmt.Sprintf("%s-%s-raspberry-pi%s", config.ProjectName, version, variant)

	// Set environment variables
	env := []string{
		"GOOS=linux",
		"GOARCH=arm",
		"GOARM=" + armVersion,
	}

	fmt.Printf("\n> Building for raspberry pi%s (arm%s)\n", variant, armVersion)

	// Build binary with custom path
	if err := gobuildWithPath(config, binaryName, buildPath, env); err != nil {
		return fmt.Errorf("failed to build for raspberry pi: %v", err)
	}

	// Copy files
	if err := copyFiles(config, binaryName, distDir); err != nil {
		return err
	}

	// Create archive
	if err := createArchive(config, version, distDir, "linux"); err != nil {
		return err
	}

	// Remove binary
	return os.Remove(binaryName)
}

// new--Sep-14-2025 
func gobuildWithPath(config *Config, output, buildPath string, env []string) error {
	args := []string{"build"}

	// Add ldflags if specified
	if config.LdFlags != "" {
		args = append(args, "-ldflags="+config.LdFlags)
	}

	// Parse and add build flags
	if config.BuildFlags != "" {
		buildFlagArgs, err := parseArguments(config.BuildFlags)
		if err != nil {
			return fmt.Errorf("failed to parse build flags: %v", err)
		}
		args = append(args, buildFlagArgs...)
	}

	// Add extra build args from command line
	if len(config.ExtraBuildArgs) > 0 {
		args = append(args, config.ExtraBuildArgs...)
	}

	// Add output flag
	args = append(args, "-o", output)

	// Add build path if specified
	if buildPath != "" && buildPath != "." {
		args = append(args, buildPath)
	}

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
// new--Sep-14-2025 

// Helper function to run go build with custom path
func gobuildWithPathOld(config *Config, output, buildPath string, env []string) error {
	args := []string{
		"build",
		"-ldflags=" + config.LdFlags,
		config.BuildFlags,
		"-o", output,
	}
	
	// Add build path if specified
	if buildPath != "" && buildPath != "." {
		args = append(args, buildPath)
	}

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Create a GitHub release
func createRelease(config *Config, note, noteFile string) error {
	// Check if GitHub CLI exists
	if err := checkGhCliExists(); err != nil {
		return err
	}

	// Check if GITHUB_TOKEN is set
	if os.Getenv("GITHUB_TOKEN") == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	// Get version
	version, err := getVersion(config)
	if err != nil {
		return err
	}

	// Check if bin directory exists and is not empty
	if _, err := os.Stat(config.BinDir); os.IsNotExist(err) {
		return fmt.Errorf("bin directory does not exist")
	}

	// Check if bin directory has files
	files, err := os.ReadDir(config.BinDir)
	if err != nil {
		return fmt.Errorf("failed to read bin directory: %v", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("bin directory is empty")
	}

	// Get the appropriate gh command
	ghCmd := getGhCommand()

	// Prepare the command arguments
	args := []string{
		"release",
		"create",
		version,
	}

	// Handle release notes
	if note != "" && noteFile != "" {
		// Create a temporary file for combined notes
		tempFile, err := os.CreateTemp("", "release-notes-*.md")
		if err != nil {
			return fmt.Errorf("failed to create temporary file for release notes: %v", err)
		}
		defer os.Remove(tempFile.Name())

		// Write the inline note
		if _, err := tempFile.WriteString(note + "\n\n"); err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to write inline note to temporary file: %v", err)
		}

		// Append the file content
		fileContent, err := os.ReadFile(noteFile)
		if err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to read release note file: %v", err)
		}

		if _, err := tempFile.Write(fileContent); err != nil {
			tempFile.Close()
			return fmt.Errorf("failed to write file content to temporary file: %v", err)
		}

		tempFile.Close()
		args = append(args, "--notes-file", tempFile.Name())
	} else if note != "" {
		// Use the --notes flag directly
		args = append(args, "--notes", note)
	} else if noteFile != "" {
		// Check if file exists
		if _, err := os.Stat(noteFile); os.IsNotExist(err) {
			return fmt.Errorf("release note file does not exist: %s", noteFile)
		}
		args = append(args, "--notes-file", noteFile)
	} else {
		// Check if default release notes file exists
		defaultNoteFile := "release_notes.md"
		if _, err := os.Stat(defaultNoteFile); os.IsNotExist(err) {
			return fmt.Errorf("no release notes specified and default file '%s' does not exist", defaultNoteFile)
		}
		args = append(args, "--notes-file", defaultNoteFile)
	}

	// Add all files from bin directory
	args = append(args, filepath.Join(config.BinDir, "*"))

	// Execute gh release create command
	fmt.Printf("Creating GitHub release %s\n", version)
	cmd := exec.Command(ghCmd, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create GitHub release: %v", err)
	}

	// List releases
	listCmd := exec.Command(ghCmd, "release", "list")
	listCmd.Stdout = os.Stdout
	listCmd.Stderr = os.Stderr

	if err := listCmd.Run(); err != nil {
		return fmt.Errorf("failed to list GitHub releases: %v", err)
	}

	fmt.Printf("GitHub release %s created successfully\n", version)
	return nil
}

// Get the GitHub CLI command with precedence:
// 1. Environment variable (if set)
// 2. Platform-specific executable name
// 3. Default "gh" command
func getGhCommand() string {
	// Check environment variable first
	if envCmd := os.Getenv("GH_CLI_PATH"); envCmd != "" {
		// Verify the specified path exists and is executable
		if _, err := os.Stat(envCmd); err == nil {
			return envCmd
		}
		// If env var is set but invalid, print a warning
		fmt.Fprintf(os.Stderr, "Warning: GH_CLI_PATH environment variable is set to '%s' but file doesn't exist or isn't executable\n", envCmd)
	}

	// Platform-specific checks
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("gh.exe"); err == nil {
			return "gh.exe"
		}
	}

	// Default to "gh"
	return "gh"
}

// Check if GitHub CLI exists (using all possible methods)
func checkGhCliExists() error {
	cmd := getGhCommand()

	if _, err := exec.LookPath(cmd); err != nil {
		return fmt.Errorf("GitHub CLI (%s) is not installed or not in PATH", cmd)
	}

	return nil
}

// Process handles the main build process (legacy single-target mode)
func process(config *Config) error {
	// Initialize
	if err := initialize(config); err != nil {
		return err
	}

	// Get version
	version, err := getVersion(config)
	if err != nil {
		return err
	}

	fmt.Printf("Building %s version %s\n", config.ProjectName, version)
	fmt.Printf("%s version %s\n", config.ProjectName, version)
	fmt.Printf("The binaries are cross compiled with %s\n", url)

	// Clean existing checksums
	checksumFile := filepath.Join(config.BinDir, fmt.Sprintf("%s-%s-%s", config.ProjectName, version, config.ChecksumsFile))
	if err := os.RemoveAll(checksumFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old checksums file: %v", err)
	}

	// Build for platforms in platforms.txt
	if err := buildForPlatforms(config, version); err != nil {
		return err
	}

	if buildForPi {
		// Build for Raspberry Pi variants
		if err := buildPi(config, version, "", "7"); err != nil { // modern pi
			return err
		}
		if err := buildPi(config, version, "-jessie", "6"); err != nil { // pi jessie
			return err
		}
	}

	fmt.Printf("Build complete. Artifacts are in %s\n", config.BinDir)
	return nil
}

// Initialize and verify required files exist
func initialize(config *Config) error {
	// Check that version file exists
	if _, err := os.Stat(config.VersionFile); os.IsNotExist(err) {
		return fmt.Errorf("version file not found: %s", config.VersionFile)
	}

	// Check that platforms file exists
	if _, err := os.Stat(config.PlatformsFile); os.IsNotExist(err) {
		return fmt.Errorf("platforms file not found: %s", config.PlatformsFile)
	}

	// Create bin directory if it doesn't exist
	if err := os.MkdirAll(config.BinDir, 0755); err != nil {
		return fmt.Errorf("could not create bin directory: %s, error: %v", config.BinDir, err)
	}

	return nil
}

// Get version from VERSION file
func getVersion(config *Config) (string, error) {
	content, err := os.ReadFile(config.VersionFile)
	if err != nil {
		return "", fmt.Errorf("failed to read version file: %v", err)
	}
	return strings.TrimSpace(string(content)), nil
}

// Print error and exit
func fail(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}

// Calculate sha256 checksum and append to checksums file
func takeChecksum(config *Config, version, archive string) error {
	checksumFilename := fmt.Sprintf("%s-%s-%s", config.ProjectName, version, config.ChecksumsFile)

	// Change to bin directory to get relative paths
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}

	if err := os.Chdir(config.BinDir); err != nil {
		return fmt.Errorf("failed to change to bin directory: %v", err)
	}
	defer os.Chdir(currentDir)

	// Read the file
	data, err := os.ReadFile(archive)
	if err != nil {
		return fmt.Errorf("failed to read archive for checksum: %v", err)
	}

	// Calculate checksum
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	// Append to checksum file
	f, err := os.OpenFile(checksumFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open checksum file: %v", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "%s  %s\n", checksum, archive); err != nil {
		return fmt.Errorf("failed to write to checksum file: %v", err)
	}

	return nil
}

// Copy required files to distribution directory
func copyFiles(config *Config, bin, distDir string) error {
	// Create dist directory
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("failed to create dist directory: %v", err)
	}

	// Copy binary
	if err := copyFile(bin, filepath.Join(distDir, filepath.Base(bin))); err != nil {
		return fmt.Errorf("failed to copy binary: %v", err)
	}

	// Copy documentation files if they exist
	docFiles := map[string]string{
		filepath.Join(filepath.Dir(config.BinDir), "README.md"):                     filepath.Join(distDir, "README.md"),
		filepath.Join(filepath.Dir(config.BinDir), "docs", config.ProjectName+".1"): filepath.Join(distDir, config.ProjectName+".1"),
		filepath.Join(filepath.Dir(config.BinDir), "LICENSE.txt"):                   filepath.Join(distDir, "LICENSE.txt"),
		filepath.Join(filepath.Dir(config.BinDir), "LICENSE"):                       filepath.Join(distDir, "LICENSE"),
		filepath.Join(filepath.Dir(config.BinDir), "platforms.txt"):                 filepath.Join(distDir, "platforms.txt"),
	}

	for src, dst := range docFiles {
		if _, err := os.Stat(src); err == nil {
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to copy %s: %v", src, err)
			}
		}
	}

	// Copy additional files if specified
	/*
	for _, filePath := range config.AdditionalFiles {
		if _, err := os.Stat(filePath); err == nil {
			destPath := filepath.Join(distDir, filepath.Base(filePath))
			if err := copyFile(filePath, destPath); err != nil {
				return fmt.Errorf("failed to copy additional file %s: %v", filePath, err)
			}
			fmt.Printf("Added additional file: %s\n", filePath)
		} else {
			fmt.Printf("Warning: additional file not found: %s\n", filePath)
		}
	}
	*/
	for _, filePath := range config.AdditionalFiles {
		if _, err := os.Stat(filePath); err == nil {
			destPath := filepath.Join(distDir, filepath.Base(filePath))
			if err := copyFile(filePath, destPath); err != nil {
				return fmt.Errorf("failed to copy additional file %s: %v", filePath, err)
			}
			fmt.Printf("Added additional file: %s\n", filePath)
		} else {
			fmt.Printf("Warning: additional file not found: %s\n", filePath)
		}
	}

	return nil
}

// Helper function to copy a file
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// Remove temporary directory
func cleanupDir(dir string) error {
	if _, err := os.Stat(dir); err == nil {
		return os.RemoveAll(dir)
	}
	return nil
}

// Create archive (zip for windows, tar.gz for others)
func createArchive(config *Config, version, distDir, goos string) error {
	var archiveName string

	// Create appropriate archive based on OS
	if goos == "windows" {
		archiveName = distDir + ".zip"
		if err := zipDir(distDir, archiveName); err != nil {
			return fmt.Errorf("failed to create zip archive: %v", err)
		}
	} else {
		archiveName = distDir + ".tar.gz"
		if err := tarGzDir(distDir, archiveName); err != nil {
			return fmt.Errorf("failed to create tar archive: %v", err)
		}
	}

	// Move archive to bin directory
	finalArchivePath := filepath.Join(config.BinDir, filepath.Base(archiveName))
	if err := moveFile(archiveName, finalArchivePath); err != nil {
		return fmt.Errorf("failed to move archive to bin directory: %v", err)
	}

	// Take checksum
	if err := takeChecksum(config, version, filepath.Base(archiveName)); err != nil {
		return fmt.Errorf("failed to take checksum: %v", err)
	}

	// Cleanup
	return cleanupDir(distDir)
}

// Helper function to move a file
func moveFile(src, dst string) error {
	return os.Rename(src, dst)
}

// Create a zip archive of a directory
func zipDir(srcDir, destZip string) error {
	zipFile, err := os.Create(destZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	return filepath.Walk(srcDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == srcDir {
			return nil
		}

		// Create a relative path as zip header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// Adjust the path in the archive
		relPath, err := filepath.Rel(filepath.Dir(srcDir), path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

// Create a tar.gz archive of a directory
func tarGzDir(srcDir, destTarGz string) error {
	tarGzFile, err := os.Create(destTarGz)
	if err != nil {
		return err
	}
	defer tarGzFile.Close()

	gzipWriter := gzip.NewWriter(tarGzFile)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(srcDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if path == srcDir {
			return nil
		}

		// Create tar header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		// Use relative path for the header name
		relPath, err := filepath.Rel(filepath.Dir(srcDir), path)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(relPath)

		// Write header
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a directory, just write the header
		if info.IsDir() {
			return nil
		}

		// If it's a file, write the contents
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})
}

// Build for Raspberry Pi (legacy single-target mode)
func buildPi(config *Config, version, variant, armVersion string) error {
	distDir := fmt.Sprintf("%s-%s-raspberry-pi%s.d", config.ProjectName, version, variant)
	binaryName := fmt.Sprintf("%s-%s-raspberry-pi%s", config.ProjectName, version, variant)

	// Set environment variables
	env := []string{
		"GOOS=linux",
		"GOARCH=arm",
		"GOARM=" + armVersion,
	}

	fmt.Printf("> Building for raspberry pi%s (arm%s)\n", variant, armVersion)

	// Build binary
	if err := gobuild(config, binaryName, env); err != nil {
		return fmt.Errorf("failed to build for raspberry pi: %v", err)
	}

	// Copy files
	if err := copyFiles(config, binaryName, distDir); err != nil {
		return err
	}

	// Create archive
	if err := createArchive(config, version, distDir, "linux"); err != nil {
		return err
	}

	// Remove binary
	return os.Remove(binaryName)
}
// new -Sep-14-2025 
func gobuild(config *Config, output string, env []string) error {
	args := []string{"build"}

	// Add ldflags if specified
	if config.LdFlags != "" {
		args = append(args, "-ldflags="+config.LdFlags)
	}

	// Parse and add build flags
	if config.BuildFlags != "" {
		buildFlagArgs, err := parseArguments(config.BuildFlags)
		if err != nil {
			return fmt.Errorf("failed to parse build flags: %v", err)
		}
		args = append(args, buildFlagArgs...)
	}

	// Add extra build args from command line
	if len(config.ExtraBuildArgs) > 0 {
		args = append(args, config.ExtraBuildArgs...)
	}

	// Add output flag
	args = append(args, "-o", output)

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
// new -Sep-14-2025 

// Helper function to run go build (legacy single-target mode)
func gobuildOld(config *Config, output string, env []string) error {
	args := []string{
		"build",
		"-ldflags=" + config.LdFlags,
		config.BuildFlags,
		"-o", output,
	}

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Build for platforms in platforms.txt (legacy single-target mode)
func buildForPlatforms(config *Config, version string) error {
	file, err := os.Open(config.PlatformsFile)
	if err != nil {
		return fmt.Errorf("failed to open platforms file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "/")
		if len(parts) < 2 {
			continue
		}

		goos := parts[0]
		goarch := parts[1]

		fmt.Printf("> Building for %s/%s\n", goos, goarch)

		distDir := fmt.Sprintf("%s-%s-%s-%s.d", config.ProjectName, version, goos, goarch)
		binaryName := fmt.Sprintf("%s-%s-%s-%s", config.ProjectName, version, goos, goarch)
		if goos == "windows" {
			binaryName += ".exe"
		}

		// Set environment variables
		env := []string{
			"GOOS=" + goos,
			"GOARCH=" + goarch,
		}

		// Build binary
		if err := gobuild(config, binaryName, env); err != nil {
			return fmt.Errorf("failed to build for %s/%s: %v", goos, goarch, err)
		}

		// Copy files
		if err := copyFiles(config, binaryName, distDir); err != nil {
			return err
		}

		// Create archive
		if err := createArchive(config, version, distDir, goos); err != nil {
			return err
		}

		// Remove binary
		if err := os.Remove(binaryName); err != nil {
			return fmt.Errorf("failed to remove binary: %v", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read platforms file: %v", err)
	}

	return nil
}

// parseArguments parses a string of build arguments, respecting quotes
func parseArguments(argStr string) ([]string, error) {
	var args []string
	var current strings.Builder
	var inQuotes bool
	var quoteChar rune

	argStr = strings.TrimSpace(argStr)
	if argStr == "" {
		return args, nil
	}

	for i, char := range argStr {
		switch {
		case char == '"' || char == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteRune(char)
			}
		case char == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(char)
		}

		// Handle end of string
		if i == len(argStr)-1 && current.Len() > 0 {
			args = append(args, current.String())
		}
	}

	if inQuotes {
		return nil, fmt.Errorf("unclosed quote in build arguments")
	}

	return args, nil
}
