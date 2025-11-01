package detection

import (
	"context"
	"encoding/json"
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
	apiKey string
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
		apiKey: apiKey,
	}, nil
}

// Name returns the provider name
func (g *GeminiProvider) Name() string {
	return "gemini"
}

// IsConfigured checks if the provider is properly configured
func (g *GeminiProvider) IsConfigured() bool {
	return g.apiKey != ""
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
	// If custom prompt provided, use it
	if opts.CustomPrompt != "" {
		return opts.CustomPrompt
	}

	// Build prompt based on features
	var prompts []string

	for _, feature := range opts.Features {
		switch feature {
		case FeatureLabels, FeatureObjects:
			prompts = append(prompts, fmt.Sprintf(
				"Identify all objects and provide labels with confidence scores (0.0-1.0). "+
					"Return JSON array: {\"labels\": [{\"name\": \"object\", \"confidence\": 0.95}]}. "+
					"Return at most %d labels with confidence >= %.2f.",
				opts.MaxResults, opts.MinConfidence,
			))

		case FeatureDescription:
			prompts = append(prompts, "Provide a detailed description of this image.")

		case FeatureText:
			prompts = append(prompts, "Extract all visible text from this image. "+
				"Return JSON: {\"text\": [{\"text\": \"extracted text\", \"confidence\": 0.95}]}")

		case FeatureFaces:
			prompts = append(prompts, "Detect any faces and describe their expressions/emotions.")

		case FeatureProperties:
			prompts = append(prompts, "Analyze image properties: dominant colors, brightness, style.")

		case FeatureLandmarks:
			prompts = append(prompts, "Identify any landmarks, monuments, or famous locations.")

		case FeatureSafeSearch:
			prompts = append(prompts, "Analyze if the image contains any adult, violent, or inappropriate content.")
		}
	}

	if len(prompts) == 0 {
		// Default prompt
		return fmt.Sprintf(
			"Analyze this image and identify all visible objects. "+
				"Return JSON: {\"labels\": [{\"name\": \"object\", \"confidence\": 0.95}], \"description\": \"...\"} "+
				"with up to %d labels having confidence >= %.2f.",
			opts.MaxResults, opts.MinConfidence,
		)
	}

	return strings.Join(prompts, "\n\n")
}

// parseResponse parses Gemini API response into DetectionResult
func (g *GeminiProvider) parseResponse(resp *genai.GenerateContentResponse, opts *DetectOptions) (*DetectionResult, error) {
	result := &DetectionResult{
		Labels:     []Label{},
		Text:       []TextBlock{},
		Properties: make(map[string]string),
	}

	if resp == nil || len(resp.Candidates) == 0 {
		return result, fmt.Errorf("empty response from API")
	}

	// Get text content from first candidate
	candidate := resp.Candidates[0]
	if candidate.Content == nil || len(candidate.Content.Parts) == 0 {
		return result, fmt.Errorf("no content in response")
	}

	// Extract text from parts
	var fullText strings.Builder
	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			fullText.WriteString(part.Text)
		}
	}

	responseText := strings.TrimSpace(fullText.String())

	// Store raw response if requested
	if opts.IncludeRawResponse {
		result.RawResponse = responseText
	}

	// Try to parse as JSON first
	if err := g.parseJSONResponse(responseText, result); err == nil {
		// Successfully parsed as JSON
		return result, nil
	}

	// Fallback: parse as natural language response
	result.Description = responseText

	// Extract labels from natural language if possible
	labels := g.extractLabelsFromText(responseText, opts)
	result.Labels = append(result.Labels, labels...)

	// Calculate overall confidence
	if len(result.Labels) > 0 {
		var totalConf float32
		for _, label := range result.Labels {
			totalConf += label.Confidence
		}
		result.Confidence = totalConf / float32(len(result.Labels))
	}

	return result, nil
}

// parseJSONResponse attempts to parse response as JSON
func (g *GeminiProvider) parseJSONResponse(text string, result *DetectionResult) error {
	// Try to extract JSON from markdown code blocks
	text = extractJSONFromMarkdown(text)

	// Define possible JSON structures
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		return err
	}

	// Extract labels
	if labelsData, ok := parsed["labels"].([]interface{}); ok {
		for _, item := range labelsData {
			if labelMap, ok := item.(map[string]interface{}); ok {
				label := Label{}
				if name, ok := labelMap["name"].(string); ok {
					label.Name = name
				}
				if conf, ok := labelMap["confidence"].(float64); ok {
					label.Confidence = float32(conf)
				}
				if label.Name != "" {
					result.Labels = append(result.Labels, label)
				}
			}
		}
	}

	// Extract description
	if desc, ok := parsed["description"].(string); ok {
		result.Description = desc
	}

	// Extract text blocks
	if textData, ok := parsed["text"].([]interface{}); ok {
		for _, item := range textData {
			if textMap, ok := item.(map[string]interface{}); ok {
				block := TextBlock{}
				if text, ok := textMap["text"].(string); ok {
					block.Text = text
				}
				if conf, ok := textMap["confidence"].(float64); ok {
					block.Confidence = float32(conf)
				}
				if block.Text != "" {
					result.Text = append(result.Text, block)
				}
			}
		}
	}

	return nil
}

// extractLabelsFromText extracts labels from natural language response
func (g *GeminiProvider) extractLabelsFromText(text string, opts *DetectOptions) []Label {
	// This is a simple heuristic-based extraction
	labels := []Label{}

	// Look for common patterns like "I see a dog", "This is a cat", etc.
	words := strings.Fields(strings.ToLower(text))
	commonObjects := []string{"dog", "cat", "person", "car", "building", "tree", "flower", "animal", "vehicle"}

	for _, word := range words {
		for _, obj := range commonObjects {
			if strings.Contains(word, obj) {
				labels = append(labels, Label{
					Name:       obj,
					Confidence: 0.7, // Default confidence for extracted labels
				})
				break
			}
		}
		if len(labels) >= opts.MaxResults {
			break
		}
	}

	return labels
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

// extractJSONFromMarkdown extracts JSON from markdown code blocks
func extractJSONFromMarkdown(text string) string {
	// Remove markdown code fences if present
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		if idx := strings.Index(text, "```"); idx != -1 {
			text = text[:idx]
		}
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		if idx := strings.Index(text, "```"); idx != -1 {
			text = text[:idx]
		}
	}
	return strings.TrimSpace(text)
}
