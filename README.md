[![Downloads](https://img.shields.io/github/downloads/muquit/go-xbuild-go/total.svg)](https://github.com/muquit/go-xbuild-go/releases)
# Table Of Contents
- [Introduction](#introduction)
- [Background and Motivation](#background-and-motivation)
- [Features](#features)
- [Synopsis](#synopsis)
- [Latest Version (v1.0.9)](#latest-version-v109)
- [Homebrew Formula Generation](#homebrew-formula-generation)
  - [When the formula is generated](#when-the-formula-is-generated)
    - [Trusting the tap ([Homebrew](https://brew.sh/) 6.0+)](#trusting-the-tap-homebrewhttpsbrewsh-60)
  - [When the formula is pushed](#when-the-formula-is-pushed)
  - [Flag summary](#flag-summary)
- [Installation](#installation)
  - [Download](#download)
  - [Verify Checksum](#verify-checksum)
  - [Extract](#extract)
  - [Install](#install)
  - [Installing using Homebrew on Mac/Linux](#installing-using-homebrew-on-maclinux)
    - [Install](#install)
    - [Upgrade](#upgrade)
    - [Uninstall](#uninstall)
    - [Remove the tap](#remove-the-tap)
  - [Building from source](#building-from-source)
- [How to use](#how-to-use)
- [Example](#example)
  - [Legacy mode (single binary)](#legacy-mode-single-binary)
  - [Multi-binary mode](#multi-binary-mode)
  - [Multi-Binary Configuration](#multi-binary-configuration)
  - [Output Structure](#output-structure)
  - [Included Files](#included-files)
  - [Config file for single binary project](#config-file-for-single-binary-project)
- [How to release your project to github (Any kind, not just golang based projects)](#how-to-release-your-project-to-github-any-kind-not-just-golang-based-projects)
- [Contributing](#contributing)
- [License](#license)
- [Authors](#authors)

# Introduction

A cross-platform program to cross compile 
[Go](https://go.dev/) projects without the complexity of [GoReleaser](https://goreleaser.com/).
The program can be used to:

- Cross compile go projects for various platforms - with ease
- **Build multi-binary Go projects** - handle projects with multiple main packages like `cmd/cli/`, `cmd/server/`, etc.
- Make releases to github - with ease. Not just go projects, **any project** can be released to github,
just copy the assets to `./bin` directory. Please look at Look at [How to release your project to github](#how-to-release-your-project-to-github-any-kind-not-just-golang-based-projects) for details.


# Background and Motivation

It was written from the frustration of using [GoReleaser](https://goreleaser.com/). I don't
release software often, every 6 to 12 months or so, and by the time
I need to release again, something about GoReleaser's config has
changed and needs updating.

I'd rather own a simple tool I understand fully than fight with
someone else's. If something breaks in `go-xbuild-go`, I can just fix
it myself.

Hope you will find it useful and fun to use.

This is a [Go](https://go.dev/) port of my bash script https://github.com/muquit/go-xbuild

Pull requests, suggestions are always welcome.



# Features
- Simple to use and maintain
- Cross compile for multiple platforms
- **NEW in v1.0.5**: Multi-binary project support with JSON configuration
- **NEW in v1.0.5**: Build multiple main packages from `cmd/` directory structure
- **NEW in v1.0.5**: Per-target customization (ldflags, build flags, output names)
- **NEW in v1.0.5**: List available build targets with `-list-targets`
- Special handling for Raspberry Pi (modern and Jessie)
- Generates checksums
- Creates archives (ZIP for Windows, tar.gz for others)
- No complex configuration files (for simple projects)
- Just uncomment platforms in platforms.txt to build for them
- Make release of the project to github
- Full backward compatibility - existing projects work unchanged

# Synopsis
```
go-xbuild-go v1.0.9
A program to cross compile go programs and release any software to github

Usage:
  go-xbuild-go [options]             # Build using defaults or config file
  go-xbuild-go -config build-config.json   # Build using custom config
  go-xbuild-go -release               # Create GitHub release from ./bin

Options:
  -additional-files string
    	Comma-separated list of additional files to include
  -brew-desc string
    	Description for Homebrew formula
  -build-args string
    	Additional go build arguments (e.g., '-tags systray')
  -config string
    	Path to build configuration file (JSON)
  -help
    	Show help information and exit
  -list-targets
    	List available build targets and exit
  -pi
    	Build for Raspberry Pi (default true)
  -platforms-file string
    	Path of platforms.txt (default "platforms.txt")
  -release
    	Create a GitHub release
  -release-note string
    	Release note text
  -release-note-file string
    	File containing release notes
  -skip-brew
    	Skip Homebrew formula generation
  -skip-brew-push
    	Skip Homebrew formula push on release
  -version
    	Show version information and exit

Environment Variables (for GitHub release):
  GITHUB_TOKEN      GitHub API token (required for -release)
  GH_CLI_PATH       Custom path to GitHub CLI executable (optional)

Config File:
  Optional JSON file for advanced, multi-target builds.

A minimal example config file (build-config.json):
{
  "project_name": "myproject",
  "version_file": "VERSION",
  "targets": [
    { "name": "cli", "path": "./cmd/cli" },
    { "name": "server", "path": "./cmd/server" }
  ]
}
```

# Latest Version (v1.0.9)
The current version is v1.0.9
Please look at [ChangeLog](ChangeLog.md) for what has changed in the current version.

# Homebrew Formula Generation

`go-xbuild-go` can generate a [Homebrew](https://brew.sh/) formula for your project and write it to
a per-project `Formula/<project>.rb` file. It also generates `docs/brew_install.md`
with ready-to-use tap/install/upgrade/uninstall instructions for that project.

## When the formula is generated

The formula is generated at the end of the **build** phase (`GenerateFormula`,
called from `internal/brew/brew.go`), every time `go-xbuild-go` runs a build,
**unless** one of these flags is set:

- `-skip-brew`: skip formula generation entirely.
- `-skip-brew-push`: also causes generation to be skipped (see note below).

Generation requires:

- A `VERSION` file (or whatever `version_file` is configured) to determine the
  formula version.
- A git remote pointing at a `github.com` repo (`git remote get-url origin`),
  used to derive the `homepage` and the release asset URLs. Generation fails
  if the remote isn't a [GitHub](https://github.com/) URL.
- A checksums file in `bin/` (`<project>-<version>-<checksums-file>`) produced
  by the build/archive step, listing the `.tar.gz` archives and their SHA-256
  sums. Only `darwin`/`linux` `.tar.gz` archives are used (Windows archives and
  the Raspberry Pi archive are excluded from the formula).

On success it writes:

- `Formula/<project>.rb`: the [Homebrew](https://brew.sh/) formula (marked `DO NOT EDIT`, regenerated
  every run).
- `docs/brew_install.md`: [Homebrew](https://brew.sh/) install/upgrade/uninstall instructions for
  the project, suitable for inclusion in `docs/main.md` via
  `@[:markdown](brew_install.md)`.

A best-effort license string (MIT, Apache-2.0, GPL-2.0/3.0, BSD-2/3-Clause) is
detected from a `LICENSE*` file in the project root and included in the formula
if found.

### Trusting the tap ([Homebrew](https://brew.sh/) 6.0+)

[Homebrew](https://brew.sh/) 6.0+ refuses to load formulae from third-party taps until the tap is
explicitly trusted, failing with `Refusing to load formula ... from untrusted
tap ...`. The generated `docs/brew_install.md` and the console instructions
printed after a build both include a `brew trust <tap>` step between `brew tap`
and `brew install` to cover this.

## When the formula is pushed

Pushing (`PushFormula`) happens at the end of the **release** phase (`-release`),
and commits and pushes `Formula/<project>.rb` to the current git repo, unless:

- `-skip-brew-push` is set: push is skipped.
- The formula file is unchanged since the last commit (`git status --porcelain`
  reports nothing): push is skipped with a message, since there's nothing new
  to commit.

If `Formula/<project>.rb` doesn't exist when `-release` runs (e.g. because the
build step was skipped or `-skip-brew` was used), the push step fails: the
formula must be generated by a prior build before it can be pushed.

## Flag summary

| Flag | Effect |
|---|---|
| `-skip-brew` | Don't generate `Formula/<project>.rb` or `docs/brew_install.md` during build. |
| `-skip-brew-push` | Generation is skipped too (build phase), and the release phase won't commit/push the formula. |
| `-brew-desc` | Sets the `desc` field in the generated formula. |

# Installation

## Download
* Download the appropriate archive for your platform from the [Releases](https://github.com/muquit/go-xbuild-go/releases) page

## Verify Checksum

```bash
# Download the checksums file
# Verify the archive
sha256sum -c go-xbuild-go-vX.X.X-darwin-arm64.d.tar.gz
```
Repeat the step for other archives

## Extract
macOS/Linux:

```bash
tar -xzf go-xbuild-go-vX.X.X-darwin-arm64.d.tar.gz
cd go-xbuild-go-vX.X.X-darwin-arm64.d
```

Repeat the step for other archives

Windows:

The tar command is available in Windows 10 (1803) and later, or you can
use the GUI (right-click → Extract All). After extracting, copy/rename the
binary somewhere in your PATH.

## Install

```bash
# macOS/Linux
sudo cp go-xbuild-go-vX.X.X-darwin-arm64 /usr/local/bin/go-xbuild-go
sudo chmod +x /usr/local/bin/go-xbuild-go
```

```bash
# Windows
copy go-xbuild-go-vX.X.X-windows-amd64.exe C:\Windows\System32\go-xbuild-go.exe
```

## Installing using Homebrew on Mac/Linux

You will need to install [Homebrew](https://brew.sh/) first.

### Install

First install the custom tap, then trust it. Homebrew 6.0+ refuses to load
formulae from third-party taps until they are explicitly trusted.

```
brew tap muquit/go-xbuild-go https://github.com/muquit/go-xbuild-go.git
brew trust muquit/go-xbuild-go
brew install go-xbuild-go
```

Or tap, trust and install in one go:
```
brew tap muquit/go-xbuild-go https://github.com/muquit/go-xbuild-go.git
brew trust muquit/go-xbuild-go
brew install muquit/go-xbuild-go/go-xbuild-go
```

### Upgrade
```
brew upgrade go-xbuild-go
```

### Uninstall
```
brew uninstall go-xbuild-go
```

### Remove the tap
```
brew untap muquit/go-xbuild-go
```

## Building from source

Install [Go](https://go.dev/) first

```bash
git clone https://github.com/muquit/go-xbuild-go
cd go-xbuild-go
go build .
or
make build
```

Please look at [Makefile](Makefile) for more info

# How to use

There are two ways to use go-xbuild-go:

**For simple projects (legacy mode):**
- Copy `go-xbuild-go` somewhere in your PATH
- Copy `platforms.txt` to your go project's root directory
- Create a `VERSION` file with your version (e.g., v1.0.1)
- Edit `platforms.txt` to uncomment the platforms you want to build for
- Run: `go-xbuild-go`

**For complex projects with multiple binaries (new in v1.0.5):**
- Copy `go-xbuild-go` somewhere in your PATH
- Create a `build-config.json` file (see [Multi-Binary Configuration](#multi-binary-configuration))
- Run: `go-xbuild-go -config build-config.json`

Look at [How to release your project to github](#how-to-release-your-project-to-github)

Note: go programs can be cross compiled for more than 40 platforms. `go tool dist list` for the list
supported by your version of go.

A few lines of platforms.txt is shown below:
```text
########################################################################
# GOOS/GOARCH
# generated by running: go tool dist list
# Uncomment or add platforms if needed
########################################################################
#aix/ppc64
#android/386
darwin/amd64
darwin/arm64
windows/amd64
#linux/386
linux/amd64
#linux/arm
#linux/arm64
...
```

# Example

## Legacy mode (single binary)
Run go-xbuild-go from the root of your project. Update VERSION file if needed.
Then, compile the binaries:

```bash
go-xbuild-go
```

The program will:
1. Detect your project name from the directory
2. Read version from VERSION file
3. Build for all uncommented platforms in platforms.txt
4. Create appropriate archives (ZIP for Windows, tar.gz for others)
5. Generate checksums for all archives
6. Place all artifacts in _./bin_ directory

## Multi-binary mode
For projects with multiple main packages (e.g., `cmd/cli/`, `cmd/server/`), create a `build-config.json` file and run:

```bash
# List available targets
go-xbuild-go -config build-config.json -list-targets

# Build all targets
go-xbuild-go -config build-config.json

# Create GitHub release
go-xbuild-go -release -release-note "Multi-binary release"
```

## Multi-Binary Configuration

The JSON configuration file supports the following structure:

```json
{
  "project_name": "myproject",
  "version_file": "VERSION",
  "platforms_file": "platforms.txt",
  "default_ldflags": "-s -w -X main.version={{.Version}} -X main.commit={{.Commit}}",
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
    },
    {
      "name": "admin",
      "path": "./cmd/admin"
    }
  ]
}
```

**Configuration options:**
- `project_name`: Project name used in archive names
- `version_file`: Path to file containing version (default: "VERSION")
- `platforms_file`: Path to platforms definition file (default: "platforms.txt")  
- `default_ldflags`: Default linker flags applied to all targets
- `default_build_flags`: Default build flags applied to all targets
- `ldflags`: Custom ldflags
- `build_flags`: Custom build flags
- `targets`: Array of build targets

**Target options:**
- `name`: Target identifier (used in `-list-targets`)
- `path`: Path to main package (e.g., "./cmd/cli")
- `output_name`: Custom binary name (optional, defaults to target name)

**Variable substitution in ldflags:**
- `{{.Version}}`: Replaced with version from VERSION file
- `{{.Commit}}`: Replaced with current git commit hash
- `{{.Date}}`: Replaced with build timestamp

**Example project structure:**
```
myproject/
├── cmd/
│   ├── cli/main.go
│   ├── server/main.go
│   └── admin/main.go
├── build-config.json
├── platforms.txt
├── VERSION
└── README.md
```

For a complete working example, see: [go-multi-main-example](https://github.com/muquit/go-multi-main-example)

## Output Structure

```
bin/
├── project-v1.0.1-darwin-amd64.d.tar.gz
├── project-v1.0.1-darwin-arm64.d.tar.gz
├── project-v1.0.1-windows-amd64.d.zip
├── project-v1.0.1-linux-amd64.d.tar.gz
├── project-v1.0.1-raspberry-pi.d.tar.gz
├── project-v1.0.1-raspberry-pi-jessie.d.tar.gz
└── project-v1.0.1-checksums.txt
```

## Included Files
The following files will be included in archives if they exist:
- Compiled binary
- README.md
- LICENSE.txt
- docs/project-name.1 (man page)
- platforms.txt
- Add extra files with `-additional-files` (Do not add these default: README.md, LICENSE.txt, LICENSE, platforms.txt, <project>.1)
## Config file for single binary project

`build-config.json` with `-config` flag can be used for single binary
project as well. Example for `go-xbuild-go` itself:
```json
{
  "project_name": "go-xbuild-go",
  "version_file": "VERSION",
  "platforms_file": "platforms.txt",
  "default_ldflags": "-s -w -X 'github.com/muquit/go-xbiuld-go/pkg/version.Version=v1.0.1' -X 'main.buildTime={{.BuildTime}}'",
  "default_build_flags": "-trimpath",
  "global_additional_files": [
    "README.md",
    "LICENSE",
    "build-config.json"
  ],
  "targets": [
    {
      "name": "go-xbuild-go",
      "path": ".",
      "output_name": "go-xbuild-go",
      "build_flags": "-trimpath",
      "build-args": "-s -w -X 'github.com/muquit/go-xbiuld-go/pkg/version.Version=v1.0.1' -X 'main.buildTime={{.BuildTime}}'"
    }
  ]
}
```


# How to release your project to github (Any kind, not just golang based projects)

Now that you cross-compiled and created archives for your go project, you 
can use go-xbuild-go to publish it to GitHub.  Note: any project can be 
released to github using this tool, not just go
projects.

1. Make sure you have the GitHub CLI [gh](https://cli.github.com/) is installed. By default, the path will be
searched to find it. However, the environment variable **GH_CLI_PATH** can be
set to  specify an alternate path.

2. Set up your GitHub token:
   * Get a GitHub token from _Profile image -> Settings -> Developer Settings_

   * Click on _Personal access tokens_

   * Select _Tokens (classic)_

   * Select the Checkbox at the left side of _repo_

   * Click on _Generate token_ at the bottom

   * Save the token securely

   * Export it: **export GITHUB_TOKEN=your_github_token**

   * Create a release notes file `release_notes.md` at the root of your
   project. The options `-release-note` or `-release-notes-file` can be used
   to specify the release notes.
3. Update `VERSION` file if needed. The Release and tag with the content 
of VERSION file will be created.

Now Run:

```
go-xbuild-go \
        -release \
        -release-note "Release v1.0.1"
```

To make a formatted release note, create a file `release_notes.md`. By default, it looks for file `release_notes.md` in the current working directory. Then type:

```
go-xbuild-go -release
```

# Contributing
Pull requests welcome! Please keep it simple.

# License
MIT License - See [LICENSE](LICENSE) file for details.

# Authors
Developed with [Claude AI Sonnet 4/4.5](https://claude.ai), [Claude Code](https://code.claude.com/docs/en/overview)


---
<sub>TOC/glossary expansion by https://github.com/muquit/markdown-toc-go v1.0.5 on Jun-21-2026</sub>
