package build

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/muquit/go-xbuild-go/internal/config"
)

// ProcessMultiTarget handles the build process for a multi-target project.
// It is EXPORTED so it can be called by the main package.
func ProcessMultiTarget(cfg *config.Config, buildForPi bool) error {
	// Initialize
	if err := initialize(cfg); err != nil {
		return err
	}

	// Get version
	version, err := getVersion(cfg)
	if err != nil {
		return err
	}

	projectConfig := cfg.ProjectConfig
	fmt.Printf("Building %s version %s with %d targets\n", cfg.ProjectName, version, len(projectConfig.Targets))

	// Build each target
	for _, target := range projectConfig.Targets {
		fmt.Printf("\n=== Building target: %s ===\n", target.Name)

		targetConfig := *cfg
		targetConfig.ProjectName = target.Name
		if target.OutputName != "" {
			targetConfig.ProjectName = target.OutputName
		}
		targetConfig.LdFlags = projectConfig.DefaultLdFlags
		if target.LdFlags != "" {
			targetConfig.LdFlags = target.LdFlags
		}
		targetConfig.BuildFlags = projectConfig.DefaultBuildFlags
		if target.BuildFlags != "" {
			targetConfig.BuildFlags = target.BuildFlags
		}
		//targetConfig.AdditionalFiles = append(projectConfig.GlobalAdditionalFiles, target.AdditionalFiles...)
		//targetConfig.AdditionalFiles = append(targetConfig.AdditionalFiles, cfg.AdditionalFiles...)
		// new code starts -- Oct-12-2025 
		// Create a new slice to prevent leaking files between targets
		numFiles := len(projectConfig.GlobalAdditionalFiles) + len(target.AdditionalFiles) + len(cfg.AdditionalFiles)
		allFiles := make([]string, 0, numFiles)

		allFiles = append(allFiles, projectConfig.GlobalAdditionalFiles...)
		allFiles = append(allFiles, target.AdditionalFiles...)
		allFiles = append(allFiles, cfg.AdditionalFiles...)

		targetConfig.AdditionalFiles = allFiles
		// new code ends --

		checksumFile := filepath.Join(cfg.BinDir, fmt.Sprintf("%s-%s-%s", targetConfig.ProjectName, version, cfg.ChecksumsFile))
		if err := os.RemoveAll(checksumFile); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove old checksums file: %v", err)
		}

		if err := buildForPlatformsWithPath(&targetConfig, version, target.Path); err != nil {
			return fmt.Errorf("failed to build target %s: %v", target.Name, err)
		}

		if buildForPi {
			if err := buildPiWithPath(&targetConfig, version, target.Path, "", "7"); err != nil {
				return fmt.Errorf("failed to build target %s for Pi: %v", target.Name, err)
			}
			if err := buildPiWithPath(&targetConfig, version, target.Path, "-jessie", "6"); err != nil {
				return fmt.Errorf("failed to build target %s for Pi Jessie: %v", target.Name, err)
			}
		}
		fmt.Printf("Target %s build complete\n", target.Name)
	}

	fmt.Printf("\nAll targets build complete. Artifacts are in %s\n", cfg.BinDir)
	return nil
}

// Process handles the main build process for a single-target project.
// It is EXPORTED so it can be called by the main package.
func Process(cfg *config.Config, buildForPi bool) error {
	if err := initialize(cfg); err != nil {
		return err
	}

	version, err := getVersion(cfg)
	if err != nil {
		return err
	}

	fmt.Printf("Building %s version %s\n", cfg.ProjectName, version)

	checksumFile := filepath.Join(cfg.BinDir, fmt.Sprintf("%s-%s-%s", cfg.ProjectName, version, cfg.ChecksumsFile))
	if err := os.RemoveAll(checksumFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove old checksums file: %v", err)
	}

	if err := buildForPlatforms(cfg, version); err != nil {
		return err
	}

	if buildForPi {
		if err := buildPi(cfg, version, "", "7"); err != nil {
			return err
		}
		if err := buildPi(cfg, version, "-jessie", "6"); err != nil {
			return err
		}
	}

	fmt.Printf("Build complete. Artifacts are in %s\n", cfg.BinDir)
	return nil
}

