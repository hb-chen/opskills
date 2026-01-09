package logger

import (
	"go.uber.org/zap"
)

var zapLogger *zap.Logger

// ReplaceLogger replaces the global logger with a zap logger
func ReplaceLogger(logger *zap.Logger) {
	zapLogger = logger
	globalLogger = newZapLogger(logger)
}

// GetLogger returns the zap logger
func GetLogger() *zap.Logger {
	if zapLogger == nil {
		zapLogger = zap.NewNop()
	}
	return zapLogger
}

// Logger interface implementation using zap
type zapLoggerImpl struct {
	logger *zap.Logger
}

func newZapLogger(logger *zap.Logger) Logger {
	return &zapLoggerImpl{logger: logger}
}

func (l *zapLoggerImpl) Debug(v ...interface{}) {
	l.logger.Sugar().Debug(v...)
}

func (l *zapLoggerImpl) Debugf(format string, v ...interface{}) {
	l.logger.Sugar().Debugf(format, v...)
}

func (l *zapLoggerImpl) Info(v ...interface{}) {
	l.logger.Sugar().Info(v...)
}

func (l *zapLoggerImpl) Infof(format string, v ...interface{}) {
	l.logger.Sugar().Infof(format, v...)
}

func (l *zapLoggerImpl) Warn(v ...interface{}) {
	l.logger.Sugar().Warn(v...)
}

func (l *zapLoggerImpl) Warnf(format string, v ...interface{}) {
	l.logger.Sugar().Warnf(format, v...)
}

func (l *zapLoggerImpl) Error(v ...interface{}) {
	l.logger.Sugar().Error(v...)
}

func (l *zapLoggerImpl) Errorf(format string, v ...interface{}) {
	l.logger.Sugar().Errorf(format, v...)
}

func (l *zapLoggerImpl) Fatal(v ...interface{}) {
	l.logger.Sugar().Fatal(v...)
}

func (l *zapLoggerImpl) Fatalf(format string, v ...interface{}) {
	l.logger.Sugar().Fatalf(format, v...)
}

