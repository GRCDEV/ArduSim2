package logger

import (
	"log"
	"os"
)

type ConsoleLogger struct {
	level string
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewConsoleLogger(level string) *ConsoleLogger {
	return &ConsoleLogger{
		level: level,
		info:  log.New(os.Stdout, "INFO: ", log.LstdFlags),
		warn:  log.New(os.Stdout, "WARN: ", log.LstdFlags),
		error: log.New(os.Stderr, "ERROR: ", log.LstdFlags),
	}
}

func (l *ConsoleLogger) Info(msg string, fields ...any) {
	if l.level == "debug" || l.level == "info" {
		l.info.Println(append([]any{msg}, fields...)...)
	}
}

func (l *ConsoleLogger) Warn(msg string, fields ...any) {
	if l.level != "error" {
		l.warn.Println(append([]any{msg}, fields...)...)
	}
}

func (l *ConsoleLogger) Error(msg string, fields ...any) {
	l.error.Println(append([]any{msg}, fields...)...)
}
