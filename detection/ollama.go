package detection

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultOllamaHost  = "http://127.0.0.1:11434"
	defaultOllamaModel = "gemma3"
)

// OllamaProvider implements the Provider interface for local Ollama models
type OllamaProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

type ollamaGenerateRequest struct {
	Model  string   `json:"model"`
	Prompt string   `json:"prompt"`
	Images []string `json:"images"`
	Format string   `json:"format"`
	Stream bool     `json:"stream"`
}

type ollamaGenerateResponse struct {
	Model         string `json:"model"`
	CreatedAt     string `json:"created_at"`
	Response      string `json:"response"`
	Error         string `json:"error"`
	Done          bool   `json:"done"`
	TotalDuration int64  `json:"total_duration"`
}

// NewOllamaProvider creates a new Ollama provider instance
func NewOllamaProvider() (*OllamaProvider, error) {
	host := strings.TrimSpace(os.Getenv("IMGX_OLLAMA_HOST"))
	if host == "" {
		host = strings.TrimSpace(os.Getenv("OLLAMA_HOST"))
	}
	if host == "" {
		host = defaultOllamaHost
	}
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	host = strings.TrimRight(host, "/")

	model := strings.TrimSpace(os.Getenv("IMGX_OLLAMA_MODEL"))
	if model == "" {
		model = defaultOllamaModel
	}

	timeoutSeconds := GetTimeout()
	if timeoutSeconds <= 0 {
		timeoutSeconds = 30
	}

	return &OllamaProvider{
		endpoint: host,
		model:    model,
		client: &http.Client{
			Timeout: time.Duration(timeoutSeconds) * time.Second,
		},
	}, nil
}

// Name returns the provider name
func (o *OllamaProvider) Name() string {
	return "ollama"
}

// IsConfigured checks if the provider is properly configured
func (o *OllamaProvider) IsConfigured() bool {
	return o.endpoint != "" && o.model != ""
}

// Detect performs detection using the configured Ollama model
func (o *OllamaProvider) Detect(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error) {
	if opts == nil {
		opts = DefaultDetectOptions()
	}

	startTime := time.Now()

	imgBytes, err := imageToJPEGBytes(img)
	if err != nil {
		return nil, NewDetectionError("ollama", "failed to encode image", err)
	}

	encoded := base64.StdEncoding.EncodeToString(imgBytes)
	prompt := o.buildPrompt(opts)

	reqBody := &ollamaGenerateRequest{
		Model:  o.model,
		Prompt: prompt,
		Images: []string{encoded},
		Format: "json",
		Stream: false,
	}

	endpoint := o.endpoint + "/api/generate"

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, NewDetectionError("ollama", "failed to marshal request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, NewDetectionError("ollama", "failed to create request", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, NewDetectionError("ollama", "API request failed", err)
	}
	defer resp.Body.Close()

	const maxResponseSize = 10 << 20 // 10 MB
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize))
	if err != nil {
		return nil, NewDetectionError("ollama", "failed to read response", err)
	}

	if resp.StatusCode != http.StatusOK {
		message := strings.TrimSpace(string(body))
		if message == "" {
			message = resp.Status
		}
		return nil, NewDetectionError("ollama", fmt.Sprintf("API returned %s: %s", resp.Status, message), nil)
	}

	var parsed ollamaGenerateResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return nil, NewDetectionError("ollama", "failed to decode response", err)
	}

	if parsed.Error != "" {
		return nil, NewDetectionError("ollama", parsed.Error, nil)
	}

	responseText := strings.TrimSpace(parsed.Response)

	result, err := o.parseResponse(responseText, opts)
	if err != nil {
		return nil, NewDetectionError("ollama", "failed to parse response", err)
	}

	result.Provider = o.Name()
	result.ProcessedAt = startTime

	if result.Properties == nil {
		result.Properties = make(map[string]string)
	}
	result.Properties["model"] = o.model

	return result, nil
}

// buildPrompt constructs the prompt based on detection options
func (o *OllamaProvider) buildPrompt(opts *DetectOptions) string {
	return buildDetectionPrompt(opts)
}

// parseResponse parses the Ollama response text into a DetectionResult
func (o *OllamaProvider) parseResponse(responseText string, opts *DetectOptions) (*DetectionResult, error) {
	return parseTextResponse(responseText, opts), nil
}

// parseJSONResponse attempts to parse response as JSON
func (o *OllamaProvider) parseJSONResponse(text string, result *DetectionResult) error {
	return parseJSONDetectionResponse(text, result)
}

// extractLabelsFromText extracts plausible labels from plain text responses
func (o *OllamaProvider) extractLabelsFromText(text string, opts *DetectOptions) []Label {
	return extractLabelsFromPlainText(text, opts)
}
