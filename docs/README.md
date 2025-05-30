## Introduction

A multi-platform program to cross compile 
[go](https://go.dev/) projects without the complexity of [GoReleaser](https://goreleaser.com/).
The program can be used to:

- Cross compile go projects for various platforms - with ease
- Make releases to github - with ease. Not just go projects, **any project** can be released to github,
just copy the assets to `./bin` directory. Please look at Look at [How to release
your project to github](#how-to-release-your-project-to-github) for details.

## Background and Motivation

It was written from the frustration of using [GoReleaser](https://goreleaser.com/). I don't 
release often, whenever the time comes to release using GoReleaser, 
something has changed.
I got tired of dealing with GoReleaser's complexity when I only release
software occasionally. When I release every 6-12 months or so, GoReleaser's
config often needs updates due to changes. This simple program just works. 
Hope you will find it useful and fun to use.

This is a [go](https://go.dev/) port of my bash script https://github.com/muquit/go-xbuild

Pull requests, suggestions are always welcome.


## How to use

There is no configuration file to edit.

- Copy `go-xbuild-go` somewhere in your PATH. 
- Copy `platforms.txt` to your go project's root directory or to the directory where you build your project from
- Create a `VERSION` file with your version (e.g., v1.0.1) at the root of your
project or where you build the project from
- Edit `platforms.txt` to uncomment the platforms you want to build for.

Then type:
```
go-build-go
```
tgz or zip archive will be created inside `./bin` directory

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

## Features
- Simple to use and maintain
- Cross compile for multiple platforms
- Special handling for Raspberry Pi (modern and Jessie)
- Generates checksums
- Creates archives (ZIP for Windows, tar.gz for others)
- No complex configuration files
- Just uncomment platforms in platforms.txt to build for them
- Make release of the project to github

## Synopsis

```
A program to cross compile go programs

Environment variables (for github release):
  GITHUB_TOKEN     GitHub API token (required for -release)
  GH_CLI_PATH      Custom path to GitHub CLI executable (optional)

Usage:
  - Copy platforms.txt at the root of your project
  - Edit platforms.txt to uncomment the platforms you want to build for
  - Create a VERSION file with your version (e.g. v1.0.1)
  - Then run ./go-xbuild-go

Options:
  -additional-files string
    	Comma-separated list of additional files to include in archives
  -help
    	Show help information and exit
  -pi
    	Build Raspberry Pi (default true)
  -release
    	Create a GitHub release
  -release-note string
    	Release note text (required if -release-note-file not specified and release_notes.md doesn't exist)
  -release-note-file string
    	File containing release notes (required if -release-note not specified and release_notes.md doesn't exist)
  -version
    	Show version information and exit
```

## Version
The current version is 1.0.4

Please look at [ChangeLog](ChangeLog.md) for what has changed in the current version.

## Installation

### Download

Download pre-compiled binaries from
[Releases](https://github.com/muquit/go-xbuild-go/releases) page

Please look at [How to use](#how-to-use) 


### Building from source

Install [go](https://go.dev/) first

```bash
git clone https://github.com/muquit/go-xbuild-go
cd go-xbuild-go
go build .
```
Please look at [How to use](#how-to-use) 


## Usage
Run go-xbuild-go from the root of your project.  Update VERSION file if needed.
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
- Add extra files with `-additional-files`

## How to release your project to github

Now that you cross-compiled and created archives for your go project, you 
can use go-xbuild-go to publish it to GitHub.  Note: any project can be 
released to github using this tool, not just go
projects.

1. Make sure you have the GitHub CLI [gh](https://cli.github.com/) is installed. By default, the path will be
searched to find it. However, the environment variable **GH_CLI_PATH** can be
seto specify an alternate path.

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
or
```
go-xbuild-go \
        -release \
        -release-note-file "release_notes.md"
```
or
```
go-xbuild-go \
        -release \
        -release-note "Release v1.0.x" \
        -release-note-file "release_notes.md"

```
or if `release_notes.md` exists in the current working directory:
```
go-xbuild-go -release
```



By default, it looks file `release_notes.md` in the current working directory. 

## Contributing
Pull requests welcome! Please keep it simple.

## License
MIT License - See LICENSE file for details.

## Author
Developed with Claude AI 3.7 Sonnet, working under my guidance and instructions.
