package imaging

import (
	"image"
	"image/color"
	"testing"

	"golang.org/x/image/font/basicfont"
)

func TestWatermark(t *testing.T) {
	testCases := []struct {
		name string
		src  image.Image
		opts WatermarkOptions
	}{
		{
			"Watermark with default options",
			image.NewNRGBA(image.Rect(0, 0, 100, 100)),
			WatermarkOptions{
				Text:     "Test",
				Position: BottomRight,
				Opacity:  0.5,
			},
		},
		{
			"Watermark top left",
			image.NewNRGBA(image.Rect(0, 0, 100, 100)),
			WatermarkOptions{
				Text:      "Copyright",
				Position:  TopLeft,
				Opacity:   0.8,
				TextColor: color.White,
				Padding:   5,
			},
		},
		{
			"Watermark center",
			image.NewNRGBA(image.Rect(0, 0, 200, 100)),
			WatermarkOptions{
				Text:      "CONFIDENTIAL",
				Position:  Center,
				Opacity:   0.3,
				TextColor: color.NRGBA{255, 0, 0, 255},
				Font:      basicfont.Face7x13,
			},
		},
		{
			"Watermark bottom right with custom color",
			image.NewNRGBA(image.Rect(0, 0, 150, 150)),
			WatermarkOptions{
				Text:      "Sample",
				Position:  BottomRight,
				Opacity:   1.0,
				TextColor: color.NRGBA{0, 255, 0, 255},
				Padding:   20,
			},
		},
		{
			"Watermark all positions",
			image.NewNRGBA(image.Rect(0, 0, 300, 200)),
			WatermarkOptions{
				Text:     "WM",
				Position: Bottom,
				Opacity:  0.7,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Watermark(tc.src, tc.opts)

			// Check that result is not nil
			if result == nil {
				t.Fatal("result should not be nil")
			}

			// Check that dimensions match input
			if result.Bounds() != tc.src.Bounds() {
				t.Fatalf("result bounds %v should match source bounds %v", result.Bounds(), tc.src.Bounds())
			}

			// Result should be of type *image.NRGBA
			if _, ok := interface{}(result).(*image.NRGBA); !ok {
				t.Fatal("result should be *image.NRGBA")
			}
		})
	}
}

func TestWatermarkEmptyText(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	opts := WatermarkOptions{
		Text:    "",
		Opacity: 0.5,
	}

	result := Watermark(src, opts)

	// Should return unchanged image when text is empty
	if !compareNRGBA(result, src, 0) {
		t.Fatal("result should match source when text is empty")
	}
}

func TestWatermarkZeroOpacity(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 100, 100))
	opts := WatermarkOptions{
		Text:    "Test",
		Opacity: 0.0,
	}

	result := Watermark(src, opts)

	// Should return unchanged image when opacity is 0
	if !compareNRGBA(result, src, 0) {
		t.Fatal("result should match source when opacity is 0")
	}
}

func TestWatermarkOpacityBounds(t *testing.T) {
	testCases := []struct {
		name    string
		opacity float64
	}{
		{"Opacity above 1.0", 2.0},
		{"Opacity negative", -0.5},
		{"Opacity at 0.0", 0.0},
		{"Opacity at 1.0", 1.0},
		{"Opacity at 0.5", 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			src := image.NewNRGBA(image.Rect(0, 0, 100, 100))
			opts := WatermarkOptions{
				Text:    "Test",
				Opacity: tc.opacity,
			}

			result := Watermark(src, opts)

			// Should handle opacity gracefully without panicking
			if result == nil {
				t.Fatal("result should not be nil")
			}
		})
	}
}

func TestWatermarkAllPositions(t *testing.T) {
	positions := []Anchor{
		TopLeft, Top, TopRight,
		Left, Center, Right,
		BottomLeft, Bottom, BottomRight,
	}

	for _, pos := range positions {
		t.Run(pos.String(), func(t *testing.T) {
			src := image.NewNRGBA(image.Rect(0, 0, 200, 200))
			opts := WatermarkOptions{
				Text:     "WM",
				Position: pos,
				Opacity:  0.5,
			}

			result := Watermark(src, opts)

			if result == nil {
				t.Fatal("result should not be nil")
			}

			if result.Bounds() != src.Bounds() {
				t.Fatalf("result bounds should match source bounds")
			}
		})
	}
}

