package imgx

import (
	"os"
	"sync"
)

// Config holds global configuration for imgx
type Config struct {
	AddMetadata bool
	mu          sync.RWMutex
}

var globalConfig = &Config{
	AddMetadata: true,
}

func init() {
	// Check environment variable
	if env := os.Getenv("IMGX_ADD_METADATA"); env == "false" || env == "0" {
		globalConfig.AddMetadata = false
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
