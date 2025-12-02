package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
)

// LogLevel represents logging levels
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Setup initializes and returns a configured slog.Logger
func Setup(level string, environment string) *slog.Logger {
	logLevel := parseLogLevel(level)
	
	var handler slog.Handler
	
	opts := &slog.HandlerOptions{
		Level: logLevel,
		AddSource: logLevel == slog.LevelDebug,
	}
	
	// Use JSON handler for production, text handler for development
	if environment == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	
	logger := slog.New(handler)
	
	// Set as default logger
	slog.SetDefault(logger)
	
	return logger
}

// parseLogLevel converts string to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// WithRequestID adds request ID to logger context
func WithRequestID(ctx context.Context, logger *slog.Logger, requestID string) *slog.Logger {
	return logger.With(slog.String("request_id", requestID))
}

// WithUserID adds user ID to logger context
func WithUserID(ctx context.Context, logger *slog.Logger, userID string) *slog.Logger {
	return logger.With(slog.String("user_id", userID))
}

// WithProjectID adds project ID to logger context
func WithProjectID(ctx context.Context, logger *slog.Logger, projectID string) *slog.Logger {
	return logger.With(slog.String("project_id", projectID))
}

// WithError adds error to logger context
func WithError(logger *slog.Logger, err error) *slog.Logger {
	return logger.With(slog.String("error", err.Error()))
}

// Fields creates a structured log field group
func Fields(args ...any) slog.Attr {
	return slog.Group("fields", args...)
}

