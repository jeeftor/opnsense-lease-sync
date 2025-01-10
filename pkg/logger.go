package pkg

import (
	"log"
	"log/syslog"
	"os"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
}

type ServiceLogger struct {
	syslog *syslog.Writer
}

type CLILogger struct {
	logger *log.Logger
}

func NewLogger() (Logger, error) {
	if os.Getppid() == 1 {
		syslogWriter, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "dhcp-adguard-sync")
		if err != nil {
			return nil, err
		}
		return &ServiceLogger{syslog: syslogWriter}, nil
	}

	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	logger.SetFlags(log.LstdFlags | log.Lshortfile)
	return &CLILogger{logger: logger}, nil
}

func (l *ServiceLogger) Info(msg string) {
	l.syslog.Info(msg)
}

func (l *ServiceLogger) Error(msg string) {
	l.syslog.Err(msg)
}

func (l *CLILogger) Info(msg string) {
	// Skip 2 frames in the call stack to show the actual caller
	l.logger.Output(2, "INFO: "+msg)
}

func (l *CLILogger) Error(msg string) {
	// Skip 2 frames in the call stack to show the actual caller
	l.logger.Output(2, "ERROR: "+msg)
}
