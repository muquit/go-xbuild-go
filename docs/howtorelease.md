## How to release your project to github (Any kind, not just golang based projects)

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