// --- Helper Functions (un-exported) ---

func initialize(cfg *config.Config) error {
	if _, err := os.Stat(cfg.VersionFile); os.IsNotExist(err) {
		return fmt.Errorf("version file not found: %s", cfg.VersionFile)
	}
	if _, err := os.Stat(cfg.PlatformsFile); os.IsNotExist(err) {
		return fmt.Errorf("platforms file not found: %s", cfg.PlatformsFile)
	}
	if err := os.MkdirAll(cfg.BinDir, 0755); err != nil {
		return fmt.Errorf("could not create bin directory: %s, error: %v", cfg.BinDir, err)
	}
	return nil
}

func getVersion(cfg *config.Config) (string, error) {
	content, err := os.ReadFile(cfg.VersionFile)
	if err != nil {
		return "", fmt.Errorf("failed to read version file: %v", err)
	}
	return strings.TrimSpace(string(content)), nil
}

func buildForPlatformsWithPath(cfg *config.Config, version, buildPath string) error {
	file, err := os.Open(cfg.PlatformsFile)
	if err != nil {
		return fmt.Errorf("failed to open platforms file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		parts := strings.Split(line, "/")
		if len(parts) < 2 {
			continue
		}
		goos, goarch := parts[0], parts[1]
		fmt.Printf("\n> Building for %s/%s\n", goos, goarch)

		binaryName := fmt.Sprintf("%s-%s-%s-%s", cfg.ProjectName, version, goos, goarch)
		if goos == "windows" {
			binaryName += ".exe"
		}

		env := []string{"GOOS=" + goos, "GOARCH=" + goarch}
		if err := gobuildWithPath(cfg, binaryName, buildPath, env); err != nil {
			return fmt.Errorf("failed to build for %s/%s: %v", goos, goarch, err)
		}

		distDir := fmt.Sprintf("%s-%s-%s-%s.d", cfg.ProjectName, version, goos, goarch)
		if err := copyFiles(cfg, binaryName, distDir); err != nil {
			return err
		}
		if err := createArchive(cfg, version, distDir, goos); err != nil {
			return err
		}
		if err := os.Remove(binaryName); err != nil {
			return fmt.Errorf("failed to remove binary: %v", err)
		}
	}
	return scanner.Err()
}

func buildPiWithPath(cfg *config.Config, version, buildPath, variant, armVersion string) error {
	binaryName := fmt.Sprintf("%s-%s-raspberry-pi%s", cfg.ProjectName, version, variant)
	env := []string{"GOOS=linux", "GOARCH=arm", "GOARM=" + armVersion}
	fmt.Printf("\n> Building for raspberry pi%s (arm%s)\n", variant, armVersion)

	if err := gobuildWithPath(cfg, binaryName, buildPath, env); err != nil {
		return fmt.Errorf("failed to build for raspberry pi: %v", err)
	}

	distDir := fmt.Sprintf("%s-%s-raspberry-pi%s.d", cfg.ProjectName, version, variant)
	if err := copyFiles(cfg, binaryName, distDir); err != nil {
		return err
	}
	if err := createArchive(cfg, version, distDir, "linux"); err != nil {
		return err
	}
	return os.Remove(binaryName)
}

func gobuildWithPath(cfg *config.Config, output, buildPath string, env []string) error {
	args := []string{"build"}
	if cfg.LdFlags != "" {
		args = append(args, "-ldflags="+cfg.LdFlags)
	}
	if cfg.BuildFlags != "" {
		buildFlagArgs, err := ParseArguments(cfg.BuildFlags)
		if err != nil {
			return fmt.Errorf("failed to parse build flags: %v", err)
		}
		args = append(args, buildFlagArgs...)
	}
	if len(cfg.ExtraBuildArgs) > 0 {
		args = append(args, cfg.ExtraBuildArgs...)
	}
	args = append(args, "-o", output)
	if buildPath != "" && buildPath != "." {
		args = append(args, buildPath)
	}

	cmd := exec.Command("go", args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func takeChecksum(cfg *config.Config, version, archive string) error {
	checksumFilename := fmt.Sprintf("%s-%s-%s", cfg.ProjectName, version, cfg.ChecksumsFile)
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %v", err)
	}
	if err := os.Chdir(cfg.BinDir); err != nil {
		return fmt.Errorf("failed to change to bin directory: %v", err)
	}
	defer os.Chdir(currentDir)

	data, err := os.ReadFile(archive)
	if err != nil {
		return fmt.Errorf("failed to read archive for checksum: %v", err)
	}
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])

	f, err := os.OpenFile(checksumFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open checksum file: %v", err)
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s  %s\n", checksum, archive)
	return err
}

func copyFiles(cfg *config.Config, bin, distDir string) error {
	if err := os.MkdirAll(distDir, 0755); err != nil {
		return fmt.Errorf("failed to create dist directory: %v", err)
	}
	if err := copyFile(bin, filepath.Join(distDir, filepath.Base(bin))); err != nil {
		return fmt.Errorf("failed to copy binary: %v", err)
	}

	// Auto-include files
	docFiles := []string{"README.md", "LICENSE.txt", "LICENSE", "platforms.txt"}
	baseDir := filepath.Dir(cfg.BinDir)
	for _, f := range docFiles {
		src := filepath.Join(baseDir, f)
		if _, err := os.Stat(src); err == nil {
			if err := copyFile(src, filepath.Join(distDir, f)); err != nil {
				return fmt.Errorf("failed to copy %s: %v", src, err)
			}
		}
	}

	// Additional files
	for _, filePath := range cfg.AdditionalFiles {
		if _, err := os.Stat(filePath); err == nil {
			destPath := filepath.Join(distDir, filepath.Base(filePath))
			if err := copyFile(filePath, destPath); err != nil {
				return fmt.Errorf("failed to copy additional file %s: %v", filePath, err)
			}
		} else {
			fmt.Printf("Warning: additional file not found: %s\n", filePath)
		}
	}
	return nil
}

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

func createArchive(cfg *config.Config, version, distDir, goos string) error {
	var archiveName string
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

	finalArchivePath := filepath.Join(cfg.BinDir, filepath.Base(archiveName))
	if err := os.Rename(archiveName, finalArchivePath); err != nil {
		return fmt.Errorf("failed to move archive to bin directory: %v", err)
	}

	if err := takeChecksum(cfg, version, filepath.Base(archiveName)); err != nil {
		return fmt.Errorf("failed to take checksum: %v", err)
	}
	return os.RemoveAll(distDir)
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

		// Create a relative path for the file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name, err = filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// Add a trailing slash for directories
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		// If it's a file, copy its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(writer, file)
			return err
		}

		return nil
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

		// Create a relative path for the file header
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name, err = filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		// If it's a file, copy its contents
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(tarWriter, file)
			return err
		}

		return nil
	})
}


// Legacy single-target functions
func buildPi(cfg *config.Config, version, variant, armVersion string) error {
	return buildPiWithPath(cfg, version, ".", variant, armVersion)
}

func gobuild(cfg *config.Config, output string, env []string) error {
	return gobuildWithPath(cfg, output, ".", env)
}

func buildForPlatforms(cfg *config.Config, version string) error {
	return buildForPlatformsWithPath(cfg, version, ".")
}

func ParseArguments(argStr string) ([]string, error) {
	var args []string
	var current strings.Builder
	inQuotes := false
	argStr = strings.TrimSpace(argStr)
	if argStr == "" {
		return args, nil
	}
	for _, char := range argStr {
		switch char {
		case '"', '\'':
			inQuotes = !inQuotes
		case ' ':
			if !inQuotes {
				if current.Len() > 0 {
					args = append(args, current.String())
					current.Reset()
				}
			} else {
				current.WriteRune(char)
			}
		default:
			current.WriteRune(char)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}
	if inQuotes {
		return nil, fmt.Errorf("unclosed quote in build arguments")
	}
	return args, nil
}
