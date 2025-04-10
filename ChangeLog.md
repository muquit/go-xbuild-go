## Contents

- [v1.0.4](#v104)
- [v1.0.3](#v103)
- [v1.0.2](#v102)
- [v1.0.1](#v101)

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
