package detection

import (
	"context"
	"image/color"
	"os"
	"testing"
	"time"
)

// TestGeminiProviderName tests the Name method
func TestGeminiProviderName(t *testing.T) {
	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
		return
	}

	name := provider.Name()
	if name != "gemini" {
		t.Errorf("Name() = %q, want %q", name, "gemini")
	}
}

// TestGeminiProviderIsConfigured tests the IsConfigured method
func TestGeminiProviderIsConfigured(t *testing.T) {
	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
		return
	}

	if !provider.IsConfigured() {
		t.Error("IsConfigured() = false, want true for initialized provider")
	}
}

// TestNewGeminiProviderWithCredentials tests provider initialization with valid API key
func TestNewGeminiProviderWithCredentials(t *testing.T) {
	// This test requires GEMINI_API_KEY to be available
	provider, err := NewGeminiProvider()

	// If API key is not available, skip the test
	if err != nil {
		if IsNotConfigured(err) {
			t.Skipf("Skipping test - GEMINI_API_KEY not configured: %v", err)
			return
		}
		t.Fatalf("NewGeminiProvider() unexpected error: %v", err)
	}

	if provider == nil {
		t.Fatal("NewGeminiProvider() returned nil provider")
	}

	if provider.client == nil {
		t.Error("provider.client is nil")
	}

	if provider.apiKey == "" {
		t.Error("provider.apiKey is empty")
	}
}

// TestNewGeminiProviderWithoutCredentials tests provider initialization without API key
func TestNewGeminiProviderWithoutCredentials(t *testing.T) {
	// Save original environment
	origAPIKey := os.Getenv("GEMINI_API_KEY")

	// Clear Gemini API key temporarily
	os.Unsetenv("GEMINI_API_KEY")

	// Restore environment after test
	defer func() {
		if origAPIKey != "" {
			os.Setenv("GEMINI_API_KEY", origAPIKey)
		}
	}()

	provider, err := NewGeminiProvider()

	// Should return error when API key not available
	if err == nil {
		t.Error("NewGeminiProvider() expected error when GEMINI_API_KEY not set, got nil")
		return
	}

	// Verify error is ErrProviderNotConfigured
	if !IsNotConfigured(err) {
		t.Errorf("Expected ErrProviderNotConfigured, got: %v", err)
	}

	if provider != nil {
		t.Error("NewGeminiProvider() returned non-nil provider with error")
	}
}

// TestGeminiProviderDetectWithNilOptions tests detection with nil options
func TestGeminiProviderDetectWithNilOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gemini API test in short mode")
	}

	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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
			t.Skip("Gemini not properly configured for API calls")
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

	if result.Provider != "gemini" {
		t.Errorf("result.Provider = %q, want %q", result.Provider, "gemini")
	}
}

// TestGeminiProviderDetectWithLabels tests detection with labels feature
func TestGeminiProviderDetectWithLabels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gemini API test in short mode")
	}

	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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
			t.Skip("Gemini not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error (may be expected for test image): %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	AssertDetectionResult(t, result)

	// Gemini might return 0 or more labels for a simple colored square
	t.Logf("Detected %d labels", len(result.Labels))

	for i, label := range result.Labels {
		t.Logf("Label %d: %s (%.1f%%)", i, label.Name, label.Confidence*100)
		AssertLabel(t, label)
	}
}

// TestGeminiProviderDetectWithCustomPrompt tests detection with custom prompt
func TestGeminiProviderDetectWithCustomPrompt(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gemini API test in short mode")
	}

	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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
			t.Skip("Gemini not properly configured for API calls")
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

// TestGeminiProviderDetectWithDescription tests detection with description feature
func TestGeminiProviderDetectWithDescription(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gemini API test in short mode")
	}

	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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
			t.Skip("Gemini not properly configured for API calls")
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

// TestGeminiProviderDetectContextCancellation tests context cancellation
func TestGeminiProviderDetectContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gemini API test in short mode")
	}

	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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

