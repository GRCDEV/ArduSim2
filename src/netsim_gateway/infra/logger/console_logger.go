package logger

import (
	"log"
	"os"
)

type ConsoleLogger struct {
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{
		info:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		warn:  log.New(os.Stdout, "WARN: ", log.LstdFlags),
		error: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}
}

func (l *ConsoleLogger) Info(msg string, fields ...any) {
	l.info.Println(append([]any{msg}, fields...)...)
}

func (l *ConsoleLogger) Warn(msg string, fields ...any) {
	l.warn.Println(append([]any{msg}, fields...)...)
}

func (l *ConsoleLogger) Error(msg string, fields ...any) {
	l.error.Println(append([]any{msg}, fields...)...)
}
