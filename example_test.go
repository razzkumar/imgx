package imgx_test

import (
	"image"
	"image/color"
	"log"

	"github.com/razzkumar/imgx"
)

func Example() {
	// Load a test image.
	src, err := imgx.Load("testdata/flower.jpg")
	if err != nil {
		log.Fatalf("failed to load image: %v", err)
	}

	// Crop the original image to 300x300px size using the center anchor.
	src = src.CropAnchor(300, 300, imgx.Center)

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	src = src.Resize(200, 0, imgx.Lanczos)

	// Create a blurred version of the image.
	img1 := src.Blur(5)

	// Create a grayscale version of the image with higher contrast and sharpness.
	img2 := src.Grayscale().AdjustContrast(20).Sharpen(2)

	// Create an inverted version of the image.
	img3 := src.Invert()

	// Create an embossed version of the image using a convolution filter.
	img4 := src.Convolve3x3(
		[9]float64{
			-1, -1, 0,
			-1, 1, 1,
			0, 1, 1,
		},
		nil,
	)

	// Create a new image and paste the four produced images into it.
	dst := imgx.NewImage(400, 400, color.NRGBA{0, 0, 0, 0})
	dst = dst.Paste(img1, image.Pt(0, 0))
	dst = dst.Paste(img2, image.Pt(0, 200))
	dst = dst.Paste(img3, image.Pt(200, 0))
	dst = dst.Paste(img4, image.Pt(200, 200))

	// Save the resulting image as JPEG.
	err = dst.Save("testdata/out_example.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
}
