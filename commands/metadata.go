package commands

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/razzkumar/imgx"
	"github.com/urfave/cli/v3"
)

// MetadataCommand creates the metadata command
func MetadataCommand() *cli.Command {
	return &cli.Command{
		Name:  "metadata",
		Usage: "Extract and display image metadata",
		Description: `Extract comprehensive image metadata using exiftool when available.

If exiftool is installed, displays detailed EXIF, IPTC, and XMP metadata including:
- Camera information (make, model, lens)
- Camera settings (ISO, aperture, shutter speed, focal length)
- GPS coordinates and altitude
- Date/time information
- Copyright and authorship
- Software and processing information

If exiftool is not available, displays basic metadata:
- File information and format
- Dimensions and aspect ratio
- File size and color model

Examples:
  # Display all available metadata
  imgx metadata photo.jpg

  # Show basic metadata only
  imgx metadata --basic photo.jpg

  # Output as JSON
  imgx metadata --json photo.jpg

Installation:
  macOS:    brew install exiftool
  Ubuntu:   sudo apt-get install libimage-exiftool-perl
  Windows:  https://exiftool.org`,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "basic",
				Aliases: []string{"b"},
				Usage:   "Show basic metadata only (skip exiftool)",
			},
			&cli.BoolFlag{
				Name:    "json",
				Aliases: []string{"j"},
				Usage:   "Output metadata as JSON",
			},
		},
		Action: metadataAction,
	}
}

func metadataAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)

	// Prepare options
	var opts []imgx.MetadataOption
	if cmd.Bool("basic") {
		opts = append(opts, imgx.WithBasicOnly())
	}

	// Extract metadata
	metadata, err := imgx.Metadata(inputPath, opts...)
	if err != nil {
		return fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Output format
	if cmd.Bool("json") {
		return outputJSON(metadata)
	}

	return outputPretty(metadata)
}

func outputJSON(metadata *imgx.ImageMetadata) error {
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func outputPretty(metadata *imgx.ImageMetadata) error {
	fmt.Println("=== Image Metadata ===")
	fmt.Println()

	// Basic information
	fmt.Println("File Information:")
	fmt.Printf("  Path:        %s\n", metadata.FilePath)
	fmt.Printf("  Format:      %s\n", metadata.Format)
	fmt.Printf("  Size:        %s\n", FormatBytes(metadata.FileSize))
	fmt.Println()

	fmt.Println("Image Properties:")
	fmt.Printf("  Dimensions:  %dx%d\n", metadata.Width, metadata.Height)
	fmt.Printf("  Aspect Ratio: %.2f\n", metadata.AspectRatio)
	fmt.Printf("  Megapixels:  %.2f MP\n", metadata.Megapixels)
	fmt.Printf("  Color Model: %s\n", metadata.ColorModel)

	// Extended metadata if available
	if metadata.HasExtended {
		if metadata.CameraMake != "" || metadata.CameraModel != "" || metadata.LensModel != "" {
			fmt.Println()
			fmt.Println("Camera Information:")
			if metadata.CameraMake != "" {
				fmt.Printf("  Make:        %s\n", metadata.CameraMake)
			}
			if metadata.CameraModel != "" {
				fmt.Printf("  Model:       %s\n", metadata.CameraModel)
			}
			if metadata.LensModel != "" {
				fmt.Printf("  Lens:        %s\n", metadata.LensModel)
			}
		}

		if metadata.FocalLength != "" || metadata.FNumber != "" ||
			metadata.ExposureTime != "" || metadata.ISO != "" {
			fmt.Println()
			fmt.Println("Camera Settings:")
			if metadata.FocalLength != "" {
				fmt.Printf("  Focal Length: %s\n", metadata.FocalLength)
			}
			if metadata.FNumber != "" {
				fmt.Printf("  Aperture:    %s\n", metadata.FNumber)
			}
			if metadata.ExposureTime != "" {
				fmt.Printf("  Shutter:     %s\n", metadata.ExposureTime)
			}
			if metadata.ISO != "" {
				fmt.Printf("  ISO:         %s\n", metadata.ISO)
			}
			if metadata.Flash != "" {
				fmt.Printf("  Flash:       %s\n", metadata.Flash)
			}
		}

		if metadata.DateTime != "" || metadata.DateTimeOriginal != "" {
			fmt.Println()
			fmt.Println("Date/Time:")
			if metadata.DateTimeOriginal != "" {
				fmt.Printf("  Original:    %s\n", metadata.DateTimeOriginal)
			}
			if metadata.DateTime != "" {
				fmt.Printf("  Modified:    %s\n", metadata.DateTime)
			}
		}

		if metadata.GPSLatitude != "" || metadata.GPSLongitude != "" {
			fmt.Println()
			fmt.Println("GPS Location:")
			if metadata.GPSLatitude != "" {
				fmt.Printf("  Latitude:    %s\n", metadata.GPSLatitude)
			}
			if metadata.GPSLongitude != "" {
				fmt.Printf("  Longitude:   %s\n", metadata.GPSLongitude)
			}
			if metadata.GPSAltitude != "" {
				fmt.Printf("  Altitude:    %s\n", metadata.GPSAltitude)
			}
		}

		if metadata.Copyright != "" || metadata.Artist != "" || metadata.Software != "" {
			fmt.Println()
			fmt.Println("Additional Information:")
			if metadata.Software != "" {
				fmt.Printf("  Software:    %s\n", metadata.Software)
			}
			if metadata.Artist != "" {
				fmt.Printf("  Artist:      %s\n", metadata.Artist)
			}
			if metadata.Copyright != "" {
				fmt.Printf("  Copyright:   %s\n", metadata.Copyright)
			}
		}
	} else {
		fmt.Println()
		fmt.Println("---")
		fmt.Println()
		fmt.Println("exiftool not found. Install exiftool for comprehensive metadata.")
		fmt.Println()
		fmt.Println("Installation:")
		fmt.Println("  macOS:    brew install exiftool")
		fmt.Println("  Ubuntu:   sudo apt-get install libimage-exiftool-perl")
		fmt.Println("  Windows:  https://exiftool.org")
	}

	return nil
}
