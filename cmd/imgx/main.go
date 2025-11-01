package main

import (
	"context"
	"fmt"
	"os"

	"github.com/razzkumar/imgx"
	"github.com/razzkumar/imgx/commands"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:                  "imgx",
		Usage:                 "A powerful command-line image processing tool",
		Version:               imgx.Version,
		EnableShellCompletion: true,
		Description: `imgx is a CLI tool for common image processing operations including:
- Resizing (resize, fit, fill, thumbnail)
- Transformations (rotate, flip, crop)
- Color adjustments (brightness, contrast, gamma, saturation, hue)
- Effects (blur, sharpen, grayscale, invert)
- Watermarking

Examples:
  imgx resize photo.jpg -w 800 -o resized.jpg
  imgx thumbnail photo.jpg -s 150 -o thumb.jpg
  imgx metadata photo.jpg  # or: imgx info photo.jpg`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "output file path (auto-generated if not specified)",
			},
			&cli.IntFlag{
				Name:    "quality",
				Aliases: []string{"q"},
				Usage:   "JPEG quality 1-100 (default: 90)",
				Value:   90,
			},
			&cli.BoolFlag{
				Name:  "auto-orient",
				Usage: "auto-orient based on EXIF data (default: true)",
				Value: true,
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "force output format (jpg, png, gif, tiff, bmp)",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "verbose output",
				Value:   false,
			},
		},
		Commands: []*cli.Command{
			// Resize operations
			commands.ResizeCommand(),
			commands.FitCommand(),
			commands.FillCommand(),
			commands.ThumbnailCommand(),

			// Transform operations
			commands.RotateCommand(),
			commands.Rotate90Command(),
			commands.Rotate180Command(),
			commands.Rotate270Command(),
			commands.FlipCommand(),
			commands.CropCommand(),
			commands.TransposeCommand(),
			commands.TransverseCommand(),

			// Color adjustments
			commands.AdjustCommand(),
			commands.GrayscaleCommand(),
			commands.InvertCommand(),

			// Effects
			commands.BlurCommand(),
			commands.SharpenCommand(),

			// Watermark
			commands.WatermarkCommand(),

			// Info/Metadata
			commands.MetadataCommand(),

			// Object Detection
			commands.DetectCommand(),

			// Completions
			commands.CompletionsCommand(),
		},
	}

	if err := app.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
