package imgx

import (
	"context"
	"fmt"
	"image"
	"time"

	"github.com/razzkumar/imgx/detection"
)

// Image represents an image with processing metadata
type Image struct {
	data     *image.NRGBA
	metadata *ProcessingMetadata
}

// ProcessingMetadata contains information about image processing operations
type ProcessingMetadata struct {
	SourcePath  string
	Operations  []OperationRecord
	Software    string // Fixed: "imgx"
	Version     string // Fixed: version from load.go
	Author      string // Customizable: artist/creator name
	ProjectURL  string // Fixed: project URL
	AddMetadata bool

	DetectionResult *detection.DetectionResult `json:"detection_result,omitempty"` // Object detection results
}

// OperationRecord represents a single image processing operation
type OperationRecord struct {
	Action     string
	Parameters string
	Timestamp  time.Time
}

// Clone creates a deep copy of ProcessingMetadata
func (m *ProcessingMetadata) Clone() *ProcessingMetadata {
	ops := make([]OperationRecord, len(m.Operations))
	copy(ops, m.Operations)
	return &ProcessingMetadata{
		SourcePath:  m.SourcePath,
		Operations:  ops,
		Software:    m.Software,
		Version:     m.Version,
		Author:      m.Author,
		ProjectURL:  m.ProjectURL,
		AddMetadata: m.AddMetadata,

		DetectionResult: m.DetectionResult, // Shallow copy is fine for DetectionResult
	}
}

// AddOperation adds a new operation record to the metadata
func (m *ProcessingMetadata) AddOperation(action, parameters string) {
	m.Operations = append(m.Operations, OperationRecord{
		Action:     action,
		Parameters: parameters,
		Timestamp:  time.Now(),
	})
}

// ToNRGBA returns the underlying NRGBA image data
func (img *Image) ToNRGBA() *image.NRGBA {
	return img.data
}

// Bounds returns the bounds of the image
func (img *Image) Bounds() image.Rectangle {
	return img.data.Bounds()
}

// GetMetadata returns the processing metadata
func (img *Image) GetMetadata() *ProcessingMetadata {
	return img.metadata
}

// SetAuthor sets the artist/creator name for the image metadata
// This overrides the default author but keeps creator_tool unchanged
func (img *Image) SetAuthor(author string) *Image {
	img.metadata.Author = author
	return img
}

// Detect performs object detection on the image using the specified provider
// and returns detection results.
//
// Supported providers:
//   - "gemini" or "google" - Google Gemini API (requires GEMINI_API_KEY)
//   - "aws" - AWS Rekognition (requires AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
//   - "openai" - OpenAI Vision (requires OPENAI_API_KEY)
//
// Example:
//
//	img, err := imgx.Load("photo.jpg")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	ctx := context.Background()
//	result, err := img.Detect(ctx, "gemini")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Access detection results
//	for _, label := range result.Labels {
//		fmt.Printf("Found: %s (%.2f%% confidence)\n", label.Name, label.Confidence*100)
//	}
//
// With custom options:
//
//	opts := &detection.DetectOptions{
//		Features:      []detection.Feature{detection.FeatureLabels, detection.FeatureText},
//		MaxResults:    15,
//		MinConfidence: 0.7,
//		CustomPrompt:  "Is there a dog in this image?",
//	}
//	result, err := img.Detect(ctx, "gemini", opts)
func (img *Image) Detect(ctx context.Context, provider string, opts ...*detection.DetectOptions) (*detection.DetectionResult, error) {
	// Use default options if none provided
	var opt *detection.DetectOptions
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = detection.DefaultDetectOptions()
	}

	// Resolve provider alias ("google" -> "gemini")
	resolvedProvider := detection.ResolveProviderAlias(provider)

	// Get provider instance via factory
	prov, err := detection.GetProvider(resolvedProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get detection provider: %w", err)
	}

	// Run detection
	result, err := prov.Detect(ctx, img.data, opt)
	if err != nil {
		return nil, fmt.Errorf("detection failed: %w", err)
	}

	// Store detection result in metadata
	img.metadata.DetectionResult = result

	// Add operation record to metadata
	params := fmt.Sprintf("provider=%s, features=%v", provider, opt.Features)
	if opt.CustomPrompt != "" {
		params = fmt.Sprintf("provider=%s, prompt=%q", provider, opt.CustomPrompt)
	}
	img.metadata.AddOperation("detect", params)

	return result, nil
}
