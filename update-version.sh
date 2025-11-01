#!/bin/bash
# update-version.sh - Update version across all files

set -e

NEW_VERSION=$1

if [ -z "$NEW_VERSION" ]; then
    echo "Usage: ./update-version.sh <version>"
    echo "Example: ./update-version.sh 1.1.0"
    echo ""
    echo "Current version: $(cat VERSION)"
    exit 1
fi

echo "Updating version to $NEW_VERSION..."
echo ""

# Update version.go
echo "→ Updating version.go"
sed -i '' "s/const Version = \".*\"/const Version = \"$NEW_VERSION\"/" version.go

# Update VERSION file
echo "→ Updating VERSION file"
echo "$NEW_VERSION" > VERSION

# Verify changes
echo ""
echo "✓ Updated to version $NEW_VERSION"
echo ""

# Show what changed
echo "Changes:"
git diff version.go VERSION

echo ""
echo "Next steps:"
echo "  1. Review changes above"
echo "  2. Test: go test && go build ./cmd/imgx && ./imgx --version"
echo "  3. Commit: git commit -am 'chore: bump version to v$NEW_VERSION'"
echo "  4. Tag: git tag -a v$NEW_VERSION -m 'Release v$NEW_VERSION'"
echo "  5. Push: git push && git push --tags"
