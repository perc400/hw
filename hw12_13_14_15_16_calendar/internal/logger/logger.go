package logger

import (
	"strings"

	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
	atomic zap.AtomicLevel
	level  string
}

func New(level string) (*Logger, error) {
	atomicLevel, err := zap.ParseAtomicLevel(strings.ToLower(level))
	if err != nil {
		return nil, err
	}

	config := zap.NewProductionConfig()
	config.Level = atomicLevel
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		logger: logger,
		atomic: atomicLevel,
		level:  level,
	}, nil
}

func (l *Logger) SetLevel(level string) error {
	atom, err := zap.ParseAtomicLevel(strings.ToLower(level))
	if err != nil {
		return err
	}
	l.level = level
	l.atomic.SetLevel(atom.Level())
	return nil
}

func (l *Logger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *Logger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *Logger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *Logger) Debug(msg string) {
	l.logger.Debug(msg)
}
