package log

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Krzysztofz01/apikit/internal/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const logFileName = "apikit.log"

func CreateBaseLogger(verbose bool) (*zap.Logger, func() error, error) {
	fileEncoderConfig := zap.NewProductionEncoderConfig()
	consoleEncoderConfig := zap.NewProductionEncoderConfig()

	fileEncoder := zapcore.NewJSONEncoder(fileEncoderConfig)
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	cwd, err := os.Getwd()
	if err != nil {
		return nil, nil, fmt.Errorf("log: failed to access the current working directory: %w", err)
	}

	logFilePath := filepath.Join(cwd, logFileName)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, nil, fmt.Errorf("log: failed to access the current log file: %w", err)
	}

	fileWriter := zapcore.AddSync(logFile)
	consoleWriter := zapcore.AddSync(os.Stdout)

	logLevel := zapcore.InfoLevel
	if verbose {
		logLevel = zapcore.DebugLevel
	}

	fileCore := zapcore.NewCore(fileEncoder, fileWriter, logLevel)
	consoleCore := zapcore.NewCore(consoleEncoder, consoleWriter, logLevel)
	core := zapcore.NewTee(fileCore, consoleCore)

	logger := zap.New(core, zap.AddCaller())

	dispose := func() error {
		logger.Sync()

		if err := logFile.Close(); err != nil {
			return fmt.Errorf("log: failed to close the log file: %w", err)
		}

		return nil
	}

	return logger, dispose, nil
}

func CreateInternalLogger(l *zap.Logger) log.Logger {
	return &internalLogger{
		l: l,
	}
}

type internalLogger struct {
	l *zap.Logger
}

func (i *internalLogger) PrefixedFormat(prefix string, format string) string {
	return fmt.Sprintf("[%s] ", prefix) + format
}

func (i *internalLogger) Debugf(prefix string, format string, args ...interface{}) {
	i.l.Sugar().Debugf(i.PrefixedFormat(prefix, format), args...)
}

func (i *internalLogger) Errorf(prefix string, format string, args ...interface{}) {
	i.l.Sugar().Errorf(i.PrefixedFormat(prefix, format), args...)
}

func (i *internalLogger) Infof(prefix string, format string, args ...interface{}) {
	i.l.Sugar().Infof(i.PrefixedFormat(prefix, format), args...)
}

func (i *internalLogger) Warnf(prefix string, format string, args ...interface{}) {
	i.l.Sugar().Warnf(i.PrefixedFormat(prefix, format), args...)
}
