package release

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/muquit/go-xbuild-go/internal/config"
)

// Create handles the entire GitHub release process using batch uploads.
func Create(cfg *config.Config, note, noteFile string) error {
	if err := checkGhCliExists(); err != nil {
		return err
	}
	if os.Getenv("GITHUB_TOKEN") == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	version, err := getVersion(cfg)
	if err != nil {
		return err
	}
	if _, err := os.Stat(cfg.BinDir); os.IsNotExist(err) {
		return fmt.Errorf("bin directory does not exist for release")
	}

	ghCmd := getGhCommand()

	// Step 1: Create the release without assets
	createArgs := []string{"release", "create", version}
	if note != "" {
		createArgs = append(createArgs, "--notes", note)
	} else if noteFile != "" {
		createArgs = append(createArgs, "--notes-file", noteFile)
	} else if _, err := os.Stat("release_notes.md"); err == nil {
		createArgs = append(createArgs, "--notes-file", "release_notes.md")
	} else {
		return fmt.Errorf("no release notes specified, and release_notes.md not found")
	}

	fmt.Printf("Creating GitHub release %s (without assets)...\n", version)
	cmd := exec.Command(ghCmd, createArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create GitHub release shell command: %v", err)
	}

	// Step 2: Collect and upload assets in batches
	fmt.Println("Uploading assets to release...")
	files, err := os.ReadDir(cfg.BinDir)
	if err != nil {
		return fmt.Errorf("failed to read bin directory: %v", err)
	}

	var assetsToUpload []string
	for _, file := range files {
		if !file.IsDir() {
			assetsToUpload = append(assetsToUpload, filepath.Join(cfg.BinDir, file.Name()))
		}
	}

	if len(assetsToUpload) == 0 {
		return fmt.Errorf("bin directory is empty, no assets to upload")
	}

	batchSize := 10
	for i := 0; i < len(assetsToUpload); i += batchSize {
		end := i + batchSize
		if end > len(assetsToUpload) {
			end = len(assetsToUpload)
		}
		batch := assetsToUpload[i:end]
		
		uploadArgs := []string{"release", "upload", version}
		uploadArgs = append(uploadArgs, batch...)

		fmt.Printf("Uploading batch %d/%d (%d files)...\n",
			(i/batchSize)+1,
			(len(assetsToUpload)+batchSize-1)/batchSize,
			len(batch))

		uploadCmd := exec.Command(ghCmd, uploadArgs...)
		uploadCmd.Stdout = os.Stdout
		uploadCmd.Stderr = os.Stderr
		if err := uploadCmd.Run(); err != nil {
			return fmt.Errorf("failed to upload assets batch: %v", err)
		}
	}

	fmt.Printf("GitHub release %s created successfully with all assets\n", version)
	return nil
}


// --- Helper Functions ---

func getVersion(cfg *config.Config) (string, error) {
	content, err := os.ReadFile(cfg.VersionFile)
	if err != nil {
		return "", fmt.Errorf("failed to read version file: %v", err)
	}
	return strings.TrimSpace(string(content)), nil
}

func getGhCommand() string {
	if envCmd := os.Getenv("GH_CLI_PATH"); envCmd != "" {
		return envCmd
	}
	if runtime.GOOS == "windows" {
		if _, err := exec.LookPath("gh.exe"); err == nil {
			return "gh.exe"
		}
	}
	return "gh"
}

func checkGhCliExists() error {
	cmd := getGhCommand()
	if _, err := exec.LookPath(cmd); err != nil {
		return fmt.Errorf("GitHub CLI (%s) is not installed or not in PATH", cmd)
	}
	return nil
}