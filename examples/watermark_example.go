package main

import (
	"image/color"
	"log"

	"github.com/disintegration/imaging"
)

func main() {
	// Open an image file
	src, err := imaging.Open("input.jpg")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	// Example 1: Simple watermark with default settings
	watermarked1 := imaging.Watermark(src, imaging.WatermarkOptions{
		Text:    "Copyright 2025",
		Opacity: 0.5,
	})

	// Save the watermarked image
	err = imaging.Save(watermarked1, "output1.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}

	// Example 2: Custom watermark with position and color
	watermarked2 := imaging.Watermark(src, imaging.WatermarkOptions{
		Text:      "CONFIDENTIAL",
		Position:  imaging.Center,
		Opacity:   0.3,
		TextColor: color.NRGBA{255, 0, 0, 255}, // Red text
		Padding:   20,
	})

	err = imaging.Save(watermarked2, "output2.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}

	// Example 3: Top-left watermark with high opacity
	watermarked3 := imaging.Watermark(src, imaging.WatermarkOptions{
		Text:      "Sample Watermark",
		Position:  imaging.TopLeft,
		Opacity:   0.8,
		TextColor: color.White,
		Padding:   10,
	})

	err = imaging.Save(watermarked3, "output3.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}

	log.Println("Watermarks applied successfully!")
}
