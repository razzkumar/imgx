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
		Name:    "metadata",
		Aliases: []string{"info"},
		Usage:   "Display image information and metadata",
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

	// File Information
	fmt.Println("File Information:")
	fmt.Printf("  Path:           %s\n", metadata.FilePath)
	fmt.Printf("  Format:         %s\n", metadata.Format)
	fmt.Printf("  Size:           %s\n", FormatBytes(metadata.FileSize))
	fmt.Println()

	// Image Properties
	fmt.Println("Image Properties:")
	fmt.Printf("  Dimensions:     %dx%d\n", metadata.Width, metadata.Height)
	fmt.Printf("  Aspect Ratio:   %s\n", metadata.AspectRatio)
	fmt.Printf("  Megapixels:     %.2f MP\n", metadata.Megapixels)
	fmt.Printf("  Color Model:    %s\n", metadata.ColorModel)

	// Extended metadata if available
	if metadata.HasExtended {
		// Image Technical Details
		showTechnical := metadata.BitDepth > 0 || metadata.ColorSpace != "" ||
			metadata.Compression != "" || metadata.XResolution > 0 || metadata.Orientation > 0
		if showTechnical {
			fmt.Println()
			fmt.Println("Technical Details:")
			if metadata.BitDepth > 0 {
				fmt.Printf("  Bit Depth:      %d\n", metadata.BitDepth)
			}
			if metadata.ColorSpace != "" {
				fmt.Printf("  Color Space:    %s\n", metadata.ColorSpace)
			}
			if metadata.Compression != "" {
				fmt.Printf("  Compression:    %s\n", metadata.Compression)
			}
			if metadata.XResolution > 0 {
				fmt.Printf("  Resolution:     %.0fx%.0f %s\n", metadata.XResolution, metadata.YResolution, metadata.ResolutionUnit)
			}
			if metadata.Orientation > 0 {
				fmt.Printf("  Orientation:    %d\n", metadata.Orientation)
			}
		}

		// Camera Information
		showCamera := metadata.CameraMake != "" || metadata.CameraModel != "" ||
			metadata.LensModel != "" || metadata.CameraSerialNumber != ""
		if showCamera {
			fmt.Println()
			fmt.Println("Camera Information:")
			if metadata.CameraMake != "" {
				fmt.Printf("  Make:           %s\n", metadata.CameraMake)
			}
			if metadata.CameraModel != "" {
				fmt.Printf("  Model:          %s\n", metadata.CameraModel)
			}
			if metadata.CameraSerialNumber != "" {
				fmt.Printf("  Serial Number:  %s\n", metadata.CameraSerialNumber)
			}
			if metadata.LensModel != "" {
				fmt.Printf("  Lens:           %s\n", metadata.LensModel)
			}
			if metadata.LensSerialNumber != "" {
				fmt.Printf("  Lens S/N:       %s\n", metadata.LensSerialNumber)
			}
			if metadata.LensFocalLengthMin != "" && metadata.LensFocalLengthMax != "" {
				fmt.Printf("  Lens Range:     %s-%s\n", metadata.LensFocalLengthMin, metadata.LensFocalLengthMax)
			}
			if metadata.FirmwareVersion != "" {
				fmt.Printf("  Firmware:       %s\n", metadata.FirmwareVersion)
			}
		}

		// Camera Settings
		showSettings := metadata.FocalLength != "" || metadata.Aperture != "" ||
			metadata.ShutterSpeed != "" || metadata.ISO != ""
		if showSettings {
			fmt.Println()
			fmt.Println("Camera Settings:")
			if metadata.FocalLength != "" {
				fmt.Printf("  Focal Length:   %s\n", metadata.FocalLength)
			}
			if metadata.Aperture != "" {
				fmt.Printf("  Aperture:       %s\n", metadata.Aperture)
			}
			if metadata.ShutterSpeed != "" {
				fmt.Printf("  Shutter Speed:  %s\n", metadata.ShutterSpeed)
			}
			if metadata.ISO != "" {
				fmt.Printf("  ISO:            %s\n", metadata.ISO)
			}
			if metadata.ExposureCompensation != "" {
				fmt.Printf("  Exp. Comp.:     %s\n", metadata.ExposureCompensation)
			}
			if metadata.ExposureMode != "" {
				fmt.Printf("  Exposure Mode:  %s\n", metadata.ExposureMode)
			}
			if metadata.ExposureProgram != "" {
				fmt.Printf("  Exp. Program:   %s\n", metadata.ExposureProgram)
			}
			if metadata.MeteringMode != "" {
				fmt.Printf("  Metering:       %s\n", metadata.MeteringMode)
			}
			if metadata.WhiteBalance != "" {
				fmt.Printf("  White Balance:  %s\n", metadata.WhiteBalance)
			}
			if metadata.Flash != "" {
				fmt.Printf("  Flash:          %s\n", metadata.Flash)
			}
			if metadata.FlashMode != "" {
				fmt.Printf("  Flash Mode:     %s\n", metadata.FlashMode)
			}
			if metadata.FocusMode != "" {
				fmt.Printf("  Focus Mode:     %s\n", metadata.FocusMode)
			}
			if metadata.SubjectDistance != "" {
				fmt.Printf("  Subject Dist.:  %s\n", metadata.SubjectDistance)
			}
		}

		// Timestamps
		showTime := metadata.DateTimeOriginal != "" || metadata.DateTime != "" ||
			metadata.CreateDate != ""
		if showTime {
			fmt.Println()
			fmt.Println("Date/Time:")
			if metadata.DateTimeOriginal != "" {
				fmt.Printf("  Taken:          %s\n", metadata.DateTimeOriginal)
			}
			if metadata.CreateDate != "" && metadata.CreateDate != metadata.DateTimeOriginal {
				fmt.Printf("  Created:        %s\n", metadata.CreateDate)
			}
			if metadata.DateTime != "" {
				fmt.Printf("  Modified:       %s\n", metadata.DateTime)
			}
			if metadata.DateTimeDigitized != "" && metadata.DateTimeDigitized != metadata.DateTimeOriginal {
				fmt.Printf("  Digitized:      %s\n", metadata.DateTimeDigitized)
			}
			if metadata.TimeZone != "" {
				fmt.Printf("  Time Zone:      %s\n", metadata.TimeZone)
			}
		}

		// GPS Location
		showGPS := metadata.GPSLatitude != "" || metadata.GPSLongitude != ""
		if showGPS {
			fmt.Println()
			fmt.Println("GPS Location:")
			if metadata.GPSLatitude != "" {
				fmt.Printf("  Latitude:       %s\n", metadata.GPSLatitude)
			}
			if metadata.GPSLongitude != "" {
				fmt.Printf("  Longitude:      %s\n", metadata.GPSLongitude)
			}
			if metadata.GPSAltitude != "" {
				fmt.Printf("  Altitude:       %s\n", metadata.GPSAltitude)
			}
			if metadata.GPSSpeed != "" {
				fmt.Printf("  Speed:          %s\n", metadata.GPSSpeed)
			}
			if metadata.GPSDirection != "" {
				fmt.Printf("  Direction:      %s\n", metadata.GPSDirection)
			}
			if metadata.GPSTimestamp != "" {
				fmt.Printf("  GPS Time:       %s\n", metadata.GPSTimestamp)
			}
			if metadata.GPSSatellites != "" {
				fmt.Printf("  Satellites:     %s\n", metadata.GPSSatellites)
			}
		}

		// Content & Authorship
		showContent := metadata.Title != "" || metadata.Subject != "" ||
			metadata.Keywords != "" || metadata.Artist != "" || metadata.Copyright != ""
		if showContent {
			fmt.Println()
			fmt.Println("Content & Authorship:")
			if metadata.Title != "" {
				fmt.Printf("  Title:          %s\n", metadata.Title)
			}
			if metadata.Subject != "" {
				fmt.Printf("  Subject:        %s\n", metadata.Subject)
			}
			if metadata.Keywords != "" {
				fmt.Printf("  Keywords:       %s\n", metadata.Keywords)
			}
			if metadata.Rating > 0 {
				fmt.Printf("  Rating:         %d stars\n", metadata.Rating)
			}
			if metadata.Artist != "" {
				fmt.Printf("  Artist:         %s\n", metadata.Artist)
			}
			if metadata.Creator != "" && metadata.Creator != metadata.Artist {
				fmt.Printf("  Creator:        %s\n", metadata.Creator)
			}
			if metadata.Copyright != "" {
				fmt.Printf("  Copyright:      %s\n", metadata.Copyright)
			}
			if metadata.Software != "" {
				fmt.Printf("  Software:       %s\n", metadata.Software)
			}
			if metadata.CreatorTool != "" && metadata.CreatorTool != metadata.Software {
				fmt.Printf("  Creator Tool:   %s\n", metadata.CreatorTool)
			}
		}

		// Image Description
		if metadata.ImageDescription != "" || metadata.UserComment != "" {
			fmt.Println()
			fmt.Println("Description:")
			if metadata.ImageDescription != "" {
				fmt.Printf("  Image Desc.:    %s\n", metadata.ImageDescription)
			}
			if metadata.UserComment != "" {
				fmt.Printf("  User Comment:   %s\n", metadata.UserComment)
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
