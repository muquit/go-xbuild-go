# go-xbuild-go

Cross compile Go projects without the complexity of [GoReleaser](https://goreleaser.com/).

It was written from the frustration of using [GoReleaser](https://goreleaser.com/). I don't 
release often, whenever the time comes to release using GoReleaser, 
something has changed.
I got tired of dealing with GoReleaser's complexity when I only release
software occasionally. When I release every 6-12 months or so, GoReleaser's
config often needs updates due to changes. This simple program just works.

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

## Download

Download pre-compiled binaries from
[Releases](https://github.com/muquit/go-xbuild-go/releases) page

## Building from source

Install [Go](https://go.dev/) first

```bash
git clone https://github.com/muquit/go-xbuild-go
cd go-xbuild-go
go build .
```

## Setup
1. Copy go-xbuild-go somewhere in your PATH
2. Copy platforms.txt to your project root
2. Create a VERSION file with your version (e.g., v1.0.1)
3. Edit platforms.txt to uncomment the platforms you want to build for

go programs can be cross compiled for more than 40 platforms. 
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
Run go-xbuild-go from the root of your project. 
```bash
go-xbuild-go
```

The program will:
1. Detect your project name from the directory
2. Read version from VERSION file
3. Build for all uncommented platforms in platforms.txt
4. Create appropriate archives (ZIP for Windows, tar.gz for others)
5. Generate checksums for all archives
6. Place all artifacts in ./bin directory

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

## Contributing
Pull requests welcome! Please keep it simple.

## License
MIT License. See LICENSE.txt

## Author
Developed with Claude AI 3.7 Sonnet, working under my guidance and instructions.

## License

MIT License - See LICENSE.txt file for details.

