package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// BlurCommand creates the blur command
func BlurCommand() *cli.Command {
	return &cli.Command{
		Name:  "blur",
		Usage: "Apply Gaussian blur to image",
		Description: `Apply a Gaussian blur effect to the image.
The sigma parameter controls the blur strength (higher = more blur).

Examples:
  imgx blur photo.jpg --sigma 2.5 -o output.jpg
  imgx blur photo.jpg -s 5.0 -o output.jpg`,
		Flags: []cli.Flag{
			&cli.FloatFlag{
				Name:     "sigma",
				Aliases:  []string{"s"},
				Usage:    "blur strength (positive number, typical range: 0.5-10)",
				Required: true,
				Validator: func(f float64) error {
					if f <= 0 {
						return fmt.Errorf("sigma must be positive")
					}
					return nil
				},
			},
		},
		Action: blurAction,
	}
}

func blurAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	sigma := cmd.Float("sigma")

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	if cmd.Bool("verbose") {
		fmt.Printf("Applying Gaussian blur with sigma: %.2f\n", sigma)
	}

	// Apply blur
	result := img.Blur(sigma)

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-blurred")
	return saveImage(cmd, result, outputPath)
}

// SharpenCommand creates the sharpen command
func SharpenCommand() *cli.Command {
	return &cli.Command{
		Name:  "sharpen",
		Usage: "Sharpen image",
		Description: `Sharpen the image using unsharp masking.
The sigma parameter controls the sharpening strength (higher = more sharpening).

Examples:
  imgx sharpen photo.jpg --sigma 1.5 -o output.jpg
  imgx sharpen photo.jpg -s 2.0 -o output.jpg`,
		Flags: []cli.Flag{
			&cli.FloatFlag{
				Name:     "sigma",
				Aliases:  []string{"s"},
				Usage:    "sharpening strength (positive number, typical range: 0.5-5)",
				Required: true,
				Validator: func(f float64) error {
					if f <= 0 {
						return fmt.Errorf("sigma must be positive")
					}
					return nil
				},
			},
		},
		Action: sharpenAction,
	}
}

func sharpenAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	sigma := cmd.Float("sigma")

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	if cmd.Bool("verbose") {
		fmt.Printf("Applying sharpening with sigma: %.2f\n", sigma)
	}

	// Apply sharpen
	result := img.Sharpen(sigma)

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-sharpened")
	return saveImage(cmd, result, outputPath)
}
