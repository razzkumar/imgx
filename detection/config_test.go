package detection

import (
	"sync"
	"testing"
)

// TestGetDefaultProvider tests getting default provider
func TestGetDefaultProvider(t *testing.T) {
	// Save original value
	original := GetDefaultProvider()
	defer SetDefaultProvider(original) // Restore after test

	// Set a test value
	SetDefaultProvider("aws")

	result := GetDefaultProvider()
	if result != "aws" {
		t.Errorf("GetDefaultProvider() = %q, want %q", result, "aws")
	}
}

// TestSetDefaultProvider tests setting default provider
func TestSetDefaultProvider(t *testing.T) {
	// Save original value
	original := GetDefaultProvider()
	defer SetDefaultProvider(original) // Restore after test

	tests := []struct {
		name     string
		provider string
	}{
		{"ollama", "ollama"},
		{"gemini", "gemini"},
		{"aws", "aws"},
		{"openai", "openai"},
		{"empty", ""},
		{"with spaces", "  test  "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultProvider(tt.provider)
			result := GetDefaultProvider()
			if result != tt.provider {
				t.Errorf("After SetDefaultProvider(%q), GetDefaultProvider() = %q, want %q",
					tt.provider, result, tt.provider)
			}
		})
	}
}

// TestGetDefaultConfidence tests getting default confidence
func TestGetDefaultConfidence(t *testing.T) {
	// Save original value
	original := GetDefaultConfidence()
	defer SetDefaultConfidence(original) // Restore after test

	// Set a test value
	SetDefaultConfidence(0.75)

	result := GetDefaultConfidence()
	if result != 0.75 {
		t.Errorf("GetDefaultConfidence() = %f, want %f", result, 0.75)
	}
}

// TestSetDefaultConfidence tests setting default confidence
func TestSetDefaultConfidence(t *testing.T) {
	// Save original value
	original := GetDefaultConfidence()
	defer SetDefaultConfidence(original) // Restore after test

	tests := []struct {
		name       string
		confidence float32
	}{
		{"minimum", 0.0},
		{"low", 0.3},
		{"default", 0.5},
		{"high", 0.8},
		{"maximum", 1.0},
		{"over maximum", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetDefaultConfidence(tt.confidence)
			result := GetDefaultConfidence()
			if result != tt.confidence {
				t.Errorf("After SetDefaultConfidence(%f), GetDefaultConfidence() = %f, want %f",
					tt.confidence, result, tt.confidence)
			}
		})
	}
}

// TestGetTimeout tests getting timeout
func TestGetTimeout(t *testing.T) {
	// Save original value
	original := GetTimeout()
	defer SetTimeout(original) // Restore after test

	// Set a test value
	SetTimeout(60)

	result := GetTimeout()
	if result != 60 {
		t.Errorf("GetTimeout() = %d, want %d", result, 60)
	}
}

// TestSetTimeout tests setting timeout
func TestSetTimeout(t *testing.T) {
	// Save original value
	original := GetTimeout()
	defer SetTimeout(original) // Restore after test

	tests := []struct {
		name    string
		timeout int
	}{
		{"zero", 0},
		{"short", 10},
		{"default", 30},
		{"long", 120},
		{"very long", 600},
		{"negative", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetTimeout(tt.timeout)
			result := GetTimeout()
			if result != tt.timeout {
				t.Errorf("After SetTimeout(%d), GetTimeout() = %d, want %d",
					tt.timeout, result, tt.timeout)
			}
		})
	}
}

// TestConcurrentProviderAccess tests concurrent access to default provider
func TestConcurrentProviderAccess(t *testing.T) {
	// Save original value
	original := GetDefaultProvider()
	defer SetDefaultProvider(original)

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2) // readers and writers

	// Start readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				_ = GetDefaultProvider()
			}
		}()
	}

	// Start writers
	providers := []string{"ollama", "gemini", "aws", "openai"}
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				SetDefaultProvider(providers[j%len(providers)])
			}
		}(i)
	}

	wg.Wait()

	// Verify we can still get a value (no panic/deadlock)
	result := GetDefaultProvider()
	if result == "" {
		t.Error("GetDefaultProvider() returned empty string after concurrent access")
	}
}

