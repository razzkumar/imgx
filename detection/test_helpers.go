package detection

import (
	"context"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	NameFunc         func() string
	IsConfiguredFunc func() bool
	DetectFunc       func(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error)
}

// Name returns the provider name
func (m *MockProvider) Name() string {
	if m.NameFunc != nil {
		return m.NameFunc()
	}
	return "mock"
}

// IsConfigured returns whether the provider is configured
func (m *MockProvider) IsConfigured() bool {
	if m.IsConfiguredFunc != nil {
		return m.IsConfiguredFunc()
	}
	return true
}

// Detect performs detection
func (m *MockProvider) Detect(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error) {
	if m.DetectFunc != nil {
		return m.DetectFunc(ctx, img, opts)
	}
	return &DetectionResult{
		Provider:    m.Name(),
		Labels:      []Label{{Name: "test", Confidence: 0.9}},
		Confidence:  0.9,
		ProcessedAt: time.Now(),
	}, nil
}

// CreateTestImage creates a solid color test image
func CreateTestImage(width, height int, c color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, c)
		}
	}
	return img
}

// CreateTestImageWithPattern creates a test image with a pattern
func CreateTestImageWithPattern(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Create a gradient pattern
			r := uint8((x * 255) / width)
			g := uint8((y * 255) / height)
			b := uint8(((x + y) * 255) / (width + height))
			img.SetNRGBA(x, y, color.NRGBA{R: r, G: g, B: b, A: 255})
		}
	}
	return img
}

// CreateTestImageWithText creates a simple image with text-like patterns
func CreateTestImageWithText(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Fill with white background
	white := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetNRGBA(x, y, white)
		}
	}

	// Draw some black horizontal lines to simulate text
	black := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	for i := 0; i < 5; i++ {
		y := (i + 1) * height / 6
		for x := 10; x < width-10; x++ {
			if y < height {
				img.SetNRGBA(x, y, black)
				if y+1 < height {
					img.SetNRGBA(x, y+1, black)
				}
			}
		}
	}

	return img
}

// LoadFixtureResponse loads a JSON fixture from testdata/responses
func LoadFixtureResponse(t *testing.T, filename string) []byte {
	t.Helper()

	path := filepath.Join("testdata", "responses", filename)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to load fixture %s: %v", filename, err)
	}

	return data
}

// AssertLabel validates a Label structure
func AssertLabel(t *testing.T, label Label) {
	t.Helper()

	if label.Name == "" {
		t.Error("Label.Name is empty")
	}

	if label.Confidence < 0 || label.Confidence > 1 {
		t.Errorf("Label.Confidence = %f, want 0.0-1.0", label.Confidence)
	}
}

// AssertTextBlock validates a TextBlock structure
func AssertTextBlock(t *testing.T, text TextBlock) {
	t.Helper()

	if text.Text == "" {
		t.Error("TextBlock.Text is empty")
	}

	if text.Confidence < 0 || text.Confidence > 1 {
		t.Errorf("TextBlock.Confidence = %f, want 0.0-1.0", text.Confidence)
	}
}

// AssertFace validates a Face structure
func AssertFace(t *testing.T, face Face) {
	t.Helper()

	if face.Confidence < 0 || face.Confidence > 1 {
		t.Errorf("Face.Confidence = %f, want 0.0-1.0", face.Confidence)
	}
}

