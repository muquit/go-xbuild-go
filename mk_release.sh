#!/bin/bash

########################################################################
# Make a release to github using gh CLI 
# muquit@muquit.com Mar-27-2025 
########################################################################

SCRIPT_VERSION="1.0.2"
usage() {
    echo "==========================================================="
    echo "A script to create a github Releases, v${SCRIPT_VERSION}"
    echo "It uses github gh CLI create the release" 
    echo ""
    echo "Usage: ${0} <release_notes.md>"
    echo " - github gh CLI must be in PATH"
    echo " - GITHUB_TOKEN env var must be set"
    echo " - VERSION file must exist"
    echo " - ./bin directory cannot be empty"
    echo "This script is part of: https://github.com/muquit/go-xbuild-go"
    echo ""
    exit 1
}
echo "$0 v${SCRIPT_VERSION}"

RELEASE_NOTES_FILE="release_notes.md"
if [[ $# -eq 1 ]]; then
    RELEASE_NOTES_FILE=$1
fi
if [[ ! -f ${RELEASE_NOTES_FILE} ]]; then
    echo "Error: Release notes file ${RELEASE_NOTES_FILE} does not exist"
    echo ""
    usage
fi

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
echo ""
echo "Processing release notes from: ${RELEASE_NOTES_FILE}"
gh release create "${VERSION}" \
    --notes-file ${RELEASE_NOTES_FILE} \
    './bin/*'

gh release list
