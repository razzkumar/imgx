package detection

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"testing"
)

// TestImageToJPEGBytes tests image to JPEG conversion
func TestImageToJPEGBytes(t *testing.T) {
	tests := []struct {
		name   string
		img    *image.NRGBA
		width  int
		height int
	}{
		{
			name:   "small image 10x10",
			img:    createTestImage(10, 10, color.NRGBA{R: 255, G: 0, B: 0, A: 255}),
			width:  10,
			height: 10,
		},
		{
			name:   "medium image 100x100",
			img:    createTestImage(100, 100, color.NRGBA{R: 0, G: 255, B: 0, A: 255}),
			width:  100,
			height: 100,
		},
		{
			name:   "rectangular 50x100",
			img:    createTestImage(50, 100, color.NRGBA{R: 0, G: 0, B: 255, A: 255}),
			width:  50,
			height: 100,
		},
		{
			name:   "1x1 image",
			img:    createTestImage(1, 1, color.NRGBA{R: 128, G: 128, B: 128, A: 255}),
			width:  1,
			height: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := imageToJPEGBytes(tt.img)
			if err != nil {
				t.Fatalf("imageToJPEGBytes() error = %v", err)
			}

			if len(data) == 0 {
				t.Error("imageToJPEGBytes() returned empty data")
			}

			// Verify it's valid JPEG by decoding it
			decoded, err := jpeg.Decode(bytes.NewReader(data))
			if err != nil {
				t.Fatalf("Failed to decode generated JPEG: %v", err)
			}

			bounds := decoded.Bounds()
			if bounds.Dx() != tt.width {
				t.Errorf("Decoded width = %d, want %d", bounds.Dx(), tt.width)
			}

			if bounds.Dy() != tt.height {
				t.Errorf("Decoded height = %d, want %d", bounds.Dy(), tt.height)
			}

			// Verify JPEG header (starts with FF D8 FF)
			if data[0] != 0xFF || data[1] != 0xD8 || data[2] != 0xFF {
				t.Error("Generated data does not have valid JPEG header")
			}
		})
	}
}

// TestImageToJPEGBytesNilImage tests handling of nil image
func TestImageToJPEGBytesNilImage(t *testing.T) {
	// Note: This will panic because the JPEG encoder doesn't handle nil images
	// We test that the panic occurs as expected
	defer func() {
		if r := recover(); r == nil {
			t.Error("imageToJPEGBytes(nil) expected panic, got nil")
		}
	}()

	imageToJPEGBytes(nil)
}

// TestImageToJPEGBytesEmptyImage tests handling of empty/zero-size image
func TestImageToJPEGBytesEmptyImage(t *testing.T) {
	// Create an image with 0x0 dimensions
	img := image.NewNRGBA(image.Rect(0, 0, 0, 0))

	data, err := imageToJPEGBytes(img)
	// This should either error or produce minimal JPEG data
	// The behavior depends on the JPEG encoder implementation
	if err == nil && len(data) == 0 {
		t.Error("imageToJPEGBytes(empty) produced empty data without error")
	}
}

// TestImageToJPEGBytesQuality tests JPEG quality
func TestImageToJPEGBytesQuality(t *testing.T) {
	img := createTestImage(100, 100, color.NRGBA{R: 255, G: 128, B: 64, A: 255})

	data, err := imageToJPEGBytes(img)
	if err != nil {
		t.Fatalf("imageToJPEGBytes() error = %v", err)
	}

	// JPEG should compress reasonably
	// A 100x100 RGB image is 30,000 bytes uncompressed
	// With quality 90, it should be significantly smaller
	if len(data) > 30000 {
		t.Errorf("JPEG data size = %d bytes, expected < 30000 (compression not working)", len(data))
	}

	if len(data) < 100 {
		t.Errorf("JPEG data size = %d bytes, suspiciously small", len(data))
	}
}

