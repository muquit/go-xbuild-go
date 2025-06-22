## Contents

- [v1.0.5](#v105)
- [v1.0.4](#v104)
- [v1.0.3](#v103)
- [v1.0.2](#v102)
- [v1.0.1](#v101)

# v1.0.5
**Major feature release**: Added support for multi-binary Go projects while maintaining full backward compatibility.

**Migration**: Existing projects work unchanged. To use multi-binary features, create a `build-config.json` file (see documentation for examples).
- **Multi-target configuration system**: New `-config` flag accepts JSON configuration file for building multiple binaries from a single project
  - Example: `go-xbuild-go -config build-config.json`
Please look at
[go-multi-main-example](https://github.com/muquit/go-multi-main-example) for
an example
- **Multiple main package support**: Can now build projects with `cmd/cli/main.go`, `cmd/server/main.go`, etc. structure
- **Flexible build paths**: Each target can specify custom build path like `./cmd/cli` or `./cmd/server`
- **Per-target customization**:
 - Individual ldflags and build flags per binary
 - Target-specific additional files (e.g., different config files per component)
 - Custom output binary names
- **Unified cross-compilation**: Single command builds all project binaries across all platforms
- **Independent archives**: Each binary gets its own platform-specific archives and checksums
- **New `-list-targets` flag**: Shows all available build targets from configuration file
- **Enhanced help system**: Displays example JSON configuration in help output
- **Backward compatibility**: Simple single-binary projects continue to work exactly as before without any configuration file

Modern Go projects using the standard `cmd/` directory layout can now build all components (CLI tools, servers, admin utilities) in one unified cross-platform build process.

(May-21-2025)

# v1.0.4
github release can be made the CLI itself. Before a script `mk_release.sh` was used.
- Added GitHub release functionality with the `-release` flag
- Added flexible release note options:
  - `-release-note` for specifying release notes directly in the command
  - `-release-note-file` for providing release notes from a file
  - Automatic use of "release_notes.md" if it exists and no other notes are specified
  - Combination of both text and file notes when both are provided

In order to make a release, the environment variable `GITHUB_TOKEN` must be
set. An optional environment variable `GH_CLI_PATH` can be set for the path
of github `gh` program.

- Added `-additional-files` flag for specifying extra files to include in all archives

  - Takes a comma-separated list of file paths
  - Files are copied into the archives if they exist
  - Provides feedback about successful additions and warnings for missing files

(Apr-09-2025)

# v1.0.3

- Add `platforms.txt` in the built archives (tgz/zip)

(Apr-06-2025)

# v1.0.2

- Add flag -pi=true/false. The default is true to build for Raspberry Pi

(Apr-02-2025)

# v1.0.1

- Initial Release

(Mar-29-2025)
