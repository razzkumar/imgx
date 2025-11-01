# Releasing imgx

Quick guide for maintainers on how to release new versions.

## Quick Release

```bash
# 1. Make sure main branch is up to date
git checkout main
git pull

# 2. Run tests
go test ./...

# 3. Create and push a tag (GitHub Actions does the rest!)
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

That's it! GitHub Actions will automatically:
- ✅ Update version in code files
- ✅ Commit changes back to main
- ✅ Build binaries for all platforms
- ✅ Create GitHub release

## What Happens Automatically

When you push a tag like `v1.1.0`, the release workflow:

1. **Extracts Version**
   - Tag: `v1.1.0` → Version: `1.1.0`

2. **Updates Files**
   - `version.go`: `const Version = "1.1.0"`
   - `VERSION`: `1.1.0`

3. **Commits Back to Main**
   ```
   chore: bump version to v1.1.0 [skip ci]
   ```

4. **Builds Binaries**
   - Linux: amd64, arm64
   - macOS: amd64, arm64
   - Windows: amd64

5. **Creates Release**
   - GitHub release with all binaries
   - Checksums and SBOMs
   - Release notes

## Pre-Release Checklist

Before creating a release tag:

- [ ] All tests pass: `go test ./...`
- [ ] CLI builds: `go build ./cmd/imgx`
- [ ] CLI works: `./imgx --version`
- [ ] Examples work: `go run examples/author/main.go`
- [ ] Documentation is updated
- [ ] CHANGELOG is updated (if you maintain one)
- [ ] No uncommitted changes: `git status`

## Version Numbers

Follow [Semantic Versioning](https://semver.org/):

- **v1.0.0 → v2.0.0** - Breaking changes
  - Example: Removed `Options` field, changed function signature

- **v1.0.0 → v1.1.0** - New features
  - Example: Added new image filter, new CLI command

- **v1.0.0 → v1.0.1** - Bug fixes
  - Example: Fixed memory leak, corrected calculation

## Release Types

### Stable Release

```bash
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

### Pre-Release (Beta, RC)

```bash
git tag -a v1.1.0-beta.1 -m "Beta release v1.1.0-beta.1"
git push origin v1.1.0-beta.1
```

Mark as pre-release on GitHub after creation.

## Manual Release (Emergency)

If GitHub Actions fails, you can release manually:

### 1. Update Version

```bash
./update-version.sh 1.1.0
git commit -am "chore: bump version to v1.1.0"
git push
```

### 2. Build Locally

```bash
# Install goreleaser
go install github.com/goreleaser/goreleaser@latest

# Create snapshot (no tag needed)
goreleaser release --snapshot --clean

# Or full release (requires tag)
git tag -a v1.1.0 -m "Release v1.1.0"
goreleaser release --clean
```

### 3. Upload to GitHub

Manually upload binaries from `dist/` to GitHub release.

## Viewing Releases

```bash
# List all tags
git tag -l

# Show latest tag
git describe --tags --abbrev=0

# View release on GitHub
gh release view v1.1.0

# List all releases
gh release list
```

## Hotfix Release

For urgent bug fixes:

```bash
# 1. Create hotfix branch
git checkout -b hotfix/1.0.1 v1.0.0

# 2. Fix the bug
git commit -am "fix: critical bug"

# 3. Tag and push
git tag -a v1.0.1 -m "Hotfix v1.0.1"
git push origin v1.0.1

# 4. Merge back
git checkout main
git merge hotfix/1.0.1
git push
```

## Rollback

If you need to roll back a release:

```bash
# 1. Delete tag locally
git tag -d v1.1.0

# 2. Delete tag on remote
git push origin :refs/tags/v1.1.0

# 3. Delete GitHub release
gh release delete v1.1.0

# 4. Revert version commit
git revert <commit-hash>
git push
```

## Testing Before Release

```bash
# Run all tests
go test ./...

# Test with race detector
go test -race ./...

# Build for all platforms
GOOS=linux GOARCH=amd64 go build ./cmd/imgx
GOOS=darwin GOARCH=arm64 go build ./cmd/imgx
GOOS=windows GOARCH=amd64 go build ./cmd/imgx

# Or use goreleaser snapshot
goreleaser release --snapshot --clean
```

## Monitoring Release

After pushing a tag:

```bash
# Watch workflow progress
gh run watch

# View workflow logs
gh run list --workflow=release.yml
gh run view <run-id> --log

# Check release
gh release view v1.1.0
```

## Troubleshooting

### Workflow fails to commit

**Problem:** Permission denied when pushing version commit.

**Solution:** Check that `GITHUB_TOKEN` has `contents: write` permission in workflow file.

### Version not updated

**Problem:** version.go still shows old version after release.

**Solution:**
```bash
git pull  # Pull the automated commit
```

### Build fails

**Problem:** GoReleaser build fails.

**Solution:** Check `.goreleaser.yml` configuration and workflow logs.

### Tag already exists

**Problem:** Tag `v1.1.0` already exists.

**Solution:**
```bash
# Delete and recreate
git tag -d v1.1.0
git push origin :refs/tags/v1.1.0
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

## Tips

- 🏷️ Always use annotated tags (`-a`), not lightweight tags
- 📝 Write clear commit messages
- ✅ Test thoroughly before tagging
- 🔄 Pull after workflow completes to get version commit
- 📊 Monitor GitHub Actions for any failures
- 🚀 Announce releases in relevant channels

## See Also

- [VERSIONING.md](VERSIONING.md) - Version management details
- [.github/workflows/release.yml](.github/workflows/release.yml) - Release workflow
- [.goreleaser.yml](.goreleaser.yml) - Build configuration
