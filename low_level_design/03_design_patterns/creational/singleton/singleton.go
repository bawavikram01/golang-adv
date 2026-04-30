// Package singleton demonstrates the Singleton pattern.
//
// INTENT: Ensure a type has only ONE instance and provide a global point of access.
//
// Go idiom: Use sync.Once to guarantee thread-safe lazy initialization.
// No mutexes, no double-checked locking — sync.Once handles it perfectly.
//
// WHEN TO USE:
//   - Configuration manager
//   - Connection pool
//   - Logger
//   - Cache
//
// WARNING: Singletons make testing harder. Prefer dependency injection
// when possible. Use singletons only for truly global, shared resources.
package singleton

import "sync"

// ──────────────────────────────────────────────
// The Singleton — Config Manager
// ──────────────────────────────────────────────

type Config struct {
	AppName    string
	Version    string
	Debug      bool
	MaxRetries int
	settings   map[string]string
}

var (
	configInstance *Config
	configOnce     sync.Once
)

// GetConfig returns the single Config instance, creating it on first call.
func GetConfig() *Config {
	configOnce.Do(func() {
		configInstance = &Config{
			AppName:    "LLD-App",
			Version:    "1.0.0",
			Debug:      false,
			MaxRetries: 3,
			settings:   make(map[string]string),
		}
	})
	return configInstance
}

func (c *Config) Set(key, value string) {
	c.settings[key] = value
}

func (c *Config) Get(key string) string {
	return c.settings[key]
}

// ──────────────────────────────────────────────
// Resettable singleton (useful for testing)
// ──────────────────────────────────────────────

// ResetConfig is ONLY for tests — resets the singleton.
func ResetConfig() {
	configOnce = sync.Once{}
	configInstance = nil
}
