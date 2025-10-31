# imgx

[![PkgGoDev](https://pkg.go.dev/badge/github.com/razzkumar/imgx)](https://pkg.go.dev/github.com/razzkumar/imgx)

Package imgx provides basic image processing functions (resize, rotate, crop, brightness/contrast adjustments, etc.).

All the image processing functions provided by the package accept any image type that implements `image.Image` interface
as an input, and return a new image of `*image.NRGBA` type (32bit RGBA colors, non-premultiplied alpha).

## Installation

    go get -u github.com/razzkumar/imgx

## Documentation

https://pkg.go.dev/github.com/razzkumar/imgx

## Usage examples

A few usage examples can be found below. See the documentation for the full list of supported functions.

### Image resizing

```go
// Resize srcImage to size = 128x128px using the Lanczos filter.
dstImage128 := imgx.Resize(srcImage, 128, 128, imgx.Lanczos)

// Resize srcImage to width = 800px preserving the aspect ratio.
dstImage800 := imgx.Resize(srcImage, 800, 0, imgx.Lanczos)

// Scale down srcImage to fit the 800x600px bounding box.
dstImageFit := imgx.Fit(srcImage, 800, 600, imgx.Lanczos)

// Resize and crop the srcImage to fill the 100x100px area.
dstImageFill := imgx.Fill(srcImage, 100, 100, imgx.Center, imgx.Lanczos)
```

Imaging supports image resizing using various resampling filters. The most notable ones:
- `Lanczos` - A high-quality resampling filter for photographic images yielding sharp results.
- `CatmullRom` - A sharp cubic filter that is faster than Lanczos filter while providing similar results.
- `MitchellNetravali` - A cubic filter that produces smoother results with less ringing artifacts than CatmullRom.
- `Linear` - Bilinear resampling filter, produces smooth output. Faster than cubic filters.
- `Box` - Simple and fast averaging filter appropriate for downscaling. When upscaling it's similar to NearestNeighbor.
- `NearestNeighbor` - Fastest resampling filter, no antialiasing.

The full list of supported filters:  NearestNeighbor, Box, Linear, Hermite, MitchellNetravali, CatmullRom, BSpline, Gaussian, Lanczos, Hann, Hamming, Blackman, Bartlett, Welch, Cosine. Custom filters can be created using ResampleFilter struct.

**Resampling filters comparison**

Original image:

![srcImage](testdata/branch.jpg)

The same image can be resized using different resampling filters.
From faster (lower quality) to slower (higher quality): `NearestNeighbor`, `Linear`, `CatmullRom`, `Lanczos`.


### Gaussian Blur

```go
dstImage := imgx.Blur(srcImage, 0.5)
```

Sigma parameter allows to control the strength of the blurring effect.

Original image:

![srcImage](testdata/flower.jpg)

### Sharpening

```go
dstImage := imgx.Sharpen(srcImage, 0.5)
```

`Sharpen` uses gaussian function internally. Sigma parameter allows to control the strength of the sharpening effect.

Original image:

![srcImage](testdata/flower.jpg)

### Gamma correction

```go
dstImage := imgx.AdjustGamma(srcImage, 0.75)
```

Original image:

![srcImage](testdata/flower.jpg)

### Contrast adjustment

```go
dstImage := imgx.AdjustContrast(srcImage, 20)
```

Original image:

![srcImage](testdata/flower.jpg)

### Brightness adjustment

```go
dstImage := imgx.AdjustBrightness(srcImage, 20)
```

Original image:

![srcImage](testdata/flower.jpg)

### Saturation adjustment

```go
dstImage := imgx.AdjustSaturation(srcImage, 20)
```

Original image:

![srcImage](testdata/flower.jpg)

### Hue adjustment

```go
dstImage := imgx.AdjustHue(srcImage, 20)
```

Original image:

![srcImage](testdata/flower.jpg)

## FAQ

### Incorrect image orientation after processing (e.g. an image appears rotated after resizing)

Most probably, the given image contains the EXIF orientation tag.
The standard `image/*` packages do not support loading and saving
this kind of information. To fix the issue, try opening images with
the `AutoOrientation` decode option. If this option is set to `true`,
the image orientation is changed after decoding, according to the
orientation tag (if present). Here's the example:

```go
img, err := imgx.Open("test.jpg", imgx.AutoOrientation(true))
```

### What's the difference between `imaging` and `gift` packages?

[imaging](https://github.com/razzkumar/imgx)
is designed to be a lightweight and simple image manipulation package.
It provides basic image processing functions and a few helper functions
such as `Open` and `Save`. It consistently returns *image.NRGBA image 
type (8 bits per channel, RGBA).

## Example code

```go
package main

import (
	"image"
	"image/color"
	"log"

	"github.com/razzkumar/imgx"
)

func main() {
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
```

## Acknowledgments

This project is based on the [imaging](https://github.com/disintegration/imaging) library created by [Grigory Dryapak](https://github.com/disintegration).

### What's New in imgx?

- **Modernized codebase** using latest Go features (1.21+):
  - Range over int (`for range 256`)
  - Built-in `min`/`max` functions
  - `WaitGroup.Go()` for goroutine management
  - `b.Loop()` for benchmarks
- **Improved code organization**: Refactored complex functions for better maintainability
- **CLI tool**: Coming soon - command-line interface for image processing
- **Active development**: New features and improvements

We're grateful to the original author for creating such a solid foundation for image processing in Go!

## License

This project maintains the original license. Please see the LICENSE file for details.
