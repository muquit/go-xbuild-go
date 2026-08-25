package brew

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/muquit/go-xbuild-go/internal/config"
	ver "github.com/muquit/go-xbuild-go/pkg/version"
)

const formulaDir = "Formula"

type platformEntry struct {
	goos    string
	goarch  string
	archive string
	sha256  string
}

// GenerateFormula generates a Homebrew formula file and writes it to Formula/{project}.rb.
// Called at the end of the build phase.
func GenerateFormula(cfg *config.Config) error {
	if cfg.SkipBrew || cfg.SkipBrewPush {
		fmt.Println("Skipping Homebrew formula generation (--skip-brew / --skip-brew-push)")
		return nil
	}

	version, err := getVersion(cfg)
	if err != nil {
		return err
	}

	homepage, err := detectHomepage()
	if err != nil {
		return err
	}

	platforms, err := readChecksums(cfg, version)
	if err != nil {
		return err
	}
	if len(platforms) == 0 {
		return fmt.Errorf("no platform entries found in checksums file")
	}

	license := detectLicense()
	content := buildFormulaContent(cfg, version, homepage, license, platforms)

	if err := os.MkdirAll(formulaDir, 0755); err != nil {
		return fmt.Errorf("failed to create Formula directory: %v", err)
	}

	outFile := filepath.Join(formulaDir, cfg.ProjectName+".rb")
	if err := os.WriteFile(outFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write formula file: %v", err)
	}

	fmt.Printf("Homebrew formula written to: %s\n", outFile)

	tapName := strings.TrimPrefix(homepage, "https://github.com/")
	tapURL := homepage + ".git"
	printBrewInstructions(cfg.ProjectName, tapName, tapURL)
	if err := writeBrewInstallDoc(cfg.ProjectName, tapName, tapURL); err != nil {
		fmt.Printf("Warning: failed to write docs/brew_install.md: %v\n", err)
	}
	return nil
}

// PushFormula commits and pushes Formula/{project}.rb to the project repo.
// Called at the end of the release phase.
func PushFormula(cfg *config.Config) error {
	if cfg.SkipBrewPush {
		fmt.Println("Skipping Homebrew formula push (--skip-brew-push)")
		return nil
	}

	version, err := getVersion(cfg)
	if err != nil {
		return err
	}

	formulaPath := filepath.Join(formulaDir, cfg.ProjectName+".rb")
	if _, err := os.Stat(formulaPath); os.IsNotExist(err) {
		return fmt.Errorf("formula file not found: %s (run build first)", formulaPath)
	}

	// Check if formula has actually changed before committing
	statusOut, err := exec.Command("git", "status", "--porcelain", formulaPath).Output()
	if err != nil {
		return fmt.Errorf("failed to check git status of formula: %v", err)
	}
	if strings.TrimSpace(string(statusOut)) == "" {
		fmt.Println("Homebrew formula unchanged, skipping commit and push")
		return nil
	}

	// Trim leading 'v' for the commit message display
	displayVersion := strings.TrimPrefix(version, "v")
	commitMsg := fmt.Sprintf("Update %s formula to %s", cfg.ProjectName, displayVersion)

	fmt.Printf("Pushing Homebrew formula (%s)...\n", formulaPath)

	steps := [][]string{
		{"git", "add", formulaPath},
		{"git", "commit", "-m", commitMsg},
		{"git", "push"},
	}

	for _, args := range steps {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to run '%s': %v", strings.Join(args, " "), err)
		}
	}

	fmt.Println("Homebrew formula pushed successfully")
	return nil
}

// --- Helpers ---

func getVersion(cfg *config.Config) (string, error) {
	content, err := os.ReadFile(cfg.VersionFile)
	if err != nil {
		return "", fmt.Errorf("failed to read version file: %v", err)
	}
	return strings.TrimSpace(string(content)), nil
}

func detectHomepage() (string, error) {
	out, err := exec.Command("git", "remote", "get-url", "origin").Output()
	if err != nil {
		return "", fmt.Errorf("could not detect git remote URL: %v", err)
	}
	remote := strings.TrimSpace(string(out))

	// git@github.com:user/repo.git -> https://github.com/user/repo
	if strings.HasPrefix(remote, "git@github.com:") {
		path := strings.TrimPrefix(remote, "git@github.com:")
		path = strings.TrimSuffix(path, ".git")
		return "https://github.com/" + path, nil
	}
	// https://github.com/user/repo.git -> https://github.com/user/repo
	if strings.HasPrefix(remote, "https://github.com/") {
		return strings.TrimSuffix(remote, ".git"), nil
	}

	return "", fmt.Errorf("unsupported git remote format: %s", remote)
}

