// Package core_logger предоставляет структурированный журнал zap с выводом в консоль и файл и передачей через контекст.
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

const (
	logDirectoryPermissions = 0o755
	logFilePermissions      = 0o644
)

var (
	key = loggerContextKey{}
)

// Logger хранит структурированный журнал и открытый файл для записи сообщений.
type Logger struct {
	*zap.Logger

	file *os.File
}

// ToContext добавляет log в контекст ctx и возвращает производный контекст.
func ToContext(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, key, log)
}

// FromContext возвращает журнал из ctx. При его отсутствии возвращает безопасный Nop-журнал.
func FromContext(ctx context.Context) *Logger {
	if ctx != nil {
		if log, ok := ctx.Value(key).(*Logger); ok && log != nil {
			return log
		}
	}

	return Nop()
}

// Nop возвращает журнал, который отбрасывает сообщения и не выполняет ввод-вывод.
func Nop() *Logger {
	return &Logger{
		Logger: zap.NewNop(),
	}
}

// NewLogger создаёт журнал zap с уровнем и каталогом из config.
// Создаёт каталог и файл с временной меткой; возвращает ошибку неверного уровня или открытия файла.
func NewLogger(config Config) (*Logger, error) {
	zaplvl := zap.NewAtomicLevel()
	if err := zaplvl.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, fmt.Errorf("unmarshal log level: %w", err)
	}

	if err := os.MkdirAll(config.Folder, logDirectoryPermissions); err != nil {
		return nil, fmt.Errorf("mkdir log folder: %w", err)
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15-04-05.000000")
	logFilePath := filepath.Join(
		config.Folder,
		fmt.Sprintf("%s.log", timestamp),
	)

	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY, logFilePermissions)
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

// Close закрывает открытый файл журнала, если он существует.
func (l *Logger) Close() {
	if l.file == nil {
		return
	}
	if err := l.file.Close(); err != nil {
		fmt.Println("failed to close application logger:", err)
	}
}

// With возвращает копию журнала с дополнительными полями field.
func (l *Logger) With(field ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(field...),
		file:   l.file,
	}
}
