package commands

import (
	"context"
	"fmt"
	"os"

	"github.com/razzkumar/imgx"
	"github.com/urfave/cli/v3"
)

// InfoCommand creates the info command
func InfoCommand() *cli.Command {
	return &cli.Command{
		Name:  "info",
		Usage: "Display image information",
		Description: `Display detailed information about an image file including:
- File path and format
- Dimensions (width x height)
- File size
- Color model

Example:
  imgx info photo.jpg`,
		Action: infoAction,
	}
}

func infoAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)

	// Get file info
	fileInfo, err := os.Stat(inputPath)
	if err != nil {
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Detect format from filename
	format, err := imgx.FormatFromFilename(inputPath)
	if err != nil {
		return fmt.Errorf("failed to detect format: %w", err)
	}

	// Open image
	img, err := imgx.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Print information
	fmt.Printf("File: %s\n", inputPath)
	fmt.Printf("Format: %s\n", FormatName(format))
	fmt.Printf("Dimensions: %dx%d\n", width, height)
	fmt.Printf("Size: %s\n", FormatBytes(fileInfo.Size()))
	fmt.Printf("Color Model: %T\n", img.ColorModel())

	return nil
}
