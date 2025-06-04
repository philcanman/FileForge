#!/bin/bash
version="1.0.7.1"

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