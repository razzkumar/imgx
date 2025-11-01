package detection

import (
	"context"
	"fmt"
	"image"
	"strings"
	"time"
)

// DetectionResult contains all detection results from a provider
type DetectionResult struct {
	Provider      string            `json:"provider"`                 // Provider name
	Labels        []Label           `json:"labels,omitempty"`         // Detected objects/labels
	Description   string            `json:"description,omitempty"`    // Natural language description
	Text          []TextBlock       `json:"text,omitempty"`           // OCR results
	Faces         []Face            `json:"faces,omitempty"`          // Face detection
	Web           *WebDetection     `json:"web,omitempty"`            // Web search results
	BoundingBoxes []BoundingBox     `json:"bounding_boxes,omitempty"` // Object locations
	Properties    map[string]string `json:"properties,omitempty"`     // Provider-specific data
	Confidence    float32           `json:"confidence"`               // Overall confidence 0.0-1.0
	Error         string            `json:"error,omitempty"`          // Error message if detection failed
	RawResponse   string            `json:"raw_response,omitempty"`   // Raw API response for debugging
	ProcessedAt   time.Time         `json:"processed_at"`             // When detection ran
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
	Text        string  `json:"text"`                  // Detected text
	Confidence  float32 `json:"confidence"`            // Confidence score
	Language    string  `json:"language,omitempty"`    // Detected language
	BoundingBox *Box    `json:"bounding_box,omitempty"` // Text location
	Type        string  `json:"type,omitempty"`        // TEXT, LINE, WORD, etc.
}

// Face represents a detected face
type Face struct {
	Confidence         float32 `json:"confidence"`                    // Detection confidence
	BoundingBox        *Box    `json:"bounding_box,omitempty"`        // Face location
	JoyLikelihood      string  `json:"joy_likelihood,omitempty"`      // Emotion: joy
	SorrowLikelihood   string  `json:"sorrow_likelihood,omitempty"`   // Emotion: sorrow
	AngerLikelihood    string  `json:"anger_likelihood,omitempty"`    // Emotion: anger
	SurpriseLikelihood string  `json:"surprise_likelihood,omitempty"` // Emotion: surprise
	Gender             string  `json:"gender,omitempty"`              // Male/Female
	AgeRange           string  `json:"age_range,omitempty"`           // Age range estimate
	Landmarks          []Landmark `json:"landmarks,omitempty"`        // Facial landmarks
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
	URL              string `json:"url"`
	Score            float32 `json:"score,omitempty"`
	PageTitle        string `json:"page_title,omitempty"`
	FullMatchingImages []WebImage `json:"full_matching_images,omitempty"`
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
	X      float32 `json:"x"`       // X coordinate (top-left)
	Y      float32 `json:"y"`       // Y coordinate (top-left)
	Width  float32 `json:"width"`   // Box width
	Height float32 `json:"height"`  // Box height
}

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
	case "aws", "rekognition":
		return NewAWSProvider()
	case "openai", "gpt4vision", "gpt-4-vision":
		return NewOpenAIProvider()
	default:
		return nil, fmt.Errorf("unknown provider: %s (valid: gemini, google, aws, openai)", name)
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
