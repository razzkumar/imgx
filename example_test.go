package imgx_test

import (
	"image"
	"image/color"
	"log"

	"github.com/razzkumar/imgx"
)

func Example() {
	// Open a test image.
	src, err := imgx.Open("testdata/flower.jpg")
	if err != nil {
		log.Fatalf("failed to open image: %v", err)
	}

	// Crop the original image to 300x300px size using the center anchor.
	src = imgx.CropAnchor(src, 300, 300, imgx.Center)

	// Resize the cropped image to width = 200px preserving the aspect ratio.
	src = imgx.Resize(src, 200, 0, imgx.Lanczos)

	// Create a blurred version of the image.
	img1 := imgx.Blur(src, 5)

	// Create a grayscale version of the image with higher contrast and sharpness.
	img2 := imgx.Grayscale(src)
	img2 = imgx.AdjustContrast(img2, 20)
	img2 = imgx.Sharpen(img2, 2)

	// Create an inverted version of the image.
	img3 := imgx.Invert(src)

	// Create an embossed version of the image using a convolution filter.
	img4 := imgx.Convolve3x3(
		src,
		[9]float64{
			-1, -1, 0,
			-1, 1, 1,
			0, 1, 1,
		},
		nil,
	)

	// Create a new image and paste the four produced images into it.
	dst := imgx.New(400, 400, color.NRGBA{0, 0, 0, 0})
	dst = imgx.Paste(dst, img1, image.Pt(0, 0))
	dst = imgx.Paste(dst, img2, image.Pt(0, 200))
	dst = imgx.Paste(dst, img3, image.Pt(200, 0))
	dst = imgx.Paste(dst, img4, image.Pt(200, 200))

	// Save the resulting image as JPEG.
	err = imgx.Save(dst, "testdata/out_example.jpg")
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
}