func detectLicense() string {
	candidates := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "LICENCE", "LICENCE.txt"}
	for _, f := range candidates {
		if _, err := os.Stat(f); err == nil {
			// Try to guess the type from the file content
			data, err := os.ReadFile(f)
			if err != nil {
				return ""
			}
			content := strings.ToUpper(string(data))
			switch {
			case strings.Contains(content, "MIT LICENSE") || strings.Contains(content, "THE MIT"):
				return "MIT"
			case strings.Contains(content, "APACHE LICENSE"):
				return "Apache-2.0"
			case strings.Contains(content, "GNU GENERAL PUBLIC LICENSE") && strings.Contains(content, "VERSION 2"):
				return "GPL-2.0"
			case strings.Contains(content, "GNU GENERAL PUBLIC LICENSE") && strings.Contains(content, "VERSION 3"):
				return "GPL-3.0"
			case strings.Contains(content, "BSD 2-CLAUSE"):
				return "BSD-2-Clause"
			case strings.Contains(content, "BSD 3-CLAUSE"):
				return "BSD-3-Clause"
			default:
				return ""
			}
		}
	}
	return ""
}

func readChecksums(cfg *config.Config, version string) ([]platformEntry, error) {
	checksumFile := filepath.Join(cfg.BinDir, fmt.Sprintf("%s-%s-%s", cfg.ProjectName, version, cfg.ChecksumsFile))
	f, err := os.Open(checksumFile)
	if err != nil {
		return nil, fmt.Errorf("checksums file not found: %s", checksumFile)
	}
	defer f.Close()

	var platforms []platformEntry
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Format: <sha256>  <filename>
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		sha, archive := parts[0], parts[1]

		// Only include darwin and linux tar.gz (skip windows, raspberry pi)
		if !strings.HasSuffix(archive, ".tar.gz") {
			continue
		}
		if strings.Contains(archive, "raspberry-pi") {
			continue
		}

		goos, goarch := detectPlatform(archive)
		if goos == "" {
			continue
		}

		platforms = append(platforms, platformEntry{
			goos:    goos,
			goarch:  goarch,
			archive: archive,
			sha256:  sha,
		})
	}
	return platforms, scanner.Err()
}

// detectPlatform extracts goos/goarch from archive filename like:
// mailsend-go-v1.0.11-darwin-arm64.d.tar.gz
func detectPlatform(archive string) (goos, goarch string) {
	known := []struct{ goos, goarch string }{
		{"darwin", "arm64"},
		{"darwin", "amd64"},
		{"linux", "arm64"},
		{"linux", "amd64"},
		{"linux", "arm"},
	}
	for _, p := range known {
		if strings.Contains(archive, "-"+p.goos+"-"+p.goarch+".") {
			return p.goos, p.goarch
		}
	}
	return "", ""
}

