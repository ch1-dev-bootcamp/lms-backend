package logger

import (
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

// Config holds logger configuration
type Config struct {
	Level  string
	Format string
	Output string
}

// Initialize sets up the logger with the given configuration
func Initialize(level, format, output string) {
	Log = logrus.New()

	// Set log level
	switch strings.ToLower(level) {
	case "debug":
		Log.SetLevel(logrus.DebugLevel)
	case "info":
		Log.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		Log.SetLevel(logrus.WarnLevel)
	case "error":
		Log.SetLevel(logrus.ErrorLevel)
	case "fatal":
		Log.SetLevel(logrus.FatalLevel)
	default:
		Log.SetLevel(logrus.InfoLevel)
	}

	// Set log format
	switch strings.ToLower(format) {
	case "json":
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	case "text":
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	default:
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output
	switch strings.ToLower(output) {
	case "stdout":
		Log.SetOutput(os.Stdout)
	case "stderr":
		Log.SetOutput(os.Stderr)
	case "file":
		// For file output, you would typically create a file
		// For now, we'll use stdout
		Log.SetOutput(os.Stdout)
	default:
		Log.SetOutput(os.Stdout)
	}

	// Add common fields
	Log = Log.WithFields(logrus.Fields{
		"service": "lms-backend",
	}).Logger
}

// Helper functions for common logging patterns
func Debug(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		Log.WithFields(fields[0]).Debug(msg)
	} else {
		Log.Debug(msg)
	}
}

func Info(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		Log.WithFields(fields[0]).Info(msg)
	} else {
		Log.Info(msg)
	}
}

func Warn(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		Log.WithFields(fields[0]).Warn(msg)
	} else {
		Log.Warn(msg)
	}
}

func Error(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		Log.WithFields(fields[0]).Error(msg)
	} else {
		Log.Error(msg)
	}
}

func Fatal(msg string, fields ...logrus.Fields) {
	if len(fields) > 0 {
		Log.WithFields(fields[0]).Fatal(msg)
	} else {
		Log.Fatal(msg)
	}
}

func WithField(key string, value interface{}) *logrus.Entry {
	return Log.WithField(key, value)
}

func WithFields(fields logrus.Fields) *logrus.Entry {
	return Log.WithFields(fields)
}
