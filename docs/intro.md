## Introduction

A multi-platform program to cross compile 
[go](https://go.dev/) projects without the complexity of [GoReleaser](https://goreleaser.com/).
The program can be used to:

- Cross compile go projects for various platforms - with ease
- **Build multi-binary Go projects** - handle projects with multiple main packages like `cmd/cli/`, `cmd/server/`, etc.
- Make releases to github - with ease. Not just go projects, **any project** can be released to github,
just copy the assets to `./bin` directory. Please look at Look at [How to release
your project to github](#how-to-release-your-project-to-github) for details.

