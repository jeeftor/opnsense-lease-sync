// pkg/appConfig.go
package pkg

import "time"

type Config struct {
	AdGuardURL           string
	LeasePath            string
	DryRun               bool
	Logger               Logger
	Username             string
	Password             string
	Scheme               string
	Timeout              int
	PreserveDeletedHosts bool
	Debug                bool
	NDPUpdateInterval    time.Duration

	// Logging configuration
	LogConfig LogConfig
}

// DefaultLogConfig returns the default logging configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:      LogLevelInfo,
		MaxSize:    100,  // 100 MB
		MaxBackups: 3,    // Keep 3 backups
		MaxAge:     28,   // 28 days
		Compress:   true, // Compress old logs
	}
}
