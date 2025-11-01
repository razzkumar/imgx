package detection

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// TestNewDetectionError tests DetectionError creation
func TestNewDetectionError(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		message  string
		err      error
	}{
		{
			name:     "with underlying error",
			provider: "gemini",
			message:  "API error",
			err:      errors.New("connection timeout"),
		},
		{
			name:     "without underlying error",
			provider: "aws",
			message:  "invalid credentials",
			err:      nil,
		},
		{
			name:     "empty provider",
			provider: "",
			message:  "test error",
			err:      errors.New("test"),
		},
		{
			name:     "empty message",
			provider: "openai",
			message:  "",
			err:      errors.New("test"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detErr := NewDetectionError(tt.provider, tt.message, tt.err)

			if detErr == nil {
				t.Fatal("NewDetectionError() returned nil")
			}

			if detErr.Provider != tt.provider {
				t.Errorf("Provider = %q, want %q", detErr.Provider, tt.provider)
			}

			if detErr.Message != tt.message {
				t.Errorf("Message = %q, want %q", detErr.Message, tt.message)
			}

			if detErr.Err != tt.err {
				t.Errorf("Err = %v, want %v", detErr.Err, tt.err)
			}
		})
	}
}

// TestDetectionErrorError tests Error() method
func TestDetectionErrorError(t *testing.T) {
	tests := []struct {
		name       string
		detErr     *DetectionError
		wantPrefix string
		wantSuffix string
	}{
		{
			name: "with underlying error",
			detErr: &DetectionError{
				Provider: "gemini",
				Message:  "API error",
				Err:      errors.New("connection timeout"),
			},
			wantPrefix: "[gemini]",
			wantSuffix: "connection timeout",
		},
		{
			name: "without underlying error",
			detErr: &DetectionError{
				Provider: "aws",
				Message:  "invalid credentials",
				Err:      nil,
			},
			wantPrefix: "[aws]",
			wantSuffix: "invalid credentials",
		},
		{
			name: "empty provider",
			detErr: &DetectionError{
				Provider: "",
				Message:  "test error",
				Err:      nil,
			},
			wantPrefix: "[]",
			wantSuffix: "test error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.detErr.Error()

			if !strings.Contains(errStr, tt.wantPrefix) {
				t.Errorf("Error() = %q, does not contain prefix %q", errStr, tt.wantPrefix)
			}

			if !strings.Contains(errStr, tt.wantSuffix) {
				t.Errorf("Error() = %q, does not contain suffix %q", errStr, tt.wantSuffix)
			}
		})
	}
}

