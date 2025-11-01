package main

import (
	"fmt"
	"log"

	"github.com/razzkumar/imgx"
)

func main() {
	imagePath := "testdata/flower.jpg"

	fmt.Println("=== Metadata Extraction Example ===")
	fmt.Println()

	// Example 1: Extract all available metadata
	fmt.Println("1. Extracting all available metadata...")
	fmt.Println("   (uses exiftool if available, falls back to basic metadata)")
	fmt.Println()

	metadata, err := imgx.Metadata(imagePath)
	if err != nil {
		log.Fatalf("Failed to extract metadata: %v", err)
	}

	// Display basic information (always available)
	fmt.Printf("File: %s\n", metadata.FilePath)
	fmt.Printf("Format: %s\n", metadata.Format)
	fmt.Printf("Dimensions: %dx%d pixels\n", metadata.Width, metadata.Height)
	fmt.Printf("Aspect Ratio: %.2f\n", metadata.AspectRatio)
	fmt.Printf("Megapixels: %.2f MP\n", metadata.Megapixels)
	fmt.Printf("File Size: %d bytes (%.2f KB)\n", metadata.FileSize, float64(metadata.FileSize)/1024)
	fmt.Printf("Color Model: %s\n", metadata.ColorModel)
	fmt.Println()

	// Display extended metadata if available (requires exiftool)
	if metadata.HasExtended {
		fmt.Println("✓ Extended metadata available (exiftool is installed)")
		fmt.Println()

		// Camera information
		if metadata.CameraMake != "" || metadata.CameraModel != "" {
			fmt.Println("Camera Information:")
			if metadata.CameraMake != "" {
				fmt.Printf("  Make: %s\n", metadata.CameraMake)
			}
			if metadata.CameraModel != "" {
				fmt.Printf("  Model: %s\n", metadata.CameraModel)
			}
			if metadata.LensModel != "" {
				fmt.Printf("  Lens: %s\n", metadata.LensModel)
			}
			fmt.Println()
		}

		// Camera settings
		if metadata.FocalLength != "" || metadata.FNumber != "" ||
			metadata.ExposureTime != "" || metadata.ISO != "" {
			fmt.Println("Camera Settings:")
			if metadata.FocalLength != "" {
				fmt.Printf("  Focal Length: %s\n", metadata.FocalLength)
			}
			if metadata.FNumber != "" {
				fmt.Printf("  Aperture: %s\n", metadata.FNumber)
			}
			if metadata.ExposureTime != "" {
				fmt.Printf("  Shutter Speed: %s\n", metadata.ExposureTime)
			}
			if metadata.ISO != "" {
				fmt.Printf("  ISO: %s\n", metadata.ISO)
			}
			if metadata.Flash != "" {
				fmt.Printf("  Flash: %s\n", metadata.Flash)
			}
			fmt.Println()
		}

		// Date/time information
		if metadata.DateTimeOriginal != "" || metadata.DateTime != "" {
			fmt.Println("Date/Time Information:")
			if metadata.DateTimeOriginal != "" {
				fmt.Printf("  Date Taken: %s\n", metadata.DateTimeOriginal)
			}
			if metadata.DateTime != "" {
				fmt.Printf("  Last Modified: %s\n", metadata.DateTime)
			}
			fmt.Println()
		}

		// GPS information
		if metadata.GPSLatitude != "" || metadata.GPSLongitude != "" {
			fmt.Println("GPS Location:")
			if metadata.GPSLatitude != "" {
				fmt.Printf("  Latitude: %s\n", metadata.GPSLatitude)
			}
			if metadata.GPSLongitude != "" {
				fmt.Printf("  Longitude: %s\n", metadata.GPSLongitude)
			}
			if metadata.GPSAltitude != "" {
				fmt.Printf("  Altitude: %s\n", metadata.GPSAltitude)
			}
			fmt.Println()
		}

		// Additional information
		if metadata.Software != "" || metadata.Artist != "" || metadata.Copyright != "" {
			fmt.Println("Additional Information:")
			if metadata.Software != "" {
				fmt.Printf("  Software: %s\n", metadata.Software)
			}
			if metadata.Artist != "" {
				fmt.Printf("  Artist: %s\n", metadata.Artist)
			}
			if metadata.Copyright != "" {
				fmt.Printf("  Copyright: %s\n", metadata.Copyright)
			}
			fmt.Println()
		}

		// Access raw extended data
		if len(metadata.Extended) > 0 {
			fmt.Printf("Total metadata fields extracted: %d\n", len(metadata.Extended))
			fmt.Println()
		}
	} else {
		fmt.Println("⚠ Extended metadata not available")
		fmt.Println("  Install exiftool for comprehensive metadata:")
		fmt.Println()
		fmt.Println("  macOS:    brew install exiftool")
		fmt.Println("  Ubuntu:   sudo apt-get install libimage-exiftool-perl")
		fmt.Println("  Windows:  https://exiftool.org")
		fmt.Println()
	}

	// Example 2: Extract basic metadata only (skip exiftool check)
	fmt.Println("---")
	fmt.Println()
	fmt.Println("2. Extracting basic metadata only (WithBasicOnly option)...")
	fmt.Println()

	basicMetadata, err := imgx.Metadata(imagePath, imgx.WithBasicOnly())
	if err != nil {
		log.Fatalf("Failed to extract basic metadata: %v", err)
	}

	fmt.Printf("File: %s\n", basicMetadata.FileName)
	fmt.Printf("Format: %s\n", basicMetadata.Format)
	fmt.Printf("Size: %dx%d\n", basicMetadata.Width, basicMetadata.Height)
	fmt.Printf("Has Extended: %v (forced to false with WithBasicOnly)\n", basicMetadata.HasExtended)
	fmt.Println()

	// Example 3: Using metadata for conditional processing
	fmt.Println("---")
	fmt.Println()
	fmt.Println("3. Conditional processing based on metadata...")
	fmt.Println()

	// Check if image needs orientation correction
	if metadata.Orientation != 0 && metadata.Orientation != 1 {
		fmt.Printf("⚠ Image has orientation flag: %d (may need auto-orientation)\n", metadata.Orientation)
	}

	// Check image size
	if metadata.Width > 4000 || metadata.Height > 4000 {
		fmt.Println("ℹ Large image detected - consider resizing for web use")
	} else {
		fmt.Println("✓ Image size is suitable for web use")
	}

	// Check file size
	maxFileSize := int64(5 * 1024 * 1024) // 5 MB
	if metadata.FileSize > maxFileSize {
		fmt.Printf("⚠ Large file size: %.2f MB (consider compression)\n", float64(metadata.FileSize)/(1024*1024))
	} else {
		fmt.Printf("✓ File size OK: %.2f KB\n", float64(metadata.FileSize)/1024)
	}
	fmt.Println()

	fmt.Println("=== Metadata extraction complete! ===")
}