// AssertDetectionResult validates a DetectionResult structure
func AssertDetectionResult(t *testing.T, result *DetectionResult) {
	t.Helper()

	if result == nil {
		t.Fatal("DetectionResult is nil")
	}

	if result.Provider == "" {
		t.Error("DetectionResult.Provider is empty")
	}

	if result.Confidence < 0 || result.Confidence > 1 {
		t.Errorf("DetectionResult.Confidence = %f, want 0.0-1.0", result.Confidence)
	}

	if result.ProcessedAt.IsZero() {
		t.Error("DetectionResult.ProcessedAt is zero")
	}

	// Validate all labels
	for i, label := range result.Labels {
		if label.Name == "" {
			t.Errorf("Label[%d].Name is empty", i)
		}
		if label.Confidence < 0 || label.Confidence > 1 {
			t.Errorf("Label[%d].Confidence = %f, want 0.0-1.0", i, label.Confidence)
		}
	}

	// Validate all text blocks
	for i, text := range result.Text {
		if text.Text == "" {
			t.Errorf("Text[%d].Text is empty", i)
		}
		if text.Confidence < 0 || text.Confidence > 1 {
			t.Errorf("Text[%d].Confidence = %f, want 0.0-1.0", i, text.Confidence)
		}
	}

	// Validate all faces
	for i, face := range result.Faces {
		if face.Confidence < 0 || face.Confidence > 1 {
			t.Errorf("Face[%d].Confidence = %f, want 0.0-1.0", i, face.Confidence)
		}
	}
}

// AssertValidConfidence checks if a confidence score is in valid range
func AssertValidConfidence(t *testing.T, confidence float32, name string) {
	t.Helper()

	if confidence < 0 || confidence > 1 {
		t.Errorf("%s confidence = %f, want 0.0-1.0", name, confidence)
	}
}

// AssertBox validates a Box structure
func AssertBox(t *testing.T, box *Box) {
	t.Helper()

	if box == nil {
		t.Fatal("Box is nil")
	}

	if box.Width < 0 {
		t.Errorf("Box.Width = %f, want >= 0", box.Width)
	}

	if box.Height < 0 {
		t.Errorf("Box.Height = %f, want >= 0", box.Height)
	}
}

// CreateMockDetectionResult creates a mock DetectionResult for testing
func CreateMockDetectionResult(provider string) *DetectionResult {
	return &DetectionResult{
		Provider: provider,
		Labels: []Label{
			{Name: "dog", Confidence: 0.95},
			{Name: "pet", Confidence: 0.88},
		},
		Description: "A dog in the image",
		Text: []TextBlock{
			{Text: "Hello World", Confidence: 0.99},
		},
		Faces: []Face{
			{Confidence: 0.92, Gender: "Male", AgeRange: "25-35"},
		},
		Properties: map[string]string{
			"brightness": "85.5",
			"contrast":   "72.3",
		},
		Confidence:  0.91,
		ProcessedAt: time.Now(),
	}
}

// SkipIfNoCredentials skips a test if the specified provider credentials are not available
func SkipIfNoCredentials(t *testing.T, provider string) {
	t.Helper()

	switch provider {
	case "gemini":
		if os.Getenv("GEMINI_API_KEY") == "" {
			t.Skip("Skipping test: GEMINI_API_KEY not set")
		}
	case "openai":
		if os.Getenv("OPENAI_API_KEY") == "" {
			t.Skip("Skipping test: OPENAI_API_KEY not set")
		}
	case "aws":
		// AWS uses credential chain, harder to check directly
		// Provider constructor will fail if not configured
	}
}

// AssertNoError is a helper to check for no error
func AssertNoError(t *testing.T, err error, msg string) {
	t.Helper()
	if err != nil {
		t.Fatalf("%s: %v", msg, err)
	}
}

// AssertError is a helper to check for an error
func AssertError(t *testing.T, err error, msg string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: expected error, got nil", msg)
	}
}

// AssertEqual checks if two values are equal
func AssertEqual(t *testing.T, got, want interface{}, msg string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", msg, got, want)
	}
}

// AssertNotNil checks if a value is not nil
func AssertNotNil(t *testing.T, value interface{}, msg string) {
	t.Helper()
	if value == nil {
		t.Errorf("%s: value is nil", msg)
	}
}

// AssertContextDeadline checks that an operation respects context deadlines
func AssertContextDeadline(t *testing.T, fn func(context.Context) error, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		if err != nil && err != context.DeadlineExceeded {
			t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
		}
	case <-time.After(timeout + 100*time.Millisecond):
		t.Error("Function did not respect context deadline")
	}
}
