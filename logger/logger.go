package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogFormat represents the output format of logs
type LogFormat string

const (
	JSONFormat LogFormat = "json"
	TextFormat LogFormat = "text"
)

// Logger is the main logger struct
type Logger struct {
	level      LogLevel
	format     LogFormat
	output     io.Writer
	mu         sync.Mutex
	fields     map[string]interface{}
	timeFormat string
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	File      string                 `json:"file,omitempty"`
	Line      int                    `json:"line,omitempty"`
}

var (
	defaultLogger *Logger
	once          sync.Once
)

// Config holds logger configuration
type Config struct {
	Level      LogLevel
	Format     LogFormat
	Output     io.Writer
	TimeFormat string
}

// New creates a new Logger instance
func New(config Config) *Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}
	if config.TimeFormat == "" {
		config.TimeFormat = time.RFC3339
	}

	return &Logger{
		level:      config.Level,
		format:     config.Format,
		output:     config.Output,
		fields:     make(map[string]interface{}),
		timeFormat: config.TimeFormat,
	}
}

// Init initializes the default logger (call once at startup)
func Init(config Config) {
	once.Do(func() {
		defaultLogger = New(config)
	})
}

// GetDefaultLogger returns the default logger instance
func GetDefaultLogger() *Logger {
	if defaultLogger == nil {
		// Initialize with default config if not already initialized
		Init(Config{
			Level:  INFO,
			Format: TextFormat,
			Output: os.Stdout,
		})
	}
	return defaultLogger
}

// WithField adds a field to the logger context (creates new logger for immutability)
// Use this when you want to preserve the original logger unchanged
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := &Logger{
		level:      l.level,
		format:     l.format,
		output:     l.output,
		fields:     make(map[string]interface{}),
		timeFormat: l.timeFormat,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

// WithFields adds multiple fields to the logger context (creates new logger for immutability)
// Use this when you want to preserve the original logger unchanged
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:      l.level,
		format:     l.format,
		output:     l.output,
		fields:     make(map[string]interface{}),
		timeFormat: l.timeFormat,
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	if level < l.level {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Format message with args if provided
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	// Copy fields to avoid race conditions if logger is reused
	fieldsCopy := make(map[string]interface{}, len(l.fields))
	for k, v := range l.fields {
		fieldsCopy[k] = v
	}

	entry := LogEntry{
		Timestamp: time.Now().Format(l.timeFormat),
		Level:     level.String(),
		Message:   message,
		Fields:    fieldsCopy,
	}

	// Add caller information for ERROR and FATAL levels
	if level >= ERROR {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			// Get just the filename, not the full path
			parts := strings.Split(file, "/")
			entry.File = parts[len(parts)-1]
			entry.Line = line
		}
	}

	var output string
	if l.format == JSONFormat {
		output = l.formatJSON(entry)
	} else {
		output = l.formatText(entry)
	}

	fmt.Fprintln(l.output, output)

	// Exit on FATAL
	if level == FATAL {
		os.Exit(1)
	}
}

// formatJSON formats the log entry as JSON
func (l *Logger) formatJSON(entry LogEntry) string {
	bytes, err := json.Marshal(entry)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal log entry: %v"}`, err)
	}
	return string(bytes)
}

// formatText formats the log entry as human-readable text
func (l *Logger) formatText(entry LogEntry) string {
	var builder strings.Builder

	// Timestamp and level with color coding for terminals
	builder.WriteString(fmt.Sprintf("[%s] %-5s ", entry.Timestamp, entry.Level))

	// Message
	builder.WriteString(entry.Message)

	// Fields
	if len(entry.Fields) > 0 {
		builder.WriteString(" | ")
		first := true
		for k, v := range entry.Fields {
			if !first {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("%s=%v", k, v))
			first = false
		}
	}

	// File and line for errors
	if entry.File != "" {
		builder.WriteString(fmt.Sprintf(" [%s:%d]", entry.File, entry.Line))
	}

	return builder.String()
}

// Debug logs a debug message
func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, message, args...)
}

// Info logs an info message
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, message, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(WARN, message, args...)
}

// Error logs an error message
func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, message, args...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(message string, args ...interface{}) {
	l.log(FATAL, message, args...)
}

// Package-level convenience functions using the default logger

// Debug logs a debug message using the default logger
func Debug(message string, args ...interface{}) {
	GetDefaultLogger().Debug(message, args...)
}

// Info logs an info message using the default logger
func Info(message string, args ...interface{}) {
	GetDefaultLogger().Info(message, args...)
}

// Warn logs a warning message using the default logger
func Warn(message string, args ...interface{}) {
	GetDefaultLogger().Warn(message, args...)
}

// Error logs an error message using the default logger
func Error(message string, args ...interface{}) {
	GetDefaultLogger().Error(message, args...)
}

// Fatal logs a fatal message and exits using the default logger
func Fatal(message string, args ...interface{}) {
	GetDefaultLogger().Fatal(message, args...)
}

// WithField adds a field using the default logger
func WithField(key string, value interface{}) *Logger {
	return GetDefaultLogger().WithField(key, value)
}

// WithFields adds multiple fields using the default logger
func WithFields(fields map[string]interface{}) *Logger {
	return GetDefaultLogger().WithFields(fields)
}

// ParseLogLevel converts a string to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN", "WARNING":
		return WARN
	case "ERROR":
		return ERROR
	case "FATAL":
		return FATAL
	default:
		return INFO
	}
}

// ParseLogFormat converts a string to LogFormat
func ParseLogFormat(format string) LogFormat {
	switch strings.ToLower(format) {
	case "json":
		return JSONFormat
	case "text":
		return TextFormat
	default:
		return TextFormat
	}
}
