package detection

import (
	"context"
	"fmt"
	"image"
)

// Detect performs object detection on an image using the specified provider
// and returns detection results.
//
// This is the primary entry point for image detection. It accepts a standard
// *image.NRGBA so that callers can use it without importing the root imgx package.
//
// Supported providers:
//   - "ollama" or "gemma3" - Local Ollama server (default, uses gemma3 model)
//   - "gemini" or "google" - Google Gemini API (requires GEMINI_API_KEY)
//   - "aws" - AWS Rekognition (uses AWS credential chain)
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
//	result, err := detection.Detect(ctx, img.ToNRGBA(), "ollama")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	for _, label := range result.Labels {
//		fmt.Printf("Found: %s (%.2f%% confidence)\n", label.Name, label.Confidence*100)
//	}
func Detect(ctx context.Context, img *image.NRGBA, provider string, opts ...*DetectOptions) (*DetectionResult, error) {
	// Use default options if none provided
	var opt *DetectOptions
	if len(opts) > 0 && opts[0] != nil {
		opt = opts[0]
	} else {
		opt = DefaultDetectOptions()
	}

	// Resolve provider alias ("google" -> "gemini")
	resolvedProvider := ResolveProviderAlias(provider)

	// Get provider instance via factory
	prov, err := GetProvider(resolvedProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to get detection provider: %w", err)
	}

	// Run detection
	result, err := prov.Detect(ctx, img, opt)
	if err != nil {
		return nil, fmt.Errorf("detection failed: %w", err)
	}

	return result, nil
}
