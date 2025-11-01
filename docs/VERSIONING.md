# Version Management

imgx uses a single source of truth for version information across the entire project.

## Version Files

### Primary Source
- **`version.go`** - The canonical version definition
  ```go
  const Version = "1.0.0"
  ```

### Secondary References
- **`VERSION`** - Plain text file for scripts/automation
- **CLI** (`cmd/imgx/main.go`) - Uses `imgx.Version` from version.go
- **Metadata** (`metadata_write.go`) - Uses `imgx.Version` for XMP metadata

## Automated Release Workflow

### Creating a Release (Automated)

The easiest way to release a new version is to push a git tag. GitHub Actions will automatically handle everything:

```bash
# 1. Commit your changes
git add .
git commit -m "feat: add new feature"
git push

# 2. Create and push a tag
git tag -a v1.1.0 -m "Release v1.1.0"
git push origin v1.1.0
```

**What happens automatically:**
1. GitHub Actions detects the new tag
2. Extracts version from tag (e.g., `v1.1.0` → `1.1.0`)
3. Updates `version.go` and `VERSION` files
4. Commits changes back to main branch with message: `chore: bump version to v1.1.0 [skip ci]`
5. Builds binaries for multiple platforms using GoReleaser
6. Creates GitHub release with binaries attached

### Manual Version Update (Optional)

If you need to update the version manually without creating a release:

#### Using the Script

```bash
./update-version.sh 1.1.0
```

#### Manual Steps

1. Edit `version.go`:
   ```go
   const Version = "1.1.0"  // Update this
   ```

2. Update `VERSION` file:
   ```bash
   echo "1.1.0" > VERSION
   ```

3. Commit:
   ```bash
   git commit -am "chore: bump version to v1.1.0"
   git push
   ```

## Version Information Location

The version appears in:

1. **Library metadata** - Embedded in XMP metadata as `creator_tool`
   ```
   XMP:CreatorTool = "imgx v1.0.0"
   ```

2. **CLI version flag**
   ```bash
   imgx --version
   # Output: imgx version 1.0.0
   ```

3. **Go module** - When imported by other projects
   ```go
   import "github.com/razzkumar/imgx"
   fmt.Println(imgx.Version)  // "1.0.0"
   ```

## Semantic Versioning

imgx follows [Semantic Versioning](https://semver.org/):

- **MAJOR** (1.x.x) - Incompatible API changes
- **MINOR** (x.1.x) - New functionality, backwards compatible
- **PATCH** (x.x.1) - Bug fixes, backwards compatible

### Examples

- `1.0.0` → `2.0.0` - Breaking change (removed Options field)
- `1.0.0` → `1.1.0` - New feature (added new image filter)
- `1.0.0` → `1.0.1` - Bug fix (fixed resize calculation)

## Release Checklist

When preparing a release:

- [ ] Update CHANGELOG.md (if applicable)
- [ ] Run tests: `go test ./...`
- [ ] Build CLI: `go build ./cmd/imgx`
- [ ] Test CLI: `./imgx --version`
- [ ] Commit all changes
- [ ] Create and push tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
- [ ] Push tag: `git push origin vX.Y.Z`
- [ ] GitHub Actions will handle the rest!

## Checking Current Version

### From Code
```go
import "github.com/razzkumar/imgx"

func main() {
    fmt.Println("imgx version:", imgx.Version)
}
```

### From Command Line
```bash
# CLI version
./imgx --version

# Library version
go list -m github.com/razzkumar/imgx

# Git tags
git describe --tags --abbrev=0

# From file
cat VERSION
```

## GitHub Actions Workflow

The release workflow (`.github/workflows/release.yml`) performs these steps:

1. **Extract Version** - Gets version from git tag (removes `v` prefix)
2. **Update Files** - Updates `version.go` and `VERSION`
3. **Commit** - Commits changes back to main with `[skip ci]` to avoid loops
4. **Build** - Uses GoReleaser to build binaries for:
   - Linux (amd64, arm64)
   - macOS (amd64, arm64)
   - Windows (amd64)
5. **Release** - Creates GitHub release with all binaries

### Workflow Triggers

The workflow triggers on any tag push:
```yaml
on:
  push:
    tags:
      - "*"
```

### Required Permissions

```yaml
permissions:
  contents: write  # To create releases and commit
  packages: write  # To publish packages
```

## Troubleshooting

### Version not updated after tag push

Check GitHub Actions workflow run:
```bash
# View workflow status
gh run list --workflow=release.yml

# View specific run
gh run view <run-id> --log
```

### Manual fix if workflow fails

```bash
# 1. Pull latest changes
git pull

# 2. Update version manually
./update-version.sh 1.1.0

# 3. Push changes
git push
```

## Notes

- Always use semantic versioning (MAJOR.MINOR.PATCH)
- Tag format: `v1.0.0` (with `v` prefix)
- Stored version: `1.0.0` (without `v` prefix)
- The `[skip ci]` in commit message prevents infinite workflow loops
- version.go is the source of truth for Go code
- VERSION file is for scripts and automation
- CLI automatically uses library version
- All processed images show the version in metadata
