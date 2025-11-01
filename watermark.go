package imgx

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// WatermarkOptions contains options for adding a text watermark to an image.
type WatermarkOptions struct {
	// Text is the watermark text to be rendered on the image.
	Text string

	// Position specifies where to place the watermark (e.g., BottomRight, TopLeft).
	// Default is BottomRight.
	Position Anchor

	// Opacity controls the transparency of the watermark text (0.0 to 1.0).
	// 0.0 is fully transparent, 1.0 is fully opaque.
	// Default is 0.5 (50% opacity).
	Opacity float64

	// TextColor is the color of the watermark text.
	// Default is white (255, 255, 255).
	TextColor color.Color

	// Font is the font face to use for rendering the text.
	// If nil, basicfont.Face7x13 is used as default.
	Font font.Face

	// Padding is the number of pixels to offset from the edge based on Position.
	// Default is 10 pixels.
	Padding int
}

// Watermark adds a text watermark to an image and returns the result.
//
// Example:
//
//	opts := imaging.WatermarkOptions{
//		Text:     "Copyright 2025",
//		Position: imaging.BottomRight,
//		Opacity:  0.5,
//	}
//	watermarkedImage := imaging.Watermark(srcImage, opts)
//
func Watermark(img image.Image, opts WatermarkOptions) *image.NRGBA {
	// Clone the input image to avoid modifying the original
	dst := Clone(img)

	// Set defaults
	if opts.Text == "" {
		return dst // Nothing to watermark
	}

	if opts.Opacity <= 0 {
		return dst // Fully transparent, nothing to draw
	}

	if opts.Opacity > 1.0 {
		opts.Opacity = 1.0
	}

	if opts.Font == nil {
		opts.Font = basicfont.Face7x13
	}

	if opts.TextColor == nil {
		opts.TextColor = color.White
	}

	if opts.Padding == 0 {
		opts.Padding = 10
	}

	// Measure the text dimensions
	textBounds, textAdvance := measureText(opts.Text, opts.Font)
	textWidth := textAdvance.Ceil()
	textHeight := textBounds.Max.Y.Ceil() - textBounds.Min.Y.Ceil()

	// Calculate position based on anchor
	bounds := dst.Bounds()
	pos := calculateWatermarkPosition(bounds, textWidth, textHeight, opts.Position, opts.Padding)

	// Create a temporary image for the text with alpha
	textImg := image.NewNRGBA(image.Rect(0, 0, textWidth, textHeight))

	// Draw text on the temporary image
	drawer := &font.Drawer{
		Dst:  textImg,
		Src:  image.NewUniform(opts.TextColor),
		Face: opts.Font,
		Dot:  fixed.Point26_6{X: 0, Y: opts.Font.Metrics().Ascent},
	}
	drawer.DrawString(opts.Text)

	// Apply opacity to the text image
	if opts.Opacity < 1.0 {
		applyOpacity(textImg, opts.Opacity)
	}

	// Overlay the text onto the destination image
	draw.Draw(dst, image.Rect(pos.X, pos.Y, pos.X+textWidth, pos.Y+textHeight), textImg, image.Point{}, draw.Over)

	return dst
}

// measureText measures the dimensions of the given text using the specified font.
func measureText(text string, face font.Face) (fixed.Rectangle26_6, fixed.Int26_6) {
	drawer := &font.Drawer{
		Face: face,
	}

	// Measure the text bounds
	var totalAdvance fixed.Int26_6
	var minX, minY, maxX, maxY fixed.Int26_6

	for i, r := range text {
		bounds, advance, ok := face.GlyphBounds(r)
		if !ok {
			continue
		}

		// Adjust bounds by current position
		bounds.Min.X += totalAdvance
		bounds.Max.X += totalAdvance

		if i == 0 {
			minX = bounds.Min.X
			minY = bounds.Min.Y
			maxX = bounds.Max.X
			maxY = bounds.Max.Y
		} else {
			if bounds.Min.X < minX {
				minX = bounds.Min.X
			}
			if bounds.Min.Y < minY {
				minY = bounds.Min.Y
			}
			if bounds.Max.X > maxX {
				maxX = bounds.Max.X
			}
			if bounds.Max.Y > maxY {
				maxY = bounds.Max.Y
			}
		}

		totalAdvance += advance
	}

	bounds := fixed.Rectangle26_6{
		Min: fixed.Point26_6{X: minX, Y: minY},
		Max: fixed.Point26_6{X: maxX, Y: maxY},
	}

	// Use the string width method from drawer for accurate measurement
	advance := drawer.MeasureString(text)

	return bounds, advance
}

// calculateWatermarkPosition calculates the top-left position for the watermark text
// based on the anchor point and padding.
func calculateWatermarkPosition(bounds image.Rectangle, textWidth, textHeight int, anchor Anchor, padding int) image.Point {
	var x, y int

	switch anchor {
	case TopLeft:
		x = bounds.Min.X + padding
		y = bounds.Min.Y + padding
	case Top:
		x = bounds.Min.X + (bounds.Dx()-textWidth)/2
		y = bounds.Min.Y + padding
	case TopRight:
		x = bounds.Max.X - textWidth - padding
		y = bounds.Min.Y + padding
	case Left:
		x = bounds.Min.X + padding
		y = bounds.Min.Y + (bounds.Dy()-textHeight)/2
	case Center:
		x = bounds.Min.X + (bounds.Dx()-textWidth)/2
		y = bounds.Min.Y + (bounds.Dy()-textHeight)/2
	case Right:
		x = bounds.Max.X - textWidth - padding
		y = bounds.Min.Y + (bounds.Dy()-textHeight)/2
	case BottomLeft:
		x = bounds.Min.X + padding
		y = bounds.Max.Y - textHeight - padding
	case Bottom:
		x = bounds.Min.X + (bounds.Dx()-textWidth)/2
		y = bounds.Max.Y - textHeight - padding
	case BottomRight:
		x = bounds.Max.X - textWidth - padding
		y = bounds.Max.Y - textHeight - padding
	default: // Default to BottomRight
		x = bounds.Max.X - textWidth - padding
		y = bounds.Max.Y - textHeight - padding
	}

	return image.Point{X: x, Y: y}
}

// applyOpacity applies an opacity value to an NRGBA image by multiplying
// the alpha channel of each pixel.
func applyOpacity(img *image.NRGBA, opacity float64) {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			i := img.PixOffset(x, y)
			// Multiply the alpha channel by the opacity
			alpha := float64(img.Pix[i+3]) * opacity
			img.Pix[i+3] = uint8(alpha + 0.5)
		}
	}
}
// Watermark adds a text watermark to the image
func (img *Image) Watermark(opts WatermarkOptions) *Image {
	newData := Watermark(img.data, opts)
	newMeta := img.metadata.Clone()
	params := fmt.Sprintf("text=%q, position=%s, opacity=%.2f", opts.Text, formatAnchorName(opts.Position), opts.Opacity)
	newMeta.AddOperation("watermark", params)
	return &Image{data: newData, metadata: newMeta}
}
