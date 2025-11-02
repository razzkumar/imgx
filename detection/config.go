package detection

import (
	"os"
	"strconv"
	"sync"
)

// Config holds global detection configuration
type Config struct {
	// DefaultProvider is the default detection provider
	DefaultProvider string

	// DefaultConfidence is the default minimum confidence threshold
	DefaultConfidence float32

	// MaxConcurrentRequests limits concurrent API requests
	MaxConcurrentRequests int

	// CacheResults enables result caching (future feature)
	CacheResults bool

	// Timeout specifies API request timeout in seconds
	Timeout int

	mu sync.RWMutex
}

var globalConfig = &Config{
	DefaultProvider:       "ollama",
	DefaultConfidence:     0.5,
	MaxConcurrentRequests: 10,
	CacheResults:          false,
	Timeout:               30,
}

func init() {
	// Load configuration from environment variables
	if provider := os.Getenv("IMGX_DETECTION_PROVIDER"); provider != "" {
		globalConfig.DefaultProvider = provider
	}

	if confidence := os.Getenv("IMGX_DETECTION_CONFIDENCE"); confidence != "" {
		if val, err := strconv.ParseFloat(confidence, 32); err == nil {
			globalConfig.DefaultConfidence = float32(val)
		}
	}

	if timeout := os.Getenv("IMGX_DETECTION_TIMEOUT"); timeout != "" {
		if val, err := strconv.Atoi(timeout); err == nil {
			globalConfig.Timeout = val
		}
	}

	if cache := os.Getenv("IMGX_DETECTION_CACHE"); cache != "" {
		globalConfig.CacheResults = cache == "true" || cache == "1"
	}
}

// GetDefaultProvider returns the default provider name
func GetDefaultProvider() string {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.DefaultProvider
}

// SetDefaultProvider sets the default provider name
func SetDefaultProvider(provider string) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.DefaultProvider = provider
}

// GetDefaultConfidence returns the default confidence threshold
func GetDefaultConfidence() float32 {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.DefaultConfidence
}

// SetDefaultConfidence sets the default confidence threshold
func SetDefaultConfidence(confidence float32) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.DefaultConfidence = confidence
}

// GetTimeout returns the API request timeout
func GetTimeout() int {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.Timeout
}

// SetTimeout sets the API request timeout
func SetTimeout(timeout int) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.Timeout = timeout
}
