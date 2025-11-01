# Author/Creator Override

By default, imgx sets the author/creator metadata to "razzkumar". However, you can easily override this while keeping the `creator_tool` (software + version) fixed as "imgx v1.0.0".

## Why Override Author?

When you process images, you may want to set the artist/creator name to:
- Your own name or company name
- A specific photographer's name
- Different names for different projects

The `creator_tool` field will always remain "imgx v1.0.0" to properly credit the software used for processing.

## Methods to Override

### 1. Per-Image at Load Time (Highest Priority)

Use the `WithAuthor()` load option:

```go
img, err := imgx.Load("photo.jpg", imgx.WithAuthor("John Doe"))
// Author: John Doe
// creator_tool: imgx v1.0.0
```

### 2. After Loading

Use the `SetAuthor()` method:

```go
img, err := imgx.Load("photo.jpg")
img.SetAuthor("Jane Smith")
// Author: Jane Smith
// creator_tool: imgx v1.0.0
```

### 3. Set Global Default

Use `SetDefaultAuthor()` to set a default for all images:

```go
imgx.SetDefaultAuthor("My Company")

img1, _ := imgx.Load("photo1.jpg")
// Author: My Company

img2, _ := imgx.Load("photo2.jpg")
// Author: My Company

// Can still override per-image
img3, _ := imgx.Load("photo3.jpg", imgx.WithAuthor("Special Artist"))
// Author: Special Artist
```

### 4. Environment Variable

Set the `IMGX_DEFAULT_AUTHOR` environment variable:

```bash
export IMGX_DEFAULT_AUTHOR="Your Name"
go run yourapp.go
```

Or inline:

```bash
IMGX_DEFAULT_AUTHOR="Your Name" ./yourapp
```

## Priority Order

When multiple methods are used, the priority is:

1. **WithAuthor()** load option (per-image) - highest priority
2. **SetAuthor()** method (after load)
3. **SetDefaultAuthor()** (global programmatic)
4. **IMGX_DEFAULT_AUTHOR** environment variable
5. **Default** "razzkumar" - lowest priority

## Complete Example

```go
package main

import (
    "github.com/razzkumar/imgx"
)

func main() {
    // Method 1: Per-image override
    img1, _ := imgx.Load("photo.jpg", imgx.WithAuthor("Alice"))
    img1.Resize(800, 0, imgx.Lanczos).Save("output1.jpg")
    // Metadata: artist=Alice, creator_tool=imgx v1.0.0

    // Method 2: After loading
    img2, _ := imgx.Load("photo.jpg")
    img2.SetAuthor("Bob").Resize(800, 0, imgx.Lanczos).Save("output2.jpg")
    // Metadata: artist=Bob, creator_tool=imgx v1.0.0

    // Method 3: Global default
    imgx.SetDefaultAuthor("My Company")
    img3, _ := imgx.Load("photo.jpg")
    img3.Resize(800, 0, imgx.Lanczos).Save("output3.jpg")
    // Metadata: artist=My Company, creator_tool=imgx v1.0.0
}
```

## Metadata Fields Affected

When you set the author, it affects these XMP metadata fields:
- `XMP:Creator` - Set to your custom author
- `DC:Creator` - Set to your custom author
- `EXIF:Artist` - Set to your custom author

The following fields remain unchanged:
- `XMP:CreatorTool` - Always "imgx v1.0.0"
- `DC:Source` - Always "https://github.com/razzkumar/imgx"

## Run Example

```bash
# Run the example
go run examples/author/main.go

# Test with environment variable
IMGX_DEFAULT_AUTHOR="MyName" go run examples/author/main.go
```

## See Also

- [README.md](README.md) - Main documentation
- [examples/author/main.go](examples/author/main.go) - Complete working example
- [Metadata tracking documentation](README.md#automatic-processing-metadata-tracking)
