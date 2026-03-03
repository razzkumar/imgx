package detection

import (
	"context"
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// GeminiProvider implements the Provider interface for Google Gemini API
type GeminiProvider struct {
	client *genai.Client
}

// NewGeminiProvider creates a new Gemini provider instance
func NewGeminiProvider() (*GeminiProvider, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("%w: GEMINI_API_KEY environment variable not set", ErrProviderNotConfigured)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, NewDetectionError("gemini", "failed to create client", err)
	}

	return &GeminiProvider{
		client: client,
	}, nil
}

// Name returns the provider name
func (g *GeminiProvider) Name() string {
	return "gemini"
}

// IsConfigured checks if the provider is properly configured
func (g *GeminiProvider) IsConfigured() bool {
	return g.client != nil
}

// Detect performs object detection using Gemini API
func (g *GeminiProvider) Detect(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error) {
	if opts == nil {
		opts = DefaultDetectOptions()
	}

	startTime := time.Now()

	// Convert image to JPEG bytes
	imgBytes, err := imageToJPEGBytes(img)
	if err != nil {
		return nil, NewDetectionError("gemini", "failed to encode image", err)
	}

	// Build prompt based on features or custom prompt
	prompt := g.buildPrompt(opts)

	// Create content parts (text + image)
	parts := []*genai.Part{
		{Text: prompt},
		{InlineData: &genai.Blob{
			Data:     imgBytes,
			MIMEType: "image/jpeg",
		}},
	}

	contents := []*genai.Content{{Parts: parts}}

	// Configure for JSON response if detecting labels
	var config *genai.GenerateContentConfig
	if opts.CustomPrompt == "" && containsFeature(opts.Features, FeatureLabels) {
		config = &genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
		}
	}

	// Generate content using gemini-2.0-flash model
	resp, err := g.client.Models.GenerateContent(ctx, "gemini-2.0-flash", contents, config)
	if err != nil {
		return nil, NewDetectionError("gemini", "API request failed", err)
	}

	// Parse response
	result, err := g.parseResponse(resp, opts)
	if err != nil {
		return nil, NewDetectionError("gemini", "failed to parse response", err)
	}

	result.Provider = "gemini"
	result.ProcessedAt = startTime

	return result, nil
}

// buildPrompt constructs the prompt based on detection options
func (g *GeminiProvider) buildPrompt(opts *DetectOptions) string {
	return buildDetectionPrompt(opts)
}

// parseResponse parses Gemini API response into DetectionResult
func (g *GeminiProvider) parseResponse(resp *genai.GenerateContentResponse, opts *DetectOptions) (*DetectionResult, error) {
	empty := &DetectionResult{
		Labels:     []Label{},
		Text:       []TextBlock{},
		Properties: make(map[string]string),
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return empty, fmt.Errorf("empty response from API")
	}

	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return empty, fmt.Errorf("no content in response")
	}

	var fullText strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			fullText.WriteString(part.Text)
		}
	}

	responseText := strings.TrimSpace(fullText.String())
	return parseTextResponse(responseText, opts), nil
}

// parseJSONResponse attempts to parse response as JSON
func (g *GeminiProvider) parseJSONResponse(text string, result *DetectionResult) error {
	return parseJSONDetectionResponse(text, result)
}

// extractLabelsFromText extracts labels from natural language response
func (g *GeminiProvider) extractLabelsFromText(text string, opts *DetectOptions) []Label {
	return extractLabelsFromPlainText(text, opts)
}

// Close closes the Gemini client (no-op as genai.Client doesn't require explicit closing)
func (g *GeminiProvider) Close() error {
	// genai.Client doesn't have a Close method
	return nil
}

// Helper functions

// containsFeature checks if a feature is in the list
func containsFeature(features []Feature, target Feature) bool {
	for _, f := range features {
		if f == target {
			return true
		}
	}
	return false
}
