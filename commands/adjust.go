package commands

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v3"
)

// AdjustCommand creates the adjust command for color adjustments
func AdjustCommand() *cli.Command {
	return &cli.Command{
		Name:  "adjust",
		Usage: "Adjust image colors (brightness, contrast, gamma, saturation, hue)",
		Description: `Adjust various color properties of an image. Multiple adjustments can be applied at once.

Examples:
  imgx adjust photo.jpg --brightness 10 --contrast 20 -o output.jpg
  imgx adjust photo.jpg --saturation -30 --hue 60 -o output.jpg
  imgx adjust photo.jpg --gamma 1.5 -o output.jpg`,
		Flags: []cli.Flag{
			&cli.FloatFlag{
				Name:  "brightness",
				Usage: "adjust brightness (-100 to 100, 0 = no change)",
				Value: 0,
			},
			&cli.FloatFlag{
				Name:  "contrast",
				Usage: "adjust contrast (-100 to 100, 0 = no change)",
				Value: 0,
			},
			&cli.FloatFlag{
				Name:  "gamma",
				Usage: "gamma correction (positive number, 1.0 = no change, <1 darkens, >1 lightens)",
				Value: 1.0,
			},
			&cli.FloatFlag{
				Name:  "saturation",
				Usage: "adjust saturation (-100 to 100, 0 = no change, -100 = grayscale)",
				Value: 0,
			},
			&cli.FloatFlag{
				Name:  "hue",
				Usage: "adjust hue in degrees (-180 to 180, 0 = no change)",
				Value: 0,
			},
		},
		Action: adjustAction,
	}
}

func adjustAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	brightness := cmd.Float("brightness")
	contrast := cmd.Float("contrast")
	gamma := cmd.Float("gamma")
	saturation := cmd.Float("saturation")
	hue := cmd.Float("hue")

	// Check if any adjustment is specified
	if brightness == 0 && contrast == 0 && gamma == 1.0 && saturation == 0 && hue == 0 {
		return fmt.Errorf("at least one adjustment parameter must be specified")
	}

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Apply adjustments in order
	result := img
	if brightness != 0 {
		if cmd.Bool("verbose") {
			fmt.Printf("Applying brightness: %.1f\n", brightness)
		}
		result = result.AdjustBrightness(brightness)
	}

	if contrast != 0 {
		if cmd.Bool("verbose") {
			fmt.Printf("Applying contrast: %.1f\n", contrast)
		}
		result = result.AdjustContrast(contrast)
	}

	if gamma != 1.0 {
		if cmd.Bool("verbose") {
			fmt.Printf("Applying gamma: %.2f\n", gamma)
		}
		result = result.AdjustGamma(gamma)
	}

	if saturation != 0 {
		if cmd.Bool("verbose") {
			fmt.Printf("Applying saturation: %.1f\n", saturation)
		}
		result = result.AdjustSaturation(saturation)
	}

	if hue != 0 {
		if cmd.Bool("verbose") {
			fmt.Printf("Applying hue shift: %.1f degrees\n", hue)
		}
		result = result.AdjustHue(hue)
	}

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-adjusted")
	return saveImage(cmd, result, outputPath)
}

// GrayscaleCommand creates the grayscale command
func GrayscaleCommand() *cli.Command {
	return &cli.Command{
		Name:  "grayscale",
		Usage: "Convert image to grayscale",
		Description: `Convert an image to grayscale using luminance weights (ITU-R BT.601).

Example:
  imgx grayscale photo.jpg -o output.jpg`,
		Action: grayscaleAction,
	}
}

func grayscaleAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Convert to grayscale
	result := img.Grayscale()

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-grayscale")
	return saveImage(cmd, result, outputPath)
}

// InvertCommand creates the invert command
func InvertCommand() *cli.Command {
	return &cli.Command{
		Name:  "invert",
		Usage: "Invert image colors (negative)",
		Description: `Invert (negate) all colors in the image.

Example:
  imgx invert photo.jpg -o output.jpg`,
		Action: invertAction,
	}
}

func invertAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Invert
	result := img.Invert()

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-inverted")
	return saveImage(cmd, result, outputPath)
}
