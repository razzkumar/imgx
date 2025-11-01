package imgx

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// ImageMetadata contains image metadata information
type ImageMetadata struct {
	// Basic metadata (always available)
	FilePath    string
	FileName    string
	Format      string
	Width       int
	Height      int
	FileSize    int64
	ColorModel  string
	AspectRatio float64
	Megapixels  float64

	// Extended metadata (only when exiftool is available)
	Extended    map[string]any // Raw exiftool data
	HasExtended bool           // Whether extended metadata is available

	// Common EXIF fields (parsed from Extended for convenience)
	CameraMake       string
	CameraModel      string
	LensModel        string
	DateTime         string
	DateTimeOriginal string
	Orientation      int
	XResolution      float64
	YResolution      float64
	ResolutionUnit   string
	Software         string
	Copyright        string
	Artist           string

	// Camera settings
	FocalLength  string
	FNumber      string
	ExposureTime string
	ISO          string
	Flash        string

	// GPS data
	GPSLatitude  string
	GPSLongitude string
	GPSAltitude  string
	GPSTimestamp string
}

// MetadataOption configures metadata extraction
type MetadataOption func(*metadataConfig)

type metadataConfig struct {
	basicOnly bool // Force basic metadata only, skip exiftool
}

// WithBasicOnly forces basic metadata extraction only (skip exiftool check)
func WithBasicOnly() MetadataOption {
	return func(c *metadataConfig) {
		c.basicOnly = true
	}
}

var (
	exiftoolCache      *bool
	exiftoolCacheMutex sync.RWMutex
)

// Metadata extracts metadata from an image file.
// If exiftool is available on the system, returns comprehensive EXIF/IPTC/XMP data.
// Otherwise, returns basic metadata extracted using Go's image package.
//
// Example:
//
//	metadata, err := imgx.Metadata("photo.jpg")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Dimensions: %dx%d\n", metadata.Width, metadata.Height)
//	if metadata.HasExtended {
//	    fmt.Printf("Camera: %s\n", metadata.CameraMake)
//	}
func Metadata(src string, options ...MetadataOption) (*ImageMetadata, error) {
	// Parse options
	config := &metadataConfig{}
	for _, opt := range options {
		opt(config)
	}

	// Extract basic metadata first
	metadata, err := extractBasicMetadata(src)
	if err != nil {
		return nil, err
	}

	// If basic only, return now
	if config.basicOnly {
		return metadata, nil
	}

	// Try to extract extended metadata with exiftool
	if isExiftoolAvailable() {
		extendedData, err := extractWithExiftool(src)
		if err == nil {
			metadata.Extended = extendedData
			metadata.HasExtended = true
			parseCommonFields(metadata, extendedData)
		}
		// If exiftool fails, we still have basic metadata
	}

	return metadata, nil
}

// isExiftoolAvailable checks if exiftool binary is available in PATH
func isExiftoolAvailable() bool {
	// Check cache first
	exiftoolCacheMutex.RLock()
	if exiftoolCache != nil {
		cached := *exiftoolCache
		exiftoolCacheMutex.RUnlock()
		return cached
	}
	exiftoolCacheMutex.RUnlock()

	// Check if exiftool is available
	exiftoolCacheMutex.Lock()
	defer exiftoolCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if exiftoolCache != nil {
		return *exiftoolCache
	}

	exiftoolPath, err := exec.LookPath("exiftool")
	available := err == nil && exiftoolPath != ""
	exiftoolCache = &available

	return available
}

// extractBasicMetadata extracts basic metadata using Go's image package
func extractBasicMetadata(src string) (*ImageMetadata, error) {
	// Get file information
	fileInfo, err := os.Stat(src)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Detect format
	format, err := FormatFromFilename(src)
	if err != nil {
		return nil, fmt.Errorf("failed to detect format: %w", err)
	}

	// Open and decode image
	img, err := Open(src)
	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	metadata := &ImageMetadata{
		FilePath:    src,
		FileName:    filepath.Base(src),
		Format:      formatToString(format),
		Width:       width,
		Height:      height,
		FileSize:    fileInfo.Size(),
		ColorModel:  fmt.Sprintf("%T", img.ColorModel()),
		AspectRatio: float64(width) / float64(height),
		Megapixels:  float64(width*height) / 1000000.0,
		HasExtended: false,
	}

	return metadata, nil
}

