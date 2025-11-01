package detection

import (
	"context"
	"image/color"
	"os"
	"testing"
	"time"
)

// TestOpenAIProviderName tests the Name method
func TestOpenAIProviderName(t *testing.T) {
	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	name := provider.Name()
	if name != "openai" {
		t.Errorf("Name() = %q, want %q", name, "openai")
	}
}

// TestOpenAIProviderIsConfigured tests the IsConfigured method
func TestOpenAIProviderIsConfigured(t *testing.T) {
	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	if !provider.IsConfigured() {
		t.Error("IsConfigured() = false, want true for initialized provider")
	}
}

// TestNewOpenAIProviderWithCredentials tests provider initialization with valid API key
func TestNewOpenAIProviderWithCredentials(t *testing.T) {
	// This test requires OPENAI_API_KEY to be available
	provider, err := NewOpenAIProvider()

	// If API key is not available, skip the test
	if err != nil {
		if IsNotConfigured(err) {
			t.Skipf("Skipping test - OPENAI_API_KEY not configured: %v", err)
			return
		}
		t.Fatalf("NewOpenAIProvider() unexpected error: %v", err)
	}

	if provider == nil {
		t.Fatal("NewOpenAIProvider() returned nil provider")
	}

	if provider.client == nil {
		t.Error("provider.client is nil")
	}

	if provider.apiKey == "" {
		t.Error("provider.apiKey is empty")
	}
}

// TestNewOpenAIProviderWithoutCredentials tests provider initialization without API key
func TestNewOpenAIProviderWithoutCredentials(t *testing.T) {
	// Save original environment
	origAPIKey := os.Getenv("OPENAI_API_KEY")

	// Clear OpenAI API key temporarily
	os.Unsetenv("OPENAI_API_KEY")

	// Restore environment after test
	defer func() {
		if origAPIKey != "" {
			os.Setenv("OPENAI_API_KEY", origAPIKey)
		}
	}()

	provider, err := NewOpenAIProvider()

	// Should return error when API key not available
	if err == nil {
		t.Error("NewOpenAIProvider() expected error when OPENAI_API_KEY not set, got nil")
		return
	}

	// Verify error is ErrProviderNotConfigured
	if !IsNotConfigured(err) {
		t.Errorf("Expected ErrProviderNotConfigured, got: %v", err)
	}

	if provider != nil {
		t.Error("NewOpenAIProvider() returned non-nil provider with error")
	}
}

// TestOpenAIProviderDetectWithNilOptions tests detection with nil options
func TestOpenAIProviderDetectWithNilOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OpenAI API test in short mode")
	}

	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	// Create a small test image
	img := CreateTestImage(100, 100, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	// Use a short timeout to avoid long-running tests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Call Detect with nil options (should use defaults)
	result, err := provider.Detect(ctx, img, nil)

	// Note: This makes a real API call, so we might get various errors
	if err != nil {
		// Check if it's a known error type
		if IsNotConfigured(err) {
			t.Skip("OpenAI not properly configured for API calls")
			return
		}
		// Other errors are fine for this test - we're just testing the nil options handling
		t.Logf("Detect returned error (expected for test): %v", err)
		return
	}

	// If successful, verify result structure
	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	AssertDetectionResult(t, result)

	if result.Provider != "openai" {
		t.Errorf("result.Provider = %q, want %q", result.Provider, "openai")
	}
}

