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
	"strconv"
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

	body, err := io.ReadAll(resp.Body)
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
	if opts.CustomPrompt != "" {
		return opts.CustomPrompt
	}

	prompts := []string{responseSchemaPrompt}

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
			prompts = append(prompts, "Analyze image properties: dominant colors, lighting, and mood.")
		case FeatureSafeSearch:
			prompts = append(prompts, "Analyze if the image contains any inappropriate content.")
		case FeatureLandmarks:
			prompts = append(prompts, "Identify any landmarks or notable scenes in the image.")
		}
	}

	if len(prompts) == 1 {
		prompts = append(prompts, fmt.Sprintf(
			"Analyze this image and identify all visible objects. "+
				"Return JSON: {\"labels\": [{\"name\": \"object\", \"confidence\": 0.95}], \"description\": \"...\"} "+
				"with up to %d labels having confidence >= %.2f.",
			opts.MaxResults, opts.MinConfidence,
		))
	}

	return strings.Join(prompts, "\n\n")
}

// parseResponse parses the Ollama response text into a DetectionResult
func (o *OllamaProvider) parseResponse(responseText string, opts *DetectOptions) (*DetectionResult, error) {
	result := &DetectionResult{
		Labels:     []Label{},
		Text:       []TextBlock{},
		Properties: make(map[string]string),
	}

	if opts.IncludeRawResponse {
		result.RawResponse = responseText
	}

	if responseText == "" {
		return result, nil
	}

	if err := o.parseJSONResponse(responseText, result); err == nil {
		return result, nil
	}

	result.Description = responseText

	labels := o.extractLabelsFromText(responseText, opts)
	result.Labels = append(result.Labels, labels...)

	if len(result.Labels) > 0 {
		var total float32
		for _, label := range result.Labels {
			total += label.Confidence
		}
		result.Confidence = total / float32(len(result.Labels))
	}

	return result, nil
}

// parseJSONResponse attempts to parse response as JSON
func (o *OllamaProvider) parseJSONResponse(text string, result *DetectionResult) error {
	text = extractJSONFromMarkdown(text)

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return err
	}

	if provider, ok := raw["provider"].(string); ok {
		result.Provider = provider
	}

	if labels, ok := raw["labels"].([]interface{}); ok {
		for _, item := range labels {
			if labelMap, ok := item.(map[string]interface{}); ok {
				label := Label{}
				if name, ok := labelMap["name"].(string); ok {
					label.Name = name
				}
				if confidence, ok := toFloat32(labelMap["confidence"]); ok {
					label.Confidence = confidence
				}
				if score, ok := toFloat32(labelMap["score"]); ok {
					label.Score = score
				}
				if mid, ok := labelMap["mid"].(string); ok {
					label.MID = mid
				}
				if categories, ok := labelMap["categories"].([]interface{}); ok {
					for _, c := range categories {
						if s, ok := c.(string); ok {
							label.Categories = append(label.Categories, s)
						}
					}
				}
				if topicID, ok := labelMap["topic_id"].(string); ok {
					label.TopicID = topicID
				}
				result.Labels = append(result.Labels, label)
			}
		}
	}

	if description, ok := raw["description"].(string); ok {
		result.Description = description
	}

	if textBlocks, ok := raw["text"].([]interface{}); ok {
		for _, item := range textBlocks {
			if textMap, ok := item.(map[string]interface{}); ok {
				block := TextBlock{}
				if textValue, ok := textMap["text"].(string); ok {
					block.Text = textValue
				}
				if confidence, ok := toFloat32(textMap["confidence"]); ok {
					block.Confidence = confidence
				}
				if language, ok := textMap["language"].(string); ok {
					block.Language = language
				}
				result.Text = append(result.Text, block)
			}
		}
	}

	if colors, ok := raw["colors"].([]interface{}); ok {
		for _, item := range colors {
			if colorMap, ok := item.(map[string]interface{}); ok {
				color := ColorInfo{}
				if name, ok := colorMap["name"].(string); ok {
					color.Name = name
				}
				if hex, ok := colorMap["hex"].(string); ok {
					color.Hex = hex
				}
				if rgb, ok := colorMap["rgb"].(string); ok {
					color.RGB = rgb
				}
				if percentage, ok := toFloat32(colorMap["percentage"]); ok {
					color.Percentage = percentage
				}
				result.Colors = append(result.Colors, color)
			}
		}
	}

	if moderation, ok := raw["moderation"].([]interface{}); ok {
		for _, item := range moderation {
			if modMap, ok := item.(map[string]interface{}); ok {
				label := ModerationLabel{}
				if name, ok := modMap["name"].(string); ok {
					label.Name = name
				}
				if parent, ok := modMap["parent"].(string); ok {
					label.Parent = parent
				}
				if confidence, ok := toFloat32(modMap["confidence"]); ok {
					label.Confidence = confidence
				}
				if severity, ok := modMap["severity"].(string); ok {
					label.Severity = severity
				}
				result.Moderation = append(result.Moderation, label)
			}
		}
	}

	if safeSearch, ok := raw["safe_search"].(map[string]interface{}); ok {
		ss := &SafeSearchSummary{}
		if labels, ok := safeSearch["labels"].([]interface{}); ok {
			for _, item := range labels {
				if labelMap, ok := item.(map[string]interface{}); ok {
					label := ModerationLabel{}
					if name, ok := labelMap["name"].(string); ok {
						label.Name = name
					}
					if severity, ok := labelMap["severity"].(string); ok {
						label.Severity = severity
					}
					if confidence, ok := toFloat32(labelMap["confidence"]); ok {
						label.Confidence = confidence
					}
					ss.Labels = append(ss.Labels, label)
				}
			}
		}
		if notes, ok := safeSearch["notes"].(string); ok {
			ss.Notes = notes
		}
		result.SafeSearch = ss
	}

	if props, ok := raw["properties"].(map[string]interface{}); ok {
		for key, value := range props {
			if str, ok := value.(string); ok {
				result.Properties[key] = str
			} else {
				result.Properties[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	if confidence, ok := toFloat32(raw["confidence"]); ok {
		result.Confidence = confidence
	} else if len(result.Labels) > 0 {
		var total float32
		for _, label := range result.Labels {
			total += label.Confidence
		}
		result.Confidence = total / float32(len(result.Labels))
	}

	return nil
}

// extractLabelsFromText extracts plausible labels from plain text responses
func (o *OllamaProvider) extractLabelsFromText(text string, opts *DetectOptions) []Label {
	lines := strings.Split(text, "\n")
	labels := make([]Label, 0, len(lines))
	added := make(map[string]struct{})

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if idx := strings.Index(line, "."); idx != -1 && idx < len(line)-1 {
			line = strings.TrimSpace(line[idx+1:])
		}

		line = strings.TrimPrefix(line, "- ")

		if line == "" {
			continue
		}

		name := line
		confidence := float32(0.7)

		if idx := strings.LastIndex(line, "("); idx != -1 && strings.HasSuffix(line, ")") {
			namePart := strings.TrimSpace(line[:idx])
			confPart := strings.TrimSuffix(line[idx+1:], ")")
			confPart = strings.TrimSuffix(confPart, "%")
			if val, err := strconv.ParseFloat(confPart, 32); err == nil {
				confidence = float32(val) / 100.0
				name = namePart
			}
		}

		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		if _, exists := added[strings.ToLower(name)]; exists {
			continue
		}
		added[strings.ToLower(name)] = struct{}{}

		labels = append(labels, Label{
			Name:       name,
			Confidence: confidence,
		})

		if len(labels) >= opts.MaxResults {
			break
		}
	}

	return labels
}