// extractWithExiftool executes exiftool and parses JSON output
func extractWithExiftool(src string) (map[string]any, error) {
	// Run: exiftool -json -G -a -s <file>
	// -json: JSON output format
	// -G: Organize output by group (EXIF, IPTC, etc.)
	// -a: Allow duplicate tags
	// -s: Short tag names
	cmd := exec.Command("exiftool", "-json", "-G", "-a", "-s", src)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("exiftool execution failed: %w", err)
	}

	// exiftool returns array with one object per file
	var result []map[string]any
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse exiftool output: %w", err)
	}

	if len(result) == 0 {
		return nil, errors.New("no metadata found")
	}

	return result[0], nil
}

// parseCommonFields extracts common EXIF fields from raw exiftool data
func parseCommonFields(metadata *ImageMetadata, data map[string]any) {
	// Camera information
	if make, ok := data["EXIF:Make"].(string); ok {
		metadata.CameraMake = make
	}
	if model, ok := data["EXIF:Model"].(string); ok {
		metadata.CameraModel = model
	}
	if lens, ok := data["EXIF:LensModel"].(string); ok {
		metadata.LensModel = lens
	}

	// Date/Time
	if dateTime, ok := data["EXIF:DateTime"].(string); ok {
		metadata.DateTime = dateTime
	}
	if dateTimeOriginal, ok := data["EXIF:DateTimeOriginal"].(string); ok {
		metadata.DateTimeOriginal = dateTimeOriginal
	}

	// Image properties
	if orientation, ok := data["EXIF:Orientation"].(float64); ok {
		metadata.Orientation = int(orientation)
	}
	if xRes, ok := data["EXIF:XResolution"].(float64); ok {
		metadata.XResolution = xRes
	}
	if yRes, ok := data["EXIF:YResolution"].(float64); ok {
		metadata.YResolution = yRes
	}
	if resUnit, ok := data["EXIF:ResolutionUnit"].(string); ok {
		metadata.ResolutionUnit = resUnit
	}

	// Software and authorship
	if software, ok := data["EXIF:Software"].(string); ok {
		metadata.Software = software
	}
	if copyright, ok := data["EXIF:Copyright"].(string); ok {
		metadata.Copyright = copyright
	}
	if artist, ok := data["EXIF:Artist"].(string); ok {
		metadata.Artist = artist
	}

	// Camera settings
	if focalLength, ok := data["EXIF:FocalLength"].(string); ok {
		metadata.FocalLength = focalLength
	}
	if fNumber, ok := data["EXIF:FNumber"].(string); ok {
		metadata.FNumber = fNumber
	}
	if exposureTime, ok := data["EXIF:ExposureTime"].(string); ok {
		metadata.ExposureTime = exposureTime
	}
	if iso, ok := data["EXIF:ISO"].(string); ok {
		metadata.ISO = iso
	} else if iso, ok := data["EXIF:ISO"].(float64); ok {
		metadata.ISO = fmt.Sprintf("%.0f", iso)
	}
	if flash, ok := data["EXIF:Flash"].(string); ok {
		metadata.Flash = flash
	}

	// GPS data
	if gpsLat, ok := data["EXIF:GPSLatitude"].(string); ok {
		metadata.GPSLatitude = gpsLat
	}
	if gpsLon, ok := data["EXIF:GPSLongitude"].(string); ok {
		metadata.GPSLongitude = gpsLon
	}
	if gpsAlt, ok := data["EXIF:GPSAltitude"].(string); ok {
		metadata.GPSAltitude = gpsAlt
	}
	if gpsTime, ok := data["EXIF:GPSTimeStamp"].(string); ok {
		metadata.GPSTimestamp = gpsTime
	}
}

// formatToString converts Format enum to string
func formatToString(format Format) string {
	switch format {
	case JPEG:
		return "JPEG"
	case PNG:
		return "PNG"
	case GIF:
		return "GIF"
	case TIFF:
		return "TIFF"
	case BMP:
		return "BMP"
	default:
		return "Unknown"
	}
}
