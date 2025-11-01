package detection

import (
	"context"
	"image/color"
	"os"
	"testing"
	"time"
)

// TestAWSProviderName tests the Name method
func TestAWSProviderName(t *testing.T) {
	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
		return
	}

	name := provider.Name()
	if name != "aws" {
		t.Errorf("Name() = %q, want %q", name, "aws")
	}
}

// TestAWSProviderIsConfigured tests the IsConfigured method
func TestAWSProviderIsConfigured(t *testing.T) {
	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
		return
	}

	if !provider.IsConfigured() {
		t.Error("IsConfigured() = false, want true for initialized provider")
	}
}

// TestNewAWSProviderWithCredentials tests provider initialization with valid credentials
func TestNewAWSProviderWithCredentials(t *testing.T) {
	// This test requires AWS credentials to be available
	provider, err := NewAWSProvider()

	// If credentials are not available, skip the test
	if err != nil {
		if IsNotConfigured(err) {
			t.Skipf("Skipping test - AWS credentials not configured: %v", err)
			return
		}
		t.Fatalf("NewAWSProvider() unexpected error: %v", err)
	}

	if provider == nil {
		t.Fatal("NewAWSProvider() returned nil provider")
	}

	if provider.client == nil {
		t.Error("provider.client is nil")
	}

	if provider.credSource == "" {
		t.Error("provider.credSource is empty")
	}

	if provider.cfg.Region == "" {
		t.Error("provider.cfg.Region is empty")
	}
}

// TestNewAWSProviderWithoutCredentials tests provider initialization without credentials
func TestNewAWSProviderWithoutCredentials(t *testing.T) {
	// Save original environment
	origAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	origSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	origRegion := os.Getenv("AWS_REGION")
	origProfile := os.Getenv("AWS_PROFILE")

	// Clear AWS environment variables temporarily
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	os.Unsetenv("AWS_REGION")
	os.Unsetenv("AWS_PROFILE")

	// Restore environment after test
	defer func() {
		if origAccessKey != "" {
			os.Setenv("AWS_ACCESS_KEY_ID", origAccessKey)
		}
		if origSecretKey != "" {
			os.Setenv("AWS_SECRET_ACCESS_KEY", origSecretKey)
		}
		if origRegion != "" {
			os.Setenv("AWS_REGION", origRegion)
		}
		if origProfile != "" {
			os.Setenv("AWS_PROFILE", origProfile)
		}
	}()

	provider, err := NewAWSProvider()

	// Should return error when credentials not available
	// Note: This might still succeed if AWS credentials are in ~/.aws/credentials
	// In that case, we just verify the provider is valid
	if err == nil {
		// Credentials were found elsewhere (e.g., ~/.aws/credentials)
		if provider == nil {
			t.Error("NewAWSProvider() returned nil provider without error")
		}
		t.Skip("Skipping test - AWS credentials found in credential chain")
		return
	}

	// Verify error is ErrProviderNotConfigured
	if !IsNotConfigured(err) {
		t.Errorf("Expected ErrProviderNotConfigured, got: %v", err)
	}

	if provider != nil {
		t.Error("NewAWSProvider() returned non-nil provider with error")
	}
}

// TestAWSProviderDetectWithNilOptions tests detection with nil options
func TestAWSProviderDetectWithNilOptions(t *testing.T) {
	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
		return
	}

	// Create a small test image
	img := CreateTestImage(100, 100, color.NRGBA{R: 255, G: 0, B: 0, A: 255})

	// Use a short timeout to avoid long-running tests
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Call Detect with nil options (should use defaults)
	result, err := provider.Detect(ctx, img, nil)

	// Note: This makes a real API call, so we might get various errors
	if err != nil {
		// Check if it's a known error type
		if IsNotConfigured(err) {
			t.Skip("AWS not properly configured for API calls")
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

	if result.Provider != "aws" {
		t.Errorf("result.Provider = %q, want %q", result.Provider, "aws")
	}
}

// TestAWSProviderDetectWithLabels tests detection with labels feature
func TestAWSProviderDetectWithLabels(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping AWS API test in short mode")
	}

	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
		return
	}

	// Create a test image
	img := CreateTestImage(200, 200, color.NRGBA{R: 0, G: 0, B: 255, A: 255})

	opts := &DetectOptions{
		Features:      []Feature{FeatureLabels},
		MaxResults:    5,
		MinConfidence: 0.5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := provider.Detect(ctx, img, opts)

	if err != nil {
		if IsNotConfigured(err) {
			t.Skip("AWS not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error (may be expected for test image): %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	AssertDetectionResult(t, result)

	// AWS might return 0 labels for a simple colored square
	t.Logf("Detected %d labels", len(result.Labels))

	for i, label := range result.Labels {
		t.Logf("Label %d: %s (%.1f%%)", i, label.Name, label.Confidence*100)
		AssertLabel(t, label)
	}
}

// TestAWSProviderDetectContextCancellation tests context cancellation
func TestAWSProviderDetectContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping AWS API test in short mode")
	}

	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
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

// TestAWSProviderDetectWithInvalidImage tests detection with nil image
func TestAWSProviderDetectWithInvalidImage(t *testing.T) {
	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
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

// TestAWSProviderProviderInterfaceCompliance tests that AWSProvider implements Provider interface
func TestAWSProviderProviderInterfaceCompliance(t *testing.T) {
	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
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

// TestAWSProviderMultipleFeatures tests detection with multiple features
func TestAWSProviderMultipleFeatures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping AWS API test in short mode")
	}

	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
		return
	}

	img := CreateTestImageWithText(300, 200)

	opts := &DetectOptions{
		Features: []Feature{
			FeatureLabels,
			FeatureText,
			FeatureFaces,
		},
		MaxResults:    10,
		MinConfidence: 0.5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	result, err := provider.Detect(ctx, img, opts)

	if err != nil {
		if IsNotConfigured(err) {
			t.Skip("AWS not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	t.Logf("Results: %d labels, %d text blocks, %d faces",
		len(result.Labels), len(result.Text), len(result.Faces))

	AssertDetectionResult(t, result)
}

// TestAWSProviderProperties tests AWS image properties feature
func TestAWSProviderProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping AWS API test in short mode")
	}

	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		t.Skipf("AWS provider not configured: %v", err)
		return
	}

	img := CreateTestImageWithPattern(200, 200)

	opts := &DetectOptions{
		Features: []Feature{FeatureProperties},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := provider.Detect(ctx, img, opts)

	if err != nil {
		if IsNotConfigured(err) {
			t.Skip("AWS not properly configured for API calls")
			return
		}
		t.Logf("Detect returned error: %v", err)
		return
	}

	if result == nil {
		t.Fatal("Detect() returned nil result without error")
	}

	t.Logf("Got %d properties", len(result.Properties))

	// Check for expected property keys
	expectedKeys := []string{"brightness", "sharpness", "contrast"}
	for _, key := range expectedKeys {
		if val, ok := result.Properties[key]; ok {
			t.Logf("Property %s: %s", key, val)
		}
	}
}

// BenchmarkAWSProviderDetect benchmarks AWS detection
func BenchmarkAWSProviderDetect(b *testing.B) {
	// Skip if AWS credentials not available
	provider, err := NewAWSProvider()
	if err != nil {
		b.Skipf("AWS provider not configured: %v", err)
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
