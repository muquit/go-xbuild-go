## Contents
- [v1.0.4](#v104)
- [v1.0.3](#v103)
- [v1.0.2](#v102)
- [v1.0.1](#v101)

# v1.0.4
- Added GitHub release functionality with the `-release` flag
- Added flexible release note options:
  - `-release-note` for specifying release notes directly in the command
  - `-release-note-file` for providing release notes from a file
  - Automatic use of "release_notes.md" if it exists and no other notes are specified
  - Combination of both text and file notes when both are provided

In order to make a release, the environment variable `GITHUB_TOKEN` must be
set. An optional environment variable `GH_CLI_PATH` can be set for the path
of github `gh` program.

  (Apr-09-2025)

# v1.0.3
* Add `platforms.txt` in the built archives (tgz/zip)
  (Apr-06-2025)

# v1.0.2
* Add flag -pi=true/false. The default is true to build for Raspberry Pi
  (Apr-02-2025)

# v1.0.1
* Initial Release
  (Mar-29-2025)
