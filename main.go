package main

/////////////////////////////////////////////////////////////////////
// Cross compile go programs for other platforms
// Developed with Claude AI 3.7 Sonnet, working under my guidance and
// instructions.
// muquit@muquit.com Mar-26-2025
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
	"strings"
)

const (
	version = "1.0.2"
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
}

func main() {
	var showVersion bool
	var showHelp bool

	flag.BoolVar(&showVersion, "version", false, "Show version information and exit")
	flag.BoolVar(&showHelp, "help", false, "Show help information and exit")
	flag.BoolVar(&buildForPi, "pi", true, "Build Raspberry Pi")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s v%s\n", os.Args[0], version)
		fmt.Fprintf(os.Stderr, "A program to cross compile go programs\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("%s v%s\n", os.Args[0], version)
		os.Exit(0)
	}

	if showHelp {
		fmt.Printf("%s v%s\n", os.Args[0], version)
		fmt.Println("A program to cross compile go programs for various platforms\n")
		fmt.Println("- Copy platforms.txt at the root of your project")
		fmt.Println("- Edit platforms.txt to uncomment the platforms you want to build for")
		fmt.Println("- Create a VERSION file with your version (e.g. v1.0.1)")
		fmt.Println("- Then run go-xbuild-go\n")
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

	fmt.Printf("Building project: %s\n", config.ProjectName)

	// Run the main process
	err = process(&config)
	if err != nil {
		fail(err.Error())
	}
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
		filepath.Join(filepath.Dir(config.BinDir), "platforms.txt"):                 filepath.Join(distDir, "platforms.txt"),
	}

	for src, dst := range docFiles {
		if _, err := os.Stat(src); err == nil {
			if err := copyFile(src, dst); err != nil {
				return fmt.Errorf("failed to copy %s: %v", src, err)
			}
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
