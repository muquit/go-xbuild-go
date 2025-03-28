#!/bin/bash

########################################################################
# Make a release to github using gh CLI 
# muquit@muquit.com Mar-27-2025 
########################################################################

# check if github gh cli exists
if ! command -v gh &> /dev/null; then
    echo "Error: GitHub CLI (gh) is not installed or not in PATH"
    echo "Please install it from https://cli.github.com/"
    exit 1
fi

if [[ ! -f "VERSION" ]]; then
    echo "Error: VERSION file does not exist"
    exit 1
fi

VERSION=$(cat VERSION)
echo "Found version: ${VERSION}"

if [[ -z "${GITHUB_TOKEN}" ]]; then
    echo "Error: GITHUB_TOKEN environment variable is not set"
    exit 1
fi

if [[ ! -d "./bin" ]]; then
    echo "Error: ./bin directory does not exist"
    exit 1
fi

if [[ -z "$(/bin/ls -A ./bin)" ]]; then
    echo "Error: ./bin directory is empty"
    exit 1
fi

if [[ $# -eq 0 ]]; then
    echo "Error: Release notes must be provided"
    echo "Usage: $0 \"Release notes here\""
    exit 1
fi

RELEASE_NOTES="$1"

echo "Creating release ${VERSION} with provided notes..."
gh release create "${VERSION}" \
    --notes "${RELEASE_NOTES}" \
    './bin/*'

gh release list