// TestParseJSON tests JSON parsing
func TestParseJSON(t *testing.T) {
	type testStruct struct {
		Name  string  `json:"name"`
		Value int     `json:"value"`
		Score float64 `json:"score"`
	}

	tests := []struct {
		name    string
		data    []byte
		want    testStruct
		wantErr bool
	}{
		{
			name: "valid JSON",
			data: []byte(`{"name":"test","value":42,"score":0.95}`),
			want: testStruct{Name: "test", Value: 42, Score: 0.95},
			wantErr: false,
		},
		{
			name: "valid JSON with spaces",
			data: []byte(`{  "name" : "test" , "value" : 42 , "score" : 0.95  }`),
			want: testStruct{Name: "test", Value: 42, Score: 0.95},
			wantErr: false,
		},
		{
			name: "partial JSON (missing fields)",
			data: []byte(`{"name":"test"}`),
			want: testStruct{Name: "test", Value: 0, Score: 0},
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			data:    []byte(`{"name":"test",`),
			want:    testStruct{},
			wantErr: true,
		},
		{
			name:    "empty JSON",
			data:    []byte(`{}`),
			want:    testStruct{},
			wantErr: false,
		},
		{
			name:    "not JSON",
			data:    []byte(`this is not json`),
			want:    testStruct{},
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			want:    testStruct{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result testStruct
			err := parseJSON(tt.data, &result)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.want {
				t.Errorf("parseJSON() = %+v, want %+v", result, tt.want)
			}
		})
	}
}

// TestParseJSONArray tests parsing JSON arrays
func TestParseJSONArray(t *testing.T) {
	data := []byte(`["apple","banana","cherry"]`)

	var result []string
	err := parseJSON(data, &result)
	if err != nil {
		t.Fatalf("parseJSON() error = %v", err)
	}

	expected := []string{"apple", "banana", "cherry"}
	if len(result) != len(expected) {
		t.Errorf("parseJSON() result length = %d, want %d", len(result), len(expected))
	}

	for i, v := range result {
		if v != expected[i] {
			t.Errorf("parseJSON() result[%d] = %q, want %q", i, v, expected[i])
		}
	}
}

// TestParseJSONNested tests parsing nested JSON structures
func TestParseJSONNested(t *testing.T) {
	type innerStruct struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	type outerStruct struct {
		Title string      `json:"title"`
		Inner innerStruct `json:"inner"`
		Tags  []string    `json:"tags"`
	}

	data := []byte(`{
		"title":"Test",
		"inner":{"id":1,"name":"Inner"},
		"tags":["a","b","c"]
	}`)

	var result outerStruct
	err := parseJSON(data, &result)
	if err != nil {
		t.Fatalf("parseJSON() error = %v", err)
	}

	if result.Title != "Test" {
		t.Errorf("parseJSON() Title = %q, want %q", result.Title, "Test")
	}

	if result.Inner.ID != 1 {
		t.Errorf("parseJSON() Inner.ID = %d, want %d", result.Inner.ID, 1)
	}

	if result.Inner.Name != "Inner" {
		t.Errorf("parseJSON() Inner.Name = %q, want %q", result.Inner.Name, "Inner")
	}

	if len(result.Tags) != 3 {
		t.Errorf("parseJSON() Tags length = %d, want %d", len(result.Tags), 3)
	}
}

// TestParseJSONNilPointer tests parsing into nil pointer
func TestParseJSONNilPointer(t *testing.T) {
	data := []byte(`{"name":"test"}`)

	err := parseJSON(data, nil)
	if err == nil {
		t.Error("parseJSON(data, nil) expected error, got nil")
	}
}

// Helper function to create test images
func createTestImage(width, height int, fill color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, fill)
		}
	}

	return img
}

// TestCreateTestImage verifies our test helper works correctly
func TestCreateTestImage(t *testing.T) {
	img := createTestImage(50, 50, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	if img == nil {
		t.Fatal("createTestImage() returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 50 {
		t.Errorf("Image width = %d, want 50", bounds.Dx())
	}

	if bounds.Dy() != 50 {
		t.Errorf("Image height = %d, want 50", bounds.Dy())
	}

	// Check that the first pixel has the correct color
	c := img.NRGBAAt(0, 0)
	expected := color.NRGBA{R: 255, G: 0, B: 0, A: 255}
	if c != expected {
		t.Errorf("Pixel color = %+v, want %+v", c, expected)
	}
}
