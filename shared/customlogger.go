package shared

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type HandlerType uint8

const (
	TextLogHandler HandlerType = iota + 1
	HandlerTypeJSON
)

type CustomLogger struct {
	*slog.Logger
}

func NewCustomLogger(filepath string, handlerType HandlerType) (*CustomLogger, *os.File, error) {
	// Open log file
	logFile, err := os.OpenFile(filepath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-handler (logs to both file & console)
	multiHandler := io.MultiWriter(os.Stdout, logFile)

	slogHandlerOptions := &slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(group []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Value = slog.StringValue(time.Now().Format(time.RFC3339))
			case slog.SourceKey:
				source, ok := a.Value.Any().(*slog.Source)
				if ok {
					a.Value = slog.StringValue(handleSourcePath(RootDir, source))
				}
			}
			return a
		},
	}

	var handler slog.Handler
	switch handlerType {
	case TextLogHandler:
		handler = slog.NewTextHandler(multiHandler, slogHandlerOptions)
	case HandlerTypeJSON:
		handler = slog.NewJSONHandler(multiHandler, slogHandlerOptions)
	default:
		return nil, nil, fmt.Errorf("unsupported handler type: %v", handlerType)
	}

	logger := slog.New(handler)

	return &CustomLogger{logger}, logFile, nil
}

func handleSourcePath(rootPath string, source *slog.Source) string {
	if source == nil {
		return ""
	}

	// Use filepath.Rel to get relative path
	relativePath, err := filepath.Rel(rootPath, source.File)
	if err != nil {
		// Fallback to original absolute path if can't compute relative path
		return fmt.Sprintf("%s:%d", source.File, source.Line)
	}

	return fmt.Sprintf("%s:%d", relativePath, source.Line)
}

var customLogger *CustomLogger

func InitializeCustomLogger(filepath string, handlerType HandlerType) (*os.File, error) {
	if customLogger != nil {
		return nil, fmt.Errorf("custom logger already initialized")
	}

	var err error
	var logFile *os.File
	customLogger, logFile, err = NewCustomLogger(filepath, handlerType)
	if err != nil {
		return logFile, fmt.Errorf("failed to initialize custom logger: %v", err)
	}

	return logFile, nil
}
func AcquireCustomLogger() *CustomLogger {
	if customLogger == nil {
		panic("custom logger not initialized, call InitializeCustomLogger first")
	}
	return customLogger
}
