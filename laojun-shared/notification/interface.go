package notification

import (
	"context"
	"fmt"
	"time"
)

// NOTIFICATION defines the interface for notification operations.
// All implementations must be thread-safe.
type NOTIFICATION interface {
	// TODO: Define your interface methods here
	// Example:
	// Get(ctx context.Context, key string) (interface{}, error)
	// Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// Config defines the configuration for notification.
type Config struct {
	Enabled bool          `yaml:"enabled" env:"NOTIFICATION_ENABLED" default:"true"`
	Debug   bool          `yaml:"debug" env:"NOTIFICATION_DEBUG" default:"false"`
	Timeout time.Duration `yaml:"timeout" env:"NOTIFICATION_TIMEOUT" default:"30s"`
	
	// TODO: Add your specific configuration fields here
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	// TODO: Add your validation logic here
	return nil
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Enabled: true,
		Debug:   false,
		Timeout: 30 * time.Second,
	}
}
