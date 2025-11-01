package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/razzkumar/imgx"
)

func main() {
	imagePath := "testdata/branch_flip_horizontal.jpg"

	fmt.Println("=== Metadata Extraction and Tracking Example ===")
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
	fmt.Printf("Aspect Ratio: %s\n", metadata.AspectRatio)
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
			if metadata.CameraSerialNumber != "" {
				fmt.Printf("  Serial Number: %s\n", metadata.CameraSerialNumber)
			}
			if metadata.LensModel != "" {
				fmt.Printf("  Lens: %s\n", metadata.LensModel)
			}
			if metadata.LensSerialNumber != "" {
				fmt.Printf("  Lens Serial: %s\n", metadata.LensSerialNumber)
			}
			if metadata.FirmwareVersion != "" {
				fmt.Printf("  Firmware: %s\n", metadata.FirmwareVersion)
			}
			fmt.Println()
		}

		// Camera settings
		if metadata.FocalLength != "" || metadata.Aperture != "" ||
			metadata.ShutterSpeed != "" || metadata.ISO != "" {
			fmt.Println("Camera Settings:")
			if metadata.FocalLength != "" {
				fmt.Printf("  Focal Length: %s\n", metadata.FocalLength)
			}
			if metadata.Aperture != "" {
				fmt.Printf("  Aperture: %s\n", metadata.Aperture)
			}
			if metadata.ShutterSpeed != "" {
				fmt.Printf("  Shutter Speed: %s\n", metadata.ShutterSpeed)
			}
			if metadata.ISO != "" {
				fmt.Printf("  ISO: %s\n", metadata.ISO)
			}
			if metadata.ExposureCompensation != "" {
				fmt.Printf("  Exposure Compensation: %s\n", metadata.ExposureCompensation)
			}
			if metadata.ExposureMode != "" {
				fmt.Printf("  Exposure Mode: %s\n", metadata.ExposureMode)
			}
			if metadata.WhiteBalance != "" {
				fmt.Printf("  White Balance: %s\n", metadata.WhiteBalance)
			}
			if metadata.Flash != "" {
				fmt.Printf("  Flash: %s\n", metadata.Flash)
			}
			if metadata.FocusMode != "" {
				fmt.Printf("  Focus Mode: %s\n", metadata.FocusMode)
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
			if metadata.GPSSpeed != "" {
				fmt.Printf("  Speed: %s\n", metadata.GPSSpeed)
			}
			if metadata.GPSDirection != "" {
				fmt.Printf("  Direction: %s\n", metadata.GPSDirection)
			}
			fmt.Println()
		}

		// Content and authorship
		if metadata.Title != "" || metadata.Artist != "" || metadata.Copyright != "" {
			fmt.Println("Content & Authorship:")
			if metadata.Title != "" {
				fmt.Printf("  Title: %s\n", metadata.Title)
			}
			if metadata.Subject != "" {
				fmt.Printf("  Subject: %s\n", metadata.Subject)
			}
			if metadata.Keywords != "" {
				fmt.Printf("  Keywords: %s\n", metadata.Keywords)
			}
			if metadata.Rating > 0 {
				fmt.Printf("  Rating: %d stars\n", metadata.Rating)
			}
			if metadata.Artist != "" {
				fmt.Printf("  Artist: %s\n", metadata.Artist)
			}
			if metadata.Copyright != "" {
				fmt.Printf("  Copyright: %s\n", metadata.Copyright)
			}
			if metadata.Software != "" {
				fmt.Printf("  Software: %s\n", metadata.Software)
			}
			fmt.Println()
		}

		// Technical details
		if metadata.ColorSpace != "" || metadata.BitDepth > 0 {
			fmt.Println("Technical Details:")
			if metadata.ColorSpace != "" {
				fmt.Printf("  Color Space: %s\n", metadata.ColorSpace)
			}
			if metadata.BitDepth > 0 {
				fmt.Printf("  Bit Depth: %d\n", metadata.BitDepth)
			}
			if metadata.Compression != "" {
				fmt.Printf("  Compression: %s\n", metadata.Compression)
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

	// Example 4: Demonstrating metadata tracking with image processing
	fmt.Println("---")
	fmt.Println()
	fmt.Println("4. Processing metadata tracking (new instance-based API)...")
	fmt.Println()

	// Load image using new instance-based API
	img, err := imgx.Load(imagePath)
	if err != nil {
		log.Fatalf("Failed to load image: %v", err)
	}

	fmt.Println("Performing image operations with automatic metadata tracking:")
	fmt.Println()

	// Chain multiple operations - metadata is tracked automatically
	processed := img.
		Resize(800, 0, imgx.Lanczos).
		AdjustContrast(20).
		AdjustBrightness(10).
		Sharpen(1.2)

	// Retrieve processing metadata
	procMetadata := processed.GetMetadata()

	fmt.Printf("Total operations performed: %d\n", len(procMetadata.Operations))
	fmt.Println()
	fmt.Println("Operations history:")
	for i, op := range procMetadata.Operations {
		fmt.Printf("  %d. %s\n", i+1, op.Action)
		if op.Parameters != "" {
			fmt.Printf("     Parameters: %s\n", op.Parameters)
		}
		fmt.Printf("     Timestamp: %s\n", op.Timestamp.Format("2006-01-02 15:04:05"))
	}
	fmt.Println()

	// Save the processed image (metadata will be written to XMP sidecar)
	outputPath := "testdata/flower_processed.jpg"
	if err := processed.Save(outputPath); err != nil {
		log.Fatalf("Failed to save processed image: %v", err)
	}
	fmt.Printf("Processed image saved to: %s\n", outputPath)
	fmt.Printf("XMP sidecar created: %s\n", outputPath+".xmp")
	fmt.Println()

	// Example 5: JSON export of all metadata
	fmt.Println("---")
	fmt.Println()
	fmt.Println("5. JSON representation of extracted metadata...")
	fmt.Println()

	jsonData, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(jsonData))
	fmt.Println()

	fmt.Println("=== Metadata extraction and tracking complete! ===")
}