// TestDetectionErrorUnwrap tests Unwrap() method
func TestDetectionErrorUnwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")

	tests := []struct {
		name    string
		detErr  *DetectionError
		wantErr error
	}{
		{
			name: "with underlying error",
			detErr: &DetectionError{
				Provider: "gemini",
				Message:  "API error",
				Err:      underlyingErr,
			},
			wantErr: underlyingErr,
		},
		{
			name: "without underlying error",
			detErr: &DetectionError{
				Provider: "aws",
				Message:  "invalid credentials",
				Err:      nil,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.detErr.Unwrap()

			if got != tt.wantErr {
				t.Errorf("Unwrap() = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

// TestErrorConstants tests that all error constants are defined
func TestErrorConstants(t *testing.T) {
	constants := []struct {
		name string
		err  error
	}{
		{"ErrProviderNotConfigured", ErrProviderNotConfigured},
		{"ErrInvalidFeature", ErrInvalidFeature},
		{"ErrImageTooLarge", ErrImageTooLarge},
		{"ErrUnsupportedFormat", ErrUnsupportedFormat},
		{"ErrAPIError", ErrAPIError},
		{"ErrRateLimit", ErrRateLimit},
		{"ErrInvalidAPIKey", ErrInvalidAPIKey},
		{"ErrNetworkError", ErrNetworkError},
		{"ErrInvalidImage", ErrInvalidImage},
		{"ErrContextCanceled", ErrContextCanceled},
	}

	for _, c := range constants {
		t.Run(c.name, func(t *testing.T) {
			if c.err == nil {
				t.Errorf("%s is nil", c.name)
			}

			if c.err.Error() == "" {
				t.Errorf("%s has empty error message", c.name)
			}
		})
	}
}

// TestIsNotConfigured tests IsNotConfigured helper function
func TestIsNotConfigured(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is not configured error",
			err:  ErrProviderNotConfigured,
			want: true,
		},
		{
			name: "wrapped not configured error",
			err:  fmt.Errorf("wrapper: %w", ErrProviderNotConfigured),
			want: true,
		},
		{
			name: "detection error wrapping not configured",
			err: &DetectionError{
				Provider: "gemini",
				Message:  "not configured",
				Err:      ErrProviderNotConfigured,
			},
			want: true,
		},
		{
			name: "different error",
			err:  ErrRateLimit,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotConfigured(tt.err)
			if got != tt.want {
				t.Errorf("IsNotConfigured(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// TestIsRateLimit tests IsRateLimit helper function
func TestIsRateLimit(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is rate limit error",
			err:  ErrRateLimit,
			want: true,
		},
		{
			name: "wrapped rate limit error",
			err:  fmt.Errorf("wrapper: %w", ErrRateLimit),
			want: true,
		},
		{
			name: "detection error wrapping rate limit",
			err: &DetectionError{
				Provider: "gemini",
				Message:  "rate limited",
				Err:      ErrRateLimit,
			},
			want: true,
		},
		{
			name: "different error",
			err:  ErrProviderNotConfigured,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRateLimit(tt.err)
			if got != tt.want {
				t.Errorf("IsRateLimit(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// TestIsInvalidAPIKey tests IsInvalidAPIKey helper function
func TestIsInvalidAPIKey(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "is invalid API key error",
			err:  ErrInvalidAPIKey,
			want: true,
		},
		{
			name: "wrapped invalid API key error",
			err:  fmt.Errorf("wrapper: %w", ErrInvalidAPIKey),
			want: true,
		},
		{
			name: "detection error wrapping invalid API key",
			err: &DetectionError{
				Provider: "openai",
				Message:  "invalid key",
				Err:      ErrInvalidAPIKey,
			},
			want: true,
		},
		{
			name: "different error",
			err:  ErrProviderNotConfigured,
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "generic error",
			err:  errors.New("some error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsInvalidAPIKey(tt.err)
			if got != tt.want {
				t.Errorf("IsInvalidAPIKey(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}

// TestErrorsIsCompatibility tests that DetectionError works with errors.Is
func TestErrorsIsCompatibility(t *testing.T) {
	baseErr := ErrProviderNotConfigured

	// Create a DetectionError wrapping the base error
	detErr := NewDetectionError("gemini", "not configured", baseErr)

	// Test that errors.Is can find the wrapped error
	if !errors.Is(detErr, baseErr) {
		t.Errorf("errors.Is(detErr, baseErr) = false, want true")
	}

	// Wrap the DetectionError further
	wrappedErr := fmt.Errorf("outer wrapper: %w", detErr)

	// Test that errors.Is can still find the base error through multiple wrappers
	if !errors.Is(wrappedErr, baseErr) {
		t.Errorf("errors.Is(wrappedErr, baseErr) = false, want true")
	}
}

// TestErrorsAsCompatibility tests that DetectionError works with errors.As
func TestErrorsAsCompatibility(t *testing.T) {
	detErr := NewDetectionError("gemini", "API error", ErrAPIError)

	// Wrap it
	wrappedErr := fmt.Errorf("wrapper: %w", detErr)

	// Test that errors.As can extract the DetectionError
	var extractedErr *DetectionError
	if !errors.As(wrappedErr, &extractedErr) {
		t.Fatal("errors.As failed to extract DetectionError")
	}

	if extractedErr.Provider != "gemini" {
		t.Errorf("extractedErr.Provider = %q, want %q", extractedErr.Provider, "gemini")
	}

	if extractedErr.Message != "API error" {
		t.Errorf("extractedErr.Message = %q, want %q", extractedErr.Message, "API error")
	}
}

// TestErrorChaining tests multiple levels of error wrapping
func TestErrorChaining(t *testing.T) {
	// Create a chain: base -> DetectionError -> fmt.Errorf -> another fmt.Errorf
	baseErr := ErrInvalidImage
	detErr := NewDetectionError("aws", "validation failed", baseErr)
	wrapped1 := fmt.Errorf("processing: %w", detErr)
	wrapped2 := fmt.Errorf("request failed: %w", wrapped1)

	// Test that errors.Is can find the base error through the entire chain
	if !errors.Is(wrapped2, baseErr) {
		t.Error("errors.Is failed to find base error through chain")
	}

	// Test that errors.As can extract the DetectionError
	var extractedErr *DetectionError
	if !errors.As(wrapped2, &extractedErr) {
		t.Error("errors.As failed to extract DetectionError from chain")
	}

	if extractedErr.Provider != "aws" {
		t.Errorf("extractedErr.Provider = %q, want %q", extractedErr.Provider, "aws")
	}
}

// TestErrorConstantsUnique tests that error constants are unique
func TestErrorConstantsUnique(t *testing.T) {
	allErrors := []error{
		ErrProviderNotConfigured,
		ErrInvalidFeature,
		ErrImageTooLarge,
		ErrUnsupportedFormat,
		ErrAPIError,
		ErrRateLimit,
		ErrInvalidAPIKey,
		ErrNetworkError,
		ErrInvalidImage,
		ErrContextCanceled,
	}

	// Check that no two errors are the same
	for i := 0; i < len(allErrors); i++ {
		for j := i + 1; j < len(allErrors); j++ {
			if errors.Is(allErrors[i], allErrors[j]) {
				t.Errorf("Error %v is identical to error %v", allErrors[i], allErrors[j])
			}
		}
	}
}

// TestDetectionErrorNilErr tests DetectionError with nil underlying error
func TestDetectionErrorNilErr(t *testing.T) {
	detErr := &DetectionError{
		Provider: "test",
		Message:  "test message",
		Err:      nil,
	}

	// Error() should not panic with nil Err
	errStr := detErr.Error()
	if errStr == "" {
		t.Error("Error() returned empty string")
	}

	// Unwrap() should return nil
	if unwrapped := detErr.Unwrap(); unwrapped != nil {
		t.Errorf("Unwrap() = %v, want nil", unwrapped)
	}

	// errors.Is with nil Err should not panic
	if errors.Is(detErr, ErrProviderNotConfigured) {
		t.Error("errors.Is returned true for unrelated error")
	}
}
