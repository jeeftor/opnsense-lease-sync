// pkg/logger.go
package pkg

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

func (l LogLevel) String() string {
	switch l {
	case LogLevelError:
		return "ERROR"
	case LogLevelWarn:
		return "WARN"
	case LogLevelInfo:
		return "INFO"
	case LogLevelDebug:
		return "DEBUG"
	default:
		return "UNKNOWN"
	}
}

func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "ERROR":
		return LogLevelError
	case "WARN":
		return LogLevelWarn
	case "INFO":
		return LogLevelInfo
	case "DEBUG":
		return LogLevelDebug
	default:
		return LogLevelInfo
	}
}

type Logger interface {
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

type LogConfig struct {
	// Log level (ERROR, WARN, INFO, DEBUG)
	Level LogLevel
	// Optional file path for logging to file
	FilePath string
	// Log rotation settings
	MaxSize    int  // megabytes
	MaxBackups int  // number of backups
	MaxAge     int  // days
	Compress   bool // compress old logs
}

type ServiceLogger struct {
	syslog *syslog.Writer
	level  LogLevel
}

type FileLogger struct {
	logger *log.Logger
	level  LogLevel
}

func NewLogger(cfg LogConfig) (Logger, error) {
	// If running as a daemon, use syslog unless file logging is explicitly configured
	if os.Getppid() == 1 && cfg.FilePath == "" {
		syslogWriter, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "dhcp-adguard-sync")
		if err != nil {
			return nil, fmt.Errorf("initializing syslog: %w", err)
		}
		return &ServiceLogger{
			syslog: syslogWriter,
			level:  cfg.Level,
		}, nil
	}

	// If file path is specified, use file logging with rotation
	if cfg.FilePath != "" {
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(cfg.FilePath), 0755); err != nil {
			return nil, fmt.Errorf("creating log directory: %w", err)
		}

		// Configure log rotation
		rotator := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,    // megabytes
			MaxBackups: cfg.MaxBackups, // number of backups
			MaxAge:     cfg.MaxAge,     // days
			Compress:   cfg.Compress,   // compress old logs
		}

		logger := log.New(rotator, "", log.LstdFlags)
		return &FileLogger{
			logger: logger,
			level:  cfg.Level,
		}, nil
	}

	// Default to stdout for CLI usage
	logger := log.New(os.Stdout, "", log.LstdFlags)
	return &FileLogger{
		logger: logger,
		level:  cfg.Level,
	}, nil
}

// ServiceLogger implementation
func (l *ServiceLogger) Error(msg string) {
	if l.level >= LogLevelError {
		l.syslog.Err(msg)
	}
}

func (l *ServiceLogger) Warn(msg string) {
	if l.level >= LogLevelWarn {
		l.syslog.Warning(msg)
	}
}

func (l *ServiceLogger) Info(msg string) {
	if l.level >= LogLevelInfo {
		l.syslog.Info(msg)
	}
}

func (l *ServiceLogger) Debug(msg string) {
	if l.level >= LogLevelDebug {
		l.syslog.Debug(msg)
	}
}

// FileLogger implementation
func (l *FileLogger) log(level LogLevel, msg string) {
	if l.level >= level {
		l.logger.Output(2, fmt.Sprintf("[%s] %s", level, msg))
	}
}

func (l *FileLogger) Error(msg string) {
	l.log(LogLevelError, msg)
}

func (l *FileLogger) Warn(msg string) {
	l.log(LogLevelWarn, msg)
}

func (l *FileLogger) Info(msg string) {
	l.log(LogLevelInfo, msg)
}

func (l *FileLogger) Debug(msg string) {
	l.log(LogLevelDebug, msg)
}
