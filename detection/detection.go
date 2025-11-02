package detection

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	"strconv"
	"strings"
	"time"
)

// DetectionResult contains all detection results from a provider
type DetectionResult struct {
	Provider      string             `json:"provider"`                 // Provider name
	Labels        []Label            `json:"labels,omitempty"`         // Detected objects/labels
	Description   string             `json:"description,omitempty"`    // Natural language description
	Text          []TextBlock        `json:"text,omitempty"`           // OCR results
	Faces         []Face             `json:"faces,omitempty"`          // Face detection
	Web           *WebDetection      `json:"web,omitempty"`            // Web search results
	BoundingBoxes []BoundingBox      `json:"bounding_boxes,omitempty"` // Object locations
	Colors        []ColorInfo        `json:"colors,omitempty"`         // Dominant colors and palettes
	ImageQuality  *ImageQuality      `json:"image_quality,omitempty"`  // Brightness/contrast metrics
	Moderation    []ModerationLabel  `json:"moderation,omitempty"`     // Safe-search or moderation labels
	Properties    map[string]string  `json:"properties,omitempty"`     // Provider-specific data
	SafeSearch    *SafeSearchSummary `json:"safe_search,omitempty"`    // Provider-safe-search summary
	Confidence    float32            `json:"confidence"`               // Overall confidence 0.0-1.0
	Error         string             `json:"error,omitempty"`          // Error message if detection failed
	RawResponse   string             `json:"raw_response,omitempty"`   // Raw API response for debugging
	ProcessedAt   time.Time          `json:"processed_at"`             // When detection ran
}

// Label represents a detected object or label
type Label struct {
	Name       string   `json:"name"`                 // Object/label name
	Confidence float32  `json:"confidence"`           // 0.0-1.0 confidence score
	Score      float32  `json:"score,omitempty"`      // Provider-specific score
	MID        string   `json:"mid,omitempty"`        // Machine ID (Google)
	Categories []string `json:"categories,omitempty"` // Category hierarchy
	TopicID    string   `json:"topic_id,omitempty"`   // Topic identifier
}

// TextBlock represents detected text
type TextBlock struct {
	Text        string  `json:"text"`                   // Detected text
	Confidence  float32 `json:"confidence"`             // Confidence score
	Language    string  `json:"language,omitempty"`     // Detected language
	BoundingBox *Box    `json:"bounding_box,omitempty"` // Text location
	Type        string  `json:"type,omitempty"`         // TEXT, LINE, WORD, etc.
}

// Face represents a detected face
type Face struct {
	Confidence         float32    `json:"confidence"`                    // Detection confidence
	BoundingBox        *Box       `json:"bounding_box,omitempty"`        // Face location
	JoyLikelihood      string     `json:"joy_likelihood,omitempty"`      // Emotion: joy
	SorrowLikelihood   string     `json:"sorrow_likelihood,omitempty"`   // Emotion: sorrow
	AngerLikelihood    string     `json:"anger_likelihood,omitempty"`    // Emotion: anger
	SurpriseLikelihood string     `json:"surprise_likelihood,omitempty"` // Emotion: surprise
	Gender             string     `json:"gender,omitempty"`              // Male/Female
	AgeRange           string     `json:"age_range,omitempty"`           // Age range estimate
	Landmarks          []Landmark `json:"landmarks,omitempty"`           // Facial landmarks
}

// Landmark represents a facial landmark point
type Landmark struct {
	Type string  `json:"type"` // LEFT_EYE, RIGHT_EYE, NOSE_TIP, etc.
	X    float32 `json:"x"`
	Y    float32 `json:"y"`
	Z    float32 `json:"z,omitempty"`
}

// WebDetection represents web search results (Google Cloud Vision)
type WebDetection struct {
	WebEntities             []WebEntity `json:"web_entities,omitempty"`               // Similar entities
	FullMatchingImages      []WebImage  `json:"full_matching_images,omitempty"`       // Exact matches
	PartialMatchingImages   []WebImage  `json:"partial_matching_images,omitempty"`    // Partial matches
	PagesWithMatchingImages []WebPage   `json:"pages_with_matching_images,omitempty"` // Pages containing image
	VisuallySimilarImages   []WebImage  `json:"visually_similar_images,omitempty"`    // Similar images
	BestGuessLabels         []string    `json:"best_guess_labels,omitempty"`          // Best guess descriptions
}

// WebEntity represents a web entity found
type WebEntity struct {
	EntityID    string  `json:"entity_id"`
	Score       float32 `json:"score"`
	Description string  `json:"description"`
}

// WebImage represents an image found on the web
type WebImage struct {
	URL   string  `json:"url"`
	Score float32 `json:"score,omitempty"`
}