func TestWatermarkDifferentImageTypes(t *testing.T) {
	testCases := []struct {
		name string
		src  image.Image
	}{
		{
			"NRGBA image",
			image.NewNRGBA(image.Rect(0, 0, 100, 100)),
		},
		{
			"RGBA image",
			image.NewRGBA(image.Rect(0, 0, 100, 100)),
		},
		{
			"Gray image",
			image.NewGray(image.Rect(0, 0, 100, 100)),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := WatermarkOptions{
				Text:    "Test",
				Opacity: 0.5,
			}

			result := Watermark(tc.src, opts)

			if result == nil {
				t.Fatal("result should not be nil")
			}

			// Result should always be *image.NRGBA
			if _, ok := interface{}(result).(*image.NRGBA); !ok {
				t.Fatal("result should be *image.NRGBA")
			}
		})
	}
}

func TestWatermarkPreservesOriginal(t *testing.T) {
	// Create source image with specific pixel data
	src := image.NewNRGBA(image.Rect(0, 0, 50, 50))
	for i := range src.Pix {
		src.Pix[i] = 128 // Fill with gray
	}

	// Create a copy to compare later
	srcCopy := Clone(src)

	opts := WatermarkOptions{
		Text:    "Test",
		Opacity: 0.5,
	}

	// Apply watermark
	_ = Watermark(src, opts)

	// Original should be unchanged (Watermark clones internally)
	if !compareNRGBA(src, srcCopy, 0) {
		t.Fatal("source image should not be modified")
	}
}

func TestCalculateWatermarkPosition(t *testing.T) {
	bounds := image.Rect(0, 0, 100, 100)
	textWidth := 20
	textHeight := 10
	padding := 5

	testCases := []struct {
		name     string
		anchor   Anchor
		expected image.Point
	}{
		{
			"TopLeft",
			TopLeft,
			image.Point{X: 5, Y: 5},
		},
		{
			"TopRight",
			TopRight,
			image.Point{X: 75, Y: 5},
		},
		{
			"BottomLeft",
			BottomLeft,
			image.Point{X: 5, Y: 85},
		},
		{
			"BottomRight",
			BottomRight,
			image.Point{X: 75, Y: 85},
		},
		{
			"Center",
			Center,
			image.Point{X: 40, Y: 45},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pos := calculateWatermarkPosition(bounds, textWidth, textHeight, tc.anchor, padding)

			if pos != tc.expected {
				t.Fatalf("got position %v, want %v", pos, tc.expected)
			}
		})
	}
}

func TestApplyOpacity(t *testing.T) {
	testCases := []struct {
		name          string
		opacity       float64
		initialAlpha  uint8
		expectedAlpha uint8
	}{
		{
			"Half opacity",
			0.5,
			255,
			128,
		},
		{
			"Full opacity",
			1.0,
			255,
			255,
		},
		{
			"Quarter opacity",
			0.25,
			200,
			50,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img := image.NewNRGBA(image.Rect(0, 0, 1, 1))
			img.Pix[3] = tc.initialAlpha // Set alpha channel

			applyOpacity(img, tc.opacity)

			if img.Pix[3] != tc.expectedAlpha {
				t.Fatalf("got alpha %d, want %d", img.Pix[3], tc.expectedAlpha)
			}
		})
	}
}

func BenchmarkWatermark(b *testing.B) {
	src := image.NewNRGBA(image.Rect(0, 0, 1024, 768))
	opts := WatermarkOptions{
		Text:      "Copyright 2025",
		Position:  BottomRight,
		Opacity:   0.5,
		TextColor: color.White,
		Padding:   10,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Watermark(src, opts)
	}
}

func BenchmarkWatermarkSmall(b *testing.B) {
	src := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	opts := WatermarkOptions{
		Text:     "WM",
		Position: BottomRight,
		Opacity:  0.5,
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		Watermark(src, opts)
	}
}

// Helper method for Anchor.String() used in tests
func (a Anchor) String() string {
	switch a {
	case Center:
		return "Center"
	case TopLeft:
		return "TopLeft"
	case Top:
		return "Top"
	case TopRight:
		return "TopRight"
	case Left:
		return "Left"
	case Right:
		return "Right"
	case BottomLeft:
		return "BottomLeft"
	case Bottom:
		return "Bottom"
	case BottomRight:
		return "BottomRight"
	default:
		return "Unknown"
	}
}
