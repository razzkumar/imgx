package imgx

import (
	"fmt"
	"image"
	"math"
)

func gaussianBlurKernel(x, sigma float64) float64 {
	return math.Exp(-(x*x)/(2*sigma*sigma)) / (sigma * math.Sqrt(2*math.Pi))
}

// Blur produces a blurred version of the image using a Gaussian function.
// Sigma parameter must be positive and indicates how much the image will be blurred.
//
// Example:
//
//	dstImage := imaging.Blur(srcImage, 3.5)
//
func Blur(img image.Image, sigma float64) *image.NRGBA {
	if sigma <= 0 {
		return Clone(img)
	}

	radius := int(math.Ceil(sigma * 3.0))
	kernel := make([]float64, radius+1)

	for i := 0; i <= radius; i++ {
		kernel[i] = gaussianBlurKernel(float64(i), sigma)
	}

	return blurVertical(blurHorizontal(img, kernel), kernel)
}

func blurHorizontal(img image.Image, kernel []float64) *image.NRGBA {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	radius := len(kernel) - 1

	parallel(0, src.h, func(ys <-chan int) {
		scanLine := make([]uint8, src.w*4)
		scanLineF := make([]float64, len(scanLine))
		for y := range ys {
			src.scan(0, y, src.w, y+1, scanLine)
			for i, v := range scanLine {
				scanLineF[i] = float64(v)
			}
			for x := 0; x < src.w; x++ {
				minX := max(0, x-radius)
				maxX := min(src.w-1, x+radius)
				var r, g, b, a, wsum float64
				for ix := minX; ix <= maxX; ix++ {
					i := ix * 4
					weight := kernel[absint(x-ix)]
					wsum += weight
					s := scanLineF[i : i+4 : i+4]
					wa := s[3] * weight
					r += s[0] * wa
					g += s[1] * wa
					b += s[2] * wa
					a += wa
				}
				if a != 0 {
					aInv := 1 / a
					j := y*dst.Stride + x*4
					d := dst.Pix[j : j+4 : j+4]
					d[0] = clamp(r * aInv)
					d[1] = clamp(g * aInv)
					d[2] = clamp(b * aInv)
					d[3] = clamp(a / wsum)
				}
			}
		}
	})

	return dst
}

func blurVertical(img image.Image, kernel []float64) *image.NRGBA {
	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	radius := len(kernel) - 1

	parallel(0, src.w, func(xs <-chan int) {
		scanLine := make([]uint8, src.h*4)
		scanLineF := make([]float64, len(scanLine))
		for x := range xs {
			src.scan(x, 0, x+1, src.h, scanLine)
			for i, v := range scanLine {
				scanLineF[i] = float64(v)
			}
			for y := 0; y < src.h; y++ {
				minY := max(0, y-radius)
				maxY := min(src.h-1, y+radius)
				var r, g, b, a, wsum float64
				for iy := minY; iy <= maxY; iy++ {
					i := iy * 4
					weight := kernel[absint(y-iy)]
					wsum += weight
					s := scanLineF[i : i+4 : i+4]
					wa := s[3] * weight
					r += s[0] * wa
					g += s[1] * wa
					b += s[2] * wa
					a += wa
				}
				if a != 0 {
					aInv := 1 / a
					j := y*dst.Stride + x*4
					d := dst.Pix[j : j+4 : j+4]
					d[0] = clamp(r * aInv)
					d[1] = clamp(g * aInv)
					d[2] = clamp(b * aInv)
					d[3] = clamp(a / wsum)
				}
			}
		}
	})

	return dst
}

// Sharpen produces a sharpened version of the image.
// Sigma parameter must be positive and indicates how much the image will be sharpened.
//
// Example:
//
//	dstImage := imaging.Sharpen(srcImage, 3.5)
//
func Sharpen(img image.Image, sigma float64) *image.NRGBA {
	if sigma <= 0 {
		return Clone(img)
	}

	src := newScanner(img)
	dst := image.NewNRGBA(image.Rect(0, 0, src.w, src.h))
	blurred := Blur(img, sigma)

	parallel(0, src.h, func(ys <-chan int) {
		scanLine := make([]uint8, src.w*4)
		for y := range ys {
			src.scan(0, y, src.w, y+1, scanLine)
			j := y * dst.Stride
			for i := 0; i < src.w*4; i++ {
				val := int(scanLine[i])<<1 - int(blurred.Pix[j])
				val = max(0, min(val, 0xff))
				dst.Pix[j] = uint8(val)
				j++
			}
		}
	})

	return dst
}
// Blur applies Gaussian blur to the image
func (img *Image) Blur(sigma float64) *Image {
	newData := Blur(img.data, sigma)
	newMeta := img.metadata.Clone()
	newMeta.AddOperation("blur", fmt.Sprintf("sigma=%.2f", sigma))
	return &Image{data: newData, metadata: newMeta}
}

// Sharpen sharpens the image
func (img *Image) Sharpen(sigma float64) *Image {
	newData := Sharpen(img.data, sigma)
	newMeta := img.metadata.Clone()
	newMeta.AddOperation("sharpen", fmt.Sprintf("sigma=%.2f", sigma))
	return &Image{data: newData, metadata: newMeta}
}
