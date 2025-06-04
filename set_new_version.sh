#!/bin/bash

# Read current version from VERSION file
current_version=$(cat VERSION)

# Prompt for new version
echo "Current version is $current_version"
read -p "Enter new version: " version

# Prompt for release notes
echo "Enter release notes (press Ctrl+D when done):"
release_notes=$(cat)

echo "$version" > VERSION
git add VERSION
git commit -m "Bump version to $version"

# Create annotated tag with release notes
git tag -a "v$version" -m "Version $version

$release_notes"

git push && git push origin "v$version"