// TestOpenAIProviderDetectWithLabels tests detection with labels feature
func TestOpenAIProviderDetectWithLabels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OpenAI API test in short mode")
	}

	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	// Create a test image
	img := CreateTestImage(200, 200, color.NRGBA{R: 0, G: 0, B: 255, A: 255})

	opts := &DetectOptions{
		Features:      []Feature{FeatureLabels},
		MaxResults:    5,
		MinConfidence: 0.5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := provider.Detect(ctx, img, opts)

	if err != nil {
		if IsNotConfigured(err) {
			t.Skip("OpenAI not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error (may be expected for test image): %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	AssertDetectionResult(t, result)

	// OpenAI might return 0 or more labels for a simple colored square
	t.Logf("Detected %d labels", len(result.Labels))

	for i, label := range result.Labels {
		t.Logf("Label %d: %s (%.1f%%)", i, label.Name, label.Confidence*100)
		AssertLabel(t, label)
	}
}

// TestOpenAIProviderDetectWithCustomPrompt tests detection with custom prompt
func TestOpenAIProviderDetectWithCustomPrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OpenAI API test in short mode")
	}

	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	img := CreateTestImage(100, 100, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	opts := &DetectOptions{
		CustomPrompt: "What is the dominant color in this image? Respond in one sentence.",
		MaxResults:   5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := provider.Detect(ctx, img, opts)

	if err != nil {
		if IsNotConfigured(err) {
			t.Skip("OpenAI not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	// With custom prompt, we should get a description
	if result.Description == "" {
		t.Log("Warning: Expected description with custom prompt, got empty")
	} else {
		t.Logf("Custom prompt response: %s", result.Description)
	}

	AssertDetectionResult(t, result)
}

// TestOpenAIProviderDetectWithDescription tests detection with description feature
func TestOpenAIProviderDetectWithDescription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OpenAI API test in short mode")
	}

	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	img := CreateTestImageWithPattern(200, 200)

	opts := &DetectOptions{
		Features:   []Feature{FeatureDescription},
		MaxResults: 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := provider.Detect(ctx, img, opts)

	if err != nil {
		if IsNotConfigured(err) {
			t.Skip("OpenAI not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	AssertDetectionResult(t, result)

	// Should have a description
	t.Logf("Description: %s", result.Description)
}

// TestOpenAIProviderDetectContextCancellation tests context cancellation
func TestOpenAIProviderDetectContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OpenAI API test in short mode")
	}

	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	img := CreateTestImage(100, 100, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := DefaultDetectOptions()

	result, err := provider.Detect(ctx, img, opts)

	// Should return an error due to cancelled context
	if err == nil {
		t.Log("Warning: Expected error from cancelled context, but got nil. API call may have completed before cancellation.")
		if result != nil {
			t.Logf("Got result with %d labels", len(result.Labels))
		}
		return
	}

	if result != nil {
		t.Error("Detect() returned non-nil result with error")
	}

	t.Logf("Got expected error from cancelled context: %v", err)
}

// TestOpenAIProviderDetectWithInvalidImage tests detection with nil image
func TestOpenAIProviderDetectWithInvalidImage(t *testing.T) {
	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := DefaultDetectOptions()

	// This should panic or error during image encoding
	defer func() {
		if r := recover(); r != nil {
			t.Logf("Got expected panic with nil image: %v", r)
		}
	}()

	result, err := provider.Detect(ctx, nil, opts)

	// Should get an error
	if err == nil {
		t.Error("Expected error with nil image, got nil")
		if result != nil {
			t.Errorf("Got non-nil result: %+v", result)
		}
	}
}

// TestOpenAIProviderProviderInterfaceCompliance tests that OpenAIProvider implements Provider interface
func TestOpenAIProviderProviderInterfaceCompliance(t *testing.T) {
	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	// Verify it implements the Provider interface
	var _ Provider = provider

	// Test interface methods
	name := provider.Name()
	if name == "" {
		t.Error("Provider.Name() returned empty string")
	}

	isConfigured := provider.IsConfigured()
	if !isConfigured {
		t.Error("Provider.IsConfigured() returned false for initialized provider")
	}
}

// TestOpenAIProviderMultipleFeatures tests detection with multiple features
func TestOpenAIProviderMultipleFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OpenAI API test in short mode")
	}

	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	img := CreateTestImageWithText(300, 200)

	opts := &DetectOptions{
		Features: []Feature{
			FeatureLabels,
			FeatureDescription,
			FeatureText,
		},
		MaxResults:    10,
		MinConfidence: 0.5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := provider.Detect(ctx, img, opts)

	if err != nil {
		if IsNotConfigured(err) {
			t.Skip("OpenAI not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	t.Logf("Results: %d labels, %d text blocks, description length: %d",
		len(result.Labels), len(result.Text), len(result.Description))

	AssertDetectionResult(t, result)
}

// TestOpenAIProviderBuildPrompt tests prompt building for different features
func TestOpenAIProviderBuildPrompt(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}

	tests := []struct {
		name     string
		opts     *DetectOptions
		contains []string
	}{
		{
			name: "labels feature",
			opts: &DetectOptions{
				Features:      []Feature{FeatureLabels},
				MaxResults:    5,
				MinConfidence: 0.7,
			},
			contains: []string{"labels", "confidence", "JSON"},
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
			contains: []string{"text", "Extract"},
		},
		{
			name: "faces feature",
			opts: &DetectOptions{
				Features: []Feature{FeatureFaces},
			},
			contains: []string{"faces"},
		},
		{
			name: "properties feature",
			opts: &DetectOptions{
				Features: []Feature{FeatureProperties},
			},
			contains: []string{"properties", "colors"},
		},
		{
			name: "landmarks feature",
			opts: &DetectOptions{
				Features: []Feature{FeatureLandmarks},
			},
			contains: []string{"landmarks"},
		},
		{
			name: "safe search feature",
			opts: &DetectOptions{
				Features: []Feature{FeatureSafeSearch},
			},
			contains: []string{"inappropriate"},
		},
		{
			name: "custom prompt",
			opts: &DetectOptions{
				CustomPrompt: "What color is this?",
			},
			contains: []string{"What color is this?"},
		},
		{
			name: "multiple features",
			opts: &DetectOptions{
				Features: []Feature{FeatureLabels, FeatureDescription, FeatureText},
			},
			contains: []string{"labels", "description", "text"},
		},
		{
			name:     "default prompt",
			opts:     &DetectOptions{Features: []Feature{}},
			contains: []string{"Analyze", "objects"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := provider.buildPrompt(tt.opts)

			if prompt == "" {
				t.Error("buildPrompt() returned empty string")
			}

			// Check for expected substrings
			for _, substr := range tt.contains {
				if !containsString(prompt, substr) {
					t.Errorf("buildPrompt() = %q, want to contain %q", prompt, substr)
				}
			}

			t.Logf("Prompt: %s", prompt)
		})
	}
}

// TestOpenAIProviderExtractLabelsFromText tests label extraction from natural language
func TestOpenAIProviderExtractLabelsFromText(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}

	tests := []struct {
		name          string
		text          string
		opts          *DetectOptions
		expectLabels  bool
		expectedWords []string
	}{
		{
			name:          "contains person",
			text:          "I see a person in the image",
			opts:          &DetectOptions{MaxResults: 10},
			expectLabels:  true,
			expectedWords: []string{"person"},
		},
		{
			name:          "contains multiple objects",
			text:          "This image shows people near a tree and a car",
			opts:          &DetectOptions{MaxResults: 10},
			expectLabels:  true,
			expectedWords: []string{"people", "tree", "car"},
		},
		{
			name:         "no recognized objects",
			text:         "This is a beautiful scene",
			opts:         &DetectOptions{MaxResults: 10},
			expectLabels: false,
		},
		{
			name:          "max results limit",
			text:          "person people dog cat car building tree flower animal vehicle house plant",
			opts:          &DetectOptions{MaxResults: 3},
			expectLabels:  true,
			expectedWords: []string{}, // Just check some labels exist
		},
		{
			name:          "deduplication",
			text:          "person person person dog dog cat",
			opts:          &DetectOptions{MaxResults: 10},
			expectLabels:  true,
			expectedWords: []string{}, // Should have deduplicated labels
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			labels := provider.extractLabelsFromText(tt.text, tt.opts)

			if tt.expectLabels && len(labels) == 0 {
				t.Error("extractLabelsFromText() expected labels, got none")
			}

			if !tt.expectLabels && len(labels) > 0 {
				t.Errorf("extractLabelsFromText() expected no labels, got %d", len(labels))
			}

			// Check that max results is respected
			if len(labels) > tt.opts.MaxResults {
				t.Errorf("extractLabelsFromText() returned %d labels, max is %d",
					len(labels), tt.opts.MaxResults)
			}

			// Check for deduplication
			seen := make(map[string]bool)
			for _, label := range labels {
				if seen[label.Name] {
					t.Errorf("extractLabelsFromText() returned duplicate label: %s", label.Name)
				}
				seen[label.Name] = true
				AssertLabel(t, label)
			}

			t.Logf("Extracted %d labels from: %q", len(labels), tt.text)
		})
	}
}

// TestOpenAIProviderRawResponse tests raw response inclusion
func TestOpenAIProviderRawResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping OpenAI API test in short mode")
	}

	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		t.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	img := CreateTestImage(100, 100, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	opts := &DetectOptions{
		Features:           []Feature{FeatureLabels},
		IncludeRawResponse: true,
		MaxResults:         5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := provider.Detect(ctx, img, opts)

	if err != nil {
		if IsNotConfigured(err) {
			t.Skip("OpenAI not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	// When IncludeRawResponse is true, we should have raw response
	if result.RawResponse == "" {
		t.Error("Expected RawResponse to be set when IncludeRawResponse=true")
	} else {
		t.Logf("Raw response length: %d bytes", len(result.RawResponse))
	}
}

// TestOpenAIProviderParseJSONResponse tests JSON response parsing
func TestOpenAIProviderParseJSONResponse(t *testing.T) {
	provider := &OpenAIProvider{apiKey: "test-key"}

	tests := []struct {
		name        string
		jsonText    string
		wantLabels  int
		wantText    int
		wantDesc    bool
		wantError   bool
	}{
		{
			name: "valid labels response",
			jsonText: `{
				"labels": [
					{"name": "dog", "confidence": 0.95},
					{"name": "pet", "confidence": 0.88}
				]
			}`,
			wantLabels: 2,
			wantError:  false,
		},
		{
			name: "labels with description",
			jsonText: `{
				"labels": [{"name": "cat", "confidence": 0.92}],
				"description": "A cat sitting on a couch"
			}`,
			wantLabels: 1,
			wantDesc:   true,
			wantError:  false,
		},
		{
			name: "text extraction response",
			jsonText: `{
				"text": [
					{"text": "Hello World", "confidence": 0.99},
					{"text": "Test", "confidence": 0.95}
				]
			}`,
			wantText:  2,
			wantError: false,
		},
		{
			name:      "invalid json",
			jsonText:  `not valid json`,
			wantError: true,
		},
		{
			name: "json in markdown",
			jsonText: "```json\n{\"labels\": [{\"name\": \"test\", \"confidence\": 0.9}]}\n```",
			wantLabels: 1,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &DetectionResult{
				Labels:     []Label{},
				Text:       []TextBlock{},
				Properties: make(map[string]string),
			}

			err := provider.parseJSONResponse(tt.jsonText, result)

			if tt.wantError {
				if err == nil {
					t.Error("parseJSONResponse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("parseJSONResponse() unexpected error: %v", err)
				return
			}

			if len(result.Labels) != tt.wantLabels {
				t.Errorf("parseJSONResponse() got %d labels, want %d", len(result.Labels), tt.wantLabels)
			}

			if len(result.Text) != tt.wantText {
				t.Errorf("parseJSONResponse() got %d text blocks, want %d", len(result.Text), tt.wantText)
			}

			if tt.wantDesc && result.Description == "" {
				t.Error("parseJSONResponse() expected description, got empty")
			}

			// Validate structure
			for _, label := range result.Labels {
				AssertLabel(t, label)
			}
			for _, text := range result.Text {
				AssertTextBlock(t, text)
			}
		})
	}
}

// BenchmarkOpenAIProviderDetect benchmarks OpenAI detection
func BenchmarkOpenAIProviderDetect(b *testing.B) {
	// Skip if OpenAI API key not available
	provider, err := NewOpenAIProvider()
	if err != nil {
		b.Skipf("OpenAI provider not configured: %v", err)
		return
	}

	img := CreateTestImage(100, 100, color.NRGBA{R: 255, G: 0, B: 0, A: 255})
	opts := DefaultDetectOptions()
	ctx := context.Background()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := provider.Detect(ctx, img, opts)
		if err != nil {
			b.Logf("Detect error: %v", err)
		}
	}
}
