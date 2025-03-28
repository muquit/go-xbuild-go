# go-xbuild-go

A [Go](https://go.dev/) program to cross compile 
[Go](https://go.dev/) projects without the complexity of [GoReleaser](https://goreleaser.com/).

It was written from the frustration of using [GoReleaser](https://goreleaser.com/). I don't 
release often, whenever the time comes to release using GoReleaser, 
something has changed.
I got tired of dealing with GoReleaser's complexity when I only release
software occasionally. When I release every 6-12 months or so, GoReleaser's
config often needs updates due to changes. This simple program just works. 
Hope you will find it useful and fun to use.

This is a [Go](https://go.dev/) port of my bash script https://github.com/muquit/go-xbuild

## Features
- Simple to use and maintain
- Cross compile for multiple platforms
- Special handling for Raspberry Pi (modern and Jessie)
- Generates checksums
- Creates archives (ZIP for Windows, tar.gz for others)
- No complex configuration files
- Just uncomment platforms in platforms.txt to build for them

## Options

```
A program to cross compile go programs
  -help
    	Show help information and exit
  -version
    	Show version information and exit
```

## Quick Start

Install [Go](https://go.dev/) first

```bash
go install github.com/muquit/go-xbuild-go@latest
go-xbuild-go -version
```
Download `platforms.txt` file. It must be copied to your pojrect's root

## Download

Download pre-compiled binaries from
[Releases](https://github.com/muquit/go-xbuild-go/releases) page

Download `platforms.txt` file. It must be copied to your pojrect's root

## Building from source

Install [Go](https://go.dev/) first

```bash
git clone https://github.com/muquit/go-xbuild-go
cd go-xbuild-go
go build .
```

## How to use

- Copy `go-xbuild-go` somewhere in your PATH. If you followed the step in *Quick Start*, it should
   be  already in your path.
- Copy `platforms.txt` to your go project's root directory.
- Create a `VERSION` file with your version (e.g., v1.0.1).
- Edit `platforms.txt` to uncomment the platforms you want to build for.

go programs can be cross compiled for more than 40 platforms. `go tool dist list` for the list
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

## How to release your project to github

Now that you cross-compiled and created archives for your go project, you can use the included release script to publish to GitHub:

1. Make sure you have the [GitHub CLI, gh](https://cli.github.com/) is installed
2. Copy `mk_release.sh` to somewhere in your PATH, just like you did with `go-xbuild-go`
3. Set up your GitHub token:
   * Get a GitHub token from _Profile image -> Settings -> Developer Settings_
   * Click on _Personal access tokens_
   * Select _Tokens (classic)_
   * Select the Checkbox at the left side of _repo_
   * Click on _Generate token_ at the bottom
   * Save the token securely
   * Export it: `export GITHUB_TOKEN=your_github_token`
5. Update `VERSION` file if needed
5. Run the release script from your project root:

```bash
   mk_release.sh "Your release notes here"
```
A Release and tag with content of VERSION file will be created

## Contributing
Pull requests welcome! Please keep it simple.

## License
MIT License - See LICENSE.txt file for details.

## Author
Developed with Claude AI 3.7 Sonnet, working under my guidance and instructions.
