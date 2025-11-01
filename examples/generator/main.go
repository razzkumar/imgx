package main

import (
	"image"
	"image/color"
	"log"

	"github.com/razzkumar/imgx"
)

func main() {
	// Open source images
	flower, err := imgx.Open("../../testdata/flower.jpg")
	if err != nil {
		log.Fatalf("failed to open flower.jpg: %v", err)
	}

	branch, err := imgx.Open("../../testdata/branch.jpg")
	if err != nil {
		log.Fatalf("failed to open branch.jpg: %v", err)
	}

	// 1. Resize examples
	log.Println("Creating resize examples...")
	resized200 := imgx.Resize(flower, 200, 0, imgx.Lanczos)
	imgx.Save(resized200, "../../testdata/flower_resized_200.jpg")

	resized800 := imgx.Resize(flower, 800, 0, imgx.Lanczos)
	imgx.Save(resized800, "../../testdata/flower_resized_800.jpg")

	// 2. Fill/Thumbnail example
	log.Println("Creating thumbnail example...")
	thumbnail := imgx.Fill(flower, 300, 300, imgx.Center, imgx.Lanczos)
	imgx.Save(thumbnail, "../../testdata/flower_thumbnail_300x300.jpg")

	// 3. Rotate examples
	log.Println("Creating rotate examples...")
	rotated90 := imgx.Rotate90(branch)
	imgx.Save(rotated90, "../../testdata/branch_rotated_90.jpg")

	rotated45 := imgx.Rotate(branch, 45, color.NRGBA{255, 255, 255, 255})
	imgx.Save(rotated45, "../../testdata/branch_rotated_45.jpg")

	// 4. Blur examples
	log.Println("Creating blur examples...")
	blurred := imgx.Blur(flower, 2.0)
	imgx.Save(blurred, "../../testdata/flower_blur_2.jpg")

	// 5. Sharpen example
	log.Println("Creating sharpen example...")
	sharpened := imgx.Sharpen(flower, 1.5)
	imgx.Save(sharpened, "../../testdata/flower_sharpen_1.5.jpg")

	// 6. Color adjustments
	log.Println("Creating color adjustment examples...")
	brightness := imgx.AdjustBrightness(flower, 30)
	imgx.Save(brightness, "../../testdata/flower_brightness_30.jpg")

	contrast := imgx.AdjustContrast(flower, 30)
	imgx.Save(contrast, "../../testdata/flower_contrast_30.jpg")

	saturation := imgx.AdjustSaturation(flower, 50)
	imgx.Save(saturation, "../../testdata/flower_saturation_50.jpg")

	grayscale := imgx.Grayscale(flower)
	imgx.Save(grayscale, "../../testdata/flower_grayscale.jpg")

	// 7. Flip examples
	log.Println("Creating flip examples...")
	flippedH := imgx.FlipH(branch)
	imgx.Save(flippedH, "../../testdata/branch_flip_horizontal.jpg")

	// 8. Watermark example (using branch as watermark on flower)
	log.Println("Creating watermark example...")
	// Resize branch to be smaller for watermark
	watermark := imgx.Resize(branch, 150, 0, imgx.Lanczos)
	// Position in bottom-right
	flowerBounds := flower.Bounds()
	wmBounds := watermark.Bounds()
	pos := image.Pt(flowerBounds.Dx()-wmBounds.Dx()-20, flowerBounds.Dy()-wmBounds.Dy()-20)
	watermarked := imgx.Overlay(flower, watermark, pos, 0.6)
	imgx.Save(watermarked, "../../testdata/flower_watermarked.jpg")

	log.Println("All example images generated successfully!")
}
