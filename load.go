package imgx

import (
	"image"
	"image/color"
)

const (
	// Version is the current version of imgx
	Version = "1.0.0"
	// Author is the author of imgx
	Author = "razzkumar"
	// ProjectURL is the GitHub URL of the project
	ProjectURL = "https://github.com/razzkumar/imgx"
)

// LoadOption configures image loading
type LoadOption func(*LoadConfig)

// LoadConfig holds configuration for loading images
type LoadConfig struct {
	AutoOrientation bool
	DisableMetadata bool
}

// DisableMetadata disables metadata tracking for this image instance
func DisableMetadata() LoadOption {
	return func(c *LoadConfig) {
		c.DisableMetadata = true
	}
}

// AutoOrient enables automatic orientation correction based on EXIF data
func AutoOrient() LoadOption {
	return func(c *LoadConfig) {
		c.AutoOrientation = true
	}
}

// Load loads an image from a file path and returns an Image instance
func Load(path string, opts ...LoadOption) (*Image, error) {
	config := &LoadConfig{
		AutoOrientation: false,
		DisableMetadata: false,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Use internal open() function
	var decodeOpts []DecodeOption
	if config.AutoOrientation {
		decodeOpts = append(decodeOpts, AutoOrientation(true))
	}

	data, err := open(path, decodeOpts...)
	if err != nil {
		return nil, err
	}

	return &Image{
		data: toNRGBA(data),
		metadata: &ProcessingMetadata{
			SourcePath:  path,
			Software:    "imgx",
			Version:     Version,
			Author:      Author,
			ProjectURL:  ProjectURL,
			AddMetadata: !config.DisableMetadata && globalConfig.AddMetadata,
		},
	}, nil
}

// FromImage creates an Image instance from an existing image.Image
func FromImage(img image.Image) *Image {
	return &Image{
		data: toNRGBA(img),
		metadata: &ProcessingMetadata{
			Software:    "imgx",
			Version:     Version,
			Author:      Author,
			ProjectURL:  ProjectURL,
			AddMetadata: globalConfig.AddMetadata,
		},
	}
}

// NewImage creates a new blank Image with the specified dimensions and fill color
func NewImage(width, height int, fillColor color.Color) *Image {
	return &Image{
		data: New(width, height, fillColor),
		metadata: &ProcessingMetadata{
			Software:    "imgx",
			Version:     Version,
			Author:      Author,
			ProjectURL:  ProjectURL,
			AddMetadata: globalConfig.AddMetadata,
		},
	}
}

