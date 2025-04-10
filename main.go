package main

/////////////////////////////////////////////////////////////////////
// Cross compile go programs for other platforms
// Developed with Claude AI 3.7 Sonnet, working under my guidance and
// instructions.
// muquit@muquit.com Mar-26-2025
/// v1.0.4 - Apr-09-2025 
/////////////////////////////////////////////////////////////////////

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	version = "1.0.4"
	url     = "https://github.com/muquit/go-xbuild-go"
)

var buildForPi = true

// Configuration constants
type Config struct {
	ProjectName   string
	BinDir        string
	VersionFile   string
	PlatformsFile string
	ChecksumsFile string
	LdFlags       string
	BuildFlags    string
	AdditionalFiles []string 
}

func main() {
	var showVersion bool
	var showHelp bool
	var makeRelease bool
	var releaseNote string
	var releaseNoteFile string
	var additionalFiles string

	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&showHelp, "help", false, "Show help information and exit")
	flag.BoolVar(&buildForPi, "pi", true, "Build Raspberry Pi")

	flag.BoolVar(&makeRelease, "release", false, "Create a GitHub release")
	flag.StringVar(&releaseNote, "release-note", "", "Release note text (required if -release-note-file not specified and release_notes.md doesn't exist)")
	flag.StringVar(&releaseNoteFile, "release-note-file", "", "File containing release notes (required if -release-note not specified and release_notes.md doesn't exist)")
	flag.StringVar(&additionalFiles, "additional-files", "", "Comma-separated list of additional files to include in archives")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s v%s\n", os.Args[0], version)
		fmt.Fprintf(os.Stderr, "A program to cross compile go programs\n\n")
		fmt.Fprintf(os.Stderr, "Environment variables (for github release):\n")
		fmt.Fprintf(os.Stderr, "  GITHUB_TOKEN     GitHub API token (required for -release)\n")
		fmt.Fprintf(os.Stderr, "  GH_CLI_PATH      Custom path to GitHub CLI executable (optional)\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  - Copy platforms.txt at the root of your project\n")
		fmt.Fprintf(os.Stderr, "  - Edit platforms.txt to uncomment the platforms you want to build for\n")
		fmt.Fprintf(os.Stderr, "  - Create a VERSION file with your version (e.g. v1.0.1)\n")
		fmt.Fprintf(os.Stderr, "  - Then run %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("%s v%s\n", os.Args[0], version)
		os.Exit(0)
	}

	if showHelp {
		fmt.Printf("%s v%s\n", os.Args[0], version)
		fmt.Println("A program to cross compile go programs for various platforms")
		fmt.Println("- Copy platforms.txt at the root of your project")
		fmt.Println("- Edit platforms.txt to uncomment the platforms you want to build for")
		fmt.Println("- Create a VERSION file with your version (e.g. v1.0.1)")
		fmt.Println("- Then run go-xbuild-go")
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

	if additionalFiles != "" {
		config.AdditionalFiles = strings.Split(additionalFiles, ",")
		// Trim spaces from each file path
		for i, file := range config.AdditionalFiles {
			config.AdditionalFiles[i] = strings.TrimSpace(file)
		}
	}

	// ./bin must have the archives
	if makeRelease {
		err = createRelease(&config, releaseNote, releaseNoteFile)
		if err != nil {
			fail(err.Error())
		}
		os.Exit(0)
	}

	// otherise, run the main process
	fmt.Printf("Building project: %s\n", config.ProjectName)
	err = process(&config)
	if err != nil {
		fail(err.Error())
	}
}

// Create a GitHub release
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

// Process handles the main build process
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
	//	checksumPath := filepath.Join(config.BinDir, checksumFilename)

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

// Build for Raspberry Pi
func buildPi(config *Config, version, variant, armVersion string) error {
	distDir := fmt.Sprintf("%s-%s-raspberry-pi%s.d", config.ProjectName, version, variant)
	binaryName := fmt.Sprintf("%s-%s-raspberry-pi%s", config.ProjectName, version, variant)

	// Set environment variables
	env := []string{
		"GOOS=linux",
		"GOARCH=arm",
		"GOARM=" + armVersion,
	}

	fmt.Printf("Building for raspberry pi%s (arm%s)\n", variant, armVersion)

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

// Helper function to run go build
func gobuild(config *Config, output string, env []string) error {
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

// Build for platforms in platforms.txt
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

		fmt.Printf("Building for %s/%s\n", goos, goarch)

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
