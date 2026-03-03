package detection

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestParseFeatures tests feature parsing from comma-separated strings
func TestParseFeatures(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Feature
	}{
		{
			name:     "empty string returns labels",
			input:    "",
			expected: []Feature{FeatureLabels},
		},
		{
			name:     "single feature",
			input:    "text",
			expected: []Feature{FeatureText},
		},
		{
			name:     "multiple features",
			input:    "labels,text,faces",
			expected: []Feature{FeatureLabels, FeatureText, FeatureFaces},
		},
		{
			name:     "with spaces",
			input:    " labels , text , faces ",
			expected: []Feature{FeatureLabels, FeatureText, FeatureFaces},
		},
		{
			name:     "uppercase converted to lowercase",
			input:    "LABELS,TEXT",
			expected: []Feature{FeatureLabels, FeatureText},
		},
		{
			name:     "mixed case",
			input:    "Labels,TeXt,FACES",
			expected: []Feature{FeatureLabels, FeatureText, FeatureFaces},
		},
		{
			name:     "trailing comma ignored",
			input:    "labels,text,",
			expected: []Feature{FeatureLabels, FeatureText},
		},
		{
			name:     "empty parts ignored",
			input:    "labels,,text",
			expected: []Feature{FeatureLabels, FeatureText},
		},
		{
			name:     "only commas returns empty",
			input:    ",,,",
			expected: []Feature{},
		},
		{
			name:  "all feature types",
			input: "labels,text,faces,web,description,properties,objects,landmarks,logos,safesearch",
			expected: []Feature{
				FeatureLabels,
				FeatureText,
				FeatureFaces,
				FeatureWeb,
				FeatureDescription,
				FeatureProperties,
				FeatureObjects,
				FeatureLandmarks,
				FeatureLogos,
				FeatureSafeSearch,
			},
		},
		{
			name:     "spaces only ignored",
			input:    "   ",
			expected: []Feature{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseFeatures(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParseFeatures(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestResolveProviderAlias tests provider name alias resolution
func TestResolveProviderAlias(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "google resolves to gemini",
			input:    "google",
			expected: "gemini",
		},
		{
			name:     "Google uppercase resolves to gemini",
			input:    "Google",
			expected: "gemini",
		},
		{
			name:     "GOOGLE uppercase resolves to gemini",
			input:    "GOOGLE",
			expected: "gemini",
		},
		{
			name:     "google with spaces",
			input:    "  google  ",
			expected: "gemini",
		},
		{
			name:     "gemini stays gemini",
			input:    "gemini",
			expected: "gemini",
		},
		{
			name:     "aws stays aws",
			input:    "aws",
			expected: "aws",
		},
		{
			name:     "local resolves to ollama",
			input:    "local",
			expected: "ollama",
		},
		{
			name:     "local-ollama resolves to ollama",
			input:    "local-ollama",
			expected: "ollama",
		},
		{
			name:     "qwen alias resolves to ollama",
			input:    "qwen",
			expected: "ollama",
		},
		{
			name:     "qwen3 alias resolves to ollama",
			input:    "qwen3",
			expected: "ollama",
		},
		{
			name:     "qwen3-vl alias resolves to ollama",
			input:    "qwen3-vl",
			expected: "ollama",
		},
		{
			name:     "gemma3 alias resolves to ollama",
			input:    "gemma3",
			expected: "ollama",
		},
		{
			name:     "gemma-3 alias resolves to ollama",
			input:    "gemma-3",
			expected: "ollama",
		},
		{
			name:     "openai stays openai",
			input:    "openai",
			expected: "openai",
		},
		{
			name:     "unknown provider unchanged",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "spaces only",
			input:    "   ",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveProviderAlias(tt.input)
			if result != tt.expected {
				t.Errorf("ResolveProviderAlias(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestGetProvider tests provider factory function
func TestGetProvider(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		wantName         string
		wantError        bool
		skipIfNoCredsFor string // Skip test if this provider's credentials aren't available
	}{
		{
			name:             "gemini provider",
			input:            "gemini",
			wantName:         "gemini",
			wantError:        false,
			skipIfNoCredsFor: "gemini",
		},
		{
			name:      "ollama provider",
			input:     "ollama",
			wantName:  "ollama",
			wantError: false,
		},
		{
			name:             "aws provider",
			input:            "aws",
			wantName:         "aws",
			wantError:        false,
			skipIfNoCredsFor: "aws",
		},
		{
			name:             "rekognition alias",
			input:            "rekognition",
			wantName:         "aws",
			wantError:        false,
			skipIfNoCredsFor: "aws",
		},
		{
			name:             "openai provider",
			input:            "openai",
			wantName:         "openai",
			wantError:        false,
			skipIfNoCredsFor: "openai",
		},
		{
			name:             "gpt4vision alias",
			input:            "gpt4vision",
			wantName:         "openai",
			wantError:        false,
			skipIfNoCredsFor: "openai",
		},
		{
			name:             "gpt-4-vision alias",
			input:            "gpt-4-vision",
			wantName:         "openai",
			wantError:        false,
			skipIfNoCredsFor: "openai",
		},
		{
			name:             "uppercase provider",
			input:            "GEMINI",
			wantName:         "gemini",
			wantError:        false,
			skipIfNoCredsFor: "gemini",
		},
		{
			name:             "provider with spaces",
			input:            "  aws  ",
			wantName:         "aws",
			wantError:        false,
			skipIfNoCredsFor: "aws",
		},
		{
			name:      "unknown provider returns error",
			input:     "unknown",
			wantName:  "",
			wantError: true,
		},
		{
			name:      "empty string returns error",
			input:     "",
			wantName:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := GetProvider(tt.input)

			if tt.wantError {
				if err == nil {
					t.Errorf("GetProvider(%q) expected error, got nil", tt.input)
				}
				return
			}

			// If credentials are required but not available, skip or expect error
			if err != nil {
				if tt.skipIfNoCredsFor != "" {
					t.Skipf("Skipping test - %s credentials not configured: %v", tt.skipIfNoCredsFor, err)
					return
				}
				t.Errorf("GetProvider(%q) unexpected error: %v", tt.input, err)
				return
			}

			if provider == nil {
				t.Errorf("GetProvider(%q) returned nil provider", tt.input)
				return
			}

			if provider.Name() != tt.wantName {
				t.Errorf("GetProvider(%q).Name() = %q, want %q", tt.input, provider.Name(), tt.wantName)
			}
		})
	}
}

// TestDefaultDetectOptions tests default options creation
func TestDefaultDetectOptions(t *testing.T) {
	opts := DefaultDetectOptions()

	if opts == nil {
		t.Fatal("DefaultDetectOptions() returned nil")
	}

	if len(opts.Features) != 1 || opts.Features[0] != FeatureLabels {
		t.Errorf("DefaultDetectOptions().Features = %v, want [labels]", opts.Features)
	}

	if opts.MaxResults != 10 {
		t.Errorf("DefaultDetectOptions().MaxResults = %d, want 10", opts.MaxResults)
	}

	if opts.MinConfidence != 0.5 {
		t.Errorf("DefaultDetectOptions().MinConfidence = %f, want 0.5", opts.MinConfidence)
	}

	if opts.CustomPrompt != "" {
		t.Errorf("DefaultDetectOptions().CustomPrompt = %q, want empty string", opts.CustomPrompt)
	}

	if opts.Language != "" {
		t.Errorf("DefaultDetectOptions().Language = %q, want empty string", opts.Language)
	}

	if opts.IncludeRawResponse {
		t.Errorf("DefaultDetectOptions().IncludeRawResponse = true, want false")
	}
}

// TestFeatureString tests Feature.String() method
func TestFeatureString(t *testing.T) {
	tests := []struct {
		feature  Feature
		expected string
	}{
		{FeatureLabels, "labels"},
		{FeatureText, "text"},
		{FeatureFaces, "faces"},
		{FeatureWeb, "web"},
		{FeatureDescription, "description"},
		{FeatureProperties, "properties"},
		{FeatureObjects, "objects"},
		{FeatureLandmarks, "landmarks"},
		{FeatureLogos, "logos"},
		{FeatureSafeSearch, "safesearch"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.feature.String()
			if result != tt.expected {
				t.Errorf("Feature(%q).String() = %q, want %q", tt.feature, result, tt.expected)
			}
		})
	}
}

// TestDetectionResultJSON tests JSON marshaling/unmarshaling of DetectionResult
func TestDetectionResultJSON(t *testing.T) {
	now := time.Now().UTC().Round(time.Second)

	original := &DetectionResult{
		Provider: "gemini",
		Labels: []Label{
			{Name: "dog", Confidence: 0.95},
			{Name: "pet", Confidence: 0.88},
		},
		Description: "A brown dog sitting on grass",
		Text: []TextBlock{
			{Text: "Hello World", Confidence: 0.99},
		},
		Faces: []Face{
			{Confidence: 0.91, Gender: "Male", AgeRange: "25-35"},
		},
		Properties: map[string]string{
			"brightness": "85.5",
			"contrast":   "72.3",
		},
		Confidence:  0.92,
		ProcessedAt: now,
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	// Unmarshal back
	var result DetectionResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	// Verify fields
	if result.Provider != original.Provider {
		t.Errorf("Provider = %q, want %q", result.Provider, original.Provider)
	}

	if len(result.Labels) != len(original.Labels) {
		t.Errorf("Labels length = %d, want %d", len(result.Labels), len(original.Labels))
	}

	if result.Description != original.Description {
		t.Errorf("Description = %q, want %q", result.Description, original.Description)
	}

	if len(result.Text) != len(original.Text) {
		t.Errorf("Text length = %d, want %d", len(result.Text), len(original.Text))
	}

	if len(result.Faces) != len(original.Faces) {
		t.Errorf("Faces length = %d, want %d", len(result.Faces), len(original.Faces))
	}

	if result.Confidence != original.Confidence {
		t.Errorf("Confidence = %f, want %f", result.Confidence, original.Confidence)
	}

	if !result.ProcessedAt.Equal(original.ProcessedAt) {
		t.Errorf("ProcessedAt = %v, want %v", result.ProcessedAt, original.ProcessedAt)
	}
}

// TestLabelJSON tests Label JSON marshaling
func TestLabelJSON(t *testing.T) {
	label := Label{
		Name:       "dog",
		Confidence: 0.95,
		Score:      95.5,
		MID:        "/m/0bt9lr",
		Categories: []string{"animal", "pet"},
		TopicID:    "topic123",
	}

	data, err := json.Marshal(label)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var result Label
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !reflect.DeepEqual(result, label) {
		t.Errorf("Label JSON roundtrip failed: got %+v, want %+v", result, label)
	}
}

// TestTextBlockJSON tests TextBlock JSON marshaling
func TestTextBlockJSON(t *testing.T) {
	textBlock := TextBlock{
		Text:       "Hello World",
		Confidence: 0.99,
		Language:   "en",
		BoundingBox: &Box{
			X:      10,
			Y:      20,
			Width:  100,
			Height: 50,
		},
		Type: "LINE",
	}

	data, err := json.Marshal(textBlock)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var result TextBlock
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !reflect.DeepEqual(result, textBlock) {
		t.Errorf("TextBlock JSON roundtrip failed: got %+v, want %+v", result, textBlock)
	}
}

// TestFaceJSON tests Face JSON marshaling
func TestFaceJSON(t *testing.T) {
	face := Face{
		Confidence:         0.91,
		BoundingBox:        &Box{X: 100, Y: 200, Width: 150, Height: 200},
		JoyLikelihood:      "LIKELY",
		SorrowLikelihood:   "UNLIKELY",
		AngerLikelihood:    "VERY_UNLIKELY",
		SurpriseLikelihood: "POSSIBLE",
		Gender:             "Male",
		AgeRange:           "25-35",
		Landmarks: []Landmark{
			{Type: "LEFT_EYE", X: 120, Y: 220},
			{Type: "RIGHT_EYE", X: 160, Y: 220},
		},
	}

	data, err := json.Marshal(face)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var result Face
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !reflect.DeepEqual(result, face) {
		t.Errorf("Face JSON roundtrip failed: got %+v, want %+v", result, face)
	}
}

// TestBoxJSON tests Box JSON marshaling
func TestBoxJSON(t *testing.T) {
	box := Box{
		X:      10.5,
		Y:      20.3,
		Width:  100.7,
		Height: 50.2,
	}

	data, err := json.Marshal(box)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var result Box
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !reflect.DeepEqual(result, box) {
		t.Errorf("Box JSON roundtrip failed: got %+v, want %+v", result, box)
	}
}

// TestBoundingBoxJSON tests BoundingBox JSON marshaling
func TestBoundingBoxJSON(t *testing.T) {
	bbox := BoundingBox{
		Label:      "dog",
		Confidence: 0.95,
		Box: Box{
			X:      10,
			Y:      20,
			Width:  100,
			Height: 50,
		},
	}

	data, err := json.Marshal(bbox)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var result BoundingBox
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !reflect.DeepEqual(result, bbox) {
		t.Errorf("BoundingBox JSON roundtrip failed: got %+v, want %+v", result, bbox)
	}
}

// TestWebDetectionJSON tests WebDetection JSON marshaling
func TestWebDetectionJSON(t *testing.T) {
	webDetection := WebDetection{
		WebEntities: []WebEntity{
			{EntityID: "/m/0bt9lr", Score: 0.95, Description: "Dog"},
		},
		FullMatchingImages: []WebImage{
			{URL: "https://example.com/image1.jpg", Score: 0.99},
		},
		BestGuessLabels: []string{"Brown dog", "Pet"},
	}

	data, err := json.Marshal(webDetection)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var result WebDetection
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !reflect.DeepEqual(result, webDetection) {
		t.Errorf("WebDetection JSON roundtrip failed: got %+v, want %+v", result, webDetection)
	}
}

// TestDetectOptionsJSON tests DetectOptions JSON marshaling
func TestDetectOptionsJSON(t *testing.T) {
	opts := &DetectOptions{
		Features:           []Feature{FeatureLabels, FeatureText},
		MaxResults:         20,
		MinConfidence:      0.7,
		CustomPrompt:       "What is in this image?",
		Language:           "en",
		IncludeRawResponse: true,
	}

	data, err := json.Marshal(opts)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}

	var result DetectOptions
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error: %v", err)
	}

	if !reflect.DeepEqual(result, *opts) {
		t.Errorf("DetectOptions JSON roundtrip failed: got %+v, want %+v", result, *opts)
	}
}

// TestBuildDetectionPrompt tests the shared prompt builder
func TestBuildDetectionPrompt(t *testing.T) {
	tests := []struct {
		name        string
		opts        *DetectOptions
		exactMatch  string
		contains    []string
		notContains []string
	}{
		{
			name: "custom prompt",
			opts: &DetectOptions{
				CustomPrompt: "describe this",
			},
			exactMatch: "describe this",
		},
		{
			name: "labels feature",
			opts: &DetectOptions{
				Features:      []Feature{FeatureLabels},
				MaxResults:    10,
				MinConfidence: 0.5,
			},
			contains: []string{"labels"},
		},
		{
			name: "description feature",
			opts: &DetectOptions{
				Features: []Feature{FeatureDescription},
			},
			contains: []string{"description"},
		},
		{
			name: "text feature",
			opts: &DetectOptions{
				Features: []Feature{FeatureText},
			},
			contains: []string{"text"},
		},
		{
			name: "multiple features",
			opts: &DetectOptions{
				Features:      []Feature{FeatureLabels, FeatureDescription, FeatureText},
				MaxResults:    10,
				MinConfidence: 0.5,
			},
			contains: []string{"labels", "description", "text"},
		},
		{
			name: "no features default",
			opts: &DetectOptions{
				Features:      []Feature{},
				MaxResults:    10,
				MinConfidence: 0.5,
			},
			contains: []string{"Analyze"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildDetectionPrompt(tt.opts)

			if tt.exactMatch != "" {
				if result != tt.exactMatch {
					t.Errorf("buildDetectionPrompt() = %q, want exactly %q", result, tt.exactMatch)
				}
				return
			}

			lower := strings.ToLower(result)
			for _, want := range tt.contains {
				if !strings.Contains(lower, strings.ToLower(want)) {
					t.Errorf("buildDetectionPrompt() result does not contain %q; got: %q", want, result)
				}
			}
		})
	}
}

// TestParseJSONDetectionResponse tests the shared JSON response parser
func TestParseJSONDetectionResponse(t *testing.T) {
	t.Run("valid labels", func(t *testing.T) {
		input := `{"labels":[{"name":"cat","confidence":0.95},{"name":"dog","confidence":0.8}]}`
		result := &DetectionResult{Labels: []Label{}, Text: []TextBlock{}, Properties: make(map[string]string)}
		if err := parseJSONDetectionResponse(input, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Labels) != 2 {
			t.Fatalf("expected 2 labels, got %d", len(result.Labels))
		}
		if result.Labels[0].Name != "cat" {
			t.Errorf("Labels[0].Name = %q, want %q", result.Labels[0].Name, "cat")
		}
		if math.Abs(float64(result.Labels[0].Confidence)-0.95) > 1e-5 {
			t.Errorf("Labels[0].Confidence = %f, want 0.95", result.Labels[0].Confidence)
		}
	})

	t.Run("valid description", func(t *testing.T) {
		input := `{"description":"A beautiful sunset"}`
		result := &DetectionResult{Labels: []Label{}, Text: []TextBlock{}, Properties: make(map[string]string)}
		if err := parseJSONDetectionResponse(input, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Description != "A beautiful sunset" {
			t.Errorf("Description = %q, want %q", result.Description, "A beautiful sunset")
		}
	})

	t.Run("ollama extended fields", func(t *testing.T) {
		input := `{"labels":[{"name":"cat","confidence":0.9,"score":0.85,"mid":"/m/01yrx","categories":["animal"],"topic_id":"123"}]}`
		result := &DetectionResult{Labels: []Label{}, Text: []TextBlock{}, Properties: make(map[string]string)}
		if err := parseJSONDetectionResponse(input, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Labels) != 1 {
			t.Fatalf("expected 1 label, got %d", len(result.Labels))
		}
		label := result.Labels[0]
		if math.Abs(float64(label.Score)-0.85) > 1e-5 {
			t.Errorf("Score = %f, want 0.85", label.Score)
		}
		if label.MID != "/m/01yrx" {
			t.Errorf("MID = %q, want %q", label.MID, "/m/01yrx")
		}
		if len(label.Categories) != 1 || label.Categories[0] != "animal" {
			t.Errorf("Categories = %v, want [\"animal\"]", label.Categories)
		}
		if label.TopicID != "123" {
			t.Errorf("TopicID = %q, want %q", label.TopicID, "123")
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		input := `not json at all`
		result := &DetectionResult{Labels: []Label{}, Text: []TextBlock{}, Properties: make(map[string]string)}
		if err := parseJSONDetectionResponse(input, result); err == nil {
			t.Error("expected error for invalid JSON, got nil")
		}
	})

	t.Run("empty labels", func(t *testing.T) {
		input := `{"labels":[]}`
		result := &DetectionResult{Labels: []Label{}, Text: []TextBlock{}, Properties: make(map[string]string)}
		if err := parseJSONDetectionResponse(input, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Labels) != 0 {
			t.Errorf("expected empty labels, got %d", len(result.Labels))
		}
	})

	t.Run("confidence calculated", func(t *testing.T) {
		input := `{"labels":[{"name":"a","confidence":0.8},{"name":"b","confidence":0.6}]}`
		result := &DetectionResult{Labels: []Label{}, Text: []TextBlock{}, Properties: make(map[string]string)}
		if err := parseJSONDetectionResponse(input, result); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := float32(0.7)
		if math.Abs(float64(result.Confidence)-float64(expected)) > 1e-5 {
			t.Errorf("Confidence = %f, want %f", result.Confidence, expected)
		}
	})
}

// TestExtractLabelsFromPlainText tests the plain-text label extractor
func TestExtractLabelsFromPlainText(t *testing.T) {
	t.Run("numbered list", func(t *testing.T) {
		input := "1. cat (92%)\n2. dog (85%)"
		opts := &DetectOptions{MaxResults: 10, MinConfidence: 0.0}
		labels := extractLabelsFromPlainText(input, opts)
		if len(labels) != 2 {
			t.Fatalf("expected 2 labels, got %d: %v", len(labels), labels)
		}
		if labels[0].Name != "cat" {
			t.Errorf("labels[0].Name = %q, want %q", labels[0].Name, "cat")
		}
		if math.Abs(float64(labels[0].Confidence)-0.92) > 1e-4 {
			t.Errorf("labels[0].Confidence = %f, want ~0.92", labels[0].Confidence)
		}
	})

	t.Run("dash list", func(t *testing.T) {
		input := "- tree\n- car"
		opts := &DetectOptions{MaxResults: 10, MinConfidence: 0.0}
		labels := extractLabelsFromPlainText(input, opts)
		if len(labels) != 2 {
			t.Fatalf("expected 2 labels, got %d: %v", len(labels), labels)
		}
		names := map[string]bool{}
		for _, l := range labels {
			names[l.Name] = true
		}
		if !names["tree"] {
			t.Errorf("expected label %q in %v", "tree", labels)
		}
		if !names["car"] {
			t.Errorf("expected label %q in %v", "car", labels)
		}
		for _, l := range labels {
			if math.Abs(float64(l.Confidence)-0.7) > 1e-5 {
				t.Errorf("label %q Confidence = %f, want 0.7 (default)", l.Name, l.Confidence)
			}
		}
	})

	t.Run("plain text with common objects", func(t *testing.T) {
		input := "A photo of a cat and dog"
		opts := &DetectOptions{MaxResults: 10, MinConfidence: 0.0}
		labels := extractLabelsFromPlainText(input, opts)
		names := map[string]bool{}
		for _, l := range labels {
			names[l.Name] = true
		}
		if !names["cat"] {
			t.Errorf("expected label %q in %v", "cat", labels)
		}
		if !names["dog"] {
			t.Errorf("expected label %q in %v", "dog", labels)
		}
	})

	t.Run("empty string", func(t *testing.T) {
		opts := &DetectOptions{MaxResults: 10, MinConfidence: 0.0}
		labels := extractLabelsFromPlainText("", opts)
		if len(labels) != 0 {
			t.Errorf("expected empty labels for empty input, got %d: %v", len(labels), labels)
		}
	})
}

// TestParseTextResponse tests the shared text-response parser skeleton
func TestParseTextResponse(t *testing.T) {
	t.Run("json input", func(t *testing.T) {
		responseText := `{"labels":[{"name":"cat","confidence":0.9}],"description":"A cat"}`
		opts := DefaultDetectOptions()
		result := parseTextResponse(responseText, opts)
		if result == nil {
			t.Fatal("parseTextResponse() returned nil")
		}
		if len(result.Labels) == 0 || result.Labels[0].Name != "cat" {
			t.Errorf("Labels[0].Name = %q, want %q; labels: %v", func() string {
				if len(result.Labels) > 0 {
					return result.Labels[0].Name
				}
				return "<none>"
			}(), "cat", result.Labels)
		}
		if result.Description != "A cat" {
			t.Errorf("Description = %q, want %q", result.Description, "A cat")
		}
	})

	t.Run("plain text input", func(t *testing.T) {
		responseText := "1. tree\n2. flower"
		opts := DefaultDetectOptions()
		result := parseTextResponse(responseText, opts)
		if result == nil {
			t.Fatal("parseTextResponse() returned nil")
		}
		if len(result.Labels) == 0 {
			t.Errorf("expected Labels length > 0, got 0")
		}
		if result.Description != responseText {
			t.Errorf("Description = %q, want %q", result.Description, responseText)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		opts := DefaultDetectOptions()
		result := parseTextResponse("", opts)
		if result == nil {
			t.Fatal("parseTextResponse() returned nil")
		}
		if len(result.Labels) != 0 {
			t.Errorf("expected empty Labels for empty input, got %d", len(result.Labels))
		}
	})
}
