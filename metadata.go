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

// ImageMetadata contains comprehensive image metadata information
type ImageMetadata struct {
	// File Information
	FilePath string `json:"file_path"`
	FileName string `json:"file_name"`
	Format   string `json:"format"`
	FileSize int64  `json:"file_size"`

	// Image Properties
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	AspectRatio string  `json:"aspect_ratio"` // Formatted as "16:9" (Width:Height)
	Megapixels  float64 `json:"megapixels"`
	ColorModel  string  `json:"color_model"`
	Orientation int     `json:"orientation,omitempty"`

	// Image Technical Details
	BitDepth           int     `json:"bit_depth,omitempty"`
	ColorSpace         string  `json:"color_space,omitempty"`
	Compression        string  `json:"compression,omitempty"`
	XResolution        float64 `json:"x_resolution,omitempty"`
	YResolution        float64 `json:"y_resolution,omitempty"`
	ResolutionUnit     string  `json:"resolution_unit,omitempty"`
	ImageDescription   string  `json:"image_description,omitempty"`
	UserComment        string  `json:"user_comment,omitempty"`

	// Extended Metadata
	Extended    map[string]any `json:"extended,omitempty"`
	HasExtended bool           `json:"has_extended"`

	// Camera Information
	CameraMake          string `json:"camera_make,omitempty"`
	CameraModel         string `json:"camera_model,omitempty"`
	CameraSerialNumber  string `json:"camera_serial_number,omitempty"`
	LensModel           string `json:"lens_model,omitempty"`
	LensSerialNumber    string `json:"lens_serial_number,omitempty"`
	LensFocalLengthMin  string `json:"lens_focal_length_min,omitempty"`
	LensFocalLengthMax  string `json:"lens_focal_length_max,omitempty"`
	FirmwareVersion     string `json:"firmware_version,omitempty"`

	// Camera Settings
	FocalLength        string `json:"focal_length,omitempty"`
	Aperture           string `json:"aperture,omitempty"` // F-number
	ShutterSpeed       string `json:"shutter_speed,omitempty"`
	ISO                string `json:"iso,omitempty"`
	ExposureCompensation string `json:"exposure_compensation,omitempty"`
	ExposureMode       string `json:"exposure_mode,omitempty"`
	ExposureProgram    string `json:"exposure_program,omitempty"`
	MeteringMode       string `json:"metering_mode,omitempty"`
	WhiteBalance       string `json:"white_balance,omitempty"`
	Flash              string `json:"flash,omitempty"`
	FlashMode          string `json:"flash_mode,omitempty"`
	LightSource        string `json:"light_source,omitempty"`
	SceneCaptureType   string `json:"scene_capture_type,omitempty"`
	SubjectDistance    string `json:"subject_distance,omitempty"`
	FocusMode          string `json:"focus_mode,omitempty"`
	DigitalZoomRatio   string `json:"digital_zoom_ratio,omitempty"`

	// Timestamps
	DateTime         string `json:"date_time,omitempty"`
	DateTimeOriginal string `json:"date_time_original,omitempty"`
	DateTimeDigitized string `json:"date_time_digitized,omitempty"`
	CreateDate       string `json:"create_date,omitempty"`
	ModifyDate       string `json:"modify_date,omitempty"`
	TimeZone         string `json:"time_zone,omitempty"`

	// GPS Location
	GPSLatitude   string `json:"gps_latitude,omitempty"`
	GPSLongitude  string `json:"gps_longitude,omitempty"`
	GPSAltitude   string `json:"gps_altitude,omitempty"`
	GPSTimestamp  string `json:"gps_timestamp,omitempty"`
	GPSSpeed      string `json:"gps_speed,omitempty"`
	GPSDirection  string `json:"gps_direction,omitempty"`
	GPSSatellites string `json:"gps_satellites,omitempty"`
	GPSDatum      string `json:"gps_datum,omitempty"`

	// Content & Authorship
	Title       string `json:"title,omitempty"`
	Subject     string `json:"subject,omitempty"`
	Keywords    string `json:"keywords,omitempty"`
	Rating      int    `json:"rating,omitempty"`
	Artist      string `json:"artist,omitempty"`
	Copyright   string `json:"copyright,omitempty"`
	Creator     string `json:"creator,omitempty"`
	CreatorTool string `json:"creator_tool,omitempty"`
	Software    string `json:"software,omitempty"`
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

// gcd calculates the greatest common divisor using Euclidean algorithm
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// formatAspectRatio formats width and height as a ratio string (e.g., "16:9", "4:3")
func formatAspectRatio(width, height int) string {
	if width == 0 || height == 0 {
		return "N/A"
	}
	divisor := gcd(width, height)
	w := width / divisor
	h := height / divisor
	return fmt.Sprintf("%d:%d", w, h)
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
		AspectRatio: formatAspectRatio(width, height),
		FileSize:    fileInfo.Size(),
		ColorModel:  fmt.Sprintf("%T", img.ColorModel()),
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
	// Helper to get string value
	getString := func(key string) string {
		if val, ok := data[key].(string); ok {
			return val
		}
		return ""
	}

	// Helper to get int value
	getInt := func(key string) int {
		if val, ok := data[key].(float64); ok {
			return int(val)
		}
		return 0
	}

	// Helper to get float value
	getFloat := func(key string) float64 {
		if val, ok := data[key].(float64); ok {
			return val
		}
		return 0
	}

	// Image Technical Details
	metadata.BitDepth = getInt("EXIF:BitsPerSample")
	metadata.ColorSpace = getString("EXIF:ColorSpace")
	if metadata.ColorSpace == "" {
		metadata.ColorSpace = getString("ICC_Profile:ColorSpaceData")
	}
	metadata.Compression = getString("EXIF:Compression")
	metadata.ImageDescription = getString("EXIF:ImageDescription")
	metadata.UserComment = getString("EXIF:UserComment")

	if orientation := getInt("EXIF:Orientation"); orientation > 0 {
		metadata.Orientation = orientation
	}
	if xRes := getFloat("EXIF:XResolution"); xRes > 0 {
		metadata.XResolution = xRes
	}
	if yRes := getFloat("EXIF:YResolution"); yRes > 0 {
		metadata.YResolution = yRes
	}
	metadata.ResolutionUnit = getString("EXIF:ResolutionUnit")

	// Camera Information
	metadata.CameraMake = getString("EXIF:Make")
	metadata.CameraModel = getString("EXIF:Model")
	metadata.CameraSerialNumber = getString("EXIF:SerialNumber")
	if metadata.CameraSerialNumber == "" {
		metadata.CameraSerialNumber = getString("EXIF:InternalSerialNumber")
	}
	metadata.LensModel = getString("EXIF:LensModel")
	if metadata.LensModel == "" {
		metadata.LensModel = getString("EXIF:LensID")
	}
	metadata.LensSerialNumber = getString("EXIF:LensSerialNumber")
	metadata.LensFocalLengthMin = getString("EXIF:MinFocalLength")
	metadata.LensFocalLengthMax = getString("EXIF:MaxFocalLength")
	metadata.FirmwareVersion = getString("EXIF:FirmwareVersion")

	// Camera Settings
	metadata.FocalLength = getString("EXIF:FocalLength")

	// Aperture (handle both string and float)
	metadata.Aperture = getString("EXIF:FNumber")
	if metadata.Aperture == "" {
		if fnum := getFloat("EXIF:FNumber"); fnum > 0 {
			metadata.Aperture = fmt.Sprintf("f/%.1f", fnum)
		}
	}
	if metadata.Aperture == "" {
		metadata.Aperture = getString("EXIF:ApertureValue")
		if metadata.Aperture == "" {
			if aval := getFloat("EXIF:ApertureValue"); aval > 0 {
				metadata.Aperture = fmt.Sprintf("f/%.1f", aval)
			}
		}
	}

	metadata.ShutterSpeed = getString("EXIF:ExposureTime")
	if metadata.ShutterSpeed == "" {
		metadata.ShutterSpeed = getString("EXIF:ShutterSpeedValue")
	}

	// ISO (handle both string and number)
	if iso := getString("EXIF:ISO"); iso != "" {
		metadata.ISO = iso
	} else if iso := getInt("EXIF:ISO"); iso > 0 {
		metadata.ISO = fmt.Sprintf("%d", iso)
	}

	metadata.ExposureCompensation = getString("EXIF:ExposureCompensation")
	metadata.ExposureMode = getString("EXIF:ExposureMode")
	metadata.ExposureProgram = getString("EXIF:ExposureProgram")
	metadata.MeteringMode = getString("EXIF:MeteringMode")
	metadata.WhiteBalance = getString("EXIF:WhiteBalance")
	metadata.Flash = getString("EXIF:Flash")
	metadata.FlashMode = getString("EXIF:FlashMode")
	metadata.LightSource = getString("EXIF:LightSource")
	metadata.SceneCaptureType = getString("EXIF:SceneCaptureType")
	metadata.SubjectDistance = getString("EXIF:SubjectDistance")
	metadata.FocusMode = getString("EXIF:FocusMode")
	if metadata.FocusMode == "" {
		metadata.FocusMode = getString("MakerNotes:FocusMode")
	}
	metadata.DigitalZoomRatio = getString("EXIF:DigitalZoomRatio")

	// Timestamps
	metadata.DateTime = getString("EXIF:DateTime")
	metadata.DateTimeOriginal = getString("EXIF:DateTimeOriginal")
	metadata.DateTimeDigitized = getString("EXIF:DateTimeDigitized")
	metadata.CreateDate = getString("EXIF:CreateDate")
	metadata.ModifyDate = getString("EXIF:ModifyDate")
	if metadata.ModifyDate == "" {
		metadata.ModifyDate = getString("File:FileModifyDate")
	}
	metadata.TimeZone = getString("EXIF:TimeZone")
	if metadata.TimeZone == "" {
		metadata.TimeZone = getString("EXIF:OffsetTime")
	}

	// GPS Location
	metadata.GPSLatitude = getString("EXIF:GPSLatitude")
	if metadata.GPSLatitude == "" {
		metadata.GPSLatitude = getString("Composite:GPSLatitude")
	}
	metadata.GPSLongitude = getString("EXIF:GPSLongitude")
	if metadata.GPSLongitude == "" {
		metadata.GPSLongitude = getString("Composite:GPSLongitude")
	}
	metadata.GPSAltitude = getString("EXIF:GPSAltitude")
	if metadata.GPSAltitude == "" {
		metadata.GPSAltitude = getString("Composite:GPSAltitude")
	}
	metadata.GPSTimestamp = getString("EXIF:GPSTimeStamp")
	metadata.GPSSpeed = getString("EXIF:GPSSpeed")
	metadata.GPSDirection = getString("EXIF:GPSImgDirection")
	if metadata.GPSDirection == "" {
		metadata.GPSDirection = getString("EXIF:GPSDestBearing")
	}
	metadata.GPSSatellites = getString("EXIF:GPSSatellites")
	metadata.GPSDatum = getString("EXIF:GPSMapDatum")

	// Content & Authorship
	metadata.Title = getString("IPTC:ObjectName")
	if metadata.Title == "" {
		metadata.Title = getString("XMP:Title")
	}
	metadata.Subject = getString("IPTC:Caption-Abstract")
	if metadata.Subject == "" {
		metadata.Subject = getString("XMP:Description")
	}
	metadata.Keywords = getString("IPTC:Keywords")
	if metadata.Keywords == "" {
		metadata.Keywords = getString("XMP:Subject")
	}
	if rating := getInt("XMP:Rating"); rating > 0 {
		metadata.Rating = rating
	}
	metadata.Artist = getString("EXIF:Artist")
	if metadata.Artist == "" {
		metadata.Artist = getString("IPTC:By-line")
	}
	if metadata.Artist == "" {
		metadata.Artist = getString("XMP:Creator")
	}
	metadata.Copyright = getString("EXIF:Copyright")
	if metadata.Copyright == "" {
		metadata.Copyright = getString("IPTC:CopyrightNotice")
	}
	if metadata.Copyright == "" {
		metadata.Copyright = getString("XMP:Rights")
	}
	metadata.Creator = getString("XMP:Creator")
	metadata.CreatorTool = getString("XMP:CreatorTool")
	metadata.Software = getString("EXIF:Software")
	if metadata.Software == "" {
		metadata.Software = getString("File:Software")
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
