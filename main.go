package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	// Import our internal packages
	"github.com/muquit/go-xbuild-go/internal/build"
	"github.com/muquit/go-xbuild-go/internal/config"
	"github.com/muquit/go-xbuild-go/internal/release"

	"github.com/muquit/go-xbuild-go/pkg/version"
)

const (
	me  = "go-xbuild-go"
	url = "https://github.com/muquit/go-xbuild-go"
)

var buildForPi = true

func main() {
	// --- 1. Define All Flags ---
	var (
		showVersion     bool
		showHelp        bool
		makeRelease     bool
		listTargets     bool
		configFile      string
		releaseNote     string
		releaseNoteFile string
		buildArgs       string
		additionalFiles string
		platformsFile   string
	)

	flag.StringVar(&buildArgs, "build-args", "", "Additional go build arguments (e.g., '-tags systray')")
	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&showHelp, "help", false, "Show help information and exit")
	flag.BoolVar(&buildForPi, "pi", true, "Build for Raspberry Pi")
	flag.BoolVar(&makeRelease, "release", false, "Create a GitHub release")
	flag.StringVar(&releaseNote, "release-note", "", "Release note text")
	flag.StringVar(&releaseNoteFile, "release-note-file", "", "File containing release notes")
	flag.StringVar(&additionalFiles, "additional-files", "", "Comma-separated list of additional files to include")
	flag.StringVar(&configFile, "config", "", "Path to build configuration file (JSON)")
	flag.StringVar(&platformsFile, "platforms-file", "platforms.txt", "Path of platforms.txt")
	flag.BoolVar(&listTargets, "list-targets", false, "List available build targets and exit")

	flag.Usage = func() {
		out := os.Stderr
		// Show help on stdout if -h, -help, or --help is used
		for _, arg := range os.Args[1:] {
			if arg == "-h" || arg == "-help" || arg == "--help" {
				out = os.Stdout
				break
			}
		}

		fmt.Fprintf(out, "%s %s\n", me, version.Get())
		fmt.Fprintf(out, "A program to cross compile go programs and release any software to github\n\n")
		fmt.Fprintf(out, "Usage:\n")
		fmt.Fprintf(out, "  %s [options]             # Build using defaults or config file\n", me)
		fmt.Fprintf(out, "  %s -config build-config.json   # Build using custom config\n", me)
		fmt.Fprintf(out, "  %s -release               # Create GitHub release from ./bin\n\n", me)

		fmt.Fprintf(out, "Options:\n")
		flag.CommandLine.SetOutput(out)
		flag.PrintDefaults()

		fmt.Fprintf(out, "\nEnvironment Variables (for GitHub release):\n")
		fmt.Fprintf(out, "  GITHUB_TOKEN      GitHub API token (required for -release)\n")
		fmt.Fprintf(out, "  GH_CLI_PATH       Custom path to GitHub CLI executable (optional)\n")

		fmt.Fprintf(out, "\nConfig File:\n")
		fmt.Fprintf(out, "  Optional JSON file for advanced, multi-target builds.\n\n")

		fmt.Fprintf(out, "A minimal example config file (build-config.json):\n")
		fmt.Fprintf(out, `{
  "project_name": "myproject",
  "version_file": "VERSION",
  "targets": [
    { "name": "cli", "path": "./cmd/cli" },
    { "name": "server", "path": "./cmd/server" }
  ]
}
`)
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

	// --- 2. Load and Prepare Configuration ---
	myDir, err := os.Getwd()
	if err != nil {
		fail("Could not get current directory: " + err.Error())
	}

	cfg := config.Config{
		ProjectName:   filepath.Base(myDir),
		BinDir:        filepath.Join(myDir, "bin"),
		VersionFile:   filepath.Join(myDir, "VERSION"),
		PlatformsFile: platformsFile,
		ChecksumsFile: "checksums.txt", // <-- THIS LINE WAS MISSING
	}

	if configFile != "" {
		projectConfig, err := config.LoadProjectConfig(configFile, myDir)
		if err != nil {
			fail("Failed to load config file: " + err.Error())
		}
		cfg.ProjectConfig = projectConfig
		if projectConfig.ProjectName != "" {
			cfg.ProjectName = projectConfig.ProjectName
		}
		if projectConfig.PlatformsFile != "" {
			cfg.PlatformsFile = projectConfig.PlatformsFile
		}
	}

	if buildArgs != "" {
		parsedArgs, err := build.ParseArguments(buildArgs)
		if err != nil {
			fail("Failed to parse build arguments: " + err.Error())
		}
		cfg.ExtraBuildArgs = parsedArgs
	}

	if additionalFiles != "" {
		cfg.AdditionalFiles = strings.Split(additionalFiles, ",")
		for i := range cfg.AdditionalFiles {
			cfg.AdditionalFiles[i] = strings.TrimSpace(cfg.AdditionalFiles[i])
		}
	}

	// --- 3. Execute Action ---
	if listTargets {
		if cfg.ProjectConfig == nil {
			fail("No project configuration file loaded. Use the -config flag.")
		}
		fmt.Printf("Available build targets for %s:\n", cfg.ProjectName)
		for _, target := range cfg.ProjectConfig.Targets {
			fmt.Printf("  - %s (path: %s)\n", target.Name, target.Path)
		}
		os.Exit(0)
	}

	if makeRelease {
		err = release.Create(&cfg, releaseNote, releaseNoteFile)
	} else {
		if cfg.ProjectConfig != nil {
			err = build.ProcessMultiTarget(&cfg, buildForPi)
		} else {
			err = build.Process(&cfg, buildForPi)
		}
	}

	if err != nil {
		fail(err.Error())
	}

	fmt.Println("Build process completed successfully.")
}

// fail is a simple helper that stays in the main package to handle program exit.
func fail(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}