func projectToClassName(name string) string {
	parts := strings.Split(name, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func buildFormulaContent(cfg *config.Config, version, homepage, license string, platforms []platformEntry) string {
	className := projectToClassName(cfg.ProjectName)
	// Strip 'v' prefix for display version in formula
	displayVersion := strings.TrimPrefix(version, "v")

	var sb strings.Builder

	sb.WriteString("# typed: false\n")
	sb.WriteString("# frozen_string_literal: true\n")
	sb.WriteString(fmt.Sprintf("# This file was generated by go-xbuild-go %s. DO NOT EDIT\n", ver.Get()))
	sb.WriteString("# https://github.com/muquit/go-xbuild-go\n")
	sb.WriteString(fmt.Sprintf("class %s < Formula\n", className))
	sb.WriteString(fmt.Sprintf("  desc \"%s\"\n", cfg.BrewDesc))
	sb.WriteString(fmt.Sprintf("  homepage \"%s\"\n", homepage))
	sb.WriteString(fmt.Sprintf("  version \"%s\"\n", displayVersion))
	if license != "" {
		sb.WriteString(fmt.Sprintf("  license \"%s\"\n", license))
	}
	sb.WriteString("\n")

	// macOS and Linux blocks using if/elsif chaining
	macArm := findPlatform(platforms, "darwin", "arm64")
	macAmd := findPlatform(platforms, "darwin", "amd64")
	linuxAmd := findPlatform(platforms, "linux", "amd64")
	linuxArm64 := findPlatform(platforms, "linux", "arm64")

	first := true
	writeBlock := func(condition, url, sha string) {
		if first {
			sb.WriteString(fmt.Sprintf("  if %s\n", condition))
			first = false
		} else {
			sb.WriteString(fmt.Sprintf("  elsif %s\n", condition))
		}
		sb.WriteString(fmt.Sprintf("    url \"%s\"\n", url))
		sb.WriteString(fmt.Sprintf("    sha256 \"%s\"\n", sha))
	}

	if macArm != nil {
		writeBlock("OS.mac? && Hardware::CPU.arm?",
			fmt.Sprintf("%s/releases/download/%s/%s", homepage, version, macArm.archive),
			macArm.sha256)
	}
	if macAmd != nil {
		writeBlock("OS.mac? && Hardware::CPU.intel?",
			fmt.Sprintf("%s/releases/download/%s/%s", homepage, version, macAmd.archive),
			macAmd.sha256)
	}
	if linuxAmd != nil {
		writeBlock("OS.linux? && Hardware::CPU.intel?",
			fmt.Sprintf("%s/releases/download/%s/%s", homepage, version, linuxAmd.archive),
			linuxAmd.sha256)
	}
	if linuxArm64 != nil {
		writeBlock("OS.linux? && Hardware::CPU.arm?",
			fmt.Sprintf("%s/releases/download/%s/%s", homepage, version, linuxArm64.archive),
			linuxArm64.sha256)
	}
	if !first {
		sb.WriteString("  end\n")
	}

	sb.WriteString("\n")
	sb.WriteString("  def install\n")

	hasMac := macArm != nil || macAmd != nil
	hasLinux := linuxAmd != nil || linuxArm64 != nil

	if hasMac && !hasLinux {
		// macOS only - use #{version} Ruby interpolation
		sb.WriteString("    if Hardware::CPU.arm?\n")
		sb.WriteString(fmt.Sprintf("      bin.install \"%s-v#{version}-darwin-arm64\" => \"%s\"\n", cfg.ProjectName, cfg.ProjectName))
		sb.WriteString("    else\n")
		sb.WriteString(fmt.Sprintf("      bin.install \"%s-v#{version}-darwin-amd64\" => \"%s\"\n", cfg.ProjectName, cfg.ProjectName))
		sb.WriteString("    end\n")
	} else {
		// macOS + Linux
		sb.WriteString("    if OS.mac?\n")
		sb.WriteString("      if Hardware::CPU.arm?\n")
		sb.WriteString(fmt.Sprintf("        bin.install \"%s-v#{version}-darwin-arm64\" => \"%s\"\n", cfg.ProjectName, cfg.ProjectName))
		sb.WriteString("      else\n")
		sb.WriteString(fmt.Sprintf("        bin.install \"%s-v#{version}-darwin-amd64\" => \"%s\"\n", cfg.ProjectName, cfg.ProjectName))
		sb.WriteString("      end\n")
		sb.WriteString("    elsif OS.linux?\n")
		sb.WriteString("      if Hardware::CPU.arm?\n")
		sb.WriteString(fmt.Sprintf("        bin.install \"%s-v#{version}-linux-arm64\" => \"%s\"\n", cfg.ProjectName, cfg.ProjectName))
		sb.WriteString("      else\n")
		sb.WriteString(fmt.Sprintf("        bin.install \"%s-v#{version}-linux-amd64\" => \"%s\"\n", cfg.ProjectName, cfg.ProjectName))
		sb.WriteString("      end\n")
		sb.WriteString("    end\n")
	}

	for _, f := range cfg.BrewExtraBinFiles {
		sb.WriteString(fmt.Sprintf("    bin.install \"%s\"\n", f))
	}
	sb.WriteString("  end\n")
	sb.WriteString("end\n")

	return sb.String()
}

func findPlatform(platforms []platformEntry, goos, goarch string) *platformEntry {
	for i := range platforms {
		if platforms[i].goos == goos && platforms[i].goarch == goarch {
			return &platforms[i]
		}
	}
	return nil
}

func printBrewInstructions(project, tapName, tapURL string) {
	fmt.Printf("\nHomebrew install instructions for %s:\n", project)
	fmt.Printf("  brew tap %s %s\n", tapName, tapURL)
	fmt.Printf("  brew install %s/%s\n", tapName, project)
	fmt.Printf("\nTo uninstall:\n")
	fmt.Printf("  brew uninstall %s\n", project)
}

func writeBrewInstallDoc(project, tapName, tapURL string) error {
	const docsDir = "docs"
	if err := os.MkdirAll(docsDir, 0755); err != nil {
		return fmt.Errorf("failed to create docs directory: %v", err)
	}

	content := fmt.Sprintf(`## Installing using Homebrew on Mac/Linux

You will need to install [Homebrew](https://brew.sh/) first.

### Install

`+"```"+`
brew tap %s %s
brew install %s/%s
`+"```"+`

### Upgrade
`+"```"+`
brew upgrade %s
`+"```"+`

### Uninstall
`+"```"+`
brew uninstall %s
`+"```"+`

### Remove the tap
`+"```"+`
brew untap %s
`+"```"+`

---
<sub>[go-xbuild-go](https://github.com/muquit/go-xbuild-go) writes the formula into the project's own `+"`Formula/project.rb`"+` rather than a central Homebrew tap repo. This keeps the formula version-controlled and committed alongside the code.</sub>

<sub>Brew install instructions and formula automatically generated by [go-xbuild-go](https://github.com/muquit/go-xbuild-go) %s on %s</sub>
`, tapName, tapURL, tapName, project, project, project, tapName, ver.Get(), time.Now().Format("Jan-02-2006"))

	outFile := filepath.Join(docsDir, "brew_install.md")
	if err := os.WriteFile(outFile, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}
	fmt.Printf("Homebrew install doc written to: %s\n", outFile)
	return nil
}
