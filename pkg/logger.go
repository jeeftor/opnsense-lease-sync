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

type CLILogger struct{}

func NewLogger() (Logger, error) {
	if os.Getppid() == 1 {
		syslogWriter, err := syslog.New(syslog.LOG_INFO|syslog.LOG_DAEMON, "dhcp-adguard-sync")
		if err != nil {
			return nil, err
		}
		return &ServiceLogger{syslog: syslogWriter}, nil
	}
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return &CLILogger{}, nil
}

func (l *ServiceLogger) Info(msg string) {
	l.syslog.Info(msg)
}

func (l *ServiceLogger) Error(msg string) {
	l.syslog.Err(msg)
}

func (l *CLILogger) Info(msg string) {
	log.Printf("INFO: %s\n", msg)
}

func (l *CLILogger) Error(msg string) {
	log.Printf("ERROR: %s\n", msg)
}
