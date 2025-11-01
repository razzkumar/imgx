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

With --extended flag (requires exiftool):
- Camera information and settings
- GPS location data
- Date/time metadata
- Copyright and authorship

Example:
  imgx info photo.jpg
  imgx info --extended photo.jpg`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "extended",
				Aliases: []string{"e"},
				Usage:   "Show extended metadata (requires exiftool)",
			},
		},
		Action: infoAction,
	}
}

func infoAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)

	// If extended flag is set, use metadata extraction
	if cmd.Bool("extended") {
		metadata, err := imgx.Metadata(inputPath)
		if err != nil {
			return fmt.Errorf("failed to extract metadata: %w", err)
		}

		// Print basic info
		fmt.Printf("File: %s\n", metadata.FilePath)
		fmt.Printf("Format: %s\n", metadata.Format)
		fmt.Printf("Dimensions: %dx%d\n", metadata.Width, metadata.Height)
		fmt.Printf("Size: %s\n", FormatBytes(metadata.FileSize))
		fmt.Printf("Color Model: %s\n", metadata.ColorModel)

		// Print extended metadata if available
		if metadata.HasExtended {
			if metadata.CameraMake != "" || metadata.CameraModel != "" {
				fmt.Println()
				fmt.Println("Camera:")
				if metadata.CameraMake != "" {
					fmt.Printf("  Make: %s\n", metadata.CameraMake)
				}
				if metadata.CameraModel != "" {
					fmt.Printf("  Model: %s\n", metadata.CameraModel)
				}
			}

			if metadata.DateTimeOriginal != "" {
				fmt.Println()
				fmt.Printf("Date Taken: %s\n", metadata.DateTimeOriginal)
			}

			if metadata.FocalLength != "" || metadata.ISO != "" {
				fmt.Println()
				fmt.Println("Settings:")
				if metadata.FocalLength != "" {
					fmt.Printf("  Focal Length: %s\n", metadata.FocalLength)
				}
				if metadata.FNumber != "" {
					fmt.Printf("  Aperture: %s\n", metadata.FNumber)
				}
				if metadata.ISO != "" {
					fmt.Printf("  ISO: %s\n", metadata.ISO)
				}
			}
		} else {
			fmt.Println()
			fmt.Println("Note: exiftool not found. Install for extended metadata.")
		}

		return nil
	}

	// Original basic info behavior
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