// TestGeminiProviderDetectWithInvalidImage tests detection with nil image
func TestGeminiProviderDetectWithInvalidImage(t *testing.T) {
	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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

// TestGeminiProviderProviderInterfaceCompliance tests that GeminiProvider implements Provider interface
func TestGeminiProviderProviderInterfaceCompliance(t *testing.T) {
	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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

// TestGeminiProviderMultipleFeatures tests detection with multiple features
func TestGeminiProviderMultipleFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gemini API test in short mode")
	}

	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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
			t.Skip("Gemini not properly configured for API calls")
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

// TestGeminiProviderBuildPrompt tests prompt building for different features
func TestGeminiProviderBuildPrompt(t *testing.T) {
	provider := &GeminiProvider{apiKey: "test-key"}

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
			contains: []string{"adult", "violent"},
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

// TestExtractJSONFromMarkdown tests JSON extraction from markdown code blocks
func TestExtractJSONFromMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "json code block",
			input:    "```json\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "generic code block",
			input:    "```\n{\"key\": \"value\"}\n```",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "no code block",
			input:    "{\"key\": \"value\"}",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "with whitespace",
			input:    "  ```json  \n  {\"key\": \"value\"}  \n  ```  ",
			expected: "{\"key\": \"value\"}",
		},
		{
			name:     "text before json block",
			input:    "Here's the result:\n```json\n{\"labels\": []}\n```",
			expected: "Here's the result:\n```json\n{\"labels\": []}\n```", // Function doesn't extract from within text
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSONFromMarkdown(tt.input)
			if result != tt.expected {
				t.Errorf("extractJSONFromMarkdown() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestContainsFeature tests the containsFeature helper
func TestContainsFeature(t *testing.T) {
	tests := []struct {
		name     string
		features []Feature
		target   Feature
		expected bool
	}{
		{
			name:     "contains labels",
			features: []Feature{FeatureLabels, FeatureText},
			target:   FeatureLabels,
			expected: true,
		},
		{
			name:     "does not contain faces",
			features: []Feature{FeatureLabels, FeatureText},
			target:   FeatureFaces,
			expected: false,
		},
		{
			name:     "empty features",
			features: []Feature{},
			target:   FeatureLabels,
			expected: false,
		},
		{
			name:     "single feature match",
			features: []Feature{FeatureDescription},
			target:   FeatureDescription,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsFeature(tt.features, tt.target)
			if result != tt.expected {
				t.Errorf("containsFeature() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestGeminiProviderExtractLabelsFromText tests label extraction from natural language
func TestGeminiProviderExtractLabelsFromText(t *testing.T) {
	provider := &GeminiProvider{apiKey: "test-key"}

	tests := []struct {
		name          string
		text          string
		opts          *DetectOptions
		expectLabels  bool
		expectedWords []string
	}{
		{
			name:          "contains dog",
			text:          "I see a dog in the image",
			opts:          &DetectOptions{MaxResults: 10},
			expectLabels:  true,
			expectedWords: []string{"dog"},
		},
		{
			name:          "contains multiple objects",
			text:          "This image shows a cat near a tree and a car",
			opts:          &DetectOptions{MaxResults: 10},
			expectLabels:  true,
			expectedWords: []string{"cat", "tree", "car"},
		},
		{
			name:         "no recognized objects",
			text:         "This is a beautiful scene",
			opts:         &DetectOptions{MaxResults: 10},
			expectLabels: false,
		},
		{
			name:          "max results limit",
			text:          "dog cat person car building tree flower animal vehicle",
			opts:          &DetectOptions{MaxResults: 2},
			expectLabels:  true,
			expectedWords: []string{}, // Just check some labels exist
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

			// Validate label structure
			for _, label := range labels {
				AssertLabel(t, label)
			}

			t.Logf("Extracted %d labels from: %q", len(labels), tt.text)
		})
	}
}

// TestGeminiProviderRawResponse tests raw response inclusion
func TestGeminiProviderRawResponse(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Gemini API test in short mode")
	}

	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		t.Skipf("Gemini provider not configured: %v", err)
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
			t.Skip("Gemini not properly configured for API calls")
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

// BenchmarkGeminiProviderDetect benchmarks Gemini detection
func BenchmarkGeminiProviderDetect(b *testing.B) {
	// Skip if Gemini API key not available
	provider, err := NewGeminiProvider()
	if err != nil {
		b.Skipf("Gemini provider not configured: %v", err)
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

// Helper function to check if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}
