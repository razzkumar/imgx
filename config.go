package imgx

import (
	"os"
	"sync"
)

// Config holds global configuration for imgx
type Config struct {
	AddMetadata   bool
	DefaultAuthor string
	mu            sync.RWMutex
}

var globalConfig = &Config{
	AddMetadata:   true,
	DefaultAuthor: "", // Empty means use Author from load.go
}

func init() {
	// Check environment variable for metadata
	if env := os.Getenv("IMGX_ADD_METADATA"); env == "false" || env == "0" {
		globalConfig.AddMetadata = false
	}

	// Check environment variable for default author
	if env := os.Getenv("IMGX_DEFAULT_AUTHOR"); env != "" {
		globalConfig.DefaultAuthor = env
	}
}

// SetAddMetadata configures whether to add metadata globally
func SetAddMetadata(enabled bool) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.AddMetadata = enabled
}

// GetAddMetadata returns the global metadata setting
func GetAddMetadata() bool {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.AddMetadata
}

// SetDefaultAuthor sets the default artist/creator name globally
// This can be overridden per-image using WithAuthor() or SetAuthor()
// Empty string means use the default "razzkumar"
func SetDefaultAuthor(author string) {
	globalConfig.mu.Lock()
	defer globalConfig.mu.Unlock()
	globalConfig.DefaultAuthor = author
}

// GetDefaultAuthor returns the global default author setting
func GetDefaultAuthor() string {
	globalConfig.mu.RLock()
	defer globalConfig.mu.RUnlock()
	return globalConfig.DefaultAuthor
}
