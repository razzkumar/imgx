package commands

import (
	"context"
	"fmt"

	"github.com/razzkumar/imgx"
	"github.com/urfave/cli/v3"
)

// WatermarkCommand creates the watermark command
func WatermarkCommand() *cli.Command {
	return &cli.Command{
		Name:  "watermark",
		Usage: "Add text watermark to image",
		Description: `Add a text watermark to an image with configurable position, opacity, color, and padding.

Examples:
  imgx watermark photo.jpg --text "Copyright 2025" -o output.jpg
  imgx watermark photo.jpg --text "DRAFT" --opacity 0.3 --anchor center
  imgx watermark photo.jpg --text "Sample" --color ff0000 --padding 20`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "text",
				Aliases:  []string{"t"},
				Usage:    "watermark text (required)",
				Required: true,
			},
			&cli.FloatFlag{
				Name:    "opacity",
				Usage:   "opacity (0.0 to 1.0)",
				Value:   0.5,
				Validator: func(f float64) error {
					if f < 0 || f > 1 {
						return fmt.Errorf("opacity must be between 0.0 and 1.0")
					}
					return nil
				},
			},
			&cli.StringFlag{
				Name:    "anchor",
				Aliases: []string{"a"},
				Usage:   "position (center, topleft, top, topright, left, right, bottomleft, bottom, bottomright)",
				Value:   "bottomright",
			},
			&cli.StringFlag{
				Name:  "color",
				Usage: "text color in hex (RGB or RGBA, e.g., ffffff or ff0000ff)",
				Value: "ffffff",
			},
			&cli.IntFlag{
				Name:  "padding",
				Usage: "padding from edges in pixels",
				Value: 10,
			},
		},
		Action: watermarkAction,
	}
}

func watermarkAction(ctx context.Context, cmd *cli.Command) error {
	if cmd.Args().Len() < 1 {
		return fmt.Errorf("input file required")
	}

	inputPath := cmd.Args().Get(0)
	text := cmd.String("text")
	opacity := cmd.Float("opacity")
	anchorName := cmd.String("anchor")
	colorStr := cmd.String("color")
	padding := cmd.Int("padding")

	// Parse anchor
	anchor, err := ParseAnchor(anchorName)
	if err != nil {
		return err
	}

	// Parse color
	textColor, err := ParseColor(colorStr)
	if err != nil {
		return err
	}

	// Load image
	img, err := loadImage(cmd, inputPath)
	if err != nil {
		return err
	}

	// Apply watermark
	opts := imgx.WatermarkOptions{
		Text:      text,
		Position:  anchor,
		Opacity:   opacity,
		TextColor: textColor,
		Padding:   padding,
	}

	result := imgx.Watermark(img, opts)

	// Save
	outputPath := getOutputPath(cmd, inputPath, "-watermarked")
	return saveImage(cmd, result, outputPath)
}
