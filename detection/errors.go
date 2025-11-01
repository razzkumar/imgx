package detection

import (
	"errors"
	"fmt"
)

// DetectionError wraps provider-specific errors
type DetectionError struct {
	Provider string // Provider name
	Message  string // Error message
	Code     string // Provider-specific error code
	Err      error  // Underlying error
}

// Error implements the error interface
func (e *DetectionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Provider, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Provider, e.Message)
}

// Unwrap returns the underlying error
func (e *DetectionError) Unwrap() error {
	return e.Err
}

// Common detection errors
var (
	// ErrProviderNotConfigured indicates missing credentials or configuration
	ErrProviderNotConfigured = errors.New("provider not configured or credentials missing")

	// ErrInvalidFeature indicates an unsupported feature was requested
	ErrInvalidFeature = errors.New("invalid or unsupported detection feature")

	// ErrImageTooLarge indicates the image exceeds provider size limits
	ErrImageTooLarge = errors.New("image exceeds provider size limits")

	// ErrUnsupportedFormat indicates the image format is not supported
	ErrUnsupportedFormat = errors.New("image format not supported by provider")

	// ErrAPIError indicates a generic API error
	ErrAPIError = errors.New("provider API error")

	// ErrRateLimit indicates API rate limit was exceeded
	ErrRateLimit = errors.New("provider API rate limit exceeded")

	// ErrInvalidAPIKey indicates authentication failed
	ErrInvalidAPIKey = errors.New("invalid API key or credentials")

	// ErrNetworkError indicates a network connectivity issue
	ErrNetworkError = errors.New("network error communicating with provider")

	// ErrInvalidImage indicates the image data is invalid
	ErrInvalidImage = errors.New("invalid image data")

	// ErrContextCanceled indicates the context was canceled
	ErrContextCanceled = errors.New("detection canceled by context")
)

// NewDetectionError creates a new DetectionError
func NewDetectionError(provider, message string, err error) *DetectionError {
	return &DetectionError{
		Provider: provider,
		Message:  message,
		Err:      err,
	}
}

// IsNotConfigured checks if error is ErrProviderNotConfigured
func IsNotConfigured(err error) bool {
	return errors.Is(err, ErrProviderNotConfigured)
}

// IsRateLimit checks if error is ErrRateLimit
func IsRateLimit(err error) bool {
	return errors.Is(err, ErrRateLimit)
}

// IsInvalidAPIKey checks if error is ErrInvalidAPIKey
func IsInvalidAPIKey(err error) bool {
	return errors.Is(err, ErrInvalidAPIKey)
}
