package detection

import (
	"context"
	"errors"
	"image"
	"image/color"
	"testing"
	"time"
)

// TestMockProvider tests the mock provider implementation
func TestMockProvider(t *testing.T) {
	t.Run("default behavior", func(t *testing.T) {
		mock := &MockProvider{}

		if mock.Name() != "mock" {
			t.Errorf("Name() = %q, want %q", mock.Name(), "mock")
		}

		if !mock.IsConfigured() {
			t.Error("IsConfigured() = false, want true")
		}

		ctx := context.Background()
		img := CreateTestImage(10, 10, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		result, err := mock.Detect(ctx, img, DefaultDetectOptions())

		if err != nil {
			t.Errorf("Detect() error = %v, want nil", err)
		}

		if result == nil {
			t.Fatal("Detect() returned nil result")
		}

		if result.Provider != "mock" {
			t.Errorf("result.Provider = %q, want %q", result.Provider, "mock")
		}
	})

	t.Run("custom name function", func(t *testing.T) {
		mock := &MockProvider{
			NameFunc: func() string {
				return "custom"
			},
		}

		if mock.Name() != "custom" {
			t.Errorf("Name() = %q, want %q", mock.Name(), "custom")
		}
	})

	t.Run("custom is configured function", func(t *testing.T) {
		mock := &MockProvider{
			IsConfiguredFunc: func() bool {
				return false
			},
		}

		if mock.IsConfigured() {
			t.Error("IsConfigured() = true, want false")
		}
	})

	t.Run("custom detect function", func(t *testing.T) {
		expectedResult := &DetectionResult{
			Provider:   "test",
			Confidence: 0.85,
		}

		mock := &MockProvider{
			DetectFunc: func(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error) {
				return expectedResult, nil
			},
		}

		ctx := context.Background()
		img := CreateTestImage(10, 10, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		result, err := mock.Detect(ctx, img, DefaultDetectOptions())

		if err != nil {
			t.Errorf("Detect() error = %v, want nil", err)
		}

		if result != expectedResult {
			t.Error("Detect() did not return expected result")
		}
	})

	t.Run("detect error", func(t *testing.T) {
		expectedErr := errors.New("test error")

		mock := &MockProvider{
			DetectFunc: func(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error) {
				return nil, expectedErr
			},
		}

		ctx := context.Background()
		img := CreateTestImage(10, 10, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
		result, err := mock.Detect(ctx, img, DefaultDetectOptions())

		if err != expectedErr {
			t.Errorf("Detect() error = %v, want %v", err, expectedErr)
		}

		if result != nil {
			t.Error("Detect() returned non-nil result on error")
		}
	})
}

// TestCreateTestImage tests test image creation
func TestCreateTestImageHelper(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
		color  color.NRGBA
	}{
		{
			name:   "small red image",
			width:  10,
			height: 10,
			color:  color.NRGBA{R: 255, G: 0, B: 0, A: 255},
		},
		{
			name:   "large green image",
			width:  100,
			height: 100,
			color:  color.NRGBA{R: 0, G: 255, B: 0, A: 255},
		},
		{
			name:   "rectangular blue image",
			width:  50,
			height: 100,
			color:  color.NRGBA{R: 0, G: 0, B: 255, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img := CreateTestImage(tt.width, tt.height, tt.color)

			if img == nil {
				t.Fatal("CreateTestImage() returned nil")
			}

			bounds := img.Bounds()
			if bounds.Dx() != tt.width {
				t.Errorf("width = %d, want %d", bounds.Dx(), tt.width)
			}

			if bounds.Dy() != tt.height {
				t.Errorf("height = %d, want %d", bounds.Dy(), tt.height)
			}

			// Check first pixel
			c := img.NRGBAAt(0, 0)
			if c != tt.color {
				t.Errorf("pixel color = %+v, want %+v", c, tt.color)
			}
		})
	}
}

// TestCreateTestImageWithPattern tests pattern image creation
func TestCreateTestImageWithPattern(t *testing.T) {
	img := CreateTestImageWithPattern(100, 100)

	if img == nil {
		t.Fatal("CreateTestImageWithPattern() returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 100 || bounds.Dy() != 100 {
		t.Errorf("image size = %dx%d, want 100x100", bounds.Dx(), bounds.Dy())
	}

	// Verify it's not a solid color (should have gradient)
	topLeft := img.NRGBAAt(0, 0)
	bottomRight := img.NRGBAAt(99, 99)
	if topLeft == bottomRight {
		t.Error("Pattern image appears to be solid color")
	}
}

// TestCreateTestImageWithText tests text image creation
func TestCreateTestImageWithText(t *testing.T) {
	img := CreateTestImageWithText(200, 100)

	if img == nil {
		t.Fatal("CreateTestImageWithText() returned nil")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 200 || bounds.Dy() != 100 {
		t.Errorf("image size = %dx%d, want 200x100", bounds.Dx(), bounds.Dy())
	}

	// Background should be mostly white
	topPixel := img.NRGBAAt(5, 5)
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if topPixel != white {
		t.Errorf("background color = %+v, want white %+v", topPixel, white)
	}
}

// TestLoadFixtureResponse tests fixture loading
func TestLoadFixtureResponse(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
	}{
		{
			name:     "gemini labels fixture",
			filename: "gemini_labels.json",
			wantErr:  false,
		},
		{
			name:     "aws labels fixture",
			filename: "aws_labels.json",
			wantErr:  false,
		},
		{
			name:     "openai labels fixture",
			filename: "openai_labels.json",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantErr {
				// Can't test error case easily with LoadFixtureResponse as it calls t.Fatalf
				t.Skip("Skipping error test case")
			}

			data := LoadFixtureResponse(t, tt.filename)

			if len(data) == 0 {
				t.Error("LoadFixtureResponse() returned empty data")
			}

			// Verify it's valid JSON
			var result map[string]interface{}
			if err := parseJSON(data, &result); err != nil {
				t.Errorf("Fixture is not valid JSON: %v", err)
			}
		})
	}
}

// TestAssertLabel tests label assertion helper
func TestAssertLabel(t *testing.T) {
	t.Run("valid label", func(t *testing.T) {
		label := Label{
			Name:       "dog",
			Confidence: 0.95,
		}

		// Should not fail
		AssertLabel(t, label)
	})

	// Note: Testing failures would require a custom testing.T implementation
	// which is complex. In real usage, these assertions help catch bugs.
}

// TestAssertDetectionResult tests result assertion helper
func TestAssertDetectionResult(t *testing.T) {
	t.Run("valid result", func(t *testing.T) {
		result := CreateMockDetectionResult("test")

		// Should not fail
		AssertDetectionResult(t, result)
	})

	// More test cases would require custom testing.T implementation
}

// TestCreateMockDetectionResult tests mock result creation
func TestCreateMockDetectionResult(t *testing.T) {
	result := CreateMockDetectionResult("test")

	if result == nil {
		t.Fatal("CreateMockDetectionResult() returned nil")
	}

	if result.Provider != "test" {
		t.Errorf("Provider = %q, want %q", result.Provider, "test")
	}

	if len(result.Labels) == 0 {
		t.Error("Labels is empty")
	}

	if len(result.Text) == 0 {
		t.Error("Text is empty")
	}

	if len(result.Faces) == 0 {
		t.Error("Faces is empty")
	}

	if len(result.Properties) == 0 {
		t.Error("Properties is empty")
	}
}

// TestAssertHelpers tests assertion helper functions
func TestAssertHelpers(t *testing.T) {
	t.Run("AssertNoError with no error", func(t *testing.T) {
		AssertNoError(t, nil, "should not fail")
	})

	t.Run("AssertError with error", func(t *testing.T) {
		err := errors.New("test error")
		AssertError(t, err, "should not fail")
	})

	t.Run("AssertEqual with equal values", func(t *testing.T) {
		AssertEqual(t, "test", "test", "should not fail")
	})

	t.Run("AssertNotNil with non-nil value", func(t *testing.T) {
		value := "test"
		AssertNotNil(t, value, "should not fail")
	})

	t.Run("AssertValidConfidence with valid confidence", func(t *testing.T) {
		AssertValidConfidence(t, 0.95, "test")
	})

	t.Run("AssertBox with valid box", func(t *testing.T) {
		box := &Box{X: 10, Y: 20, Width: 100, Height: 50}
		AssertBox(t, box)
	})
}

// TestAssertContextDeadline tests context deadline assertion
func TestAssertContextDeadline(t *testing.T) {
	t.Run("respects deadline", func(t *testing.T) {
		fn := func(ctx context.Context) error {
			select {
			case <-time.After(10 * time.Millisecond):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		// This should complete within the timeout
		AssertContextDeadline(t, fn, 50*time.Millisecond)
	})
}