// TestConcurrentConfidenceAccess tests concurrent access to default confidence
func TestConcurrentConfidenceAccess(t *testing.T) {
	// Save original value
	original := GetDefaultConfidence()
	defer SetDefaultConfidence(original)

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Start readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				_ = GetDefaultConfidence()
			}
		}()
	}

	// Start writers
	confidences := []float32{0.3, 0.5, 0.7, 0.9}
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				SetDefaultConfidence(confidences[j%len(confidences)])
			}
		}(i)
	}

	wg.Wait()

	// Verify we can still get a value
	result := GetDefaultConfidence()
	if result < 0 || result > 1 {
		t.Errorf("GetDefaultConfidence() returned invalid value %f after concurrent access", result)
	}
}

// TestConcurrentTimeoutAccess tests concurrent access to timeout
func TestConcurrentTimeoutAccess(t *testing.T) {
	// Save original value
	original := GetTimeout()
	defer SetTimeout(original)

	const numGoroutines = 100
	const numOperations = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Start readers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				_ = GetTimeout()
			}
		}()
	}

	// Start writers
	timeouts := []int{10, 30, 60, 120}
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				SetTimeout(timeouts[j%len(timeouts)])
			}
		}(i)
	}

	wg.Wait()

	// Verify we can still get a value
	result := GetTimeout()
	if result <= 0 {
		t.Errorf("GetTimeout() returned invalid value %d after concurrent access", result)
	}
}

// TestConcurrentMixedAccess tests concurrent access to all config values
func TestConcurrentMixedAccess(t *testing.T) {
	// Save original values
	originalProvider := GetDefaultProvider()
	originalConfidence := GetDefaultConfidence()
	originalTimeout := GetTimeout()
	defer func() {
		SetDefaultProvider(originalProvider)
		SetDefaultConfidence(originalConfidence)
		SetTimeout(originalTimeout)
	}()

	const numGoroutines = 50
	const numOperations = 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // Provider, Confidence, Timeout operations

	// Provider operations
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			providers := []string{"ollama", "gemini", "aws", "openai"}
			for j := 0; j < numOperations; j++ {
				if j%2 == 0 {
					SetDefaultProvider(providers[j%len(providers)])
				} else {
					_ = GetDefaultProvider()
				}
			}
		}(i)
	}

	// Confidence operations
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			confidences := []float32{0.3, 0.5, 0.7, 0.9}
			for j := 0; j < numOperations; j++ {
				if j%2 == 0 {
					SetDefaultConfidence(confidences[j%len(confidences)])
				} else {
					_ = GetDefaultConfidence()
				}
			}
		}(i)
	}

	// Timeout operations
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			timeouts := []int{10, 30, 60, 120}
			for j := 0; j < numOperations; j++ {
				if j%2 == 0 {
					SetTimeout(timeouts[j%len(timeouts)])
				} else {
					_ = GetTimeout()
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify all values are still accessible
	if GetDefaultProvider() == "" {
		t.Error("GetDefaultProvider() returned empty after mixed concurrent access")
	}
	if conf := GetDefaultConfidence(); conf < 0 || conf > 2 {
		t.Errorf("GetDefaultConfidence() returned invalid value %f after mixed concurrent access", conf)
	}
	if timeout := GetTimeout(); timeout < 0 {
		t.Errorf("GetTimeout() returned invalid value %d after mixed concurrent access", timeout)
	}
}

// TestConfigDefaultValues tests that default configuration values are reasonable
func TestConfigDefaultValues(t *testing.T) {
	// Test default provider
	provider := GetDefaultProvider()
	if provider == "" {
		t.Error("Default provider should not be empty")
	}

	// Test default confidence
	confidence := GetDefaultConfidence()
	if confidence < 0 || confidence > 1 {
		t.Errorf("Default confidence %f should be between 0 and 1", confidence)
	}

	// Test default timeout
	timeout := GetTimeout()
	if timeout <= 0 {
		t.Errorf("Default timeout %d should be positive", timeout)
	}
}

// TestConfigIsolation tests that config changes don't affect other tests
func TestConfigIsolation(t *testing.T) {
	t.Run("first", func(t *testing.T) {
		original := GetDefaultProvider()
		defer SetDefaultProvider(original)

		SetDefaultProvider("test1")
		if GetDefaultProvider() != "test1" {
			t.Error("Failed to set provider to test1")
		}
	})

	t.Run("second", func(t *testing.T) {
		original := GetDefaultProvider()
		defer SetDefaultProvider(original)

		SetDefaultProvider("test2")
		if GetDefaultProvider() != "test2" {
			t.Error("Failed to set provider to test2")
		}
	})

	// After both subtests, config should still work
	if GetDefaultProvider() == "" {
		t.Error("Provider should not be empty after subtests")
	}
}
