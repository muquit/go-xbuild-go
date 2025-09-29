## Example

### Legacy mode (single binary)
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

### Multi-binary mode
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

