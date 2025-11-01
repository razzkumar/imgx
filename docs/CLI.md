# imgx CLI Documentation

A powerful command-line tool for image processing operations including resizing, transformations, color adjustments, effects, and watermarking.

## Table of Contents

- [Installation](#installation)
- [Shell Completion](#shell-completion)
- [Quick Start](#quick-start)
- [Global Options](#global-options)
- [Commands](#commands)
  - [Resize Operations](#resize-operations)
  - [Transform Operations](#transform-operations)
  - [Color Adjustments](#color-adjustments)
  - [Effects](#effects)
  - [Watermarking](#watermarking)
  - [Image Information](#image-information)
  - [Object Detection](#object-detection)
- [Common Use Cases](#common-use-cases)
- [Tips & Tricks](#tips--tricks)

## Installation

```bash
# Build from source
go build -o imgx ./cmd/imgx

# Or install
go install github.com/razzkumar/imgx/cmd/imgx@latest
```

## Shell Completion

imgx supports shell completion for Bash, Zsh, Fish, and PowerShell. This enables tab completion for commands, flags, and options.

### Bash

**Temporary (current session only):**
```bash
source <(imgx completion bash)
```

**Permanent:**
```bash
# Save to completion directory
imgx completion bash > ~/.bash_completion.d/imgx
source ~/.bash_completion.d/imgx

# Or add to .bashrc
echo 'source <(imgx completion bash)' >> ~/.bashrc
```

### Zsh

**Add to `.zshrc`:**
```bash
# Enable completions if not already enabled
autoload -Uz compinit
compinit

# Load imgx completions
source <(imgx completion zsh)

# Or add this line to .zshrc
echo 'source <(imgx completion zsh)' >> ~/.zshrc
```

**Alternative (using completion directory):**
```bash
# Save to zsh completion directory
imgx completion zsh > "${fpath[1]}/_imgx"
```

### Fish

```bash
# Save to fish completion directory
imgx completion fish > ~/.config/fish/completions/imgx.fish
```

### PowerShell

```powershell
# Generate completion script
imgx completion pwsh > imgx.ps1

# Load completions (add to your PowerShell profile)
& path\to\imgx.ps1
```

**To find your PowerShell profile location:**
```powershell
echo $PROFILE
```

### Testing Completions

After setting up completions, test them by typing:

```bash
imgx <TAB>          # Shows all available commands
imgx resize <TAB>   # Shows subcommands and help
imgx --<TAB>        # Shows global flags (--output, --quality, etc.)
```

### What Completions Support

The shell completions currently support:

✅ **Command completion** - Complete command names (resize, adjust, blur, etc.)
✅ **Subcommand completion** - Navigate through command hierarchy
✅ **Global flag completion** - Complete global flags at the root level (--output, --quality, --verbose)
⚠️ **Subcommand flags** - Limited support for flags within subcommands (use `imgx <command> --help` to see available flags)

### ⚠️ Important: Shell Completion Quirk

When you type `imgx <command> <TAB>`, the completion will suggest `help` as an option. However:

- **DON'T use:** `imgx thumbnail help` ❌ (will fail with "Required flag not set" error)
- **DO use instead:** `imgx help thumbnail` ✅ or `imgx thumbnail --help` ✅

The completion system suggests `help` because it's technically a subcommand, but it doesn't work correctly with commands that have required flags. Always use one of the correct help syntaxes shown above.

**Note:** To see all available flags for a specific command, use the help system:
```bash
imgx adjust --help     # Shows all adjust command flags
imgx resize --help     # Shows all resize command flags
```

## Quick Start

```bash
# Resize an image
imgx resize photo.jpg -w 800 -o resized.jpg

# Create a thumbnail
imgx thumbnail photo.jpg -s 150 -o thumb.jpg

# Adjust colors
imgx adjust photo.jpg --brightness 10 --contrast 20 -o adjusted.jpg

# Get image info
imgx info photo.jpg
```

## Global Options

Global options can be used with any command:

| Flag | Description | Default |
|------|-------------|---------|
| `-o, --output <path>` | Output file path | Auto-generated |
| `-q, --quality <1-100>` | JPEG quality | 95 |
| `--auto-orient` | Auto-orient based on EXIF data | false |
| `--format <fmt>` | Force output format (jpg, png, gif, tiff, bmp) | Detected from filename |
| `-v, --verbose` | Verbose output | false |
| `--help, -h` | Show help | |
| `--version` | Show version | |

**Examples:**

```bash
# Save as JPEG with quality 90
imgx resize photo.png -w 800 -o output.jpg --quality 90

# Force PNG format with verbose output
imgx grayscale photo.jpg --format png -v -o output.png

# Auto-orient image based on EXIF before processing
imgx resize photo.jpg -w 800 --auto-orient -o output.jpg
```

## Commands

### Resize Operations

#### `resize` - Resize to specific dimensions

Resize an image to the specified width and height. If one dimension is 0, the aspect ratio is preserved.

```bash
imgx resize <input> [options]
```

**Options:**
- `-w, --width <int>` - Target width (0 to preserve aspect ratio)
- `-h, --height <int>` - Target height (0 to preserve aspect ratio)
- `-f, --filter <name>` - Resampling filter (default: lanczos)

**Available Filters:**
`lanczos`, `catmullrom`, `mitchellnetravali`, `linear`, `box`, `nearest`, `hermite`, `bspline`, `gaussian`, `hann`, `hamming`, `blackman`, `bartlett`, `welch`, `cosine`

**Examples:**

```bash
# Resize to 800x600
imgx resize photo.jpg -w 800 -h 600 -o output.jpg

# Resize to width 800, preserve aspect ratio
imgx resize photo.jpg -w 800 -o output.jpg

# Resize with different filter
imgx resize photo.jpg -w 800 -f catmullrom -o output.jpg
```

#### `fit` - Scale to fit within bounds

Scale the image to fit within the specified dimensions while preserving aspect ratio.

```bash
imgx fit <input> -w <width> -h <height> [options]
```

**Options:**
- `-w, --width <int>` - Maximum width (required)
- `-h, --height <int>` - Maximum height (required)
- `-f, --filter <name>` - Resampling filter (default: lanczos)

**Example:**

```bash
# Fit image within 800x600 bounding box
imgx fit photo.jpg -w 800 -h 600 -o output.jpg
```

#### `fill` - Crop and resize to exact dimensions

Resize and crop the image to fill the specified dimensions exactly. The image is scaled to cover the target size, then cropped to fit.

```bash
imgx fill <input> -w <width> -h <height> [options]
```

**Options:**
- `-w, --width <int>` - Target width (required)
- `-h, --height <int>` - Target height (required)
- `-a, --anchor <pos>` - Anchor position (default: center)
- `-f, --filter <name>` - Resampling filter (default: lanczos)

**Anchor Positions:**
`center`, `topleft`, `top`, `topright`, `left`, `right`, `bottomleft`, `bottom`, `bottomright`

**Examples:**

```bash
# Fill 800x600 with center crop
imgx fill photo.jpg -w 800 -h 600 -o output.jpg

# Fill with top-left anchor
imgx fill photo.jpg -w 800 -h 600 --anchor topleft -o output.jpg
```

#### `thumbnail` - Create square thumbnail

Create a square thumbnail by cropping and resizing.

```bash
imgx thumbnail <input> -s <size> [options]
```

**Options:**
- `-s, --size <int>` - Thumbnail size (width and height) (required)
- `-f, --filter <name>` - Resampling filter (default: lanczos)

**Example:**

```bash
imgx thumbnail photo.jpg -s 150 -o thumb.jpg
```

### Transform Operations

#### `rotate` - Rotate by angle

Rotate an image by the specified angle in degrees. Positive angles rotate counter-clockwise, negative angles rotate clockwise. Rotations of 90, 180, and 270 degrees are lossless.

```bash
imgx rotate <input> -a <angle> [options]
```

**Options:**
- `-a, --angle <float>` - Rotation angle in degrees (required)
- `--bg <color>` - Background color for empty areas (default: 00000000 = transparent)

**Color Format:** RGB hex (`ffffff`) or RGBA hex (`ff0000ff`)

**Examples:**

```bash
# Rotate 90 degrees counter-clockwise
imgx rotate photo.jpg -a 90 -o output.jpg

# Rotate 45 degrees with white background
imgx rotate photo.jpg -a 45 --bg ffffff -o output.jpg

# Rotate 30 degrees clockwise (negative angle)
imgx rotate photo.jpg -a -30 -o output.jpg
```

#### Quick Rotation Commands

For common rotations, use these shortcuts:

```bash
# 90 degrees counter-clockwise
imgx rotate90 photo.jpg -o output.jpg

# 180 degrees
imgx rotate180 photo.jpg -o output.jpg

# 270 degrees counter-clockwise (90 clockwise)
imgx rotate270 photo.jpg -o output.jpg
```

#### `flip` - Flip horizontally or vertically

Flip an image horizontally (left-right), vertically (top-bottom), or both.

```bash
imgx flip <input> [options]
```

**Options:**
- `--horizontal, -H` - Flip horizontally (left-right)
- `--vertical, -V` - Flip vertically (top-bottom)

**Examples:**

```bash
# Flip horizontally
imgx flip photo.jpg --horizontal -o output.jpg

# Flip vertically
imgx flip photo.jpg --vertical -o output.jpg

# Flip both (same as rotate 180)
imgx flip photo.jpg --horizontal --vertical -o output.jpg
```

#### `crop` - Crop to region

Crop an image to a specific region using either anchor positioning or exact coordinates.

```bash
imgx crop <input> -w <width> -h <height> [options]
```

**Options:**
- `-w, --width <int>` - Crop width (required)
- `-h, --height <int>` - Crop height (required)
- `-a, --anchor <pos>` - Anchor position (default: center)
- `-x <int>` - X coordinate (left edge, exclusive with --anchor)
- `-y <int>` - Y coordinate (top edge, exclusive with --anchor)

**Examples:**

```bash
# Crop 500x400 from center
imgx crop photo.jpg -w 500 -h 400 --anchor center -o output.jpg

# Crop from specific coordinates
imgx crop photo.jpg -x 100 -y 100 -w 500 -h 400 -o output.jpg

# Crop from top-left
imgx crop photo.jpg -w 500 -h 400 --anchor topleft -o output.jpg
```

#### `transpose` / `transverse` - Advanced transforms

Special transformation operations:

```bash
# Transpose: flip horizontally + rotate 90° CCW
imgx transpose photo.jpg -o output.jpg

# Transverse: flip vertically + rotate 90° CCW
imgx transverse photo.jpg -o output.jpg
```

### Color Adjustments

#### `adjust` - Adjust colors

Adjust various color properties of an image. Multiple adjustments can be applied at once and are processed in order: brightness → contrast → gamma → saturation → hue.

```bash
imgx adjust <input> [options]
```

**Options:**
- `--brightness <float>` - Brightness adjustment (-100 to 100, 0 = no change)
- `--contrast <float>` - Contrast adjustment (-100 to 100, 0 = no change)
- `--gamma <float>` - Gamma correction (positive number, 1.0 = no change)
- `--saturation <float>` - Saturation adjustment (-100 to 100, 0 = no change)
- `--hue <float>` - Hue shift in degrees (-180 to 180, 0 = no change)

**Examples:**

```bash
# Increase brightness and contrast
imgx adjust photo.jpg --brightness 10 --contrast 20 -o output.jpg

# Adjust saturation and hue
imgx adjust photo.jpg --saturation -30 --hue 60 -o output.jpg

# Apply gamma correction
imgx adjust photo.jpg --gamma 1.5 -o output.jpg

# Multiple adjustments at once
imgx adjust photo.jpg --brightness 10 --contrast 15 --saturation 20 --gamma 1.2 -o output.jpg
```

#### `grayscale` - Convert to grayscale

Convert an image to grayscale using ITU-R BT.601 luminance weights.

```bash
imgx grayscale <input> [options]
```

**Example:**

```bash
imgx grayscale photo.jpg -o output.jpg
```

#### `invert` - Invert colors

Invert (negate) all colors in the image to create a negative effect.

```bash
imgx invert <input> [options]
```

**Example:**

```bash
imgx invert photo.jpg -o output.jpg
```

### Effects

#### `blur` - Gaussian blur

Apply a Gaussian blur effect to the image. Higher sigma values produce stronger blur.

```bash
imgx blur <input> -s <sigma> [options]
```

**Options:**
- `-s, --sigma <float>` - Blur strength (required, positive number, typical: 0.5-10)

**Examples:**

```bash
# Subtle blur
imgx blur photo.jpg --sigma 1.5 -o output.jpg

# Strong blur
imgx blur photo.jpg -s 5.0 -o output.jpg
```

#### `sharpen` - Sharpen image

Sharpen the image using unsharp masking. Higher sigma values produce stronger sharpening.

```bash
imgx sharpen <input> -s <sigma> [options]
```

**Options:**
- `-s, --sigma <float>` - Sharpening strength (required, positive number, typical: 0.5-5)

**Examples:**

```bash
# Moderate sharpening
imgx sharpen photo.jpg --sigma 1.5 -o output.jpg

# Strong sharpening
imgx sharpen photo.jpg -s 3.0 -o output.jpg
```

### Watermarking

#### `watermark` - Add text watermark

Add a text watermark to an image with configurable position, opacity, color, and padding.

```bash
imgx watermark <input> -t <text> [options]
```

**Options:**
- `-t, --text <string>` - Watermark text (required)
- `--opacity <float>` - Opacity (0.0 to 1.0, default: 0.5)
- `-a, --anchor <pos>` - Position (default: bottomright)
- `--color <color>` - Text color in hex (default: ffffff = white)
- `--padding <int>` - Padding from edges in pixels (default: 10)

**Color Format:** RGB hex (`ffffff`) or RGBA hex (`ff0000ff`)

**Examples:**

```bash
# Simple copyright watermark
imgx watermark photo.jpg --text "Copyright 2025" -o output.jpg

# Draft watermark in center with red color
imgx watermark photo.jpg --text "DRAFT" --opacity 0.3 --anchor center --color ff0000 -o output.jpg

# Top-left watermark with custom padding
imgx watermark photo.jpg --text "Sample" --anchor topleft --padding 20 -o output.jpg

# Semi-transparent watermark with RGBA color
imgx watermark photo.jpg --text "Watermark" --color ff000080 -o output.jpg
```

### Image Information

#### `info` - Display image information

Display detailed information about an image file.

```bash
imgx info <input> [options]
```

**Options:**
- `-e, --extended` - Show extended metadata (requires exiftool)

**Output includes:**
- File path
- Image format (JPEG, PNG, GIF, TIFF, BMP)
- Dimensions (width × height)
- File size
- Color model

**With `--extended` flag (requires exiftool):**
- Camera information and settings
- GPS location data
- Date/time metadata
- Copyright and authorship

**Examples:**

```bash
# Basic info
imgx info photo.jpg

# Extended metadata
imgx info photo.jpg --extended
```

**Sample basic output:**

```
File: photo.jpg
Format: JPEG
Dimensions: 1920x1080
Size: 245.3 KB
Color Model: *color.modelFunc
```

**Sample extended output:**

```
File: photo.jpg
Format: JPEG
Dimensions: 1920x1080
Size: 245.3 KB
Color Model: *color.modelFunc

Camera:
  Make: Canon
  Model: Canon EOS 5D Mark IV

Date Taken: 2024:10:15 14:23:45

Settings:
  Focal Length: 50.0 mm
  Aperture: f/1.8
  ISO: 400
```

#### `metadata` - Extract comprehensive metadata

Extract and display comprehensive image metadata including EXIF, IPTC, and XMP data.

```bash
imgx metadata <input> [options]
```

**Options:**
- `-b, --basic` - Show basic metadata only (skip exiftool)
- `-j, --json` - Output metadata as JSON

**Features:**

When **exiftool is installed**, displays:
- **Camera Information:** Make, model, lens
- **Camera Settings:** ISO, aperture (f-number), shutter speed, focal length, flash
- **GPS Location:** Latitude, longitude, altitude
- **Date/Time:** Original capture date, modification date
- **Additional Info:** Software, artist, copyright
- **File Details:** Format, dimensions, aspect ratio, megapixels, file size, color model

When **exiftool is NOT installed**, displays basic metadata:
- File path, format, and size
- Image dimensions and aspect ratio
- Megapixels and color model
- Warning message with installation instructions

**Examples:**

```bash
# Display all available metadata
imgx metadata photo.jpg

# Show basic metadata only (skip exiftool)
imgx metadata photo.jpg --basic

# Output as JSON for parsing
imgx metadata photo.jpg --json > metadata.json
```

**Sample output (with exiftool):**

```
=== Image Metadata ===

File Information:
  Path:        photo.jpg
  Format:      JPEG
  Size:        2.3 MB

Image Properties:
  Dimensions:  4000x3000
  Aspect Ratio: 1.33
  Megapixels:  12.00 MP
  Color Model: *color.modelFunc

Camera Information:
  Make:        Canon
  Model:       Canon EOS 5D Mark IV
  Lens:        EF50mm f/1.8 STM

Camera Settings:
  Focal Length: 50.0 mm
  Aperture:    f/1.8
  Shutter:     1/250
  ISO:         400
  Flash:       No Flash

Date/Time:
  Original:    2024:10:15 14:23:45
  Modified:    2024:10:20 09:15:30

GPS Location:
  Latitude:    37.7749 N
  Longitude:   122.4194 W
  Altitude:    15.0 m

Additional Information:
  Software:    Adobe Photoshop CC 2024
  Copyright:   © 2024 John Doe
```

**Sample output (without exiftool):**

```
=== Image Metadata ===

File Information:
  Path:        photo.jpg
  Format:      JPEG
  Size:        2.3 MB

Image Properties:
  Dimensions:  4000x3000
  Aspect Ratio: 1.33
  Megapixels:  12.00 MP
  Color Model: *color.modelFunc

---

exiftool not found. Install exiftool for comprehensive metadata.

Installation:
  macOS:    brew install exiftool
  Ubuntu:   sudo apt-get install libimage-exiftool-perl
  Windows:  https://exiftool.org
```

**JSON Output Example:**

```bash
imgx metadata photo.jpg --json
```

```json
{
  "FilePath": "photo.jpg",
  "Format": "JPEG",
  "Width": 4000,
  "Height": 3000,
  "FileSize": 2415919,
  "Megapixels": 12.00,
  "AspectRatio": 1.33,
  "HasExtended": true,
  "CameraMake": "Canon",
  "CameraModel": "Canon EOS 5D Mark IV",
  "ISO": "400",
  "FocalLength": "50.0 mm",
  "DateTimeOriginal": "2024:10:15 14:23:45"
}
```

**Installing exiftool:**

```bash
# macOS
brew install exiftool

# Ubuntu/Debian
sudo apt-get install libimage-exiftool-perl

# Fedora/RHEL
sudo dnf install perl-Image-ExifTool

# Windows
# Download from https://exiftool.org and add to PATH

# Verify installation
exiftool -ver
```

### Object Detection

#### `detect` - AI-powered object detection

Detect objects, text, faces, and image properties using AI vision APIs from Google Gemini, AWS Rekognition, and OpenAI Vision.

```bash
imgx detect <input> [options]
```

**Options:**
- `-p, --provider string` - Detection provider: `gemini`, `google` (alias), `aws`, `openai` (default: `gemini`)
- `-f, --features string` - Features to detect: `labels,text,faces,web,description,properties` (comma-separated, default: `labels`)
- `-m, --max-results int` - Maximum number of labels to return (default: 10)
- `-c, --confidence float` - Minimum confidence threshold 0.0-1.0 (default: 0.5)
- `--prompt string` - Custom prompt for Gemini/OpenAI (overrides --features)
- `-j, --json` - Output results as JSON (includes colors, quality, moderation when available)
- `--raw` - Include raw API response in output

**Supported Providers:**
- **gemini** (Google Gemini API) - Requires `GEMINI_API_KEY`
- **aws** (AWS Rekognition) - Requires AWS credentials
- **openai** (OpenAI Vision) - Requires `OPENAI_API_KEY`

**Setup:**

```bash
# Gemini: Get API key from https://aistudio.google.com/
export GEMINI_API_KEY="your-api-key"

# AWS: Configure via AWS CLI or environment variables
aws configure
# OR
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
export AWS_REGION="us-east-1"

# OpenAI: Get API key from https://platform.openai.com/
export OPENAI_API_KEY="sk-..."
```

**Available Features:**
- `labels` - Detect objects and labels
- `text` - Extract text (OCR)
- `faces` - Detect faces and attributes
- `description` - Get natural language description (Gemini/OpenAI)
- `web` - Web entities and similar images (Gemini only)
- `landmarks` - Detect famous landmarks (Gemini only)
- `properties` - Image quality, colors, sharpness (AWS only)
- `safesearch` - Content moderation

**Examples:**

```bash
# Basic detection with Gemini (simplest)
imgx detect photo.jpg

# Use specific provider
imgx detect photo.jpg --provider aws

# Multiple features
imgx detect document.jpg --features labels,text,faces

# AWS image properties (colors, quality)
imgx detect photo.jpg --provider aws --features properties

# AWS labels + properties (charged for both)
imgx detect photo.jpg --provider aws --features labels,properties

# Custom prompt with Gemini
imgx detect dog.jpg --provider gemini --prompt "What breed is this dog?"

# Higher confidence threshold
imgx detect photo.jpg --confidence 0.8

# JSON output
imgx detect photo.jpg --json

# Compare providers
imgx detect photo.jpg --provider gemini
imgx detect photo.jpg --provider aws
imgx detect photo.jpg --provider openai
```

**Sample Output (pretty format):**

```
=== Object Detection Results (aws) ===

Labels:
  1. Dog (99.0% confidence)
  2. Pet (97.6% confidence)
  3. Animal (96.3% confidence)

Detected Text:
  1. "Backyard" (88.0% confidence)

Properties:
  brightness: 84.50
  contrast: 76.20
  foreground_color: Green

Dominant Colors:
  - Green (#8FB94B, rgb(143,185,75), 41.3%)
  - Brown (#6B4E2E, rgb(107,78,46), 18.6%)

Image Quality:
  Brightness: 84.50
  Contrast: 76.20
  Sharpness: 80.10
  Foreground:
    Brightness: 81.20
    Color: Green

Moderation:
  - Violence (severity: VERY_UNLIKELY, confidence: 0.5%)

Safe Search Summary:
  - Violence (VERY_UNLIKELY, 0.5%)

Overall Confidence: 97.6%

Processed at: 2024-11-01 16:30:45
```

**JSON Output Example:**

```json
{
  "provider": "gemini",
  "labels": [
    {
      "name": "Dog",
      "confidence": 0.985,
      "categories": ["Animal", "Pet"]
    },
    {
      "name": "Golden Retriever",
      "confidence": 0.942,
      "categories": ["Dog", "Breed"]
    }
  ],
  "confidence": 0.934,
  "processed_at": "2024-11-01T16:30:45Z"
}
```

**AWS Image Properties Output:**

```bash
imgx detect photo.jpg --provider aws --features properties

# Output:
Provider: aws
Processed: 2024-11-01 16:30:45

Image Properties:
  brightness: 75.23
  sharpness: 89.15
  contrast: 62.45
  dominant_colors: Black(45.2%), White(23.1%), Blue(15.7%)
  color_1_hex: #0C0F12
  color_1_rgb: rgb(12,15,18)
  color_2_hex: #F5F5F5
  color_2_rgb: rgb(245,245,245)
  foreground_color: Black
  background_color: White
```

**Text Extraction Example:**

```bash
imgx detect document.jpg --features text

# Output:
Provider: gemini
Processed: 2024-11-01 16:30:45

Text Detected:
  - "Invoice" (confidence: 99.2%)
  - "Total: $152.50" (confidence: 98.7%)
  - "Date: 2024-10-15" (confidence: 97.5%)
```

**Face Detection Example:**

```bash
imgx detect group-photo.jpg --features faces

# Output:
Provider: aws
Processed: 2024-11-01 16:30:45

Faces Detected: 3
  Face 1: confidence 99.8%, age 25-35, smiling
  Face 2: confidence 99.5%, age 30-40, neutral
  Face 3: confidence 98.9%, age 20-30, happy
```

**Pricing Notes:**

- **Gemini**: Free tier available, then pay-as-you-go
- **AWS Rekognition**:
  - `--features labels` → Standard DetectLabels pricing (~$0.001/image)
  - `--features properties` → Image Properties pricing only
  - `--features labels,properties` → **Charged for BOTH APIs**
- **OpenAI Vision**: Per-request pricing based on GPT-4o

**Troubleshooting:**

```bash
# Test Gemini credentials
echo $GEMINI_API_KEY

# Test AWS credentials
aws sts get-caller-identity

# Test OpenAI credentials
echo $OPENAI_API_KEY
```

**See Also:**
- Detailed API documentation: [docs/DETECTION.md](./DETECTION.md)
- Example code: [examples/detection/main.go](../examples/detection/main.go)

## Common Use Cases

### Web Optimization

```bash
# Resize for web and optimize JPEG quality
imgx resize photo.jpg -w 1200 --quality 85 -o web.jpg

# Create responsive image set
imgx resize photo.jpg -w 1920 -o photo-large.jpg
imgx resize photo.jpg -w 1200 -o photo-medium.jpg
imgx resize photo.jpg -w 800 -o photo-small.jpg
```

### Social Media

```bash
# Instagram post (1080x1080)
imgx fill photo.jpg -w 1080 -h 1080 -o instagram.jpg

# Facebook cover (1200x630)
imgx fill photo.jpg -w 1200 -h 630 -o facebook-cover.jpg

# Twitter header (1500x500)
imgx fill photo.jpg -w 1500 -h 500 -o twitter-header.jpg
```

### Photo Enhancement

```bash
# Quick enhancement
imgx adjust photo.jpg --brightness 5 --contrast 10 --saturation 15 -o enhanced.jpg

# Fix dark photo
imgx adjust photo.jpg --brightness 20 --gamma 1.3 -o brightened.jpg

# Boost colors
imgx adjust photo.jpg --saturation 25 --contrast 10 -o vibrant.jpg
```

### Thumbnails

```bash
# Square thumbnails
imgx thumbnail photo.jpg -s 150 -o thumb-150.jpg
imgx thumbnail photo.jpg -s 300 -o thumb-300.jpg

# Preserve aspect ratio thumbnails
imgx resize photo.jpg -w 300 -o thumb.jpg
```

### Watermarking

```bash
# Copyright watermark
imgx watermark photo.jpg --text "© 2025 Your Name" --opacity 0.4 -o watermarked.jpg

# Draft stamp
imgx watermark document.jpg --text "DRAFT" --anchor center --opacity 0.3 --color ff0000 -o draft.jpg

# Bottom-left attribution
imgx watermark photo.jpg --text "Photo by You" --anchor bottomleft --padding 15 -o attributed.jpg
```

### Creative Effects

```bash
# Soft focus effect
imgx blur photo.jpg --sigma 3.0 -o soft-focus.jpg

# High contrast black and white
imgx grayscale photo.jpg -o bw.jpg
imgx adjust bw.jpg --contrast 30 -o high-contrast-bw.jpg

# Vintage look
imgx adjust photo.jpg --saturation -20 --contrast -10 --gamma 1.2 -o vintage.jpg

# Inverted colors (negative)
imgx invert photo.jpg -o negative.jpg
```

## Tips & Tricks

### Auto-generated Output Names

If you don't specify an output file with `-o`, imgx automatically generates one by adding a suffix:

```bash
imgx resize photo.jpg -w 800
# Creates: photo-resized.jpg

imgx thumbnail photo.jpg -s 150
# Creates: photo-thumb.jpg

imgx grayscale photo.jpg
# Creates: photo-grayscale.jpg
```

### Format Conversion

Convert between formats using the `--format` flag:

```bash
# PNG to JPEG
imgx resize photo.png -w 800 --format jpg --quality 90 -o output.jpg

# JPEG to PNG
imgx resize photo.jpg -w 800 --format png -o output.png
```

### Preserving Quality

For JPEG output, use high quality settings to preserve detail:

```bash
# High quality (larger file)
imgx resize photo.jpg -w 1920 --quality 95 -o output.jpg

# Balanced quality
imgx resize photo.jpg -w 1920 --quality 85 -o output.jpg

# Web optimized
imgx resize photo.jpg -w 1920 --quality 75 -o output.jpg
```

### Chaining Operations

While command chaining is planned for a future release, you can currently chain operations using shell pipes or by saving intermediate files:

```bash
# Method 1: Multiple commands
imgx resize photo.jpg -w 800 -o temp.jpg
imgx adjust temp.jpg --brightness 10 --contrast 20 -o final.jpg
rm temp.jpg

# Method 2: Multiple adjustments in one command
imgx adjust photo.jpg --brightness 10 --contrast 20 --saturation 15 -o adjusted.jpg
```

### Verbose Mode

Use verbose mode (`-v`) to see what operations are being performed:

```bash
imgx adjust photo.jpg --brightness 10 --contrast 20 -v -o output.jpg
# Output:
# Loaded: photo.jpg (1920x1080)
# Applying brightness: 10.0
# Applying contrast: 20.0
# Saving: output.jpg (1920x1080)
# Saved: output.jpg
```

### Resampling Filters

Choose the right filter for your use case:

- **Lanczos** (default): Best quality for most photographic images
- **CatmullRom**: Fast, sharp results similar to Lanczos
- **MitchellNetravali**: Smoother results with less ringing
- **Linear**: Fast bilinear resampling, good for quick previews
- **Box**: Fast averaging, good for downscaling
- **Nearest**: Fastest, no antialiasing (pixelated for photos)

```bash
# High quality photographic resize
imgx resize photo.jpg -w 800 -f lanczos -o output.jpg

# Fast resize for previews
imgx resize photo.jpg -w 800 -f linear -o preview.jpg

# Pixel art (no smoothing)
imgx resize pixelart.png -w 800 -f nearest -o scaled-pixelart.png
```

## Getting Help

imgx provides comprehensive help at every level.

### General Help

```bash
# Show all commands
imgx --help
imgx -h

# Version information
imgx --version
```

### Command-Specific Help

There are two ways to get help for a specific command:

**Method 1: `help` before the command name**
```bash
imgx help resize
imgx help adjust
imgx help watermark
```

**Method 2: `--help` flag after the command name** (also shows global options)
```bash
imgx resize --help
imgx adjust --help
imgx watermark --help
```

**Note:** Don't use `imgx <command> help` (help after the command) - this syntax doesn't work correctly.

### Shell Completion Scripts

```bash
imgx completion bash
imgx completion zsh
imgx completion fish
imgx completion pwsh
```

## Exit Codes

- `0` - Success
- `1` - Error (invalid arguments, file not found, processing error, etc.)

## Supported Formats

**Input formats:** JPEG, PNG, GIF, TIFF, BMP
**Output formats:** JPEG, PNG, GIF, TIFF, BMP

Format is automatically detected from file extension or can be forced with `--format` flag.
