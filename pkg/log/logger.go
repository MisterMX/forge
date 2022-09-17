package log

import (
	"fmt"
	"io"
	"os"
)

// A Logger logs messages. Messages may be supplemented by structured data.
type Logger interface {
	// Info logs a message.
	Info(msg string)
	// Infof logs a message that is formatted with fmt.Sprintf.
	Infof(format string, values ...any)
	// Debug logs a debug message.
	Debug(msg string)
	// Debugf logs a debug message that is formatted with fmt.Sprintf.
	Debugf(format string, values ...any)
}

// NewNoopLogger returns a Logger that does nothing.
func NewNoopLogger() Logger { return &noopLogger{} }

type noopLogger struct{}

func (l noopLogger) Info(msg string)                     {}
func (l noopLogger) Infof(format string, values ...any)  {}
func (l noopLogger) Debug(msg string)                    {}
func (l noopLogger) Debugf(format string, values ...any) {}

type logger struct {
	out   io.Writer
	level int
}

// NewInfoLogger creates a new Logger that only writes info messages to stderr.
func NewInfoLogger() Logger {
	return logger{
		out:   os.Stderr,
		level: 0,
	}
}

// NewInfoLogger creates a new Logger that writes all messages to stderr.
func NewDebugLogger() Logger {
	return logger{
		out:   os.Stderr,
		level: 1,
	}
}

func (l logger) log(level int, msg string) {
	if level <= l.level {
		fmt.Fprint(l.out, msg)
	}
}

func (l logger) logf(level int, format string, values ...any) {
	if level <= l.level {
		fmt.Fprintf(l.out, format, values...)
	}
}

func (l logger) Info(msg string) {
	l.log(0, msg)
}

func (l logger) Infof(format string, values ...any) {
	l.logf(0, format, values...)
}

func (l logger) Debug(msg string) {
	l.log(1, msg)
}

func (l logger) Debugf(format string, values ...any) {
	l.logf(1, format, values...)
}
