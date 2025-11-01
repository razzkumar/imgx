package detection

import (
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"os"
	"strings"
	"time"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// OpenAIProvider implements the Provider interface for OpenAI Vision
type OpenAIProvider struct {
	client *openai.Client
	apiKey string
}

// NewOpenAIProvider creates a new OpenAI Vision provider instance
func NewOpenAIProvider() (*OpenAIProvider, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("%w: OPENAI_API_KEY environment variable not set", ErrProviderNotConfigured)
	}

	client := openai.NewClient(
		option.WithAPIKey(apiKey),
	)

	return &OpenAIProvider{
		client: &client,
		apiKey: apiKey,
	}, nil
}

// Name returns the provider name
func (o *OpenAIProvider) Name() string {
	return "openai"
}

// IsConfigured checks if the provider is properly configured
func (o *OpenAIProvider) IsConfigured() bool {
	return o.apiKey != ""
}

// Detect performs object detection using OpenAI Vision
func (o *OpenAIProvider) Detect(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error) {
	if opts == nil {
		opts = DefaultDetectOptions()
	}

	startTime := time.Now()

	// Convert image to JPEG bytes
	imgBytes, err := imageToJPEGBytes(img)
	if err != nil {
		return nil, NewDetectionError("openai", "failed to encode image", err)
	}

	// Encode to base64
	base64Image := base64.StdEncoding.EncodeToString(imgBytes)

	// Build prompt based on features or use custom prompt
	prompt := o.buildPrompt(opts)

	// Create chat completion request with vision
	chatCompletion, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage([]openai.ChatCompletionContentPartUnionParam{
				openai.TextContentPart(prompt),
				openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL:    fmt.Sprintf("data:image/jpeg;base64,%s", base64Image),
					Detail: "auto",
				}),
			}),
		},
		Model:     openai.ChatModelGPT4o,
		MaxTokens: openai.Int(500),
	})

	if err != nil {
		return nil, NewDetectionError("openai", "API request failed", err)
	}

	// Parse response
	result, err := o.parseResponse(chatCompletion, opts)
	if err != nil {
		return nil, NewDetectionError("openai", "failed to parse response", err)
	}

	result.Provider = "openai"
	result.ProcessedAt = startTime

	return result, nil
}

// buildPrompt constructs the prompt based on detection options
func (o *OpenAIProvider) buildPrompt(opts *DetectOptions) string {
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
				"Identify all objects in this image and provide labels with confidence scores (0.0-1.0). "+
					"Return JSON: {\"labels\": [{\"name\": \"object\", \"confidence\": 0.95}]}. "+
					"Return at most %d labels with confidence >= %.2f.",
				opts.MaxResults, opts.MinConfidence,
			))

		case FeatureDescription:
			prompts = append(prompts, "Provide a detailed description of this image.")

		case FeatureText:
			prompts = append(prompts, "Extract all visible text from this image. "+
				"Return JSON: {\"text\": [{\"text\": \"extracted text\", \"confidence\": 0.95}]}")

		case FeatureFaces:
			prompts = append(prompts, "Detect any faces and describe their count, expressions, and emotions.")

		case FeatureProperties:
			prompts = append(prompts, "Analyze image properties: dominant colors, lighting, mood, style.")

		case FeatureLandmarks:
			prompts = append(prompts, "Identify any landmarks, monuments, or famous locations.")

		case FeatureSafeSearch:
			prompts = append(prompts, "Analyze if the image contains any inappropriate content.")
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

// parseResponse parses OpenAI API response into DetectionResult
func (o *OpenAIProvider) parseResponse(resp *openai.ChatCompletion, opts *DetectOptions) (*DetectionResult, error) {
	result := &DetectionResult{
		Labels:     []Label{},
		Text:       []TextBlock{},
		Properties: make(map[string]string),
	}

	if len(resp.Choices) == 0 {
		return result, fmt.Errorf("empty response from API")
	}

	responseText := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Store raw response if requested
	if opts.IncludeRawResponse {
		result.RawResponse = responseText
	}

	// Try to parse as JSON first
	if err := o.parseJSONResponse(responseText, result); err == nil {
		// Successfully parsed as JSON
		return result, nil
	}

	// Fallback: treat as natural language description
	result.Description = responseText

	// Try to extract labels from natural language
	labels := o.extractLabelsFromText(responseText, opts)
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
func (o *OpenAIProvider) parseJSONResponse(text string, result *DetectionResult) error {
	// Try to extract JSON from markdown code blocks
	text = extractJSONFromMarkdown(text)

	// Try to parse JSON
	var parsed map[string]interface{}
	if err := parseJSON([]byte(text), &parsed); err != nil {
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
func (o *OpenAIProvider) extractLabelsFromText(text string, opts *DetectOptions) []Label {
	// Simple heuristic-based extraction
	labels := []Label{}

	// Look for common object patterns
	words := strings.Fields(strings.ToLower(text))
	commonObjects := []string{"person", "people", "dog", "cat", "car", "building", "tree", "flower", "animal", "vehicle", "house", "plant"}

	seen := make(map[string]bool)
	for _, word := range words {
		for _, obj := range commonObjects {
			if strings.Contains(word, obj) && !seen[obj] {
				labels = append(labels, Label{
					Name:       obj,
					Confidence: 0.7, // Default confidence for extracted labels
				})
				seen[obj] = true
				break
			}
		}
		if len(labels) >= opts.MaxResults {
			break
		}
	}

	return labels
}

// Close closes the OpenAI client (no-op)
func (o *OpenAIProvider) Close() error {
	// OpenAI client doesn't require explicit cleanup
	return nil
}
