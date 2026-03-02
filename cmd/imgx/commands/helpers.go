package commands

import (
	"fmt"

	"github.com/razzkumar/imgx"
	"github.com/urfave/cli/v3"
)

// loadImage loads an image from the specified path, respecting global flags
func loadImage(cmd *cli.Command, path string) (*imgx.Image, error) {
	autoOrient := cmd.Bool("auto-orient")

	var img *imgx.Image
	var err error

	if autoOrient {
		img, err = imgx.Load(path, imgx.Options{AutoOrient: true})
	} else {
		img, err = imgx.Load(path)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to open image: %w", err)
	}

	if cmd.Bool("verbose") {
		bounds := img.Bounds()
		fmt.Printf("Loaded: %s (%dx%d)\n", path, bounds.Dx(), bounds.Dy())
	}

	return img, nil
}

// saveImage saves an image to the specified path, respecting global flags
func saveImage(cmd *cli.Command, img *imgx.Image, path string) error {
	quality := cmd.Int("quality")
	formatName := cmd.String("format")

	var opts []imgx.SaveOption

	// Add quality option for JPEG
	if quality > 0 {
		opts = append(opts, imgx.WithJPEGQuality(quality))
	}

	// If format is specified, ensure output path has correct extension
	if formatName != "" {
		format, err := ParseFormat(formatName)
		if err != nil {
			return err
		}

		// Change extension if needed
		path = changeExtension(path, format)
	}

	if cmd.Bool("verbose") {
		bounds := img.Bounds()
		fmt.Printf("Saving: %s (%dx%d)\n", path, bounds.Dx(), bounds.Dy())
	}

	err := img.Save(path, opts...)
	if err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	if cmd.Bool("verbose") {
		fmt.Printf("Saved: %s\n", path)
	}

	return nil
}

// getOutputPath determines the output path from flags or generates one
func getOutputPath(cmd *cli.Command, inputPath, suffix string) string {
	output := cmd.String("output")
	if output != "" {
		return output
	}
	return GenerateOutputPath(inputPath, suffix)
}

// changeExtension changes the file extension based on format
func changeExtension(path string, format imgx.Format) string {
	var ext string
	switch format {
	case imgx.JPEG:
		ext = ".jpg"
	case imgx.PNG:
		ext = ".png"
	case imgx.GIF:
		ext = ".gif"
	case imgx.TIFF:
		ext = ".tiff"
	case imgx.BMP:
		ext = ".bmp"
	default:
		return path
	}

	// Replace extension
	base := path
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			base = path[:i]
			break
		}
		if path[i] == '/' {
			break
		}
	}

	return base + ext
}
