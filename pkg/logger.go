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

// LogLevel represents the severity of a log message
type LogLevel int

const (
	LogLevelError LogLevel = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// String returns the string representation of the log level
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

// ParseLogLevel converts a string to LogLevel
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

// Logger defines the interface for logging operations
type Logger interface {
	Error(msg string)
	Warn(msg string)
	Info(msg string)
	Debug(msg string)
}

type LogConfig struct {
	Level          LogLevel
	FilePath       string
	SyslogFacility string
	SyslogOnly     bool
	MaxSize        int
	MaxBackups     int
	MaxAge         int
	Compress       bool
}

type DualLogger struct {
	fileLogger *log.Logger
	sysLogger  *syslog.Writer
	level      LogLevel
	rotator    *lumberjack.Logger
}

func NewLogger(cfg LogConfig) (Logger, error) {
	var dl DualLogger
	dl.level = cfg.Level

	// Default to local3 facility if not specified
	facility := cfg.SyslogFacility
	if facility == "" {
		facility = "local3"
	}

	// Setup syslog first
	priority := syslog.LOG_INFO | getFacility(facility)
	sysLogger, err := syslog.New(priority, "dhcpsync")
	if err != nil {
		return nil, fmt.Errorf("initializing syslog: %w", err)
	}
	dl.sysLogger = sysLogger

	// Setup file/stdout logging based on configuration
	if cfg.SyslogOnly {
		// Syslog-only mode: no file or stdout logging
		dl.fileLogger = nil
	} else if cfg.FilePath != "" {
		// File logging mode: log to specified file
		logDir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return nil, fmt.Errorf("creating log directory: %w", err)
		}

		// Setup file logging with rotation
		dl.rotator = &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}

		// Create file logger
		dl.fileLogger = log.New(dl.rotator, "", log.LstdFlags)
	} else {
		// Default mode: stdout + syslog (dual logging)
		dl.fileLogger = log.New(os.Stdout, "", log.LstdFlags)
	}

	return &dl, nil
}

func (l *DualLogger) log(level LogLevel, msg string) {
	if l.level >= level {
		// Log to file/stdout if we have a file logger (unless syslog-only mode)
		if l.fileLogger != nil {
			l.fileLogger.Output(2, fmt.Sprintf("[%s] %s", level, msg))
		}

		// Always log to syslog with appropriate level
		if l.sysLogger != nil {
			switch level {
			case LogLevelError:
				l.sysLogger.Err(msg)
			case LogLevelWarn:
				l.sysLogger.Warning(msg)
			case LogLevelInfo:
				l.sysLogger.Info(msg)
			case LogLevelDebug:
				l.sysLogger.Debug(msg)
			}
		}
	}
}

func (l *DualLogger) Error(msg string) {
	l.log(LogLevelError, msg)
}

func (l *DualLogger) Warn(msg string) {
	l.log(LogLevelWarn, msg)
}

func (l *DualLogger) Info(msg string) {
	l.log(LogLevelInfo, msg)
}

func (l *DualLogger) Debug(msg string) {
	l.log(LogLevelDebug, msg)
}

// getFacility converts a facility string to syslog.Priority
func getFacility(facility string) syslog.Priority {
	switch strings.ToLower(facility) {
	case "local0":
		return syslog.LOG_LOCAL0
	case "local1":
		return syslog.LOG_LOCAL1
	case "local2":
		return syslog.LOG_LOCAL2
	case "local3":
		return syslog.LOG_LOCAL3
	case "local4":
		return syslog.LOG_LOCAL4
	case "local5":
		return syslog.LOG_LOCAL5
	case "local6":
		return syslog.LOG_LOCAL6
	case "local7":
		return syslog.LOG_LOCAL7
	default:
		return syslog.LOG_LOCAL3 // Default to local3
	}
}