// WebPage represents a web page containing the image
type WebPage struct {
	URL                   string     `json:"url"`
	Score                 float32    `json:"score,omitempty"`
	PageTitle             string     `json:"page_title,omitempty"`
	FullMatchingImages    []WebImage `json:"full_matching_images,omitempty"`
	PartialMatchingImages []WebImage `json:"partial_matching_images,omitempty"`
}

// BoundingBox represents object location in the image
type BoundingBox struct {
	Label      string  `json:"label"`      // Object label
	Confidence float32 `json:"confidence"` // Detection confidence
	Box        Box     `json:"box"`        // Bounding box coordinates
}

// Box represents rectangular coordinates
type Box struct {
	X      float32 `json:"x"`      // X coordinate (top-left)
	Y      float32 `json:"y"`      // Y coordinate (top-left)
	Width  float32 `json:"width"`  // Box width
	Height float32 `json:"height"` // Box height
}

// ColorInfo describes a dominant color detected in the image
type ColorInfo struct {
	Name       string  `json:"name,omitempty"`       // Human-friendly color name
	Hex        string  `json:"hex,omitempty"`        // Hex value (e.g. #AABBCC)
	RGB        string  `json:"rgb,omitempty"`        // RGB tuple string
	Percentage float32 `json:"percentage,omitempty"` // Coverage percentage (0.0-100.0)
}

// ImageQuality captures brightness/contrast metrics for the image (and segments)
type ImageQuality struct {
	Brightness           float32 `json:"brightness,omitempty"`
	Sharpness            float32 `json:"sharpness,omitempty"`
	Contrast             float32 `json:"contrast,omitempty"`
	ForegroundBrightness float32 `json:"foreground_brightness,omitempty"`
	ForegroundSharpness  float32 `json:"foreground_sharpness,omitempty"`
	ForegroundColor      string  `json:"foreground_color,omitempty"`
	BackgroundBrightness float32 `json:"background_brightness,omitempty"`
	BackgroundSharpness  float32 `json:"background_sharpness,omitempty"`
	BackgroundColor      string  `json:"background_color,omitempty"`
}

// ModerationLabel stores safe-search / moderation flags reported by providers
type ModerationLabel struct {
	Name       string  `json:"name"`                 // Moderation label, e.g. Adult, Violence
	Parent     string  `json:"parent,omitempty"`     // Parent category where available
	Confidence float32 `json:"confidence,omitempty"` // Confidence score 0.0-1.0
	Severity   string  `json:"severity,omitempty"`   // Provider-specific severity/likelihood text
}

// SafeSearchSummary provides a high-level summary of provider ratings
type SafeSearchSummary struct {
	Labels []ModerationLabel `json:"labels,omitempty"`
	Notes  string            `json:"notes,omitempty"`
}

const responseSchemaPrompt = "Format the response strictly as JSON (no markdown fences). Use keys like `labels` (array of {name, confidence}), `description` (string), `text` (array of {text, confidence}), `colors` (array of {name, hex, rgb, percentage}), `image_quality` (object with brightness, sharpness, contrast, foreground_*, background_* fields), `moderation` (array of {name, parent, confidence, severity}), `safe_search` (object with labels array and optional notes), and `properties` (object of additional key/value strings). Omit keys that you cannot populate."

// Provider defines the interface all detection providers must implement
type Provider interface {
	// Detect analyzes an image and returns detection results
	Detect(ctx context.Context, img *image.NRGBA, opts *DetectOptions) (*DetectionResult, error)

	// Name returns the provider name
	Name() string

	// IsConfigured returns true if provider has required credentials
	IsConfigured() bool
}

// DetectOptions contains detection configuration options
type DetectOptions struct {
	// Features specifies what to detect (labels, text, faces, web, etc.)
	Features []Feature `json:"features,omitempty"`

	// MaxResults limits the number of labels to return
	MaxResults int `json:"max_results,omitempty"`

	// MinConfidence sets the minimum confidence threshold (0.0-1.0)
	MinConfidence float32 `json:"min_confidence,omitempty"`

	// CustomPrompt for Gemini/OpenAI to ask custom questions
	CustomPrompt string `json:"custom_prompt,omitempty"`

	// Language hint for text detection
	Language string `json:"language,omitempty"`

	// IncludeRawResponse includes raw API response in result
	IncludeRawResponse bool `json:"include_raw_response,omitempty"`
}

// Feature represents a detection feature type
type Feature string

