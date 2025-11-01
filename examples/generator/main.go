package main

import (
	"image"
	"image/color"
	"log"

	"github.com/razzkumar/imgx"
)

func main() {
	// Load source images using new instance-based API
	flower, err := imgx.Load("testdata/flower.jpg")
	if err != nil {
		log.Fatalf("failed to load flower.jpg: %v", err)
	}

	branch, err := imgx.Load("testdata/branch.jpg")
	if err != nil {
		log.Fatalf("failed to load branch.jpg: %v", err)
	}

	// 1. Resize examples
	log.Println("Creating resize examples...")
	if err := flower.Resize(200, 0, imgx.Lanczos).Save("testdata/flower_resized_200.jpg"); err != nil {
		log.Fatalf("failed to save flower_resized_200.jpg: %v", err)
	}

	if err := flower.Resize(800, 0, imgx.Lanczos).Save("testdata/flower_resized_800.jpg"); err != nil {
		log.Fatalf("failed to save flower_resized_800.jpg: %v", err)
	}

	// 2. Fill/Thumbnail example
	log.Println("Creating thumbnail example...")
	if err := flower.Fill(300, 300, imgx.Center, imgx.Lanczos).Save("testdata/flower_thumbnail_300x300.jpg"); err != nil {
		log.Fatalf("failed to save thumbnail: %v", err)
	}

	// 3. Rotate examples
	log.Println("Creating rotate examples...")
	if err := branch.Rotate90().Save("testdata/branch_rotated_90.jpg"); err != nil {
		log.Fatalf("failed to save rotated image: %v", err)
	}

	if err := branch.Rotate(45, color.NRGBA{255, 255, 255, 255}).Save("testdata/branch_rotated_45.jpg"); err != nil {
		log.Fatalf("failed to save rotated image: %v", err)
	}

	// 4. Blur examples
	log.Println("Creating blur examples...")
	if err := flower.Blur(2.0).Save("testdata/flower_blur_2.jpg"); err != nil {
		log.Fatalf("failed to save blurred image: %v", err)
	}

	// 5. Sharpen example
	log.Println("Creating sharpen example...")
	if err := flower.Sharpen(1.5).Save("testdata/flower_sharpen_1.5.jpg"); err != nil {
		log.Fatalf("failed to save sharpened image: %v", err)
	}

	// 6. Color adjustments - demonstrating method chaining
	log.Println("Creating color adjustment examples...")
	if err := flower.AdjustBrightness(30).Save("testdata/flower_brightness_30.jpg"); err != nil {
		log.Fatalf("failed to save brightness adjusted image: %v", err)
	}

	if err := flower.AdjustContrast(30).Save("testdata/flower_contrast_30.jpg"); err != nil {
		log.Fatalf("failed to save contrast adjusted image: %v", err)
	}

	if err := flower.AdjustSaturation(50).Save("testdata/flower_saturation_50.jpg"); err != nil {
		log.Fatalf("failed to save saturation adjusted image: %v", err)
	}

	if err := flower.Grayscale().Save("testdata/flower_grayscale.jpg"); err != nil {
		log.Fatalf("failed to save grayscale image: %v", err)
	}

	// 7. Flip examples
	log.Println("Creating flip examples...")
	if err := branch.FlipH().Save("testdata/branch_flip_horizontal.jpg"); err != nil {
		log.Fatalf("failed to save flipped image: %v", err)
	}

	// 8. Watermark example with method chaining
	log.Println("Creating watermark example...")
	// Resize branch to be smaller for watermark
	watermark := branch.Resize(150, 0, imgx.Lanczos)
	// Position in bottom-right
	flowerBounds := flower.Bounds()
	wmBounds := watermark.Bounds()
	pos := image.Pt(flowerBounds.Dx()-wmBounds.Dx()-20, flowerBounds.Dy()-wmBounds.Dy()-20)
	if err := flower.Overlay(watermark, pos, 0.6).Save("testdata/flower_watermarked.jpg"); err != nil {
		log.Fatalf("failed to save watermarked image: %v", err)
	}

	// 9. Demonstrate complex method chaining
	log.Println("Creating complex chained transformation...")
	if err := flower.
		Resize(400, 0, imgx.Lanczos).
		AdjustContrast(20).
		AdjustSaturation(30).
		Sharpen(1.0).
		Save("testdata/flower_complex_chain.jpg"); err != nil {
		log.Fatalf("failed to save chained transformation: %v", err)
	}

	log.Println("All example images generated successfully!")
}
