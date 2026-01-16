// Package logger provides structured logging for the Platform Orchestrator.
// It uses zerolog for high-performance, structured JSON logging.
package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	zl zerolog.Logger
}

// ctxKey is used for storing logger in context
type ctxKey struct{}

// New creates a new Logger instance
func New(level string, format string, output io.Writer) *Logger {
	if output == nil {
		output = os.Stdout
	}

	// Parse log level
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	// Configure output format
	var zl zerolog.Logger
	if format == "console" {
		zl = zerolog.New(zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}).Level(lvl).With().Timestamp().Caller().Logger()
	} else {
		zl = zerolog.New(output).Level(lvl).With().Timestamp().Caller().Logger()
	}

	return &Logger{zl: zl}
}

// WithContext returns a new context with the logger attached
func (l *Logger) WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext retrieves the logger from context
func FromContext(ctx context.Context) *Logger {
	if l, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return l
	}
	return New("info", "json", os.Stdout)
}

// With returns a new logger with the given fields
func (l *Logger) With() *LoggerContext {
	return &LoggerContext{ctx: l.zl.With()}
}

// LoggerContext is a helper for building log context
type LoggerContext struct {
	ctx zerolog.Context
}

// Str adds a string field
func (c *LoggerContext) Str(key, value string) *LoggerContext {
	c.ctx = c.ctx.Str(key, value)
	return c
}

// Int adds an integer field
func (c *LoggerContext) Int(key string, value int) *LoggerContext {
	c.ctx = c.ctx.Int(key, value)
	return c
}

// Int64 adds an int64 field
func (c *LoggerContext) Int64(key string, value int64) *LoggerContext {
	c.ctx = c.ctx.Int64(key, value)
	return c
}

// Float64 adds a float64 field
func (c *LoggerContext) Float64(key string, value float64) *LoggerContext {
	c.ctx = c.ctx.Float64(key, value)
	return c
}

// Bool adds a boolean field
func (c *LoggerContext) Bool(key string, value bool) *LoggerContext {
	c.ctx = c.ctx.Bool(key, value)
	return c
}

// Err adds an error field
func (c *LoggerContext) Err(err error) *LoggerContext {
	c.ctx = c.ctx.Err(err)
	return c
}

// Interface adds an interface{} field
func (c *LoggerContext) Interface(key string, value interface{}) *LoggerContext {
	c.ctx = c.ctx.Interface(key, value)
	return c
}

// Logger returns a new logger with the context applied
func (c *LoggerContext) Logger() *Logger {
	return &Logger{zl: c.ctx.Logger()}
}

// Debug logs at debug level
func (l *Logger) Debug() *LogEvent {
	return &LogEvent{event: l.zl.Debug()}
}

// Info logs at info level
func (l *Logger) Info() *LogEvent {
	return &LogEvent{event: l.zl.Info()}
}

// Warn logs at warn level
func (l *Logger) Warn() *LogEvent {
	return &LogEvent{event: l.zl.Warn()}
}

// Error logs at error level
func (l *Logger) Error() *LogEvent {
	return &LogEvent{event: l.zl.Error()}
}

// Fatal logs at fatal level and exits
func (l *Logger) Fatal() *LogEvent {
	return &LogEvent{event: l.zl.Fatal()}
}

// LogEvent wraps a zerolog.Event
type LogEvent struct {
	event *zerolog.Event
}

// Str adds a string field
func (e *LogEvent) Str(key, value string) *LogEvent {
	e.event = e.event.Str(key, value)
	return e
}

// Int adds an integer field
func (e *LogEvent) Int(key string, value int) *LogEvent {
	e.event = e.event.Int(key, value)
	return e
}

// Int64 adds an int64 field
func (e *LogEvent) Int64(key string, value int64) *LogEvent {
	e.event = e.event.Int64(key, value)
	return e
}

// Float64 adds a float64 field
func (e *LogEvent) Float64(key string, value float64) *LogEvent {
	e.event = e.event.Float64(key, value)
	return e
}

// Bool adds a boolean field
func (e *LogEvent) Bool(key string, value bool) *LogEvent {
	e.event = e.event.Bool(key, value)
	return e
}

// Err adds an error field
func (e *LogEvent) Err(err error) *LogEvent {
	e.event = e.event.Err(err)
	return e
}

// Interface adds an interface{} field
func (e *LogEvent) Interface(key string, value interface{}) *LogEvent {
	e.event = e.event.Interface(key, value)
	return e
}

// Dur adds a duration field
func (e *LogEvent) Dur(key string, d time.Duration) *LogEvent {
	e.event = e.event.Dur(key, d)
	return e
}

// Time adds a time field
func (e *LogEvent) Time(key string, t time.Time) *LogEvent {
	e.event = e.event.Time(key, t)
	return e
}

// Msg sends the log event with a message
func (e *LogEvent) Msg(msg string) {
	e.event.Msg(msg)
}

// Msgf sends the log event with a formatted message
func (e *LogEvent) Msgf(format string, v ...interface{}) {
	e.event.Msgf(format, v...)
}

// Send sends the log event without a message
func (e *LogEvent) Send() {
	e.event.Send()
}
