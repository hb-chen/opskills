package logger

// Logger interface for logging
type Logger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})
	Info(v ...interface{})
	Infof(format string, v ...interface{})
	Warn(v ...interface{})
	Warnf(format string, v ...interface{})
	Error(v ...interface{})
	Errorf(format string, v ...interface{})
	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})
}

var globalLogger Logger

func init() {
	// Initialize with a no-op logger until zap is configured
	globalLogger = &noOpLogger{}
}

// Debug logs a debug message
func Debug(v ...interface{}) {
	globalLogger.Debug(v...)
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	globalLogger.Debugf(format, v...)
}

// Info logs an info message
func Info(v ...interface{}) {
	globalLogger.Info(v...)
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	globalLogger.Infof(format, v...)
}

// Warn logs a warning message
func Warn(v ...interface{}) {
	globalLogger.Warn(v...)
}

// Warnf logs a formatted warning message
func Warnf(format string, v ...interface{}) {
	globalLogger.Warnf(format, v...)
}

// Error logs an error message
func Error(v ...interface{}) {
	globalLogger.Error(v...)
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	globalLogger.Errorf(format, v...)
}

// Fatal logs a fatal message and exits
func Fatal(v ...interface{}) {
	globalLogger.Fatal(v...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, v ...interface{}) {
	globalLogger.Fatalf(format, v...)
}

// noOpLogger is a no-op logger implementation
type noOpLogger struct{}

func (l *noOpLogger) Debug(v ...interface{})                 {}
func (l *noOpLogger) Debugf(format string, v ...interface{}) {}
func (l *noOpLogger) Info(v ...interface{})                 {}
func (l *noOpLogger) Infof(format string, v ...interface{}) {}
func (l *noOpLogger) Warn(v ...interface{})                 {}
func (l *noOpLogger) Warnf(format string, v ...interface{}) {}
func (l *noOpLogger) Error(v ...interface{})               {}
func (l *noOpLogger) Errorf(format string, v ...interface{}) {}
func (l *noOpLogger) Fatal(v ...interface{})                {}
func (l *noOpLogger) Fatalf(format string, v ...interface{}) {}

