package commands

import (
	"context"
	"fmt"

	"github.com/razzkumar/imgx"
	"github.com/urfave/cli/v3"
)

// RotateCommand creates the rotate command
func RotateCommand() *cli.Command {
	return &cli.Command{
		Name:  "rotate",
		Usage: "Rotate image by specified angle",
		Description: `Rotate an image by the specified angle in degrees (counter-clockwise).
For 90-degree increments (90, 180, 270), the rotation is lossless.
For other angles, bilinear interpolation is used.

Examples:
  imgx rotate photo.jpg -a 90 -o output.jpg           # 90 degrees
  imgx rotate photo.jpg -a 45 --bg ffffff -o output.jpg  # 45 degrees with white background
  imgx rotate photo.jpg -a -30 --bg 00000000           # -30 degrees (clockwise) with transparent background`,
		Flags: []cli.Flag{
			&cli.FloatFlag{
				Name:     "angle",
				Aliases:  []string{"a"},
				Usage:    "rotation angle in degrees (positive = counter-clockwise, negative = clockwise)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "bg",
				Usage: "background color for empty areas in hex (RGB or RGBA, e.g., ffffff or 00000000)",
				Value: "00000000", // Transparent by default
			},
		},
		Action: rotateAction,
	}
}

func rotateAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	angle := cmd.Float("angle")
	bgColorStr := cmd.String("bg")

	// Parse background color
	bgColor, err := ParseColor(bgColorStr)
	if err != nil {
		return err
	}

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Rotate
	result := img.Rotate(angle, bgColor)

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-rotated")
	return saveImage(cmd, result, outputPath)
}

// FlipCommand creates the flip command
func FlipCommand() *cli.Command {
	return &cli.Command{
		Name:  "flip",
		Usage: "Flip image horizontally and/or vertically",
		Description: `Flip an image horizontally (left-right), vertically (top-bottom), or both.

Examples:
  imgx flip photo.jpg --horizontal -o output.jpg
  imgx flip photo.jpg --vertical -o output.jpg
  imgx flip photo.jpg --horizontal --vertical -o output.jpg  # Same as rotate 180`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "horizontal",
				Aliases: []string{"H"},
				Usage:   "flip horizontally (left-right)",
			},
			&cli.BoolFlag{
				Name:    "vertical",
				Aliases: []string{"V"},
				Usage:   "flip vertically (top-bottom)",
			},
		},
		Action: flipAction,
	}
}

func flipAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	horizontal := cmd.Bool("horizontal")
	vertical := cmd.Bool("vertical")

	if !horizontal && !vertical {
		return fmt.Errorf("at least one of --horizontal or --vertical must be specified")
	}

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Apply flips
	var result *imgx.Image
	if horizontal && vertical {
		// Both flips = 180 degree rotation
		result = img.Rotate180()
	} else if horizontal {
		result = img.FlipH()
	} else {
		result = img.FlipV()
	}

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-flipped")
	return saveImage(cmd, result, outputPath)
}

// CropCommand creates the crop command
func CropCommand() *cli.Command {
	return &cli.Command{
		Name:  "crop",
		Usage: "Crop image to specified region",
		Description: `Crop an image to a specific region. You can either specify an anchor position
or exact coordinates.

Examples:
  imgx crop photo.jpg -w 500 -h 400 --anchor center -o output.jpg
  imgx crop photo.jpg -w 500 -h 400 --anchor topleft -o output.jpg
  imgx crop photo.jpg -x 100 -y 100 -w 500 -h 400 -o output.jpg`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:     "width",
				Aliases:  []string{"w"},
				Usage:    "crop width",
				Required: true,
			},
			&cli.IntFlag{
				Name:     "height",
				Aliases:  []string{"h"},
				Usage:    "crop height",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "x",
				Usage: "X coordinate (left edge, exclusive with --anchor)",
				Value: -1,
			},
			&cli.IntFlag{
				Name:  "y",
				Usage: "Y coordinate (top edge, exclusive with --anchor)",
				Value: -1,
			},
			&cli.StringFlag{
				Name:    "anchor",
				Aliases: []string{"a"},
				Usage:   "anchor position (center, topleft, top, topright, left, right, bottomleft, bottom, bottomright)",
				Value:   "center",
			},
		},
		Action: cropAction,
	}
}

func cropAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	width := cmd.Int("width")
	height := cmd.Int("height")
	x := cmd.Int("x")
	y := cmd.Int("y")
	anchorName := cmd.String("anchor")

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	var result *imgx.Image

	// Check if coordinates are specified
	if x >= 0 && y >= 0 {
		// Use exact coordinates
		bounds := img.Bounds()
		rect := bounds.Intersect(bounds)
		rect.Min.X = x
		rect.Min.Y = y
		rect.Max.X = x + width
		rect.Max.Y = y + height
		result = img.Crop(rect)
	} else {
		// Use anchor
		anchor, err := ParseAnchor(anchorName)
		if err != nil {
			return err
		}
		result = img.CropAnchor(width, height, anchor)
	}

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-cropped")
	return saveImage(cmd, result, outputPath)
}

// TransposeCommand creates the transpose command
func TransposeCommand() *cli.Command {
	return &cli.Command{
		Name:  "transpose",
		Usage: "Transpose image (flip horizontally and rotate 90° counter-clockwise)",
		Description: `Transpose flips the image horizontally and then rotates it 90 degrees counter-clockwise.

Example:
  imgx transpose photo.jpg -o output.jpg`,
		Action: transposeAction,
	}
}

func transposeAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Transpose
	result := img.Transpose()

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-transposed")
	return saveImage(cmd, result, outputPath)
}

// TransverseCommand creates the transverse command
func TransverseCommand() *cli.Command {
	return &cli.Command{
		Name:  "transverse",
		Usage: "Transverse image (flip vertically and rotate 90° counter-clockwise)",
		Description: `Transverse flips the image vertically and then rotates it 90 degrees counter-clockwise.

Example:
  imgx transverse photo.jpg -o output.jpg`,
		Action: transverseAction,
	}
}

func transverseAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Transverse
	result := img.Transverse()

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-transversed")
	return saveImage(cmd, result, outputPath)
}

// Rotate90Command creates shortcuts for common rotations
func Rotate90Command() *cli.Command {
	return &cli.Command{
		Name:  "rotate90",
		Usage: "Rotate image 90 degrees counter-clockwise",
		Description: `Quickly rotate an image 90 degrees counter-clockwise (lossless).

Example:
  imgx rotate90 photo.jpg -o output.jpg`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() < 1 {
				return fmt.Errorf("input file required")
			}
			inputPath := cmd.Args().Get(0)
			img, err := loadImage(cmd, inputPath)
			if err != nil {
				return err
			}
			result := img.Rotate90()
			outputPath := getOutputPath(cmd, inputPath, "-rot90")
			return saveImage(cmd, result, outputPath)
		},
	}
}

func Rotate180Command() *cli.Command {
	return &cli.Command{
		Name:  "rotate180",
		Usage: "Rotate image 180 degrees",
		Description: `Quickly rotate an image 180 degrees (lossless).

Example:
  imgx rotate180 photo.jpg -o output.jpg`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() < 1 {
				return fmt.Errorf("input file required")
			}
			inputPath := cmd.Args().Get(0)
			img, err := loadImage(cmd, inputPath)
			if err != nil {
				return err
			}
			result := img.Rotate180()
			outputPath := getOutputPath(cmd, inputPath, "-rot180")
			return saveImage(cmd, result, outputPath)
		},
	}
}

func Rotate270Command() *cli.Command {
	return &cli.Command{
		Name:  "rotate270",
		Usage: "Rotate image 270 degrees counter-clockwise (90 clockwise)",
		Description: `Quickly rotate an image 270 degrees counter-clockwise / 90 degrees clockwise (lossless).

Example:
  imgx rotate270 photo.jpg -o output.jpg`,
		Action: func(ctx context.Context, cmd *cli.Command) error {
			if cmd.Args().Len() < 1 {
				return fmt.Errorf("input file required")
			}
			inputPath := cmd.Args().Get(0)
			img, err := loadImage(cmd, inputPath)
			if err != nil {
				return err
			}
			result := img.Rotate270()
			outputPath := getOutputPath(cmd, inputPath, "-rot270")
			return saveImage(cmd, result, outputPath)
		},
	}
}
