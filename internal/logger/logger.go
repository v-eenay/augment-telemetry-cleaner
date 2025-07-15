package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
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
	default:
		return "UNKNOWN"
	}
}

// Logger represents a logger instance
type Logger struct {
	logger   *log.Logger
	level    LogLevel
	file     *os.File
	callback func(level LogLevel, message string)
}

// NewLogger creates a new logger instance
func NewLogger(logDir string, callback func(LogLevel, string)) (*Logger, error) {
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create log file with timestamp
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	logFile := filepath.Join(logDir, fmt.Sprintf("augment_cleaner_%s.log", timestamp))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(file, os.Stdout)

	logger := &Logger{
		logger:   log.New(multiWriter, "", log.LstdFlags),
		level:    INFO,
		file:     file,
		callback: callback,
	}

	logger.Info("Logger initialized")
	return logger, nil
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// log writes a log message with the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	message := fmt.Sprintf(format, args...)
	logEntry := fmt.Sprintf("[%s] %s", level.String(), message)
	
	l.logger.Println(logEntry)
	
	// Call callback if provided (for GUI updates)
	if l.callback != nil {
		l.callback(level, message)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// LogOperation logs the start of an operation
func (l *Logger) LogOperation(operation string) {
	l.Info("=== Starting operation: %s ===", operation)
}

// LogOperationResult logs the result of an operation
func (l *Logger) LogOperationResult(operation string, success bool, details string) {
	if success {
		l.Info("=== Operation completed successfully: %s ===", operation)
		if details != "" {
			l.Info("Details: %s", details)
		}
	} else {
		l.Error("=== Operation failed: %s ===", operation)
		if details != "" {
			l.Error("Error details: %s", details)
		}
	}
}

// LogBackupCreated logs when a backup is created
func (l *Logger) LogBackupCreated(originalPath, backupPath string) {
	l.Info("Backup created: %s -> %s", originalPath, backupPath)
}

// LogFileOperation logs file operations
func (l *Logger) LogFileOperation(operation, path string, success bool, err error) {
	if success {
		l.Info("File %s: %s", operation, path)
	} else {
		l.Error("Failed to %s file %s: %v", operation, path, err)
	}
}