const (
	// FeatureLabels detects objects and labels
	FeatureLabels Feature = "labels"

	// FeatureText performs OCR text detection
	FeatureText Feature = "text"

	// FeatureFaces detects faces and emotions
	FeatureFaces Feature = "faces"

	// FeatureWeb searches for similar images on the web (Google only)
	FeatureWeb Feature = "web"

	// FeatureDescription generates natural language descriptions
	FeatureDescription Feature = "description"

	// FeatureProperties extracts image properties (colors, etc.)
	FeatureProperties Feature = "properties"

	// FeatureObjects detects objects with bounding boxes
	FeatureObjects Feature = "objects"

	// FeatureLandmarks detects landmarks
	FeatureLandmarks Feature = "landmarks"

	// FeatureLogos detects logos
	FeatureLogos Feature = "logos"

	// FeatureSafeSearch detects adult/violent content
	FeatureSafeSearch Feature = "safesearch"
)

// String returns the string representation of a Feature
func (f Feature) String() string {
	return string(f)
}

// ParseFeatures parses a comma-separated string into Feature slice
func ParseFeatures(s string) []Feature {
	if s == "" {
		return []Feature{FeatureLabels}
	}

	parts := strings.Split(s, ",")
	features := make([]Feature, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(strings.ToLower(part))
		if part != "" {
			features = append(features, Feature(part))
		}
	}
	return features
}

// ResolveProviderAlias resolves provider name aliases
// "google" -> "gemini" (Google's AI Studio API)
func ResolveProviderAlias(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	switch name {
	case "google":
		return "gemini" // Google AI Studio / Gemini API
	case "local", "local-ollama":
		return "ollama"
	case "qwen", "qwen3", "qwen3-vl":
		return "ollama"
	case "gemma3", "gemma-3":
		return "ollama"
	default:
		return name
	}
}

// GetProvider returns a provider instance by name
func GetProvider(name string) (Provider, error) {
	name = strings.ToLower(strings.TrimSpace(name))

	switch name {
	case "gemini":
		return NewGeminiProvider()
	case "ollama":
		return NewOllamaProvider()
	case "aws", "rekognition":
		return NewAWSProvider()
	case "openai", "gpt4vision", "gpt-4-vision":
		return NewOpenAIProvider()
	default:
		return nil, fmt.Errorf("unknown provider: %s (valid: gemini, google, ollama, aws, openai)", name)
	}
}

// DefaultDetectOptions returns default detection options
func DefaultDetectOptions() *DetectOptions {
	return &DetectOptions{
		Features:      []Feature{FeatureLabels},
		MaxResults:    10,
		MinConfidence: 0.5,
	}
}

// --- Shared parsing helpers -------------------------------------------------

func extractJSONFromMarkdown(text string) string {
	// Remove markdown code fences if present
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```json") {
		text = strings.TrimPrefix(text, "```json")
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
		if idx := strings.LastIndex(text, "```"); idx != -1 {
			text = text[:idx]
		}
	}
	return strings.TrimSpace(text)
}

func toFloat32(value interface{}) (float32, bool) {
	switch v := value.(type) {
	case float32:
		return v, true
	case float64:
		return float32(v), true
	case int:
		return float32(v), true
	case int32:
		return float32(v), true
	case int64:
		return float32(v), true
	case string:
		clean := strings.TrimSpace(strings.TrimSuffix(v, "%"))
		if clean == "" {
			return 0, false
		}
		if parsed, err := strconv.ParseFloat(clean, 32); err == nil {
			return float32(parsed), true
		}
	}
	return 0, false
}

func parseImageQualityFromInterface(value interface{}) *ImageQuality {
	data, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}

	quality := &ImageQuality{}
	hasValue := false

	if v, ok := toFloat32(data["brightness"]); ok {
		quality.Brightness = v
		hasValue = true
	}
	if v, ok := toFloat32(data["sharpness"]); ok {
		quality.Sharpness = v
		hasValue = true
	}
	if v, ok := toFloat32(data["contrast"]); ok {
		quality.Contrast = v
		hasValue = true
	}
	if v, ok := toFloat32(data["foreground_brightness"]); ok {
		quality.ForegroundBrightness = v
		hasValue = true
	}
	if v, ok := toFloat32(data["foreground_sharpness"]); ok {
		quality.ForegroundSharpness = v
		hasValue = true
	}
	if s, ok := data["foreground_color"].(string); ok && s != "" {
		quality.ForegroundColor = s
		hasValue = true
	}
	if v, ok := toFloat32(data["background_brightness"]); ok {
		quality.BackgroundBrightness = v
		hasValue = true
	}
	if v, ok := toFloat32(data["background_sharpness"]); ok {
		quality.BackgroundSharpness = v
		hasValue = true
	}
	if s, ok := data["background_color"].(string); ok && s != "" {
		quality.BackgroundColor = s
		hasValue = true
	}

	// Support nested structures
	if fg, ok := data["foreground"].(map[string]interface{}); ok {
		if v, ok := toFloat32(fg["brightness"]); ok {
			quality.ForegroundBrightness = v
			hasValue = true
		}
		if v, ok := toFloat32(fg["sharpness"]); ok {
			quality.ForegroundSharpness = v
			hasValue = true
		}
		if s, ok := fg["color"].(string); ok && s != "" {
			quality.ForegroundColor = s
			hasValue = true
		}
	}

	if bg, ok := data["background"].(map[string]interface{}); ok {
		if v, ok := toFloat32(bg["brightness"]); ok {
			quality.BackgroundBrightness = v
			hasValue = true
		}
		if v, ok := toFloat32(bg["sharpness"]); ok {
			quality.BackgroundSharpness = v
			hasValue = true
		}
		if s, ok := bg["color"].(string); ok && s != "" {
			quality.BackgroundColor = s
			hasValue = true
		}
	}

	if !hasValue {
		return nil
	}
	return quality
}

