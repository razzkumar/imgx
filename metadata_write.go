package imgx

import (
	"fmt"
	"image/png"
	"os/exec"
	"time"
)

// SaveOption configures image saving
type SaveOption func(*SaveConfig)

// SaveConfig holds configuration for saving images
type SaveConfig struct {
	DisableMetadata bool
	JPEGQuality     int
	PNGCompression  png.CompressionLevel
	GIFNumColors    int
	// Add other encode options as needed
}

// WithoutMetadata disables metadata writing for this save operation
func WithoutMetadata() SaveOption {
	return func(c *SaveConfig) {
		c.DisableMetadata = true
	}
}

// WithJPEGQuality sets the JPEG quality (1-100)
func WithJPEGQuality(quality int) SaveOption {
	return func(c *SaveConfig) {
		c.JPEGQuality = quality
	}
}

// WithPNGCompression sets the PNG compression level
func WithPNGCompression(level png.CompressionLevel) SaveOption {
	return func(c *SaveConfig) {
		c.PNGCompression = level
	}
}

// WithGIFNumColors sets the number of colors for GIF encoding
func WithGIFNumColors(numColors int) SaveOption {
	return func(c *SaveConfig) {
		c.GIFNumColors = numColors
	}
}

// Save saves the image to the specified path with optional metadata injection
func (img *Image) Save(path string, opts ...SaveOption) error {
	config := &SaveConfig{
		DisableMetadata: false,
		JPEGQuality:     90,
		PNGCompression:  png.DefaultCompression,
		GIFNumColors:    256,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Convert SaveConfig to EncodeOptions
	var encodeOpts []EncodeOption
	if config.JPEGQuality > 0 {
		encodeOpts = append(encodeOpts, JPEGQuality(config.JPEGQuality))
	}
	if config.PNGCompression != png.DefaultCompression {
		encodeOpts = append(encodeOpts, PNGCompressionLevel(config.PNGCompression))
	}
	if config.GIFNumColors != 256 {
		encodeOpts = append(encodeOpts, GIFNumColors(config.GIFNumColors))
	}

	// Save image using internal save() function
	if err := save(img.data, path, encodeOpts...); err != nil {
		return err
	}

	// Write metadata if enabled
	shouldWriteMetadata := img.metadata.AddMetadata && !config.DisableMetadata
	if shouldWriteMetadata {
		if err := img.writeXMPMetadata(path); err != nil {
			// Don't fail the save if metadata writing fails
			// Just silently skip (exiftool might not be available)
			return nil
		}
	}

	return nil
}

// writeXMPMetadata writes XMP metadata to the image file using exiftool
func (img *Image) writeXMPMetadata(path string) error {
	if !isExiftoolAvailable() {
		return nil // Silently skip if exiftool not available
	}

	args := []string{
		"-overwrite_original",
		fmt.Sprintf("-XMP:CreatorTool=%s v%s", img.metadata.Software, img.metadata.Version),
		fmt.Sprintf("-XMP:Creator=%s", img.metadata.Author),
		fmt.Sprintf("-DC:Source=%s", img.metadata.ProjectURL),
		fmt.Sprintf("-XMP:ModifyDate=%s", time.Now().Format(time.RFC3339)),
	}

	// Add history entries
	for _, op := range img.metadata.Operations {
		historyEntry := fmt.Sprintf(
			"action=converted, when=%s, softwareAgent=%s v%s, parameters=%s: %s",
			op.Timestamp.Format(time.RFC3339),
			img.metadata.Software,
			img.metadata.Version,
			op.Action,
			op.Parameters,
		)
		args = append(args, fmt.Sprintf("-XMP-xmpMM:History+=%s", historyEntry))
	}

	args = append(args, path)
	cmd := exec.Command("exiftool", args...)
	return cmd.Run()
}
