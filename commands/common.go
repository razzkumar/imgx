package commands

import (
	"fmt"
	"image/color"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/razzkumar/imgx"
)

// ParseColor parses a hex color string (RGB or RGBA) to color.NRGBA
// Supported formats: "ffffff", "ff0000ff", "#ffffff", "#ff0000ff"
func ParseColor(s string) (color.NRGBA, error) {
	s = strings.TrimPrefix(s, "#")
	s = strings.ToLower(s)

	var r, g, b, a uint8 = 0, 0, 0, 255

	switch len(s) {
	case 6: // RGB
		val, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid color format: %s", s)
		}
		r = uint8(val >> 16)
		g = uint8(val >> 8)
		b = uint8(val)
	case 8: // RGBA
		val, err := strconv.ParseUint(s, 16, 32)
		if err != nil {
			return color.NRGBA{}, fmt.Errorf("invalid color format: %s", s)
		}
		r = uint8(val >> 24)
		g = uint8(val >> 16)
		b = uint8(val >> 8)
		a = uint8(val)
	default:
		return color.NRGBA{}, fmt.Errorf("invalid color format: %s (expected 6 or 8 hex digits)", s)
	}

	return color.NRGBA{R: r, G: g, B: b, A: a}, nil
}

// ParseFilter converts a filter name string to imgx.ResampleFilter
func ParseFilter(name string) (imgx.ResampleFilter, error) {
	name = strings.ToLower(name)
	switch name {
	case "nearest", "nearestneighbor":
		return imgx.NearestNeighbor, nil
	case "box":
		return imgx.Box, nil
	case "linear":
		return imgx.Linear, nil
	case "hermite":
		return imgx.Hermite, nil
	case "mitchellnetravali", "mitchell":
		return imgx.MitchellNetravali, nil
	case "catmullrom", "catrom":
		return imgx.CatmullRom, nil
	case "bspline":
		return imgx.BSpline, nil
	case "gaussian":
		return imgx.Gaussian, nil
	case "lanczos":
		return imgx.Lanczos, nil
	case "hann":
		return imgx.Hann, nil
	case "hamming":
		return imgx.Hamming, nil
	case "blackman":
		return imgx.Blackman, nil
	case "bartlett":
		return imgx.Bartlett, nil
	case "welch":
		return imgx.Welch, nil
	case "cosine":
		return imgx.Cosine, nil
	default:
		return imgx.Lanczos, fmt.Errorf("unknown filter: %s", name)
	}
}

// ParseAnchor converts an anchor name string to imgx.Anchor
func ParseAnchor(name string) (imgx.Anchor, error) {
	name = strings.ToLower(name)
	switch name {
	case "center", "centre":
		return imgx.Center, nil
	case "topleft", "top-left":
		return imgx.TopLeft, nil
	case "top":
		return imgx.Top, nil
	case "topright", "top-right":
		return imgx.TopRight, nil
	case "left":
		return imgx.Left, nil
	case "right":
		return imgx.Right, nil
	case "bottomleft", "bottom-left":
		return imgx.BottomLeft, nil
	case "bottom":
		return imgx.Bottom, nil
	case "bottomright", "bottom-right":
		return imgx.BottomRight, nil
	default:
		return imgx.Center, fmt.Errorf("unknown anchor: %s", name)
	}
}

// GenerateOutputPath generates an output path if not provided
// Adds a suffix before the extension: input.jpg -> input-processed.jpg
func GenerateOutputPath(inputPath, suffix string) string {
	ext := filepath.Ext(inputPath)
	base := strings.TrimSuffix(inputPath, ext)
	return fmt.Sprintf("%s%s%s", base, suffix, ext)
}

// ParseFormat converts a format name to imgx.Format
func ParseFormat(name string) (imgx.Format, error) {
	name = strings.ToLower(name)
	switch name {
	case "jpg", "jpeg":
		return imgx.JPEG, nil
	case "png":
		return imgx.PNG, nil
	case "gif":
		return imgx.GIF, nil
	case "tif", "tiff":
		return imgx.TIFF, nil
	case "bmp":
		return imgx.BMP, nil
	default:
		return imgx.JPEG, fmt.Errorf("unknown format: %s", name)
	}
}

// FormatName returns the string name of a format
func FormatName(format imgx.Format) string {
	switch format {
	case imgx.JPEG:
		return "JPEG"
	case imgx.PNG:
		return "PNG"
	case imgx.GIF:
		return "GIF"
	case imgx.TIFF:
		return "TIFF"
	case imgx.BMP:
		return "BMP"
	default:
		return "Unknown"
	}
}

// FormatBytes formats a byte count as a human-readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