func parseColorsFromInterface(value interface{}) []ColorInfo {
	rawSlice, ok := value.([]interface{})
	if !ok {
		return nil
	}

	colors := make([]ColorInfo, 0, len(rawSlice))
	for _, item := range rawSlice {
		switch elem := item.(type) {
		case map[string]interface{}:
			color := ColorInfo{}
			if name, ok := elem["name"].(string); ok {
				color.Name = name
			}
			if hex, ok := elem["hex"].(string); ok {
				color.Hex = hex
			}
			if rgb, ok := elem["rgb"].(string); ok {
				color.RGB = rgb
			}
			if pct, ok := toFloat32(elem["percentage"]); ok {
				color.Percentage = pct
			}
			if color.Name != "" || color.Hex != "" || color.RGB != "" || color.Percentage > 0 {
				colors = append(colors, color)
			}
		case string:
			if elem != "" {
				colors = append(colors, ColorInfo{Name: elem})
			}
		}
	}

	if len(colors) == 0 {
		return nil
	}
	return colors
}

func parseModerationFromInterface(value interface{}) []ModerationLabel {
	rawSlice, ok := value.([]interface{})
	if !ok {
		return nil
	}

	moderation := make([]ModerationLabel, 0, len(rawSlice))
	for _, item := range rawSlice {
		if m, ok := item.(map[string]interface{}); ok {
			label := ModerationLabel{}
			if name, ok := m["name"].(string); ok {
				label.Name = name
			}
			if parent, ok := m["parent"].(string); ok {
				label.Parent = parent
			}
			if severity, ok := m["severity"].(string); ok {
				label.Severity = severity
			}
			if conf, ok := toFloat32(m["confidence"]); ok {
				if conf > 1 {
					conf = conf / 100
				}
				label.Confidence = conf
			}
			if label.Name != "" {
				moderation = append(moderation, label)
			}
		}
	}

	if len(moderation) == 0 {
		return nil
	}
	return moderation
}

func parseSafeSearchFromInterface(value interface{}) *SafeSearchSummary {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}
		return &SafeSearchSummary{Notes: v}
	case []interface{}:
		labels := parseModerationFromInterface(v)
		if len(labels) == 0 {
			return nil
		}
		return &SafeSearchSummary{Labels: labels}
	case map[string]interface{}:
		summary := &SafeSearchSummary{}
		if notes, ok := v["notes"].(string); ok && notes != "" {
			summary.Notes = notes
		}
		if labels := parseModerationFromInterface(v["labels"]); len(labels) > 0 {
			summary.Labels = labels
		} else {
			for key, raw := range v {
				if key == "notes" {
					continue
				}
				label := ModerationLabel{Name: key}
				switch typed := raw.(type) {
				case string:
					label.Severity = typed
				default:
					if conf, ok := toFloat32(typed); ok {
						if conf > 1 {
							conf = conf / 100
						}
						label.Confidence = conf
					}
				}
				if label.Name != "" {
					summary.Labels = append(summary.Labels, label)
				}
			}
		}
		if len(summary.Labels) == 0 && summary.Notes == "" {
			return nil
		}
		return summary
	default:
		return nil
	}
}

func parsePropertiesFromInterface(dest map[string]string, value interface{}) map[string]string {
	if dest == nil {
		dest = make(map[string]string)
	}
	data, ok := value.(map[string]interface{})
	if !ok {
		return dest
	}
	for key, raw := range data {
		switch v := raw.(type) {
		case string:
			dest[key] = v
		case float64:
			dest[key] = fmt.Sprintf("%.2f", v)
		case float32:
			dest[key] = fmt.Sprintf("%.2f", v)
		case int:
			dest[key] = fmt.Sprintf("%d", v)
		case int64:
			dest[key] = fmt.Sprintf("%d", v)
		default:
			bytes, err := json.Marshal(v)
			if err == nil {
				dest[key] = string(bytes)
			}
		}
	}
	return dest
}
