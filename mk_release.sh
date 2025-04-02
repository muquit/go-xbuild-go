#!/bin/bash

########################################################################
# Make a release to github using gh CLI 
# muquit@muquit.com Mar-27-2025 
########################################################################

usage() {
    echo "==========================================================="
    echo "A script to upload a project in github Releases page"
    echo "It uses github gh CLI create the release" 
    echo "Create markdown file notes.md in the current working directy"
    echo "Usage: ${0} "
    echo " - github gh CLI must be in PATH"
    echo " - GITHUB_TOKEN env var must be set"
    echo " - VERSION file must exist"
    echo " - ./bin directory cannot be empty"
    echo "This script is part of: https://github.com/muquit/go-xbuild-go"
    echo ""
    exit 1
}
# check if github gh cli exists
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed or not in PATH"
    echo "Please install it from https://cli.github.com/"
    echo ""
    usage
fi

if [[ ! -f "VERSION" ]]; then
    echo "Error: VERSION file does not exist"
    echo ""
    usage
fi

VERSION=$(cat VERSION)
echo "Found version: ${VERSION}"

if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "Error: GITHUB_TOKEN environment variable is not set"
    echo ""
    usage
fi

if [[ ! -d "./bin" ]]; then
    echo "Error: ./bin directory does not exist"
    echo ""
    usage
fi

if [[ -z "$(/bin/ls -A ./bin)" ]]; then
    echo "Error: ./bin directory is empty"
    echo ""
    usage
fi

if [[ ! -f "notes.md" ]]; then
    echo "Error: release notes file notes.md not found"
    echo "Please create nots.md with your release notes"
    echo ""
    usage
fi

echo "Creating release ${VERSION} with provided notes in ./notes.md"
pwd
gh release create "${VERSION}" \
    --notes-file ./notes.md \
    './bin/*'

gh release list
