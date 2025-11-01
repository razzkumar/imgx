package imgx

import (
	"image"
	"image/color"
)

// Version, Author, and ProjectURL are defined in version.go

// Options holds configuration for loading and processing images
type Options struct {
	// AutoOrient enables automatic orientation correction based on EXIF data
	AutoOrient bool

	// DisableMetadata disables metadata tracking for this image instance
	DisableMetadata bool

	// Author sets a custom artist/creator name for the image metadata
	// Empty string uses the default author
	Author string
}

// Load loads an image from a file path and returns an Image instance
// Optionally pass Options to configure loading behavior
//
// Examples:
//   img, err := imgx.Load("photo.jpg")  // use defaults
//   img, err := imgx.Load("photo.jpg", imgx.Options{AutoOrient: true})
//   img, err := imgx.Load("photo.jpg", imgx.Options{Author: "John Doe"})
func Load(path string, opts ...Options) (*Image, error) {
	// Use defaults if no opts provided
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Use internal open() function
	var decodeOpts []DecodeOption
	if opt.AutoOrient {
		decodeOpts = append(decodeOpts, AutoOrientation(true))
	}

	data, err := open(path, decodeOpts...)
	if err != nil {
		return nil, err
	}

	// Determine author - priority: per-image option > global config > default
	author := Author
	if opt.Author != "" {
		author = opt.Author
	} else if globalAuthor := GetDefaultAuthor(); globalAuthor != "" {
		author = globalAuthor
	}

	return &Image{
		data: toNRGBA(data),
		metadata: &ProcessingMetadata{
			SourcePath:  path,
			Software:    "imgx",
			Version:     Version,
			Author:      author,
			ProjectURL:  ProjectURL,
			AddMetadata: !opt.DisableMetadata && globalConfig.AddMetadata,
		},
	}, nil
}

// FromImage creates an Image instance from an existing image.Image
// Optionally pass Options to configure metadata
//
// Examples:
//   img := imgx.FromImage(stdImg)  // use defaults
//   img := imgx.FromImage(stdImg, imgx.Options{Author: "Jane Doe"})
func FromImage(img image.Image, opts ...Options) *Image {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Determine author
	author := Author
	if opt.Author != "" {
		author = opt.Author
	} else if globalAuthor := GetDefaultAuthor(); globalAuthor != "" {
		author = globalAuthor
	}

	return &Image{
		data: toNRGBA(img),
		metadata: &ProcessingMetadata{
			Software:    "imgx",
			Version:     Version,
			Author:      author,
			ProjectURL:  ProjectURL,
			AddMetadata: !opt.DisableMetadata && globalConfig.AddMetadata,
		},
	}
}

// NewImage creates a new blank Image with the specified dimensions and fill color
// Optionally pass Options to configure metadata
//
// Examples:
//   img := imgx.NewImage(800, 600, color.White)  // use defaults
//   img := imgx.NewImage(800, 600, color.White, imgx.Options{Author: "Bob"})
func NewImage(width, height int, fillColor color.Color, opts ...Options) *Image {
	var opt Options
	if len(opts) > 0 {
		opt = opts[0]
	}

	// Determine author
	author := Author
	if opt.Author != "" {
		author = opt.Author
	} else if globalAuthor := GetDefaultAuthor(); globalAuthor != "" {
		author = globalAuthor
	}

	return &Image{
		data: New(width, height, fillColor),
		metadata: &ProcessingMetadata{
			Software:    "imgx",
			Version:     Version,
			Author:      author,
			ProjectURL:  ProjectURL,
			AddMetadata: !opt.DisableMetadata && globalConfig.AddMetadata,
		},
	}
}

