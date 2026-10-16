// Package core_logger provides structured logging using zap, supporting console and file outputs,
// context propagation, and request-scoped field enrichment.
package core_logger

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type loggerContextKey struct{}

var (
	key = loggerContextKey{}
)

// Logger wraps a zap.Logger and optionally an open file descriptor for tee-logging.
type Logger struct {
	*zap.Logger

	file *os.File
}

// ToContext attaches the provided Logger instance to the context.
//
// Parameters:
//   - ctx: Parent context.
//   - log: Logger instance to inject.
//
// Returns:
//   - context.Context: Context holding the logger reference.
func ToContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, key, log)
}

// FromContext retrieves the Logger associated with the context. If no logger is present,
// it safely falls back to a no-op logger (Nop) instead of panicking.
//
// Parameters:
//   - ctx: Context to extract the logger from.
//
// Returns:
//   - *Logger: Attached logger or a safe no-op logger.
func FromContext(ctx context.Context) *Logger {
	if ctx != nil {
		if log, ok := ctx.Value(key).(*Logger); ok && log != nil {
			return log
		}
	}

	return Nop()
}

// Nop returns a discard logger that suppresses all log entries without performing I/O.
//
// Returns:
//   - *Logger: Safe discard logger instance.
func Nop() *Logger {
	return &Logger{
		Logger: zap.NewNop(),
	}
}

// NewLogger instantiates and configures a dual-output (stdout and file) zap logger based on Config.
//
// Parameters:
//   - config: Configuration specifying logging level (debug, info, warn, error) and output directory.
//
// Returns:
//   - *Logger: Initialized logger with caller reporting.
//   - error: Non-nil if level parsing or folder/file creation fails.
//
// Side effects:
//   - Creates directory at config.Folder if absent.
//   - Opens a timestamped log file in write-only mode.
func NewLogger(config Config) (*Logger, error) {
	zaplvl := zap.NewAtomicLevel()
	if err := zaplvl.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("unmarshal log level: %w", err)
	}

	if err := os.MkdirAll(config.Folder, 0755); err != nil {
		return nil, fmt.Errorf("mkdir log folder: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
	logFilePath := filepath.Join(
		config.Folder,
		fmt.Sprintf("%s.log", timestamp),
	)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("open log file: %w", err)
	}

	zapConfig := zap.NewDevelopmentEncoderConfig()
	zapConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02T15:04:05.000000")

	zapEncoder := zapcore.NewConsoleEncoder(zapConfig)

	core := zapcore.NewTee(
		zapcore.NewCore(zapEncoder, zapcore.AddSync(os.Stdout), zaplvl),
		zapcore.NewCore(zapEncoder, zapcore.AddSync(logFile), zaplvl),
	)

	zaplogger := zap.New(core, zap.AddCaller())

	return &Logger{
		Logger: zaplogger,
		file:   logFile,
	}, nil
}

// Close flushes and closes the underlying log file descriptor.
//
// Side effects:
//   - Closes open file handle if present.
func (l *Logger) Close() {
	if l.file == nil {
		return
	}
	if err := l.file.Close(); err != nil {
		fmt.Println("failed to close application logger:", err)
	}
}

// With clones the logger and appends additional contextual zap fields.
//
// Parameters:
//   - field: Variadic list of zap.Field values (e.g., request_id, user_id).
//
// Returns:
//   - *Logger: Cloned logger with enriched fields.
func (l *Logger) With(field ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(field...),
		file:   l.file,
	}
}
