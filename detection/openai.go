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
	}, nil
}

// Name returns the provider name
func (o *OpenAIProvider) Name() string {
	return "openai"
}

// IsConfigured checks if the provider is properly configured
func (o *OpenAIProvider) IsConfigured() bool {
	return o.client != nil
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
	return buildDetectionPrompt(opts)
}

// parseResponse parses OpenAI API response into DetectionResult
func (o *OpenAIProvider) parseResponse(resp *openai.ChatCompletion, opts *DetectOptions) (*DetectionResult, error) {
	empty := &DetectionResult{
		Labels:     []Label{},
		Text:       []TextBlock{},
		Properties: make(map[string]string),
	}

	if len(resp.Choices) == 0 {
		return empty, fmt.Errorf("empty response from API")
	}

	responseText := strings.TrimSpace(resp.Choices[0].Message.Content)
	return parseTextResponse(responseText, opts), nil
}

// parseJSONResponse attempts to parse response as JSON
func (o *OpenAIProvider) parseJSONResponse(text string, result *DetectionResult) error {
	return parseJSONDetectionResponse(text, result)
}

// extractLabelsFromText extracts labels from natural language response
func (o *OpenAIProvider) extractLabelsFromText(text string, opts *DetectOptions) []Label {
	return extractLabelsFromPlainText(text, opts)
}

// Close closes the OpenAI client (no-op)
func (o *OpenAIProvider) Close() error {
	// OpenAI client doesn't require explicit cleanup
	return nil
}
