package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// ResizeCommand creates the resize command
func ResizeCommand() *cli.Command {
	return &cli.Command{
		Name:  "resize",
		Usage: "Resize image to specific dimensions",
		Description: `Resize an image to the specified width and height.
If one dimension is 0, the aspect ratio is preserved.

Examples:
  imgx resize input.jpg -w 800 -h 600 -o output.jpg
  imgx resize input.jpg -w 800                        # preserve aspect ratio
  imgx resize input.jpg -h 600 -f catmullrom          # with different filter`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "width",
				Aliases: []string{"w"},
				Usage:   "target width (0 to preserve aspect ratio)",
				Value:   0,
			},
			&cli.IntFlag{
				Name:    "height",
				Aliases: []string{"h"},
				Usage:   "target height (0 to preserve aspect ratio)",
				Value:   0,
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "resampling filter (nearest, box, linear, hermite, mitchellnetravali, catmullrom, bspline, gaussian, lanczos, hann, hamming, blackman, bartlett, welch, cosine)",
				Value:   "lanczos",
			},
		},
		Action: resizeAction,
	}
}

func resizeAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	width := cmd.Int("width")
	height := cmd.Int("height")
	filterName := cmd.String("filter")

	if width == 0 && height == 0 {
		return fmt.Errorf("at least one dimension (width or height) must be specified")
	}

	// Parse filter
	filter, err := ParseFilter(filterName)
	if err != nil {
		return err
	}

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Resize
	result := img.Resize(width, height, filter)

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-resized")
	return saveImage(cmd, result, outputPath)
}

// FitCommand creates the fit command
func FitCommand() *cli.Command {
	return &cli.Command{
		Name:  "fit",
		Usage: "Scale image to fit within bounds",
		Description: `Scale the image down to fit within the specified maximum dimensions
while preserving the aspect ratio.

Example:
  imgx fit input.jpg -w 800 -h 600 -o output.jpg`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "width",
				Aliases:  []string{"w"},
				Usage:    "maximum width",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "height",
				Aliases:  []string{"h"},
				Usage:    "maximum height",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "resampling filter",
				Value:   "lanczos",
			},
		},
		Action: fitAction,
	}
}

func fitAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	width := cmd.Int("width")
	height := cmd.Int("height")
	filterName := cmd.String("filter")

	filter, err := ParseFilter(filterName)
	if err != nil {
		return err
	}

	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	result := img.Fit(width, height, filter)

	outputPath := getOutputPath(cmd, inputPath, "-fit")
	return saveImage(cmd, result, outputPath)
}

// FillCommand creates the fill command
func FillCommand() *cli.Command {
	return &cli.Command{
		Name:  "fill",
		Usage: "Crop and resize to fill exact dimensions",
		Description: `Resize and crop the image to fill the specified dimensions.
The image is scaled to cover the target size, then cropped to fit.

Example:
  imgx fill input.jpg -w 800 -h 600 --anchor center -o output.jpg`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "width",
				Aliases:  []string{"w"},
				Usage:    "target width",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "height",
				Aliases:  []string{"h"},
				Usage:    "target height",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "anchor",
				Aliases: []string{"a"},
				Usage:   "anchor position (center, topleft, top, topright, left, right, bottomleft, bottom, bottomright)",
				Value:   "center",
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "resampling filter",
				Value:   "lanczos",
			},
		},
		Action: fillAction,
	}
}

func fillAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	width := cmd.Int("width")
	height := cmd.Int("height")
	anchorName := cmd.String("anchor")
	filterName := cmd.String("filter")

	anchor, err := ParseAnchor(anchorName)
	if err != nil {
		return err
	}

	filter, err := ParseFilter(filterName)
	if err != nil {
		return err
	}

	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	result := img.Fill(width, height, anchor, filter)

	outputPath := getOutputPath(cmd, inputPath, "-fill")
	return saveImage(cmd, result, outputPath)
}

// ThumbnailCommand creates the thumbnail command
func ThumbnailCommand() *cli.Command {
	return &cli.Command{
		Name:  "thumbnail",
		Usage: "Create a square thumbnail",
		Description: `Create a square thumbnail by cropping and resizing.
This is a convenience command equivalent to 'fill' with a square size.

Example:
  imgx thumbnail input.jpg -s 150 -o thumb.jpg`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "size",
				Aliases:  []string{"s"},
				Usage:    "thumbnail size (width and height)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "filter",
				Aliases: []string{"f"},
				Usage:   "resampling filter",
				Value:   "lanczos",
			},
		},
		Action: thumbnailAction,
	}
}

func thumbnailAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	size := cmd.Int("size")
	filterName := cmd.String("filter")

	filter, err := ParseFilter(filterName)
	if err != nil {
		return err
	}

	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	result := img.Thumbnail(size, size, filter)

	outputPath := getOutputPath(cmd, inputPath, "-thumb")
	return saveImage(cmd, result, outputPath)
}
